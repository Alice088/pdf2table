package pdf2table

import (
	"testing"

	"github.com/ledongthuc/pdf"
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

func TestParseSpecialty(t *testing.T) {
	cases := []struct {
		in   string
		code string
		name string
		ok   bool
	}{
		{"09.02.12 ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ И СОПРОВОЖДЕНИЕ ИНФОРМАЦИОННЫХ СИСТЕМ", "09.02.12", "ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ И СОПРОВОЖДЕНИЕ ИНФОРМАЦИОННЫХ СИСТЕМ", true},
		{"Направление 11.02.15 ИНФОКОММУНИКАЦИОННЫЕ СЕТИ И СИСТЕМЫ СВЯЗИ", "11.02.15", "ИНФОКОММУНИКАЦИОННЫЕ СЕТИ И СИСТЕМЫ СВЯЗИ", true},
		{"18.02.13ТЕХНОЛОГИЯ ПРОИЗВОДСТВА ИЗДЕЛИЙ ИЗ ПОЛИМЕРНЫХ КОМПОЗИТОВ", "18.02.13", "ТЕХНОЛОГИЯ ПРОИЗВОДСТВА ИЗДЕЛИЙ ИЗ ПОЛИМЕРНЫХ КОМПОЗИТОВ", true},
		{"План Учебный план ППССЗ СПО '09.02.12_26_00.04.plx', код специальности 09.02.12", "", "", false},
		{"№ 184 от 10.03.2025", "", "", false},
		{"Протокол № 5 от 07.05.2026", "", "", false},
		{"Основы спутникового метеорологического обеспечения", "", "", false},
	}
	for _, c := range cases {
		code, name, ok := ParseSpecialty(c.in)
		if ok != c.ok || code != c.code || name != c.name {
			t.Errorf("parseSpecialty(%q)=(%q,%q,%v) want (%q,%q,%v)", c.in, code, name, ok, c.code, c.name, c.ok)
		}
	}
}

func TestMetaPairs(t *testing.T) {
	if got := (*Meta)(nil).Pairs(); got != nil {
		t.Fatalf("nil meta pairs=%v", got)
	}
	got := (&Meta{Code: "09.02.12", Name: "ТЕХНИЧЕСКАЯ"}).Pairs()
	if len(got) != 2 || got[0][0] != "Код специальности" || got[0][1] != "09.02.12" || got[1][0] != "Название специальности" || got[1][1] != "ТЕХНИЧЕСКАЯ" {
		t.Fatalf("pairs=%v", got)
	}
}

func TestLargestTable(t *testing.T) {
	tables := []*Table{{Rows: 30, Cols: 9}, {Rows: 99, Cols: 93}, {Rows: 5, Cols: 5}}
	if got := Largest(tables); got != 2 {
		t.Fatalf("largest=%d want 2", got)
	}
}
