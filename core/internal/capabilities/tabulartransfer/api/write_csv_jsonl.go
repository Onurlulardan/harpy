package api

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

func writeCSV(destination io.Writer, columns []string,
	source RowSource) (WriteResult, error) {
	writer := csv.NewWriter(destination)
	if err := writer.Write(columns); err != nil {
		return WriteResult{}, err
	}
	var result WriteResult
	for {
		cells, err := source.Next()
		if err == io.EOF {
			break
		}
		if err != nil || validateCells(columns, cells) != nil {
			return WriteResult{}, ErrFormat
		}
		values := make([]string, len(cells))
		for index, cell := range cells {
			if cell.Kind != CellNull {
				values[index] = cell.Text
			}
		}
		if err = writer.Write(values); err != nil {
			return WriteResult{}, err
		}
		result.Rows++
	}
	writer.Flush()
	return result, writer.Error()
}

func writeJSONL(destination io.Writer, columns []string,
	source RowSource) (WriteResult, error) {
	encoder := json.NewEncoder(destination)
	encoder.SetEscapeHTML(false)
	var result WriteResult
	for {
		cells, err := source.Next()
		if err == io.EOF {
			return result, nil
		}
		if err != nil || validateCells(columns, cells) != nil {
			return WriteResult{}, ErrFormat
		}
		document := make(map[string]any, len(cells))
		for index, cell := range cells {
			value, decodeErr := jsonCellValue(cell)
			if decodeErr != nil {
				return WriteResult{}, decodeErr
			}
			document[columns[index]] = value
		}
		if err = encoder.Encode(document); err != nil {
			return WriteResult{}, err
		}
		result.Rows++
	}
}

func jsonCellValue(cell Cell) (any, error) {
	switch cell.Kind {
	case CellNull:
		return nil, nil
	case CellText:
		return cell.Text, nil
	case CellNumber:
		return json.Number(cell.Text), nil
	case CellBoolean:
		return cell.Text == "true", nil
	case CellJSON:
		return json.RawMessage(cell.Text), nil
	default:
		return nil, ErrFormat
	}
}
