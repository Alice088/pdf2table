package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

func TestUniqueSorted(t *testing.T) {
	got := uniqueSorted([]float64{3, 1, 1.1, 2, 3.05, 0.9}, 0.25)
	want := []float64{0.9, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

func TestBand(t *testing.T) {
	bounds := []float64{10, 20, 30, 40}
	cases := []struct {
		v    float64
		want int
	}{
		{9, -1},
		{10, 0},
		{15, 0},
		{20, 1},
		{39.9, 2},
		{40, -1},
		{41, -1},
	}
	for _, c := range cases {
		if got := band(c.v, bounds); got != c.want {
			t.Errorf("band(%v)=%d want %d", c.v, got, c.want)
		}
	}
}

func TestCleanText(t *testing.T) {
	got := cleanText([]string{"  Привет ", "", " мир "})
	if got != "Привет мир" {
		t.Fatalf("got %q", got)
	}
}

func TestPageRunsGroupsChars(t *testing.T) {
	c := pdf.Content{Text: []pdf.Text{
		{X: 1, Y: 2, S: "П"},
		{X: 1, Y: 2, S: "П"},
		{X: 1, Y: 2, S: ".0"},
		{X: 1, Y: 2, S: "7\uFFFD"},
		{X: 9, Y: 2, S: "X"},
	}}
	runs := pageRuns(c)
	if len(runs) != 2 {
		t.Fatalf("runs=%d (%+v)", len(runs), runs)
	}
	if runs[0].Text != "ПП.07" {
		t.Fatalf("run0=%q", runs[0].Text)
	}
	if runs[1].X != 9 || runs[1].Text != "X" {
		t.Fatalf("run1=%+v", runs[1])
	}
}

func samplePDF() string {
	candidates := []string{
		os.Getenv("PDF2TABLE_SAMPLE"),
		"testdata/document.pdf",
		"/home/fworld/document.pdf",
		"../document.pdf",
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func loadTables(t *testing.T) []*Table {
	t.Helper()
	path := samplePDF()
	if path == "" {
		t.Skip("sample PDF not found")
	}
	f, r, err := pdf.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", filepath.Base(path), err)
	}
	t.Cleanup(func() { f.Close() })
	return buildTables(r, false)
}

func TestSampleTables(t *testing.T) {
	tables := loadTables(t)
	if len(tables) != 2 {
		t.Fatalf("tables=%d want 2", len(tables))
	}
	plan := tables[1]
	if plan.Rows != 99 || plan.Cols != 93 {
		t.Fatalf("plan %dx%d want 99x93", plan.Rows, plan.Cols)
	}
	area := 0
	for _, c := range plan.Cells {
		area += c.RowSpan * c.ColSpan
	}
	if area != plan.Rows*plan.Cols {
		t.Fatalf("merged cells overlap or leave gaps: area=%d full=%d", area, plan.Rows*plan.Cols)
	}
}

func TestSplitTables(t *testing.T) {
	path := samplePDF()
	if path == "" {
		t.Skip("sample PDF not found")
	}
	f, r, err := pdf.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	merged := buildTables(r, false)
	if len(merged) != 2 {
		t.Fatalf("merged tables=%d want 2", len(merged))
	}
	split := buildTables(r, true)
	if len(split) != 3 {
		t.Fatalf("split tables=%d want 3", len(split))
	}
	third := split[2]
	if third.Rows != 8 || third.Cols != 93 {
		t.Fatalf("page 4 table %dx%d want 8x93", third.Rows, third.Cols)
	}
	header, rows := third.normalize()
	_ = header
	found := false
	for _, row := range rows {
		for _, v := range row {
			if v == "ПП.07" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("page 4 table does not contain ПП.07")
	}
	_, lastRows := merged[1].normalize()
	if len(lastRows) == 0 {
		t.Fatal("merged plan has no data rows")
	}
}

func TestLargestTable(t *testing.T) {
	tables := []*Table{{Rows: 30, Cols: 9}, {Rows: 99, Cols: 93}, {Rows: 5, Cols: 5}}
	if got := largestTable(tables); got != 2 {
		t.Fatalf("largest=%d want 2", got)
	}
}

func TestNormalize(t *testing.T) {
	tables := loadTables(t)
	if len(tables) < 2 {
		t.Skip("plan table not available")
	}
	plan := tables[1]
	hdr, rows := plan.normalize()
	h := headerRowCount(plan)
	if len(rows) != plan.Rows-h {
		t.Fatalf("data rows=%d want %d", len(rows), plan.Rows-h)
	}
	idx := map[string]int{}
	for i, name := range hdr {
		idx[name] = i
	}
	for _, want := range []string{"Индекс", "Наименование", "Курс 1 / Семестр 1 / Итого", "Курс 4 / Семестр 8 / ПАтт"} {
		if _, ok := idx[want]; !ok {
			t.Fatalf("header %q missing in %v", want, hdr)
		}
	}
	seen := false
	for _, r := range rows {
		if r[idx["Индекс"]] == "ОУП.01" {
			seen = true
			if r[idx["Наименование"]] != "Русский язык" {
				t.Fatalf("ОУП.01 name=%q", r[idx["Наименование"]])
			}
			if r[idx["Итого акад.часов / По плану"]] != "102" {
				t.Fatalf("ОУП.01 plan=%q", r[idx["Итого акад.часов / По плану"]])
			}
		}
	}
	if !seen {
		t.Fatal("ОУП.01 row not found")
	}
	for c := range hdr {
		empty := true
		for _, r := range rows {
			if r[c] != "" {
				empty = false
				break
			}
		}
		if empty {
			t.Fatalf("normalized column %d (%q) is empty", c, hdr[c])
		}
	}
}

func TestWriteXLSX(t *testing.T) {
	tables := loadTables(t)
	if len(tables) < 2 {
		t.Skip("plan table not available")
	}
	plan := tables[1]

	path := filepath.Join(t.TempDir(), "plan.xlsx")
	if err := writeXLSX(path, plan); err != nil {
		t.Fatalf("writeXLSX: %v", err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer f.Close()

	for _, sheet := range []string{layoutSheet, flatSheet} {
		if idx, err := f.GetSheetIndex(sheet); err != nil || idx < 0 {
			t.Fatalf("sheet %q missing", sheet)
		}
	}
	merges, err := f.GetMergeCells(layoutSheet)
	if err != nil {
		t.Fatalf("merges: %v", err)
	}
	if len(merges) < 20 {
		t.Fatalf("too few merged cells: %d", len(merges))
	}
	flat, err := f.GetRows(flatSheet)
	if err != nil {
		t.Fatalf("flat rows: %v", err)
	}
	if len(flat) != 97 {
		t.Fatalf("flat rows=%d want 97", len(flat))
	}
	found := false
	for _, row := range flat {
		if len(row) > 2 && row[1] == "ОУП.01" && row[2] == "Русский язык" {
			found = true
		}
	}
	if !found {
		t.Fatal("ОУП.01 / Русский язык not found in Flat sheet")
	}
}

func TestSamplePlanValues(t *testing.T) {
	tables := loadTables(t)
	if len(tables) < 2 {
		t.Skip("plan table not available")
	}
	plan := tables[1]
	want := map[string]string{
		"ОУП.01":    "Русский язык",
		"ОГСЭ.03":   "Иностранный язык в профессиональной деятельности",
		"МДК.07.02": "Основы спутникового метеорологического обеспечения",
		"ГИА.01":    "Государственная итоговая аттестация",
	}
	codeAt := map[string]*Cell{}
	for _, c := range plan.Cells {
		codeAt[c.Text] = c
	}
	for code, name := range want {
		base := codeAt[code]
		if base == nil {
			t.Fatalf("code %q not found", code)
		}
		var nearest *Cell
		for _, cand := range plan.Cells {
			if cand.Row == base.Row && cand.Col > base.Col {
				if nearest == nil || cand.Col < nearest.Col {
					nearest = cand
				}
			}
		}
		if nearest == nil || nearest.Text != name {
			t.Fatalf("%s neighbour=%v want %q", code, nearest, name)
		}
	}
}
