package sequence

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestDiagram_AddTitle(t *testing.T) {
	model, err := aDiagram().
		AddTitle("title").
		BuildModel()

	assert.Nil(t, err)
	assert.Contains(t, model.Title, "title")
}

func TestDiagram_AddSubTitle(t *testing.T) {
	model, err := aDiagram().AddSubTitle("subTitle").BuildModel()

	assert.Nil(t, err)
	assert.Contains(t, model.SubTitle, "subTitle")
}

func TestDiagram_AddHttpRequest(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "http://example.com/abcdef", bytes.NewBufferString(`{"a":123}`))
	request.Header.Set("Content-Type", "application/json")
	aRequest := HttpRequest{Value: request}

	model, err := NewDiagram().AddHttpRequest(aRequest).AddHttpResponse(aResponse()).BuildModel()

	assert.Nil(t, err)
	e := model.LogEntries[0]
	assert.Contains(t, e.Body, `"a": 123`)
	assert.Contains(t, e.Header, "GET /abcdef")
	assert.Contains(t, e.Header, "Host: example.com")
	assert.Contains(t, e.Header, "Content-Type: application/json")
}

func TestDiagram_AddHttpRequest_HandlesEmptyBody(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "http://example.com/abcdef", nil)
	request.Header.Set("Content-Type", "application/json")

	model, err := aDiagram().
		AddHttpRequest(HttpRequest{Value: request}).
		AddHttpResponse(aResponse()).
		BuildModel()

	assert.Nil(t, err)
	assert.Contains(t, model.LogEntries[0].Body, "")
	assert.Contains(t, model.LogEntries[0].Header, "GET /abcdef")
}

func TestDiagram_AddHttpResponse(t *testing.T) {
	aResponse := HttpResponse{Value: &http.Response{StatusCode: http.StatusNoContent}}

	model, _ := NewDiagram().AddHttpResponse(aResponse).BuildModel()

	assert.Contains(t, model.LogEntries[0].Header, "204 No Content")
}

func TestDiagram_AddMessageRequest(t *testing.T) {
	message := MessageRequest{Header: "H", Body: "B"}

	model, _ := NewDiagram().AddMessageRequest(message).AddHttpResponse(aResponse()).BuildModel()

	assert.Equal(t, model.LogEntries[0], LogEntry{Header: "H", Body: "B"})
}

func TestDiagram_AddMessageResponse(t *testing.T) {
	message := MessageResponse{Header: "H", Body: "B"}

	model, err := aDiagram().AddMessageResponse(message).BuildModel()

	assert.Nil(t, err)
	assert.Equal(t, model.LogEntries[len(model.LogEntries)-1], LogEntry{Header: "H", Body: "B"})
}

func TestDiagram_BuildModel_ErrorIfNoEventsDefined(t *testing.T) {
	_, err := NewDiagram().BuildModel()

	assert.EqualError(t, err, "no events are defined")
}

func TestDiagram_BuildModel_ErrorIfResponseTypeNotFinalEvent(t *testing.T) {
	_, err := aDiagram().AddHttpRequest(aRequest()).BuildModel()

	assert.EqualError(t, err, "final event should be a response type")
}

func TestDiagram_SetsResponseStatus(t *testing.T) {
	aResponse := HttpResponse{Value: &http.Response{StatusCode: http.StatusNoContent}}

	model, _ := NewDiagram().AddHttpResponse(aResponse).BuildModel()

	assert.Equal(t, model.StatusCode, http.StatusNoContent)
}

func TestDiagram_BadgeCSSClass(t *testing.T) {
	tests := []struct {
		status int
		class  string
	}{
		{status: http.StatusOK, class: "badge badge-success"},
		{status: http.StatusInternalServerError, class: "badge badge-danger"},
		{status: http.StatusBadRequest, class: "badge badge-warning"},
	}
	for _, test := range tests {
		t.Run(test.class, func(t *testing.T) {
			aResponse := HttpResponse{Value: &http.Response{StatusCode: test.status}}
			model, _ := NewDiagram().AddHttpResponse(aResponse).BuildModel()
			assert.Equal(t, test.class, model.BadgeClass)
		})
	}
}

func TestFormatContent_PrettyPrintsJSON(t *testing.T) {
	buffer := ioutil.NopCloser(strings.NewReader(`{"a":"b"}`))

	content, err := formatContent(buffer, "application/json")

	assert.Nil(t, err)
	assert.Equal(t, "{\n    \"a\": \"b\"\n}", content)
}

func TestFormatContent_FormatsPlainText(t *testing.T) {
	buffer := ioutil.NopCloser(strings.NewReader(`abcdef`))

	content, err := formatContent(buffer, "text/plain")

	assert.Nil(t, err)
	assert.Equal(t, "abcdef", content)
}

func TestFormatContent_HandlesEmptyBody(t *testing.T) {
	buffer := ioutil.NopCloser(strings.NewReader(""))

	content, err := formatContent(buffer, "application/json")

	assert.Nil(t, err)
	assert.Equal(t, "", content)
}

func TestDocument_SupportsMultipleDiagrams(t *testing.T) {
	document := NewDocument().
		AddDiagram(aDiagram()).
		AddDiagram(aDiagram())

	assert.Len(t, document.Diagrams, 2)
}

func TestDocument_AddTitle(t *testing.T) {
	title := "My document"

	document := NewDocument().AddTitle(title)

	assert.Equal(t, document.Title, title)
}

func TestDocument_AddDescription(t *testing.T) {
	description := "My description"

	document := NewDocument().AddDescription(description)

	assert.Equal(t, document.Description, description)
}

func TestDocument_BuildModel(t *testing.T) {
	document := NewDocument().
		AddTitle("title").
		AddDescription("description").
		AddDiagram(aDiagram())

	model, err := document.BuildModel()

	assert.Nil(t, err)
	assert.Len(t, model.Diagrams, 1)
	assert.Equal(t, model.Title, "title")
	assert.Equal(t, model.Description, "description")
}

func TestDocument_BuildModel_ErrorIfDiagramInvalid(t *testing.T) {
	diagram := aDiagram().AddHttpRequest(aRequest())
	document := NewDocument().
		AddDiagram(diagram)

	_, err := document.BuildModel()

	assert.Error(t, err)
}

func TestDocument_AddsMeta(t *testing.T) {
	document := NewDocument().
		AddDiagram(aDiagram()).
		AddMeta(`{"a": 123}`)

	html, err := document.RenderHTML()

	assert.Nil(t, err)
	assert.Contains(t, html, `<script type="application/json" id="metaJson">{"a": 123}</script>`)
}

func aDiagram() *Diagram {
	return NewDiagram().
		AddHttpRequest(aRequest()).
		AddHttpResponse(aResponse())
}

func aRequest() HttpRequest {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/abcdef", nil)
	req.Header.Set("Content-Type", "application/json")
	return HttpRequest{Value: req}
}

func aResponse() HttpResponse {
	return HttpResponse{Value: &http.Response{StatusCode: http.StatusNoContent}}
}
