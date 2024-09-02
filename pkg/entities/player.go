package entities

type Player struct {
	ID       string
	Username string
	Password string
}

type PlayerAction struct {
	PlayerId       string
	PlayerLocation string
	PlayerMessage  string
	PlayerVision   string
}
