package pdf2table

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicAPI(t *testing.T) {
	dir, docs := rangeSamples(t)
	if len(docs) == 0 {
		t.Skip("no samples")
	}
	path := filepath.Join(dir, docs[0].file)

	tables, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(tables) == 0 {
		t.Fatal("ParseFile returned no tables")
	}
	if got := Largest(tables); got < 1 || got > len(tables) {
		t.Fatalf("Largest=%d", got)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fromReader, err := ParseReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}
	if len(fromReader) != len(tables) {
		t.Fatalf("ParseReader tables=%d want %d", len(fromReader), len(tables))
	}
	fromSplit, err := ParseReader(bytes.NewReader(raw), int64(len(raw)), WithSplit(true))
	if err != nil {
		t.Fatalf("ParseReader split: %v", err)
	}
	if len(fromSplit) < len(fromReader) {
		t.Fatalf("split tables=%d < merged=%d", len(fromSplit), len(fromReader))
	}

	tb := tables[Largest(tables)-1]
	if tb.HeaderRows() < 1 {
		t.Fatal("HeaderRows < 1")
	}
	hdr, rows := tb.Normalize()
	if len(hdr) == 0 || len(rows) == 0 {
		t.Fatal("empty normalize")
	}
	if len(tb.Occupancy()) != tb.Rows {
		t.Fatal("Occupancy shape mismatch")
	}
	if len(tb.Spans()) == 0 {
		t.Fatal("no spans")
	}
	if len(tb.Meta.Pairs()) == 0 {
		t.Fatal("no meta pairs")
	}

	for _, format := range []string{"csv", "md", "markdown", "html", "htm", "json"} {
		var buf bytes.Buffer
		if err := tb.Write(&buf, format, WithFill(true), WithRaw(true)); err != nil {
			t.Fatalf("Write(%s): %v", format, err)
		}
		if buf.Len() == 0 {
			t.Fatalf("Write(%s) produced nothing", format)
		}
	}
	var buf bytes.Buffer
	if err := tb.Write(&buf, "nope"); err == nil {
		t.Fatal("Write(nope) did not fail")
	}
	if err := tb.WriteXLSX(&buf); err != nil {
		t.Fatalf("WriteXLSX: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("WriteXLSX produced nothing")
	}

	out := t.TempDir()
	for _, name := range []string{"a.csv", "a.md", "a.html", "a.json", "a.xlsx"} {
		if err := tb.WriteFile(filepath.Join(out, name)); err != nil {
			t.Fatalf("WriteFile(%s): %v", name, err)
		}
	}
	if err := tb.WriteFile(filepath.Join(out, "a.bin")); err == nil {
		t.Fatal("WriteFile(.bin) did not fail")
	}

	code, name, ok := ParseSpecialty("09.02.12 ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ")
	if !ok || code != "09.02.12" || name != "ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ" {
		t.Fatalf("ParseSpecialty=(%q,%q,%v)", code, name, ok)
	}
}
