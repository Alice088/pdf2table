package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

func main() {
	outDir := flag.String("o", "", "output directory, directory mode only (default: next to the input file)")
	formats := flag.String("format", "xlsx,csv,md,html,json", "comma-separated formats, directory mode only")
	fill := flag.Bool("fill", false, "repeat merged cell values in every spanned cell (csv/md)")
	sel := flag.Int("table", 0, "table number to export in single-file mode (default: the largest)")
	raw := flag.Bool("raw", false, "keep the raw grid (multi-row header, empty columns) in csv/md")
	split := flag.Bool("split", false, "do not merge continuation pages; one table per PDF page")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] input.pdf [output.csv]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  input.pdf            export every table as input.tableN.{csv,md,html,json}\n")
		fmt.Fprintf(os.Stderr, "  input.pdf out.xlsx   export one table to out.xlsx (.xlsx/.csv/.md/.html/.json)\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(2)
	}
	var err error
	if flag.NArg() == 2 {
		err = runSingle(flag.Arg(0), flag.Arg(1), *fill, *raw, *split, *sel)
	} else {
		err = runAll(flag.Arg(0), *outDir, *formats, *fill, *raw, *split)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func openTables(input string, split bool) (*os.File, []*Table, error) {
	f, r, err := pdf.Open(input)
	if err != nil {
		return nil, nil, err
	}
	tables := buildTables(r, split)
	if len(tables) == 0 {
		f.Close()
		return nil, nil, fmt.Errorf("no tables found in %s", input)
	}
	return f, tables, nil
}

func runSingle(input, out string, fill, raw, split bool, sel int) error {
	f, tables, err := openTables(input, split)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(out), "."))
	if ext == "" {
		ext = "csv"
	}
	switch ext {
	case "xlsx", "csv", "md", "html", "json":
	default:
		return fmt.Errorf("unsupported output extension %q: use .xlsx, .csv, .md, .html or .json", filepath.Ext(out))
	}
	idx := sel
	if idx == 0 {
		idx = largestTable(tables)
	}
	if idx < 1 || idx > len(tables) {
		return fmt.Errorf("table %d out of range: found %d", sel, len(tables))
	}
	t := tables[idx-1]
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := writeOne(out, ext, t, fill, raw); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "table %d/%d: %d rows x %d cols -> %s\n", idx, len(tables), t.Rows, t.Cols, out)
	return nil
}

func runAll(input, outDir, formats string, fill, raw, split bool) error {
	f, tables, err := openTables(input, split)
	if err != nil {
		return err
	}
	defer f.Close()

	if outDir == "" {
		outDir = filepath.Dir(input)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	var want []string
	seen := map[string]bool{}
	for _, p := range strings.Split(formats, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" && !seen[p] {
			seen[p] = true
			want = append(want, p)
		}
	}
	for i, t := range tables {
		fmt.Fprintf(os.Stderr, "table %d: %d rows x %d cols\n", i+1, t.Rows, t.Cols)
		name := fmt.Sprintf("%s.table%d", base, i+1)
		for _, ext := range want {
			path := filepath.Join(outDir, name+"."+ext)
			if err := writeOne(path, ext, t, fill, raw); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "  wrote", path)
		}
	}
	return nil
}

func largestTable(tables []*Table) int {
	best, bestArea := 1, -1
	for i, t := range tables {
		if area := t.Rows * t.Cols; area > bestArea {
			best, bestArea = i+1, area
		}
	}
	return best
}

func writeOne(path, ext string, t *Table, fill, raw bool) error {
	if ext == "xlsx" {
		return writeXLSX(path, t)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	switch ext {
	case "csv":
		return writeCSV(f, t, fill, raw)
	case "md":
		return writeMarkdown(f, t, fill, raw)
	case "html":
		return writeHTML(f, t)
	case "json":
		return writeJSON(f, t)
	}
	return fmt.Errorf("unknown format %q", ext)
}
