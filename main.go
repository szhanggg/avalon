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
var fragments = template.Must(template.ParseFS(files, "templates/fragments.html"))

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

		fmt.Printf("Compiled Template: %s\n", name)
	}
}

// page data
type PageData struct {
	Room   *game.Room
	Player *game.Player
}

// cookie
func safeCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //only during local
		SameSite: http.SameSiteStrictMode,
	}
}

// room setup logic

func createRoom(w http.ResponseWriter, r *http.Request) {

	room := s.NewRoom()

	player := room.NewPlayer(nil, "guest")

	http.SetCookie(w, safeCookie("uuid", player.ID))
	w.Header().Set("HX-Redirect", "/room/"+room.ID)
	w.WriteHeader(http.StatusNoContent)

}

func getRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	room, ok := s.GetRoom(roomID)

	if !ok {
		// TODO: Handle room not found
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var player *game.Player
	uuid, err := r.Cookie("uuid")
	if err == nil {
		player, ok = room.GetPlayer(uuid.Value)
	}
	if err == http.ErrNoCookie || !ok {
		player = room.NewPlayer(nil, "guest")
		http.SetCookie(w, safeCookie("uuid", player.ID))
	}

	templates["room.html"].ExecuteTemplate(w, "base", PageData{
		Room:   room,
		Player: player,
	})

}

func joinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("roomID")
	_, ok := s.GetRoom(roomID)

	if !ok {
		// TODO: Room not found
		fragments.ExecuteTemplate(w, "error", "That room doesn't exist!")
		return
	}

	w.Header().Set("HX-Redirect", "/room/"+roomID)
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	room, ok := s.GetRoom(roomID)
	if !ok {
		// TODO: Room not found
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	uuid, err := r.Cookie("uuid")
	if err != nil {
		return
	}
	log.Printf("Path: %s | Cookie: %v", r.URL.Path, uuid.Value)

	player, ok := room.GetPlayer(uuid.Value)
	if !ok {
		// TODO: Player can't be found
		return
	}
	player.AttachSocket(ws)

	go player.ReadPump()
	player.WritePump()

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
	r.Get("/room/{roomID}", getRoom)
	r.Get("/join", joinRoom)
	r.Post("/create", createRoom)
	r.Get("/ws/{roomID}", socketHandler)

	log.Fatal(http.ListenAndServe(("localhost:8080"), r))

}
