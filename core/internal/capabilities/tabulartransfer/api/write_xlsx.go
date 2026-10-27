package api

import (
	"archive/zip"
	"fmt"
	"html"
	"io"
	"strings"
	"time"
)

const xlsxContentTypes = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
	`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
	`<Default Extension="xml" ContentType="application/xml"/>` +
	`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
	`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`

const xlsxRootRels = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`

const xlsxWorkbook = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
	`<sheets><sheet name="Export" sheetId="1" r:id="rId1"/></sheets></workbook>`

const xlsxWorkbookRels = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`

func writeXLSX(destination io.Writer, columns []string,
	source RowSource) (WriteResult, error) {
	archive := zip.NewWriter(destination)
	for _, document := range [][2]string{{"[Content_Types].xml", xlsxContentTypes},
		{"_rels/.rels", xlsxRootRels}, {"xl/workbook.xml", xlsxWorkbook},
		{"xl/_rels/workbook.xml.rels", xlsxWorkbookRels}} {
		if err := writeXLSXDocument(archive, document[0], document[1]); err != nil {
			return WriteResult{}, err
		}
	}
	worksheet, err := createXLSXDocument(archive, "xl/worksheets/sheet1.xml")
	if err != nil {
		return WriteResult{}, err
	}
	if _, err = io.WriteString(worksheet,
		`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`+
			xlsxRow(1, textCells(columns))); err != nil {
		return WriteResult{}, err
	}
	var result WriteResult
	for {
		cells, readErr := source.Next()
		if readErr == io.EOF {
			break
		}
		if readErr != nil || validateCells(columns, cells) != nil || result.Rows >= 1_000_000 {
			return WriteResult{}, ErrFormat
		}
		if _, err = io.WriteString(worksheet, xlsxRow(result.Rows+2, cells)); err != nil {
			return WriteResult{}, err
		}
		result.Rows++
	}
	if _, err = io.WriteString(worksheet, `</sheetData></worksheet>`); err != nil {
		return WriteResult{}, err
	}
	return result, archive.Close()
}

func writeXLSXDocument(archive *zip.Writer, name, body string) error {
	destination, err := createXLSXDocument(archive, name)
	if err == nil {
		_, err = io.WriteString(destination, body)
	}
	return err
}

func createXLSXDocument(archive *zip.Writer, name string) (io.Writer, error) {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetModTime(time.Unix(0, 0).UTC())
	return archive.CreateHeader(header)
}

func textCells(values []string) []Cell {
	result := make([]Cell, len(values))
	for index, value := range values {
		result[index] = Cell{Kind: CellText, Text: value}
	}
	return result
}

func xlsxRow(number int64, cells []Cell) string {
	var document strings.Builder
	fmt.Fprintf(&document, `<row r="%d">`, number)
	for index, cell := range cells {
		reference := xlsxColumn(index+1) + fmt.Sprint(number)
		if cell.Kind == CellNumber || cell.Kind == CellBoolean {
			fmt.Fprintf(&document, `<c r="%s"><v>%s</v></c>`, reference, cell.Text)
		} else if cell.Kind != CellNull {
			fmt.Fprintf(&document, `<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`,
				reference, html.EscapeString(cell.Text))
		}
	}
	return document.String() + `</row>`
}

func xlsxColumn(number int) string {
	var result string
	for number > 0 {
		number--
		result = string(rune('A'+number%26)) + result
		number /= 26
	}
	return result
}
