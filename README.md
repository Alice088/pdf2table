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
pdf2table input.pdf                # every table -> input.tableN.{xlsx,csv,md,html,json}
pdf2table input.pdf output.xlsx    # one table -> output.xlsx
pdf2table input.pdf output.csv     # format is taken from the extension
```

In single-file mode the largest table is exported; `-table N` picks another one.

Flags:

| Flag | Default | Meaning |
|------|---------|---------|
| `-table` | `0` | table number to export in single-file mode (0 = largest) |
| `-o` | directory of the input file | output directory (directory mode) |
| `-format` | `xlsx,csv,md,html,json` | formats (directory mode) |
| `-split` | `false` | do not merge continuation pages; one table per PDF page |
| `-raw` | `false` | keep the raw grid (multi-row header, empty columns) in csv/md |
| `-fill` | `false` | `-raw` only: repeat merged cell values in every spanned cell |

The output format is inferred from the extension: `.xlsx`, `.csv`, `.md`, `.html`, `.json`.

Examples:

```sh
pdf2table document.pdf document.csv
# table 2/2: 99 rows x 93 cols -> document.csv

pdf2table -table 1 document.pdf cover.html
pdf2table -fill document.pdf
```

## Outputs

`csv` and `md` are **normalized**: the multi-level header is flattened into a
single header row (`Курс 1 / Семестр 1 / Лек`), merged values are repeated in
every row they span, and columns that are empty everywhere are dropped.
Pass `-raw` to get the untouched grid instead (multi-row header, every grid
column, merged values only in the top-left cell).

* `xlsx` — Excel workbook with two sheets:
  * **Layout** — the table exactly as drawn in the PDF: every grid cell, real
    merged cells (`MergeCell`), borders, bold frozen header.
  * **Flat** — the normalized single-header table.
* `csv`  — one row per table row, one column per logical column.
* `md`   — GitHub-flavoured Markdown.
* `html` — full fidelity, preserves `rowspan`/`colspan` and the header block.
* `json` — `{specialty_code, specialty_name, rows, cols, cells:[{row,col,rowspan,colspan,text}]}`.

The specialty code and title are parsed from the plan heading
(`09.02.12 ТЕХНИЧЕСКАЯ ЭКСПЛУАТАЦИЯ И СОПРОВОЖДЕНИЕ ИНФОРМАЦИОННЫХ СИСТЕМ`,
optionally prefixed with `Направление`) on the cover page and emitted for every
table of the document:

* `csv`, `md`, `html` and the `Flat` xlsx sheet — two rows above the table:
  `Код специальности` / `Название специальности`;
* `json` — the top-level `specialty_code` / `specialty_name` fields.

If the heading is absent, no metadata rows are emitted. In `csv` the metadata
rows are padded with empty fields to the table width so every record has the
same number of fields.

## Tests

The suite runs against the 13 golden PDFs in `../energydocs` (override with
`PDF2TABLE_SAMPLES=<dir>`). For each document it checks:

* the specialty code and title metadata;
* that no text run inside a table grid is dropped;
* that the main table merges tile the grid exactly (no overlaps or gaps);
* that normalized columns are never empty;
* golden curriculum rows for `09.02.12`;
* `-split` mode;
* every output format (csv/md/html/json/xlsx), including metadata and
determinism.

```sh
PDF2TABLE_SAMPLES=../energydocs go test ./...
```

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
5. The `код специальности` / specialty title heading on the cover page is read
   as document metadata and prepended to every table output.
6. Pages of the same table (equal column grid) are concatenated; a repeated
   header block on continuation pages is dropped. Pass `-split` to keep every
   PDF page as its own table instead.

## Notes and limitations

* Requires a vector PDF with a drawn grid. Scanned/raster PDFs need OCR first
  (the grid cannot be recovered). Text-only PDFs without ruling lines are not
  supported.
* Extremely irregular pages (several tables sharing one page with independent,
  slightly offset grids) may merge into one table; cells are then resolved
  greedily (smallest cell wins).
* `-fill` is handy for spreadsheets that do not understand merged cells.
# pdf2table
