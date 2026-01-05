package game

type GameState struct {
	started bool
}

func NewGameState() *GameState {
	return &GameState{
		started: false,
	}
}
