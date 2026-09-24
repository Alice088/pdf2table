# pdf2table

Converts tables embedded in vector PDFs into normal, machine-readable tables
(CSV, Markdown, HTML, JSON). Pure Go, single static binary.

The converter does **not** guess the table structure from text positions.
It reads the actual grid: every cell border drawn in the PDF (thin filled
rectangles or stroked/rect cell borders) is extracted, the unique horizontal
and vertical grid lines are collected, and the table is rebuilt as a real
matrix. Merged cells are recovered by union-find: wherever an internal border
is missing, adjacent grid cells are fused into one cell with `rowspan`/`colspan`.

## Build

```sh
cd pdf2table
go build -o pdf2table .
```

## Usage

```sh
pdf2table input.pdf                # every table -> input.tableN.{csv,md,html,json}
pdf2table input.pdf output.csv     # one table -> output.csv
pdf2table input.pdf output.html    # format is taken from the extension
```

In single-file mode the largest table is exported; `-table N` picks another one.

Flags:

| Flag | Default | Meaning |
|------|---------|---------|
| `-table` | `0` | table number to export in single-file mode (0 = largest) |
| `-o` | directory of the input file | output directory (directory mode) |
| `-format` | `csv,md,html,json` | formats (directory mode) |
| `-fill` | `false` | repeat merged cell values in every spanned cell (csv/md) |

The output format is inferred from the extension: `.csv`, `.md`, `.html`, `.json`.

Examples:

```sh
pdf2table document.pdf document.csv
# table 2/2: 99 rows x 93 cols -> document.csv

pdf2table -table 1 document.pdf cover.html
pdf2table -fill document.pdf
```

## Outputs

* `csv`  — one row per table row, one column per grid column; merged values
  appear once (in the top-left cell) unless `-fill` is given.
* `md`   — GitHub-flavoured Markdown.
* `html` — full fidelity, preserves `rowspan`/`colspan` and header rows.
* `json` — `{rows, cols, cells:[{row,col,rowspan,colspan,text}]}`.

## How it works

1. Every rectangle in the page content stream is classified:
   * full-page rectangles are ignored;
   * thin rectangles become horizontal/vertical border segments;
   * if a page has no thin borders, thicker rectangles are treated as cell
     rectangles and their four edges become border segments.
2. Border coordinates are clustered (0.25 pt tolerance) into grid lines.
3. Adjacent grid cells are merged whenever the shared border is absent,
   producing the exact merged-cell layout.
4. Text is read run by run, positioned by its baseline, and placed into the
   merged cell that contains it.
5. Pages of the same table (equal column grid) are concatenated; a repeated
   header block on continuation pages is dropped.

## Notes and limitations

* Requires a vector PDF with a drawn grid. Scanned/raster PDFs need OCR first
  (the grid cannot be recovered). Text-only PDFs without ruling lines are not
  supported.
* Extremely irregular pages (several tables sharing one page with independent,
  slightly offset grids) may merge into one table; cells are then resolved
  greedily (smallest cell wins).
* `-fill` is handy for spreadsheets that do not understand merged cells.
# pdf2table
