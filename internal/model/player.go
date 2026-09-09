package model

type Player struct {
	UID string
}

var (
	PlayerOne = &Player{
		UID: "player-1",
	}
	PlayerTwo = &Player{
		UID: "player-2",
	}
)
