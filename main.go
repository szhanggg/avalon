package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/szhanggg/avalon/internal/game"
)

//go:embed templates
var files embed.FS

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var templates = make(map[string]*template.Template)

// setup game server
var s = game.NewServer()

func initTemplates() {
	pages, err := fs.Glob(files, "templates/pages/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		name := path.Base(page)
		tmpl, err := template.ParseFS(files, page, "templates/base.html")
		if err != nil {
			log.Fatal(err)
		}
		templates[name] = tmpl

		fmt.Printf("%s\n", name)
	}
}

func createRoom(w http.ResponseWriter, r *http.Request) {

}

func getRoom(w http.ResponseWriter, r *http.Request) {

}

func main() {

	r := chi.NewRouter()

	// Setup middleware

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	// Setup templates
	initTemplates()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templates["index.html"].ExecuteTemplate(w, "base", nil)
	})
	r.Post("/create", createRoom)

	log.Fatal(http.ListenAndServe(("localhost:8080"), r))

}
