package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	content []byte
	player  *Player
}

type Room struct {
	ID         string
	Players    map[string]*Player
	Host       *Player
	State      *GameState
	broadcast  chan Message
	register   chan *Player
	unregister chan *Player
	mu         sync.RWMutex
}

func (r *Room) run() {

	for {
		select {
		case player := <-r.register:
			// TODO: RECONNECT PLAYERS
			if r.State.started {
				close(player.send)
			} else {
				fmt.Printf("%s added to Room %s\n", player.Name, r.ID)
				r.mu.Lock()
				r.Players[player.ID] = player
				r.mu.Unlock()
			}
		case player := <-r.unregister:
			if _, ok := r.Players[player.ID]; ok {
				player.ws = nil
				fmt.Printf("%s disconnected from Room %s\n", player.Name, r.ID)
			}
		case message := <-r.broadcast:
			// For now just echo the message to all players
			fmt.Printf("%s\n", message.content)
			for _, player := range r.Players {
				select {
				case player.send <- message.content:
				default:
				}
			}
		}
	}
}

func (r *Room) NewPlayer(ws *websocket.Conn, name string) *Player {
	ID := uuid.New()

	player := &Player{
		room: r,
		ID:   ID.String(),
		Name: name,
		send: make(chan []byte, 256),
		ws:   ws,
	}
	r.register <- player

	// go player.readPump()
	// go player.writePump()

	return player
}

func (r *Room) GetPlayer(uuid string) (*Player, bool) {
	defer r.mu.RUnlock()
	r.mu.RLock()
	player, ok := r.Players[uuid]
	return player, ok
}
