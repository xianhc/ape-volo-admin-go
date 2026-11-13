// Package excel 封装 Excel 导出。
//
// 替代已停维护的 tealeg/xlsx：底层使用 excelize/v2，
// 对外只暴露「表头 + 逐行数据」这一种用法，避免 11 个导出点
// 各写一套样式与列宽逻辑。
package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

const (
	// defaultSheetName 是未指定工作表名时使用的名称
	defaultSheetName = "Sheet1"
	// minColWidth / maxColWidth 限制自动列宽范围（字符数）
	minColWidth = 8
	maxColWidth = 60
)

// File 是一个待导出的工作簿。
type File struct {
	book      *excelize.File
	sheet     string
	rowCursor int
}

// NewFile 创建只含一个工作表的工作簿。
func NewFile(sheetName string) *File {
	if sheetName == "" {
		sheetName = defaultSheetName
	}
	// Excel 限制工作表名不超过 31 个字符
	if utf8.RuneCountInString(sheetName) > 31 {
		sheetName = string([]rune(sheetName)[:31])
	}

	book := excelize.NewFile()
	defaultSheet := book.GetSheetName(0)
	if defaultSheet != sheetName {
		// 重命名失败时保留默认表名，不影响导出内容
		_ = book.SetSheetName(defaultSheet, sheetName)
	}

	return &File{book: book, sheet: sheetName, rowCursor: 0}
}

// SetHeader 写入表头并套用统一样式（蓝底白字加粗、居中）。
func (f *File) SetHeader(header []string) {
	f.rowCursor++
	row := f.rowCursor

	style, err := f.book.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1E90FF"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, title := range header {
		cell, cellErr := excelize.CoordinatesToCellName(i+1, row)
		if cellErr != nil {
			continue
		}
		_ = f.book.SetCellValue(f.sheet, cell, title)
		if err == nil {
			_ = f.book.SetCellStyle(f.sheet, cell, cell, style)
		}
	}
}

// AddRow 追加一行数据，值的类型由 excelize 负责转换。
func (f *File) AddRow(values ...any) {
	f.rowCursor++
	row := f.rowCursor
	for i, v := range values {
		cell, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			continue
		}
		_ = f.book.SetCellValue(f.sheet, cell, v)
	}
}

// SetColWidths 按内容长度自动设置列宽。
//
// 在写入完所有行之后调用；超过 maxColWidth 的列会被截断，
// 避免长文本（如 JSON、错误堆栈）把表格撑开。
func (f *File) SetColWidths() {
	rows, err := f.book.GetRows(f.sheet)
	if err != nil {
		return
	}

	widths := make(map[int]int)
	for _, row := range rows {
		for i, cell := range row {
			n := utf8.RuneCountInString(cell)
			if n > widths[i] {
				widths[i] = n
			}
		}
	}

	for i, w := range widths {
		if w < minColWidth {
			w = minColWidth
		}
		if w > maxColWidth {
			w = maxColWidth
		}
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			continue
		}
		_ = f.book.SetColWidth(f.sheet, col, col, float64(w)+2)
	}
}

// Save 将工作簿写入指定路径，父目录不存在时自动创建。
func (f *File) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建导出目录失败: %w", err)
	}
	f.SetColWidths()
	if err := f.book.SaveAs(path); err != nil {
		return fmt.Errorf("保存 Excel 失败: %w", err)
	}
	return nil
}

// Close 释放工作簿占用的内存与临时文件。
func (f *File) Close() error {
	if f.book == nil {
		return nil
	}
	return f.book.Close()
}

// SafeName 去掉文件名中不适合做文件名的字符。
func SafeName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(name)
}
