package pdf2table

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

var ErrUnsupportedFormat = errors.New("pdf2table: unsupported output format")

type Option func(*Options)

type Options struct {
	Split bool
}

func WithSplit(split bool) Option {
	return func(o *Options) { o.Split = split }
}

func ParseFile(path string, opts ...Option) ([]*Table, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(r, opts...), nil
}

func ParseReader(ra io.ReaderAt, size int64, opts ...Option) ([]*Table, error) {
	r, err := pdf.NewReader(ra, size)
	if err != nil {
		return nil, err
	}
	return Parse(r, opts...), nil
}

func Parse(r *pdf.Reader, opts ...Option) []*Table {
	cfg := Options{}
	for _, o := range opts {
		o(&cfg)
	}
	return buildTables(r, cfg.Split)
}

func Largest(tables []*Table) int {
	best, bestArea := 1, -1
	for i, t := range tables {
		if area := t.Rows * t.Cols; area > bestArea {
			best, bestArea = i+1, area
		}
	}
	return best
}

func (t *Table) HeaderRows() int {
	return headerRowCount(t)
}

type WriteOptions struct {
	Fill bool
	Raw  bool
}

type WriteOption func(*WriteOptions)

func WithFill(fill bool) WriteOption {
	return func(o *WriteOptions) { o.Fill = fill }
}

func WithRaw(raw bool) WriteOption {
	return func(o *WriteOptions) { o.Raw = raw }
}

func writeOptions(opts []WriteOption) WriteOptions {
	var o WriteOptions
	for _, f := range opts {
		f(&o)
	}
	return o
}

func (t *Table) WriteCSV(w io.Writer, opts ...WriteOption) error {
	o := writeOptions(opts)
	return writeCSV(w, t, o.Fill, o.Raw)
}

func (t *Table) WriteMarkdown(w io.Writer, opts ...WriteOption) error {
	o := writeOptions(opts)
	return writeMarkdown(w, t, o.Fill, o.Raw)
}

func (t *Table) WriteHTML(w io.Writer) error {
	return writeHTML(w, t)
}

func (t *Table) WriteJSON(w io.Writer) error {
	return writeJSON(w, t)
}

func (t *Table) Write(w io.Writer, format string, opts ...WriteOption) error {
	switch strings.ToLower(strings.TrimPrefix(format, ".")) {
	case "csv":
		return t.WriteCSV(w, opts...)
	case "md", "markdown":
		return t.WriteMarkdown(w, opts...)
	case "html", "htm":
		return t.WriteHTML(w)
	case "json":
		return t.WriteJSON(w)
	}
	return fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
}

func (t *Table) WriteFile(path string, opts ...WriteOption) error {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "" {
		ext = "csv"
	}
	if ext == "xlsx" {
		return t.SaveXLSX(path)
	}
	switch ext {
	case "csv", "md", "markdown", "html", "htm", "json":
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedFormat, ext)
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Write(f, ext, opts...)
}
