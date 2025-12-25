package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

type ExportColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type ExportRequest struct {
	Filename string                   `json:"filename"`
	Columns  []ExportColumn           `json:"columns"`
	Data     []map[string]interface{} `json:"data"`
}

type ExportHandlers struct{}

func NewExportHandlers() *ExportHandlers {
	return &ExportHandlers{}
}

// ExportToExcel handles Excel export requests
func (h *ExportHandlers) ExportToExcel(c echo.Context) error {
	var req ExportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if len(req.Data) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No data to export")
	}

	if len(req.Columns) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No columns specified")
	}

	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create sheet")
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Create header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4CAF50"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create header style")
	}

	// Write headers
	for colIdx, col := range req.Columns {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, col.Label)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Write data
	for rowIdx, row := range req.Data {
		for colIdx, col := range req.Columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			value := row[col.Key]

			// Handle different types
			switch v := value.(type) {
			case nil:
				f.SetCellValue(sheetName, cell, "")
			case float64:
				f.SetCellValue(sheetName, cell, v)
			case int:
				f.SetCellValue(sheetName, cell, v)
			case bool:
				if v {
					f.SetCellValue(sheetName, cell, "Yes")
				} else {
					f.SetCellValue(sheetName, cell, "No")
				}
			default:
				f.SetCellValue(sheetName, cell, fmt.Sprintf("%v", v))
			}
		}
	}

	// Auto-fit columns
	for colIdx := range req.Columns {
		colName, _ := excelize.ColumnNumberToName(colIdx + 1)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// Set response headers
	filename := req.Filename
	if filename == "" {
		filename = fmt.Sprintf("export_%s.xlsx", time.Now().Format("20060102_150405"))
	} else if len(filename) < 5 || filename[len(filename)-5:] != ".xlsx" {
		filename = filename + ".xlsx"
	}

	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Write to response
	if err := f.Write(c.Response().Writer); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to write Excel file")
	}

	return nil
}
