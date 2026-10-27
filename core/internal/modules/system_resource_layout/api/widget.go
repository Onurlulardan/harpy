package api

type Widget string

const (
	WidgetAuto             Widget = "auto"
	WidgetCheckbox         Widget = "checkbox"
	WidgetDate             Widget = "date"
	WidgetDateTime         Widget = "datetime"
	WidgetDocumentSelector Widget = "document_selector"
	WidgetDocumentUploader Widget = "document_uploader"
	WidgetEmail            Widget = "email"
	WidgetJSON             Widget = "json"
	WidgetLookup           Widget = "lookup"
	WidgetMultiSelect      Widget = "multi_select"
	WidgetNumber           Widget = "number"
	WidgetPassword         Widget = "password"
	WidgetRadio            Widget = "radio"
	WidgetSelect           Widget = "select"
	WidgetTelephone        Widget = "telephone"
	WidgetText             Widget = "text"
	WidgetTextarea         Widget = "textarea"
	WidgetTime             Widget = "time"
	WidgetToggle           Widget = "toggle"
	WidgetURL              Widget = "url"
)

func (widget Widget) Valid() bool {
	switch widget {
	case WidgetAuto, WidgetCheckbox, WidgetDate, WidgetDateTime,
		WidgetDocumentSelector, WidgetDocumentUploader, WidgetEmail, WidgetJSON,
		WidgetLookup, WidgetMultiSelect, WidgetNumber, WidgetPassword, WidgetRadio,
		WidgetSelect, WidgetTelephone, WidgetText, WidgetTextarea, WidgetTime,
		WidgetToggle, WidgetURL:
		return true
	default:
		return false
	}
}
