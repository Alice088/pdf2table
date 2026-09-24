package main

import (
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

type Cell struct {
	Row, Col         int
	RowSpan, ColSpan int
	Text             string
}

type Table struct {
	Rows  int
	Cols  int
	Xs    []float64
	Grid  [][]*Cell
	Cells []*Cell
}

type box struct {
	minI, minJ, maxI, maxJ int
}

func buildPageTable(p pdf.Page) *Table {
	c := p.Content()
	hs, vs := pageLines(c)
	if len(hs) < 3 || len(vs) < 3 {
		hs, vs = pageRects(c)
	}
	g := newGrid(hs, vs)
	if g == nil {
		return nil
	}
	boxes := map[int]*box{}
	for j := 0; j < g.height; j++ {
		for i := 0; i < g.width; i++ {
			root := g.find(g.cell(i, j))
			b := boxes[root]
			if b == nil {
				boxes[root] = &box{minI: i, minJ: j, maxI: i, maxJ: j}
				continue
			}
			if i < b.minI {
				b.minI = i
			}
			if i > b.maxI {
				b.maxI = i
			}
			if j < b.minJ {
				b.minJ = j
			}
			if j > b.maxJ {
				b.maxJ = j
			}
		}
	}
	texts := map[int][]string{}
	for _, rn := range pageRuns(c) {
		i := band(rn.X, g.Xs)
		j := band(rn.Y, g.Ys)
		if i < 0 || j < 0 {
			continue
		}
		root := g.find(g.cell(i, j))
		texts[root] = append(texts[root], rn.Text)
	}
	t := &Table{Rows: g.height, Cols: g.width, Xs: g.Xs}
	t.Grid = make([][]*Cell, t.Rows)
	for r := range t.Grid {
		t.Grid[r] = make([]*Cell, t.Cols)
	}
	for root, b := range boxes {
		cell := &Cell{
			Row:     (g.height - 1) - b.maxJ,
			Col:     b.minI,
			RowSpan: b.maxJ - b.minJ + 1,
			ColSpan: b.maxI - b.minI + 1,
			Text:    cleanText(texts[root]),
		}
		if cell.Row < 0 || cell.Col < 0 || cell.Row >= t.Rows || cell.Col >= t.Cols {
			continue
		}
		t.Grid[cell.Row][cell.Col] = cell
		t.Cells = append(t.Cells, cell)
	}
	return t
}

func cleanText(parts []string) string {
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

func (t *Table) rowSig(r int) string {
	if r < 0 || r >= t.Rows {
		return ""
	}
	var sb strings.Builder
	for c := 0; c < t.Cols; c++ {
		cell := t.Grid[r][c]
		if cell == nil {
			continue
		}
		sb.WriteString(strings.ToLower(strings.Join(strings.Fields(cell.Text), "")))
		sb.WriteByte('|')
	}
	return sb.String()
}

func (t *Table) signatures() map[string]bool {
	m := map[string]bool{}
	for r := 0; r < t.Rows; r++ {
		if s := t.rowSig(r); s != "" {
			m[s] = true
		}
	}
	return m
}

func sameColumns(a, b *Table) bool {
	if a == nil || b == nil || a.Cols != b.Cols || len(a.Xs) != len(b.Xs) {
		return false
	}
	for i := range a.Xs {
		if absf(a.Xs[i]-b.Xs[i]) > bandTol {
			return false
		}
	}
	return true
}

func appendTable(dst, src *Table, skip map[string]bool) {
	skipping := true
	for r := 0; r < src.Rows; r++ {
		sig := src.rowSig(r)
		if skipping && sig != "" && skip[sig] {
			continue
		}
		skipping = false
		row := make([]*Cell, src.Cols)
		for c := 0; c < src.Cols; c++ {
			cell := src.Grid[r][c]
			if cell == nil {
				continue
			}
			nc := *cell
			nc.Row = dst.Rows
			row[c] = &nc
			dst.Cells = append(dst.Cells, &nc)
		}
		dst.Grid = append(dst.Grid, row)
		dst.Rows++
	}
}

func buildTables(r *pdf.Reader, split bool) []*Table {
	var tables []*Table
	var cur *Table
	for p := 1; p <= r.NumPage(); p++ {
		t := buildPageTable(r.Page(p))
		if t == nil || len(t.Cells) == 0 || t.Rows == 0 || t.Cols == 0 {
			continue
		}
		if !split && cur != nil && sameColumns(cur, t) {
			appendTable(cur, t, cur.signatures())
			continue
		}
		cur = t
		tables = append(tables, cur)
	}
	return tables
}

func (t *Table) occupancy() [][]*Cell {
	occ := make([][]*Cell, t.Rows)
	for r := range occ {
		occ[r] = make([]*Cell, t.Cols)
	}
	cs := make([]*Cell, len(t.Cells))
	copy(cs, t.Cells)
	sort.SliceStable(cs, func(i, j int) bool {
		return cs[i].RowSpan*cs[i].ColSpan > cs[j].RowSpan*cs[j].ColSpan
	})
	for _, cl := range cs {
		for r := cl.Row; r < cl.Row+cl.RowSpan && r < t.Rows; r++ {
			for c := cl.Col; c < cl.Col+cl.ColSpan && c < t.Cols; c++ {
				occ[r][c] = cl
			}
		}
	}
	return occ
}

func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

type span struct {
	r0, c0, r1, c1 int
}

func (t *Table) spans() []span {
	occ := t.occupancy()
	skip := make([][]bool, t.Rows)
	for r := range skip {
		skip[r] = make([]bool, t.Cols)
	}
	out := make([]span, 0, len(t.Cells))
	for r := 0; r < t.Rows; r++ {
		for c := 0; c < t.Cols; c++ {
			if skip[r][c] {
				continue
			}
			cell := occ[r][c]
			if cell == nil {
				skip[r][c] = true
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
			out = append(out, span{r, c, r + rs - 1, c + cs - 1})
		}
	}
	return out
}

func (t *Table) normalize() ([]string, [][]string) {
	h := headerRowCount(t)
	occ := t.occupancy()

	paths := make([]string, t.Cols)
	for c := 0; c < t.Cols; c++ {
		var parts []string
		for r := 0; r < h; r++ {
			cell := occ[r][c]
			if cell == nil {
				continue
			}
			txt := collapse(cell.Text)
			if txt == "" || txt == "-" {
				continue
			}
			if len(parts) > 0 && parts[len(parts)-1] == txt {
				continue
			}
			parts = append(parts, txt)
		}
		paths[c] = strings.Join(parts, " / ")
	}

	keep := make([]int, 0, t.Cols)
	for c := 0; c < t.Cols; c++ {
		hasData := false
		for r := h; r < t.Rows; r++ {
			if cell := occ[r][c]; cell != nil && collapse(cell.Text) != "" {
				hasData = true
				break
			}
		}
		if hasData {
			keep = append(keep, c)
		}
	}

	nameCol := -1
	for _, c := range keep {
		if paths[c] == "Наименование" {
			nameCol = c
		}
	}
	repCol := func(cell *Cell) int {
		if cell.ColSpan <= 1 {
			return cell.Col
		}
		if nameCol >= cell.Col && nameCol < cell.Col+cell.ColSpan {
			return nameCol
		}
		for _, c := range keep {
			if c >= cell.Col && c < cell.Col+cell.ColSpan {
				return c
			}
		}
		return cell.Col
	}
	rows := make([][]string, 0, t.Rows-h)
	for r := h; r < t.Rows; r++ {
		row := make([]string, len(keep))
		for i, c := range keep {
			if cell := occ[r][c]; cell != nil && c == repCol(cell) {
				row[i] = collapse(cell.Text)
			}
		}
		rows = append(rows, row)
	}
	used := make([]bool, len(keep))
	for _, row := range rows {
		for i, v := range row {
			if v != "" {
				used[i] = true
			}
		}
	}
	finalHeader := make([]string, 0, len(keep))
	for i, c := range keep {
		if used[i] {
			finalHeader = append(finalHeader, paths[c])
		}
	}
	finalRows := make([][]string, len(rows))
	for r, row := range rows {
		out := make([]string, 0, len(finalHeader))
		for i := range keep {
			if used[i] {
				out = append(out, row[i])
			}
		}
		finalRows[r] = out
	}
	return finalHeader, finalRows
}
