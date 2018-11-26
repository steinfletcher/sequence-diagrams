package main

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type App struct {
	Router     *mux.Router
	httpClient *http.Client
}

func main() {
	NewApp().Start()
}

func NewApp() *App {
	app := &App{
		Router:     mux.NewRouter(),
		httpClient: &http.Client{Timeout: time.Duration(5) * time.Second},
	}
	app.registerRoutes()
	return app
}

func (a *App) registerRoutes() {
	a.Router.HandleFunc("/post", GetPosts(a.httpClient)).Methods(http.MethodGet)
	a.Router.HandleFunc("/post/{id}", DeletePost(a.httpClient)).Methods(http.MethodDelete)
	a.Router.HandleFunc("/post", CreatePost(a.httpClient)).Methods(http.MethodPost)
}

func (a *App) Start() {
	log.Fatal(http.ListenAndServe(":9999", a.Router))
}

func GetPosts(httpClient *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com/posts", nil)
		res, err := httpClient.Do(req)
		if err != nil || res.StatusCode >= 400 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, _ := ioutil.ReadAll(res.Body)
		w.Write(body)
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeletePost(httpClient *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		req, _ := http.NewRequest(http.MethodDelete, "http://example.com/posts/"+id, nil)
		res, err := httpClient.Do(req)
		if err != nil || res.StatusCode >= 400 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func CreatePost(httpClient *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest(http.MethodPost, "http://example.com/posts", r.Body)
		req.Header.Set("Content-Type", "application/json")
		res, err := httpClient.Do(req)
		if err != nil || res.StatusCode >= 400 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
