package main

import (
	"github.com/h2non/gock"
	"net/http"
	"testing"
)

func TestApi_GetPosts(t *testing.T) {
	testCase := TestCase{
		RequestURL:             "/post",
		RequestMethod:          http.MethodGet,
		ExpectedResponseStatus: http.StatusOK,
		ExpectedResponseBody:   `{"userId":1, "title": "go rulez", "body": "say no more"}`,
		Before: func(test *ApiTest) {
			gock.New("http://example.com").
				Get("/posts").
				Reply(http.StatusOK).
				BodyString(`{"userId":1, "title": "go rulez", "body": "say no more"}`)
		},
	}

	NewApiTest(NewApp()).Run(t, testCase)
}

func TestApi_DeletePost(t *testing.T) {
	testCase := TestCase{
		RequestURL:             "/post/1",
		RequestMethod:          http.MethodDelete,
		ExpectedResponseStatus: http.StatusNoContent,
		Before: func(test *ApiTest) {
			gock.New("http://example.com").
				Delete("/posts/1").
				Reply(http.StatusNoContent)
		},
	}

	NewApiTest(NewApp()).Run(t, testCase)
}

func TestApi_CreatePost(t *testing.T) {
	testCase := TestCase{
		RequestURL:             "/post",
		RequestMethod:          http.MethodPost,
		RequestBody:            `{"userId":1, "title": "go rulez", "body": "say no more"}`,
		ExpectedResponseStatus: http.StatusCreated,
		Before: func(test *ApiTest) {
			gock.New("http://example.com").
				Post("/posts").
				MatchHeader("Content-Type", "application/json").
				BodyString(`{"userId":1, "title": "go rulez", "body": "say no more"}`).
				Reply(http.StatusCreated)
		},
	}

	NewApiTest(NewApp()).Run(t, testCase)
}
