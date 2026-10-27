package dashboardexport

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const pdfLinesPerPage = 45

func writePDF(destination io.Writer, pin Pin, items []renderedWidget) error {
	lines := []string{"Harpy dashboard export", "dashboard=" + pin.DashboardID.String(),
		"version=" + pin.DashboardVersionID.String(), "format=pdf"}
	for _, item := range items {
		lines = append(lines, "widget="+item.WidgetID.String()+" query="+
			item.QueryVersionID.String()+fmt.Sprintf(" rows=%d", item.Rows))
		encoded := base64.RawStdEncoding.EncodeToString(item.Body)
		for len(encoded) > 96 {
			lines = append(lines, encoded[:96])
			encoded = encoded[96:]
		}
		lines = append(lines, encoded)
	}
	pages := (len(lines) + pdfLinesPerPage - 1) / pdfLinesPerPage
	objects := make([][]byte, 3+pages*2)
	objects[0] = []byte("<< /Type /Catalog /Pages 2 0 R >>")
	kids := make([]string, pages)
	for index := range pages {
		kids[index] = fmt.Sprintf("%d 0 R", 4+index*2)
	}
	objects[1] = []byte(fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>",
		pages, strings.Join(kids, " ")))
	objects[2] = []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	for page := range pages {
		start := page * pdfLinesPerPage
		end := min(start+pdfLinesPerPage, len(lines))
		stream := pdfTextStream(lines[start:end])
		pageObject := 4 + page*2
		contentObject := pageObject + 1
		objects[pageObject-1] = []byte(fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] "+
				"/Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>", contentObject))
		objects[contentObject-1] = []byte(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream",
			len(stream), stream))
	}
	return writePDFObjects(destination, objects)
}

func pdfTextStream(lines []string) string {
	var value strings.Builder
	value.WriteString("BT\n/F1 9 Tf\n36 806 Td\n11 TL\n")
	for _, line := range lines {
		value.WriteString("(")
		value.WriteString(pdfEscape(line))
		value.WriteString(") Tj\nT*\n")
	}
	value.WriteString("ET")
	return value.String()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `(`, `\(`)
	return strings.ReplaceAll(value, `)`, `\)`)
}

func writePDFObjects(destination io.Writer, objects [][]byte) error {
	var document bytes.Buffer
	document.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = document.Len()
		fmt.Fprintf(&document, "%d 0 obj\n", index+1)
		document.Write(object)
		document.WriteString("\nendobj\n")
	}
	xref := document.Len()
	fmt.Fprintf(&document, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&document, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&document, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(offsets), xref)
	_, err := destination.Write(document.Bytes())
	return err
}
