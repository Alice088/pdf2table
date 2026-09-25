package pdf2table

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"strings"
)

var headerWords = []string{
	"наименование", "индекс", "итого", "семестр", "курс", "компетенции",
	"объём", "объем", "счита", "виды деятельности", "форма", "квалификация",
}

func rowIsHeader(t *Table, r int) bool {
	var sb strings.Builder
	for c := 0; c < t.Cols; c++ {
		if cell := t.Grid[r][c]; cell != nil {
			sb.WriteString(strings.Join(strings.Fields(strings.ToLower(cell.Text)), ""))
			sb.WriteByte(' ')
		}
	}
	s := sb.String()
	for _, w := range headerWords {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

func headerRowCount(t *Table) int {
	n := 0
	for r := 0; r < t.Rows; r++ {
		if rowIsHeader(t, r) {
			n = r + 1
		} else {
			break
		}
	}
	if n == 0 && t.Rows > 0 {
		n = 1
	}
	return n
}

func writeCSV(w io.Writer, t *Table, fill, raw bool) error {
	cw := csv.NewWriter(w)
	var hdr []string
	var rows [][]string
	width := t.Cols
	if !raw {
		hdr, rows = t.Normalize()
		width = len(hdr)
	}
	if width < 2 {
		width = 2
	}
	for _, p := range t.Meta.Pairs() {
		rec := make([]string, width)
		rec[0], rec[1] = p[0], p[1]
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	if !raw {
		if err := cw.Write(hdr); err != nil {
			return err
		}
		for _, r := range rows {
			if err := cw.Write(r); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	}
	occ := t.Occupancy()
	rec := make([]string, width)
	for r := 0; r < t.Rows; r++ {
		for c := 0; c < t.Cols; c++ {
			cell := occ[r][c]
			if cell == nil {
				rec[c] = ""
				continue
			}
			if fill || (cell.Row == r && cell.Col == c) {
				rec[c] = collapse(cell.Text)
			} else {
				rec[c] = ""
			}
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func mdEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", "<br>")
	return strings.TrimSpace(s)
}

func writeMarkdown(w io.Writer, t *Table, fill, raw bool) error {
	var header []string
	var rows [][]string
	if !raw {
		header, rows = t.Normalize()
	} else {
		occ := t.Occupancy()
		header = make([]string, t.Cols)
		for c := 0; c < t.Cols; c++ {
			header[c] = ""
		}
		rows = make([][]string, 0, t.Rows)
		for r := 0; r < t.Rows; r++ {
			row := make([]string, t.Cols)
			for c := 0; c < t.Cols; c++ {
				cell := occ[r][c]
				if cell != nil && (fill || (cell.Row == r && cell.Col == c)) {
					row[c] = collapse(cell.Text)
				}
			}
			rows = append(rows, row)
		}
	}
	writeRow := func(cells []string) error {
		var sb strings.Builder
		sb.WriteString("|")
		for _, v := range cells {
			sb.WriteString(" ")
			sb.WriteString(mdEscape(v))
			sb.WriteString(" |")
		}
		sb.WriteByte('\n')
		_, err := io.WriteString(w, sb.String())
		return err
	}
	if pairs := t.Meta.Pairs(); len(pairs) > 0 {
		if err := writeRow(pairs[0][:]); err != nil {
			return err
		}
		if err := writeRow([]string{"---", "---"}); err != nil {
			return err
		}
		for _, p := range pairs[1:] {
			if err := writeRow(p[:]); err != nil {
				return err
			}
		}
		io.WriteString(w, "\n")
	}
	if err := writeRow(header); err != nil {
		return err
	}
	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	if err := writeRow(sep); err != nil {
		return err
	}
	for _, r := range rows {
		if err := writeRow(r); err != nil {
			return err
		}
	}
	return nil
}

func writeHTML(w io.Writer, t *Table) error {
	occ := t.Occupancy()
	hdr := headerRowCount(t)
	skip := make([][]bool, t.Rows)
	for r := range skip {
		skip[r] = make([]bool, t.Cols)
	}
	if pairs := t.Meta.Pairs(); len(pairs) > 0 {
		io.WriteString(w, "<table border=\"1\" cellspacing=\"0\" cellpadding=\"2\">\n")
		for _, p := range pairs {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>\n", html.EscapeString(p[0]), html.EscapeString(p[1]))
		}
		io.WriteString(w, "</table>\n<br>\n")
	}
	io.WriteString(w, "<table border=\"1\" cellspacing=\"0\" cellpadding=\"2\">\n")
	for r := 0; r < t.Rows; r++ {
		if r == 0 && hdr > 0 {
			io.WriteString(w, "<thead>\n")
		}
		if r == hdr && hdr > 0 {
			io.WriteString(w, "</thead>\n<tbody>\n")
		}
		io.WriteString(w, "<tr>")
		for c := 0; c < t.Cols; c++ {
			if skip[r][c] {
				continue
			}
			cell := occ[r][c]
			if cell == nil {
				io.WriteString(w, "<td></td>")
				continue
			}
			cs := 1
			for c+cs < t.Cols && !skip[r][c+cs] && occ[r][c+cs] == cell {
				cs++
			}
			rs := 1
			for r+rs < t.Rows {
				ok := true
				for k := 0; k < cs; k++ {
					if skip[r+rs][c+k] || occ[r+rs][c+k] != cell {
						ok = false
						break
					}
				}
				if !ok {
					break
				}
				rs++
			}
			for rr := 0; rr < rs; rr++ {
				for cc := 0; cc < cs; cc++ {
					skip[r+rr][c+cc] = true
				}
			}
			span := ""
			if rs > 1 {
				span += fmt.Sprintf(" rowspan=\"%d\"", rs)
			}
			if cs > 1 {
				span += fmt.Sprintf(" colspan=\"%d\"", cs)
			}
			txt := ""
			if cell.Row == r && cell.Col == c {
				txt = html.EscapeString(strings.Join(strings.Fields(cell.Text), " "))
			}
			fmt.Fprintf(w, "<td%s>%s</td>", span, txt)
		}
		io.WriteString(w, "</tr>\n")
	}
	if hdr > 0 {
		io.WriteString(w, "</tbody>\n")
	}
	io.WriteString(w, "</table>\n")
	return nil
}

type JSONCell struct {
	Row     int    `json:"row"`
	Col     int    `json:"col"`
	RowSpan int    `json:"rowspan"`
	ColSpan int    `json:"colspan"`
	Text    string `json:"text"`
}

type JSONTable struct {
	SpecialtyCode string     `json:"specialty_code,omitempty"`
	SpecialtyName string     `json:"specialty_name,omitempty"`
	Rows          int        `json:"rows"`
	Cols          int        `json:"cols"`
	Cells         []JSONCell `json:"cells"`
}

func writeJSON(w io.Writer, t *Table) error {
	jt := JSONTable{Rows: t.Rows, Cols: t.Cols}
	if t.Meta != nil {
		jt.SpecialtyCode = t.Meta.Code
		jt.SpecialtyName = t.Meta.Name
	}
	for _, c := range t.Cells {
		jt.Cells = append(jt.Cells, JSONCell{
			Row:     c.Row,
			Col:     c.Col,
			RowSpan: c.RowSpan,
			ColSpan: c.ColSpan,
			Text:    strings.Join(strings.Fields(c.Text), " "),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jt)
}
