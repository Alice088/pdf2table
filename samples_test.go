package pdf2table

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

type sampleDoc struct {
	file string
	code string
	name string
}

var goldenSamples = []sampleDoc{
	{"09.02.12_26_00.04-эцп.pdf", "09.02.12", "ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ И СОПРОВОЖДЕНИЕ ИНФОРМАЦИОННЫХ СИСТЕМ"},
	{"11.02.15_26_00.plx Проф-эцп.pdf", "11.02.15", "ИНФОКОММУНИКАЦИОННЫЕ СЕТИ И СИСТЕМЫ СВЯЗИ"},
	{"11.02.17_2026.plx Проф-эцп.pdf", "11.02.17", "РАЗРАБОТКА ЭЛЕКТРОННЫХ УСТРОЙСТВ И СИСТЕМ"},
	{"13.02.12_26_00.plx Проф_ЦИТП-эцп.pdf", "13.02.12", "ЭЛЕКТРИЧЕСКИЕ СТАНЦИИ, СЕТИ, ИХ РЕЛЕЙНАЯ ЗАЩИТА И АВТОМАТИЗАЦИЯ"},
	{"15.02.09_2026.plx Проф-эцп.pdf", "15.02.09", "АДДИТИВНЫЕ ТЕХНОЛОГИИ"},
	{"15.02.10_26_00.plx Проф_ЦИТП-эцп.pdf", "15.02.10", "МЕХАТРОНИКА И РОБОТОТЕХНИКА (ПО ОТРАСЛЯМ)"},
	{"15.02.16_2026.plx-эцп_ПРОФ.pdf", "15.02.16", "ТЕХНОЛОГИЯ МАШИНОСТРОЕНИЯ"},
	{"15.02.17_26_00.plx_Проф-эцп.pdf", "15.02.17", "МОНТАЖ, ТЕХНИЧЕСКОЕ ОБСЛУЖИВАНИЕ, ЭКСПЛУАТАЦИЯ И РЕМОНТ ПРОМЫШЛЕННОГО ОБОРУДОВАНИЯ (ПО ОТРАСЛЯМ)"},
	{"15.02.18_26_00.plx Проф_ЦИТП-эцп.pdf", "15.02.18", "ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ И ОБСЛУЖИВАНИЕ РОБОТИЗИРОВАННОГО ПРОИЗВОДСТВА (ПО ОТРАСЛЯМ)"},
	{"15.02.19_2026.plx Проф-эцп.pdf", "15.02.19", "СВАРОЧНОЕ ПРОИЗВОДСТВО"},
	{"18.02.13_2026.plx Проф_ЦОПП-эцп.pdf", "18.02.13", "ТЕХНОЛОГИЯ ПРОИЗВОДСТВА ИЗДЕЛИЙ ИЗ ПОЛИМЕРНЫХ КОМПОЗИТОВ"},
	{"20.02.02_26.plx-эцп_ПРОФ.pdf", "20.02.02", "ЗАЩИТА В ЧРЕЗВЫЧАЙНЫХ СИТУАЦИЯХ"},
	{"25.02.07_2026_ПРОФ.plx-эцп.pdf", "25.02.07", "ТЕХНИЧЕСКОЕ ОБСЛУЖИВАНИЕ АВИАЦИОННЫХ ДВИГАТЕЛЕЙ"},
}

type docData struct {
	path   string
	meta   *Meta
	merged []*Table
	pages  int
	losses []string
}

var (
	sampleDirOnce sync.Once
	sampleDirPath string

	docMu    sync.Mutex
	docCache = map[string]*docData{}
)

func findSampleDir() string {
	sampleDirOnce.Do(func() {
		for _, d := range []string{os.Getenv("PDF2TABLE_SAMPLES"), "../energydocs", "testdata"} {
			if d == "" {
				continue
			}
			if st, err := os.Stat(d); err == nil && st.IsDir() {
				if m, _ := filepath.Glob(filepath.Join(d, "*.pdf")); len(m) > 0 {
					sampleDirPath = d
					return
				}
			}
		}
	})
	return sampleDirPath
}

func rangeSamples(t *testing.T) (string, []sampleDoc) {
	t.Helper()
	dir := findSampleDir()
	if dir == "" {
		t.Skip("no sample PDF directory found; set PDF2TABLE_SAMPLES")
	}
	out := make([]sampleDoc, 0, len(goldenSamples))
	for _, g := range goldenSamples {
		if _, err := os.Stat(filepath.Join(dir, g.file)); err != nil {
			t.Errorf("golden sample missing: %s", filepath.Join(dir, g.file))
			continue
		}
		out = append(out, g)
	}
	return dir, out
}

func loadDoc(t *testing.T, path string) *docData {
	t.Helper()
	docMu.Lock()
	defer docMu.Unlock()
	if d, ok := docCache[path]; ok {
		return d
	}
	f, r, err := pdf.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", filepath.Base(path), err)
	}
	d := &docData{path: path, meta: extractMeta(r), pages: r.NumPage(), merged: buildTables(r, false)}
	for p := 1; p <= d.pages; p++ {
		c := r.Page(p).Content()
		hs, vs := pageLines(c)
		if len(hs) < 3 || len(vs) < 3 {
			hs, vs = pageRects(c)
		}
		g := newGrid(hs, vs)
		if g == nil {
			continue
		}
		minX, maxX := g.Xs[0], g.Xs[len(g.Xs)-1]
		minY, maxY := g.Ys[0], g.Ys[len(g.Ys)-1]
		for _, rn := range pageRuns(c) {
			if band(rn.X, g.Xs) >= 0 && band(rn.Y, g.Ys) >= 0 {
				continue
			}
			if rn.X >= minX && rn.X <= maxX && rn.Y >= minY && rn.Y <= maxY {
				d.losses = append(d.losses, fmt.Sprintf("page %d (%.1f,%.1f) %q", p, rn.X, rn.Y, rn.Text))
			}
		}
	}
	f.Close()
	docCache[path] = d
	return d
}

func largest(t *testing.T, d *docData) *Table {
	t.Helper()
	if len(d.merged) == 0 {
		t.Fatal("document has no tables")
	}
	return d.merged[Largest(d.merged)-1]
}

func TestSampleMetadata(t *testing.T) {
	dir, docs := rangeSamples(t)
	for _, g := range docs {
		g := g
		t.Run(g.file, func(t *testing.T) {
			d := loadDoc(t, filepath.Join(dir, g.file))
			if d.meta == nil {
				t.Fatalf("specialty metadata not extracted")
			}
			if d.meta.Code != g.code || d.meta.Name != g.name {
				t.Fatalf("meta=%q / %q want %q / %q", d.meta.Code, d.meta.Name, g.code, g.name)
			}
			if !strings.Contains(g.file, g.code) {
				t.Fatalf("code %q not present in filename", g.code)
			}
		})
	}
}

func TestSampleTextIsNotLost(t *testing.T) {
	dir, docs := rangeSamples(t)
	for _, g := range docs {
		g := g
		t.Run(g.file, func(t *testing.T) {
			d := loadDoc(t, filepath.Join(dir, g.file))
			if len(d.losses) > 0 {
				t.Fatalf("%d text run(s) inside a table grid were dropped:\n%s", len(d.losses), strings.Join(d.losses, "\n"))
			}
		})
	}
}

func TestSampleTablesAndNormalize(t *testing.T) {
	dir, docs := rangeSamples(t)
	for _, g := range docs {
		g := g
		t.Run(g.file, func(t *testing.T) {
			d := loadDoc(t, filepath.Join(dir, g.file))
			if len(d.merged) == 0 {
				t.Fatal("no tables found")
			}
			for i, tb := range d.merged {
				if tb.Rows <= 0 || tb.Cols <= 0 {
					t.Fatalf("table%d empty grid %dx%d", i+1, tb.Rows, tb.Cols)
				}
				for _, c := range tb.Cells {
					if c.Row < 0 || c.Col < 0 || c.Row+c.RowSpan > tb.Rows || c.Col+c.ColSpan > tb.Cols {
						t.Fatalf("table%d cell out of bounds: %+v", i+1, c)
					}
				}
			}
			plan := largest(t, d)
			area := 0
			for _, c := range plan.Cells {
				area += c.RowSpan * c.ColSpan
			}
			if area != plan.Rows*plan.Cols {
				t.Fatalf("largest table merges overlap or leave gaps: area=%d full=%d", area, plan.Rows*plan.Cols)
			}
			h := headerRowCount(plan)
			hdr, rows := plan.Normalize()
			if len(rows) != plan.Rows-h {
				t.Fatalf("normalized rows=%d want %d", len(rows), plan.Rows-h)
			}
			for i := range hdr {
				empty := true
				for _, rw := range rows {
					if rw[i] != "" {
						empty = false
						break
					}
				}
				if empty {
					t.Fatalf("normalized column %d (%q) is empty", i, hdr[i])
				}
			}
		})
	}
}

func TestSampleGoldenPlan(t *testing.T) {
	dir, _ := rangeSamples(t)
	path := filepath.Join(dir, "09.02.12_26_00.04-эцп.pdf")
	if _, err := os.Stat(path); err != nil {
		t.Skip("golden plan sample not available")
	}
	d := loadDoc(t, path)
	plan := largest(t, d)
	hdr, rows := plan.Normalize()
	idx := map[string]int{}
	for i, h := range hdr {
		idx[h] = i
	}
	for _, want := range []string{"Индекс", "Наименование", "Итого акад.часов / По плану", "Курс 1 / Семестр 1 / Итого"} {
		if _, ok := idx[want]; !ok {
			t.Fatalf("header %q missing", want)
		}
	}
	want := map[string]string{
		"ОУП.01": "Русский язык",
		"ГИА.01": "Государственная итоговая аттестация",
	}
	seen := map[string]bool{}
	for _, r := range rows {
		code := r[idx["Индекс"]]
		name, ok := want[code]
		if !ok {
			continue
		}
		if got := r[idx["Наименование"]]; got != name {
			t.Fatalf("%s = %q want %q", code, got, name)
		}
		seen[code] = true
	}
	for code := range want {
		if !seen[code] {
			t.Fatalf("row %s not found", code)
		}
	}
}

func TestSampleSplit(t *testing.T) {
	dir, docs := rangeSamples(t)
	for _, g := range docs {
		g := g
		t.Run(g.file, func(t *testing.T) {
			f, r, err := pdf.Open(filepath.Join(dir, g.file))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			merged := buildTables(r, false)
			split := buildTables(r, true)
			if len(split) < len(merged) {
				t.Fatalf("split tables=%d < merged=%d", len(split), len(merged))
			}
			for i, tb := range split {
				if tb == nil || tb.Rows == 0 || tb.Cols == 0 {
					t.Fatalf("split table%d empty", i+1)
				}
			}
		})
	}
}

func TestSampleWriters(t *testing.T) {
	dir, docs := rangeSamples(t)
	for _, g := range docs {
		g := g
		t.Run(g.file, func(t *testing.T) {
			d := loadDoc(t, filepath.Join(dir, g.file))
			plan := largest(t, d)
			pairs := plan.Meta.Pairs()
			_, wantRows := plan.Normalize()
			out := t.TempDir()

			csvPath := filepath.Join(out, "t.csv")
			if err := writeOne(csvPath, "csv", plan, false, false); err != nil {
				t.Fatalf("csv: %v", err)
			}
			recs := readCSV(t, csvPath)
			if len(recs) != len(pairs)+1+len(wantRows) {
				t.Fatalf("csv records=%d want %d", len(recs), len(pairs)+1+len(wantRows))
			}
			if recs[0][0] != "Код специальности" || recs[0][1] != g.code {
				t.Fatalf("csv meta code=%v", recs[0])
			}
			if recs[1][0] != "Название специальности" || recs[1][1] != g.name {
				t.Fatalf("csv meta name=%v", recs[1])
			}

			if err := writeOne(filepath.Join(out, "t.md"), "md", plan, false, false); err != nil {
				t.Fatalf("md: %v", err)
			}
			if body := readFile(t, filepath.Join(out, "t.md")); !strings.Contains(body, g.name) || !strings.Contains(body, g.code) {
				t.Fatal("md loses specialty metadata")
			}

			if err := writeOne(filepath.Join(out, "t.html"), "html", plan, false, false); err != nil {
				t.Fatalf("html: %v", err)
			}
			if body := readFile(t, filepath.Join(out, "t.html")); !strings.Contains(body, g.name) || !strings.Contains(body, g.code) {
				t.Fatal("html loses specialty metadata")
			}

			jsonPath := filepath.Join(out, "t.json")
			if err := writeOne(jsonPath, "json", plan, false, false); err != nil {
				t.Fatalf("json: %v", err)
			}
			var jt JSONTable
			if err := json.Unmarshal([]byte(readFile(t, jsonPath)), &jt); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if jt.SpecialtyCode != g.code || jt.SpecialtyName != g.name {
				t.Fatalf("json meta=%q/%q", jt.SpecialtyCode, jt.SpecialtyName)
			}
			if jt.Rows != plan.Rows || jt.Cols != plan.Cols || len(jt.Cells) != len(plan.Cells) {
				t.Fatalf("json shape=%dx%d cells=%d want %dx%d cells=%d", jt.Rows, jt.Cols, len(jt.Cells), plan.Rows, plan.Cols, len(plan.Cells))
			}

			xlsxPath := filepath.Join(out, "t.xlsx")
			if err := writeOne(xlsxPath, "xlsx", plan, false, false); err != nil {
				t.Fatalf("xlsx: %v", err)
			}
			xf, err := excelize.OpenFile(xlsxPath)
			if err != nil {
				t.Fatalf("open xlsx: %v", err)
			}
			defer xf.Close()
			for _, sheet := range []string{layoutSheet, flatSheet} {
				if idx, err := xf.GetSheetIndex(sheet); err != nil || idx < 0 {
					t.Fatalf("sheet %q missing", sheet)
				}
			}
			merges, err := xf.GetMergeCells(layoutSheet)
			if err != nil {
				t.Fatalf("layout merges: %v", err)
			}
			if len(merges) == 0 {
				t.Fatal("layout sheet has no merged cells")
			}
			flat, err := xf.GetRows(flatSheet)
			if err != nil {
				t.Fatalf("flat rows: %v", err)
			}
			if len(flat) < len(pairs)+2 {
				t.Fatalf("flat rows=%d too few", len(flat))
			}
			if flat[0][0] != "Код специальности" || flat[0][1] != g.code {
				t.Fatalf("flat meta code=%v", flat[0])
			}
			if flat[1][0] != "Название специальности" || flat[1][1] != g.name {
				t.Fatalf("flat meta name=%v", flat[1])
			}

			if err := writeOne(filepath.Join(out, "t2.csv"), "csv", plan, false, false); err != nil {
				t.Fatalf("csv2: %v", err)
			}
			if a, b := readFile(t, csvPath), readFile(t, filepath.Join(out, "t2.csv")); a != b {
				t.Fatal("csv output is not deterministic")
			}
			if err := writeOne(filepath.Join(out, "t2.json"), "json", plan, false, false); err != nil {
				t.Fatalf("json2: %v", err)
			}
			if a, b := readFile(t, jsonPath), readFile(t, filepath.Join(out, "t2.json")); a != b {
				t.Fatal("json output is not deterministic")
			}
		})
	}
}

func writeOne(path, ext string, t *Table, fill, raw bool) error {
	return t.WriteFile(path, WithFill(fill), WithRaw(raw))
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	recs, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return recs
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes.ToValidUTF8(b, []byte("?")))
}
