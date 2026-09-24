package main

import (
	"github.com/xuri/excelize/v2"
)

const (
	layoutSheet = "Layout"
	flatSheet   = "Flat"
)

func writeXLSX(path string, t *Table) error {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", layoutSheet); err != nil {
		return err
	}
	if _, err := f.NewSheet(flatSheet); err != nil {
		return err
	}
	if err := writeLayoutSheet(f, t); err != nil {
		return err
	}
	if err := writeFlatSheet(f, t); err != nil {
		return err
	}
	return f.SaveAs(path)
}

func writeLayoutSheet(f *excelize.File, t *Table) error {
	occ := t.occupancy()
	spans := t.spans()
	widths := make([]int, t.Cols)
	for _, s := range spans {
		cell := occ[s.r0][s.c0]
		if cell == nil {
			continue
		}
		name, err := excelize.CoordinatesToCellName(s.c0+1, s.r0+1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(layoutSheet, name, collapse(cell.Text)); err != nil {
			return err
		}
		if s.r1 > s.r0 || s.c1 > s.c0 {
			end, err := excelize.CoordinatesToCellName(s.c1+1, s.r1+1)
			if err != nil {
				return err
			}
			if err := f.MergeCell(layoutSheet, name, end); err != nil {
				return err
			}
		}
		if s.c0 == s.c1 {
			if n := len([]rune(collapse(cell.Text))); n > widths[s.c0] {
				widths[s.c0] = n
			}
		}
	}
	if err := styleSheet(f, layoutSheet, t.Rows, t.Cols, headerRowCount(t), widths); err != nil {
		return err
	}
	return nil
}

func writeFlatSheet(f *excelize.File, t *Table) error {
	header, rows := t.normalize()
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len([]rune(h))
		if err := f.SetCellValue(flatSheet, cellName(i+1, 1), h); err != nil {
			return err
		}
	}
	for r, row := range rows {
		for i, v := range row {
			if v == "" {
				continue
			}
			if n := len([]rune(v)); n > widths[i] {
				widths[i] = n
			}
			if err := f.SetCellValue(flatSheet, cellName(i+1, r+2), v); err != nil {
				return err
			}
		}
	}
	if err := styleSheet(f, flatSheet, len(rows)+1, len(header), 1, widths); err != nil {
		return err
	}
	return nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func styleSheet(f *excelize.File, sheet string, rows, cols, headerRows int, widths []int) error {
	if rows < 1 || cols < 1 {
		return nil
	}
	border, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "BFBFBF", Style: 1},
			{Type: "right", Color: "BFBFBF", Style: 1},
			{Type: "top", Color: "BFBFBF", Style: 1},
			{Type: "bottom", Color: "BFBFBF", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, cellName(1, 1), cellName(cols, rows), border); err != nil {
		return err
	}
	head, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return err
	}
	if headerRows > 0 {
		if err := f.SetCellStyle(sheet, cellName(1, 1), cellName(cols, headerRows), head); err != nil {
			return err
		}
	}
	for i := 0; i < cols && i < len(widths); i++ {
		w := float64(widths[i])*1.1 + 2
		if w < 6 {
			w = 6
		}
		if w > 45 {
			w = 45
		}
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}
		if err := f.SetColWidth(sheet, col, col, w); err != nil {
			return err
		}
	}
	pane := &excelize.Panes{Freeze: true, TopLeftCell: cellName(1, headerRows+1), ActivePane: "bottomLeft"}
	if headerRows > 0 {
		pane.YSplit = headerRows
	}
	if rows > 1 {
		if err := f.SetPanes(sheet, pane); err != nil {
			return err
		}
	}
	if err := f.SetRowHeight(sheet, 1, 15); err != nil {
		return err
	}
	return nil
}
