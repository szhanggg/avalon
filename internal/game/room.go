package game

type Message struct {
	content []byte
	player  *Player
}

type Room struct {
	ID         string
	Players    map[string]*Player
	State      *GameState
	broadcast  chan Message
	register   chan *Player
	unregister chan *Player
}

func (r *Room) run() {

	for {
		select {
		case player := <-r.register:
			if r.State.started {
				close(player.send)
			} else {
				r.Players[player.ID] = player
			}
		case player := <-r.unregister:
			if _, ok := r.Players[player.ID]; ok {
				delete(r.Players, player.ID)
				close(player.send)
			}
		case message := <-r.broadcast:
			// For now just echo the message to all players
			for _, player := range r.Players {
				select {
				case player.send <- message.content:
				default:
					delete(r.Players, player.ID)
					close(player.send)
				}
			}
		}
	}

}
