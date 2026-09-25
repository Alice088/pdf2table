package main

import (
	"regexp"

	"github.com/ledongthuc/pdf"
)

type Meta struct {
	Code string
	Name string
}

var specialtyRe = regexp.MustCompile(`^\s*(?:Направление\s+)?(\d{2}\.\d{2}\.\d{2})\s*([А-ЯЁA-Z].*?)\s*$`)

func parseSpecialty(s string) (string, string, bool) {
	sub := specialtyRe.FindStringSubmatch(s)
	if sub == nil {
		return "", "", false
	}
	return sub[1], collapse(sub[2]), true
}

func extractMeta(r *pdf.Reader) *Meta {
	for p := 1; p <= r.NumPage(); p++ {
		for _, rn := range pageRuns(r.Page(p).Content()) {
			code, name, ok := parseSpecialty(rn.Text)
			if !ok {
				continue
			}
			return &Meta{Code: code, Name: name}
		}
	}
	return nil
}

func (m *Meta) pairs() [][2]string {
	if m == nil {
		return nil
	}
	var out [][2]string
	if m.Code != "" {
		out = append(out, [2]string{"Код специальности", m.Code})
	}
	if m.Name != "" {
		out = append(out, [2]string{"Название специальности", m.Name})
	}
	return out
}
