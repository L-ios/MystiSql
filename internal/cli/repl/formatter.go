package repl

import (
	"MystiSql/pkg/types"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) FormatTable(result *types.QueryResult) string {
	if len(result.Columns) == 0 {
		return "Empty set\r\n"
	}

	colWidths := f.calculateColumnWidths(result)

	var output strings.Builder

	f.writeSeparator(&output, colWidths)
	f.writeHeader(&output, result.Columns, colWidths)
	f.writeSeparator(&output, colWidths)

	for _, row := range result.Rows {
		f.writeRow(&output, row, colWidths)
	}

	f.writeSeparator(&output, colWidths)

	output.WriteString(fmt.Sprintf("%d row(s) in set (%.3f sec)\r\n\r\n", result.RowCount, result.ExecutionTime.Seconds()))

	return output.String()
}

func (f *Formatter) calculateColumnWidths(result *types.QueryResult) []int {
	widths := make([]int, len(result.Columns))

	for i, col := range result.Columns {
		widths[i] = utf8.RuneCountInString(col.Name)
	}

	for _, row := range result.Rows {
		for i, val := range row {
			str := f.formatValue(val)
			runeLen := utf8.RuneCountInString(str)
			if runeLen > widths[i] {
				widths[i] = runeLen
			}
		}
	}

	return widths
}

func (f *Formatter) writeSeparator(output *strings.Builder, widths []int) {
	output.WriteString("+")
	for _, w := range widths {
		output.WriteString(strings.Repeat("-", w+2))
		output.WriteString("+")
	}
	output.WriteString("\r\n")
}

func (f *Formatter) writeHeader(output *strings.Builder, columns []types.ColumnInfo, widths []int) {
	output.WriteString("|")
	for i, col := range columns {
		output.WriteString(" ")
		output.WriteString(f.padRight(col.Name, widths[i]))
		output.WriteString(" |")
	}
	output.WriteString("\r\n")
}

func (f *Formatter) writeRow(output *strings.Builder, row []interface{}, widths []int) {
	output.WriteString("|")
	for i, val := range row {
		output.WriteString(" ")
		str := f.formatValue(val)
		if f.isNumeric(val) {
			output.WriteString(f.padLeft(str, widths[i]))
		} else {
			output.WriteString(f.padRight(str, widths[i]))
		}
		output.WriteString(" |")
	}
	output.WriteString("\r\n")
}

func (f *Formatter) formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", val)
}

func (f *Formatter) isNumeric(val interface{}) bool {
	if val == nil {
		return false
	}
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(val.(string), 64)
		return err == nil
	default:
		return false
	}
}

func (f *Formatter) padRight(s string, width int) string {
	runeLen := utf8.RuneCountInString(s)
	if runeLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-runeLen)
}

func (f *Formatter) padLeft(s string, width int) string {
	runeLen := utf8.RuneCountInString(s)
	if runeLen >= width {
		return s
	}
	return strings.Repeat(" ", width-runeLen) + s
}

func (f *Formatter) FormatVertical(result *types.QueryResult) string {
	if len(result.Columns) == 0 {
		return "Empty set\r\n"
	}

	var output strings.Builder

	for rowIdx, row := range result.Rows {
		output.WriteString(fmt.Sprintf("*************************** %d. row ***************************\r\n", rowIdx+1))
		for i, col := range result.Columns {
			value := f.formatValue(row[i])
			output.WriteString(fmt.Sprintf("%s: %s\r\n", col.Name, value))
		}
	}

	output.WriteString(fmt.Sprintf("\r\n%d row(s) in set (%.3f sec)\r\n\r\n", result.RowCount, result.ExecutionTime.Seconds()))

	return output.String()
}

func (f *Formatter) FormatCSV(result *types.QueryResult) string {
	if len(result.Columns) == 0 {
		return ""
	}

	var output strings.Builder
	writer := csv.NewWriter(&output)

	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col.Name
	}
	writer.Write(headers)

	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, val := range row {
			if val == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		writer.Write(record)
	}

	writer.Flush()
	return output.String()
}

func (f *Formatter) FormatJSON(result *types.QueryResult) string {
	if len(result.Columns) == 0 {
		return "[]\r\n"
	}

	rows := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		rowMap := make(map[string]interface{})
		for j, col := range result.Columns {
			rowMap[col.Name] = row[j]
		}
		rows[i] = rowMap
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v\r\n", err)
	}

	return string(data) + "\r\n"
}

func (f *Formatter) FormatError(err error) string {
	return fmt.Sprintf("ERROR: %v\r\n", err)
}

func (f *Formatter) FormatEmpty() string {
	return "Empty set\r\n"
}

func (f *Formatter) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (f *Formatter) FormatStatus(r *REPL) string {
	var output strings.Builder

	output.WriteString("--------------\r\n")
	output.WriteString("MystiSql Status\r\n")
	output.WriteString("--------------\r\n")
	output.WriteString(fmt.Sprintf("Current instance: %s\r\n", r.currentInstance))
	output.WriteString(fmt.Sprintf("Total instances: %d\r\n", len(r.instances)))

	if len(r.instances) > 0 {
		output.WriteString("Available instances:\r\n")
		for _, inst := range r.instances {
			marker := " "
			if inst.Name == r.currentInstance {
				marker = "*"
			}
			output.WriteString(fmt.Sprintf("  %s %s (%s)\r\n", marker, inst.Name, inst.Type))
		}
	}

	output.WriteString(fmt.Sprintf("Prompt: %s\r\n", r.prompt))
	output.WriteString(fmt.Sprintf("History entries: %d\r\n", r.history.Count()))
	output.WriteString("--------------\r\n")

	return output.String()
}
