package pdf2table

import (
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

type HLine struct {
	Y, X0, X1 float64
}

type VLine struct {
	X, Y0, Y1 float64
}

type TextRun struct {
	X, Y float64
	Text string
}

const (
	lineTol  = 2.0
	coordTol = 0.25
	bandTol  = 0.9
)

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func uniqueSorted(vals []float64, tol float64) []float64 {
	sort.Float64s(vals)
	out := make([]float64, 0, len(vals))
	for _, v := range vals {
		if len(out) == 0 || v-out[len(out)-1] > tol {
			out = append(out, v)
		}
	}
	return out
}

func band(v float64, bounds []float64) int {
	i := sort.Search(len(bounds), func(i int) bool { return bounds[i] > v }) - 1
	if i < 0 || i > len(bounds)-2 {
		return -1
	}
	return i
}

func nearest(bounds []float64, v, tol float64) int {
	best, bestDist := -1, tol
	for i, b := range bounds {
		if d := absf(b - v); d <= bestDist {
			best, bestDist = i, d
		}
	}
	return best
}

func pageRuns(c pdf.Content) []TextRun {
	out := make([]TextRun, 0, len(c.Text)/8)
	for i := 0; i < len(c.Text); {
		t := c.Text[i]
		j := i
		var sb strings.Builder
		for j < len(c.Text) && c.Text[j].X == t.X && c.Text[j].Y == t.Y {
			sb.WriteString(c.Text[j].S)
			j++
		}
		i = j
		s := strings.TrimRight(sb.String(), "\uFFFD")
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, TextRun{X: t.X, Y: t.Y, Text: s})
		}
	}
	return out
}

func pageLines(c pdf.Content) ([]HLine, []VLine) {
	var hs []HLine
	var vs []VLine
	for _, rc := range c.Rect {
		w := absf(rc.Max.X - rc.Min.X)
		h := absf(rc.Max.Y - rc.Min.Y)
		if w > 700 && h > 500 {
			continue
		}
		x0, x1 := rc.Min.X, rc.Max.X
		y0, y1 := rc.Min.Y, rc.Max.Y
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		switch {
		case h < lineTol && w >= h:
			hs = append(hs, HLine{Y: (y0 + y1) / 2, X0: x0, X1: x1})
		case w < lineTol && h > w:
			vs = append(vs, VLine{X: (x0 + x1) / 2, Y0: y0, Y1: y1})
		}
	}
	return hs, vs
}

func pageRects(c pdf.Content) ([]HLine, []VLine) {
	var hs []HLine
	var vs []VLine
	for _, rc := range c.Rect {
		w := absf(rc.Max.X - rc.Min.X)
		h := absf(rc.Max.Y - rc.Min.Y)
		if w < 2 || h < 2 || (w > 700 && h > 500) {
			continue
		}
		x0, x1 := rc.Min.X, rc.Max.X
		y0, y1 := rc.Min.Y, rc.Max.Y
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		hs = append(hs, HLine{Y: y0, X0: x0, X1: x1}, HLine{Y: y1, X0: x0, X1: x1})
		vs = append(vs, VLine{X: x0, Y0: y0, Y1: y1}, VLine{X: x1, Y0: y0, Y1: y1})
	}
	return hs, vs
}

type Grid struct {
	Xs     []float64
	Ys     []float64
	parent []int
	width  int
	height int
}

func newGrid(hs []HLine, vs []VLine) *Grid {
	var xs, ys []float64
	for _, v := range vs {
		xs = append(xs, v.X)
	}
	for _, h := range hs {
		ys = append(ys, h.Y)
	}
	xs = uniqueSorted(xs, coordTol)
	ys = uniqueSorted(ys, coordTol)
	if len(xs) < 2 || len(ys) < 2 {
		return nil
	}
	g := &Grid{Xs: xs, Ys: ys, width: len(xs) - 1, height: len(ys) - 1}
	g.parent = make([]int, g.width*g.height)
	for i := range g.parent {
		g.parent[i] = i
	}

	vCov := make([][]bool, len(xs))
	for i := range vCov {
		vCov[i] = make([]bool, g.height)
	}
	for _, v := range vs {
		ix := nearest(xs, v.X, bandTol)
		if ix < 0 {
			continue
		}
		j0 := nearest(ys, v.Y0, bandTol)
		j1 := nearest(ys, v.Y1, bandTol)
		if j0 < 0 || j1 < 0 {
			continue
		}
		if j0 > j1 {
			j0, j1 = j1, j0
		}
		for j := j0; j < j1 && j < g.height; j++ {
			vCov[ix][j] = true
		}
	}
	hCov := make([][]bool, len(ys))
	for i := range hCov {
		hCov[i] = make([]bool, g.width)
	}
	for _, h := range hs {
		iy := nearest(ys, h.Y, bandTol)
		if iy < 0 {
			continue
		}
		i0 := nearest(xs, h.X0, bandTol)
		i1 := nearest(xs, h.X1, bandTol)
		if i0 < 0 || i1 < 0 {
			continue
		}
		if i0 > i1 {
			i0, i1 = i1, i0
		}
		for i := i0; i < i1 && i < g.width; i++ {
			hCov[iy][i] = true
		}
	}

	idx := func(i, j int) int { return j*g.width + i }
	for j := 0; j < g.height; j++ {
		for i := 0; i < g.width; i++ {
			if i+1 < g.width && !vCov[i+1][j] {
				g.union(idx(i, j), idx(i+1, j))
			}
			if j+1 < g.height && !hCov[j+1][i] {
				g.union(idx(i, j), idx(i, j+1))
			}
		}
	}
	return g
}

func (g *Grid) find(a int) int {
	for g.parent[a] != a {
		g.parent[a] = g.parent[g.parent[a]]
		a = g.parent[a]
	}
	return a
}

func (g *Grid) union(a, b int) {
	ra, rb := g.find(a), g.find(b)
	if ra != rb {
		g.parent[ra] = rb
	}
}

func (g *Grid) cell(i, j int) int { return j*g.width + i }
