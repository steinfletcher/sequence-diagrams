package sequence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type (
	Document struct {
		Diagrams    []*Diagram
		Title       string
		Description string
		MetaJSON    template.JS
	}

	Diagram struct {
		Title    string
		SubTitle string
		Events   []interface{}
	}

	DocumentHtmlModel struct {
		Title       string
		Description string
		Diagrams    []DiagramHtmlModel
		MetaJSON    template.JS
	}

	DiagramHtmlModel struct {
		WebSequenceDSL string
		Title          string
		SubTitle       string
		BadgeClass     string
		StatusCode     int
		LogEntries     []LogEntry
	}

	LogEntry struct {
		Header string
		Body   string
	}

	MessageRequest struct {
		Source string
		Target string
		Header string
		Body   string
	}

	MessageResponse struct {
		Source string
		Target string
		Header string
		Body   string
	}

	HttpRequest struct {
		Source string
		Target string
		Value  *http.Request
	}

	HttpResponse struct {
		Source string
		Target string
		Value  *http.Response
	}
)

func NewDocument() *Document {
	return &Document{}
}

func (r *Document) AddTitle(title string) *Document {
	r.Title = title
	return r
}

func (r *Document) AddDescription(description string) *Document {
	r.Description = description
	return r
}

func (r *Document) AddDiagram(diagram *Diagram) *Document {
	r.Diagrams = append(r.Diagrams, diagram)
	return r
}

func (r *Document) AddMeta(jsonData template.JS) *Document {
	r.MetaJSON = jsonData
	return r
}

func NewDiagram() *Diagram {
	return &Diagram{}
}

func (r *Diagram) AddHttpRequest(req HttpRequest) *Diagram {
	r.Events = append(r.Events, req)
	return r
}

func (r *Diagram) AddHttpResponse(req HttpResponse) *Diagram {
	r.Events = append(r.Events, req)
	return r
}

func (r *Diagram) AddMessageRequest(m MessageRequest) *Diagram {
	r.Events = append(r.Events, m)
	return r
}

func (r *Diagram) AddMessageResponse(m MessageResponse) *Diagram {
	r.Events = append(r.Events, m)
	return r
}

func (r *Diagram) AddTitle(title string) *Diagram {
	r.Title = title
	return r
}

func (r *Diagram) AddSubTitle(subTitle string) *Diagram {
	r.SubTitle = subTitle
	return r
}

func (r *Diagram) responseStatus() (int, error) {
	if len(r.Events) == 0 {
		return -1, errors.New("no events are defined")
	}

	switch v := r.Events[len(r.Events)-1].(type) {
	case HttpResponse:
		return v.Value.StatusCode, nil
	case MessageResponse:
		return -1, nil
	default:
		return -1, errors.New("final event should be a response type")
	}
}

func badgeCSSClass(status int) string {
	class := "badge badge-success"
	if status >= 400 && status < 500 {
		class = "badge badge-warning"
	} else if status >= 500 {
		class = "badge badge-danger"
	}
	return class
}

func (r *Document) BuildModel() (DocumentHtmlModel, error) {
	var diagrams []DiagramHtmlModel
	for _, d := range r.Diagrams {
		model, err := d.BuildModel()
		if err != nil {
			return DocumentHtmlModel{}, err
		}
		diagrams = append(diagrams, model)
	}

	return DocumentHtmlModel{
		Title:       r.Title,
		Description: r.Description,
		Diagrams:    diagrams,
		MetaJSON:    r.MetaJSON,
	}, nil
}

func (r *Diagram) BuildModel() (DiagramHtmlModel, error) {
	if len(r.Events) == 0 {
		return DiagramHtmlModel{}, errors.New("no events are defined")
	}

	var logs []LogEntry
	webSequenceDiagram := &WebSequenceDiagram{}
	for _, event := range r.Events {
		switch v := event.(type) {
		case HttpRequest:
			httpReq := v.Value
			webSequenceDiagram.AddRequestRow(v.Source, v.Target, fmt.Sprintf("%s %s", httpReq.Method, httpReq.URL))
			entry, err := newHttpRequestLogModel(httpReq)
			if err != nil {
				return DiagramHtmlModel{}, err
			}
			logs = append(logs, entry)
		case HttpResponse:
			webSequenceDiagram.AddResponseRow(v.Source, v.Target, strconv.Itoa(v.Value.StatusCode))
			entry, err := newHttpResponseLogModel(v.Value)
			if err != nil {
				return DiagramHtmlModel{}, err
			}
			logs = append(logs, entry)
		case MessageRequest:
			webSequenceDiagram.AddRequestRow(v.Source, v.Target, v.Header)
			logs = append(logs, LogEntry{Header: v.Header, Body: v.Body})
		case MessageResponse:
			webSequenceDiagram.AddResponseRow(v.Source, v.Target, v.Header)
			logs = append(logs, LogEntry{Header: v.Header, Body: v.Body})
		default:
			panic("received unknown event type")
		}
	}

	status, err := r.responseStatus()
	if err != nil {
		return DiagramHtmlModel{}, err
	}

	return DiagramHtmlModel{
		LogEntries: logs,
		Title:      r.Title,
		SubTitle:   r.SubTitle,
		StatusCode: status,
		BadgeClass: badgeCSSClass(status),
	}, nil
}

func (r *Document) RenderHTML() (string, error) {
	htmlModel, err := r.BuildModel()
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("sequenceDiagram").
		Funcs(*incTemplateFunc).
		Parse(t)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, htmlModel)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func newHttpRequestLogModel(req *http.Request) (LogEntry, error) {
	reqHeader, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return LogEntry{}, err
	}
	body, err := formatContent(req.Body, req.Header.Get("Content-Type"))
	if err != nil {
		return LogEntry{}, err
	}
	return LogEntry{Header: string(reqHeader), Body: body}, err
}

func newHttpResponseLogModel(res *http.Response) (LogEntry, error) {
	resDump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return LogEntry{}, err
	}
	body, err := formatContent(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		return LogEntry{}, err
	}
	return LogEntry{Header: string(resDump), Body: body}, err
}

func formatContent(bodyReadCloser io.ReadCloser, contentType string) (string, error) {
	if bodyReadCloser == nil {
		return "", nil
	}

	body, err := ioutil.ReadAll(bodyReadCloser)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if contentType == "application/json" {
		if len(body) > 0 {
			err := json.Indent(buf, body, "", "    ")
			if err != nil {
				return "", err
			}
		}
		return "", nil
	} else {
		_, err := buf.Write(body)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

var incTemplateFunc = &template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
}
