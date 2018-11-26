package main

import (
	"bytes"
	"fmt"
	"github.com/steinfletcher/sequence-diagrams"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/h2non/gock"
)

type Interaction struct {
	Request *http.Request
	Mock    gock.Mock
}

type ApiTest struct {
	App                    *App
	GockInteractions       []Interaction
	capturedInitialRequest *http.Request
	capturedFinalResponse  *http.Response
}

type TestCase struct {
	Name                   string
	RequestMethod          string
	RequestURL             string
	RequestBody            string
	ExpectedResponseBody   string
	ExpectedResponseStatus int
	Before                 func(test *ApiTest)
}

func NewApiTest(app *App) *ApiTest {
	return &ApiTest{App: app, GockInteractions: []Interaction{}}
}

func (a *ApiTest) Run(t *testing.T, spec TestCase) {
	defer gock.OffAll()
	defer a.renderSequenceDiagram(spec)
	gock.Observe(func(request *http.Request, mock gock.Mock) {
		gock.DumpRequest(request, mock)
		a.GockInteractions = append(a.GockInteractions, Interaction{Request: request, Mock: mock})
	})

	gock.Intercept()

	if spec.Before != nil {
		spec.Before(a)
	}

	res := a.runTest(spec)

	assertUnmatchedRequests(t)
	assertResponse(t, spec, res)
}

func (a *ApiTest) runTest(testCase TestCase) *httptest.ResponseRecorder {
	req := buildRequestFromTestCase(testCase)
	a.capturedInitialRequest = buildRequestFromTestCase(testCase)
	res := httptest.NewRecorder()
	a.App.Router.ServeHTTP(res, req)
	a.capturedFinalResponse = res.Result()
	return res
}

func buildRequestFromTestCase(testCase TestCase) *http.Request {
	req, _ := http.NewRequest(testCase.RequestMethod, testCase.RequestURL, bytes.NewBufferString(testCase.RequestBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func assertUnmatchedRequests(t *testing.T) {
	requests := gock.GetUnmatchedRequests()
	if len(requests) > 0 {
		t.Fatalf("Unmatched gock requests: %#+v", requests[0])
	}
}

func assertResponse(t *testing.T, test TestCase, res *httptest.ResponseRecorder) {
	assert.Equal(t, test.ExpectedResponseStatus, res.Code, test.Name)
	if test.ExpectedResponseBody != "" {
		assert.JSONEq(t, test.ExpectedResponseBody, res.Body.String(), test.Name)
	}
}

func (a *ApiTest) renderSequenceDiagram(spec TestCase) {
	diagram, err := a.diagramFromInteractions(spec)
	if err != nil {
		panic(err)
	}

	html, err := sequence.NewDocument().
		AddTitle(fmt.Sprintf("%s %s", spec.RequestMethod, spec.RequestURL)).
		AddDescription(spec.Name).
		AddDiagram(diagram).
		RenderHTML()

	if err != nil {
		panic(err)
	}

	fmt.Println(html)
}

func (a *ApiTest) diagramFromInteractions(spec TestCase) (*sequence.Diagram, error) {
	patchedURL, _ := url.Parse(fmt.Sprintf("http://consumer/%s", a.capturedInitialRequest.URL))
	a.capturedInitialRequest.URL = patchedURL

	// add the initial http request into the app under test
	d := sequence.NewDiagram().
		AddTitle(fmt.Sprintf("%s %s", a.capturedInitialRequest.Method, a.capturedInitialRequest.URL.Path)).
		AddSubTitle(spec.Name).
		AddHttpRequest(sequence.HttpRequest{
			Value:  a.capturedInitialRequest,
			Source: "consumer",
			Target: "app",
		})

	// add all gock interactions
	for _, interaction := range a.GockInteractions {
		d.AddHttpRequest(sequence.HttpRequest{
			Value:  interaction.Request,
			Source: "app",
			Target: interaction.Request.Host,
		})
		if interaction.Mock != nil {
			d.AddHttpResponse(sequence.HttpResponse{
				Value:  buildResponseFromGockMockResponse(interaction.Mock.Response()),
				Source: interaction.Request.Host,
				Target: "app",
			})
		}
	}

	// add the final response to the consumer (this is not returned from gock)
	return d.AddHttpResponse(sequence.HttpResponse{
		Source: "app",
		Target: "consumer",
		Value:  a.capturedFinalResponse,
	}), nil
}

func buildResponseFromGockMockResponse(gockResponse *gock.Response) *http.Response {
	return &http.Response{
		Body:          ioutil.NopCloser(bytes.NewReader(gockResponse.BodyBuffer)),
		Header:        gockResponse.Header,
		StatusCode:    gockResponse.StatusCode,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(gockResponse.BodyBuffer)),
	}
}
