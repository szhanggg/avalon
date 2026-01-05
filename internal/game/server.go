package game

import (
	"math/rand/v2"
	"sync"
)

type Server struct {
	mu    sync.RWMutex
	Rooms map[string]*Room
}

func NewServer() *Server {
	return &Server{
		mu:    sync.RWMutex{},
		Rooms: make(map[string]*Room),
	}
}

func (s *Server) GetRoom(id string) (*Room, bool) {
	defer s.mu.Unlock()
	s.mu.Lock()
	room, ok := s.Rooms[id]
	return room, ok
}

func (s *Server) AddRoom(r *Room) {
	s.mu.Lock()
	s.Rooms[r.ID] = r
	s.mu.Unlock()
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRoomID() string {
	b := make([]byte, 4)
	for i := range b {
		// rand.IntN is fast and non-biased
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func (s *Server) NewRoom() *Room {
	for {
		newId := generateRoomID()
		if _, ok := s.GetRoom(newId); !ok {
			room := &Room{
				ID:         newId,
				Players:    make(map[string]*Player),
				State:      NewGameState(),
				broadcast:  make(chan Message),
				register:   make(chan *Player),
				unregister: make(chan *Player),
			}
			s.AddRoom(room)
			return room
		}
	}
}
