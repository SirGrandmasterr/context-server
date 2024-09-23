package entities

type PlayerActionReceive struct {
	PlayerId       string
	PlayerLocation string
	PlayerMessage  string
	PlayerVision   string
}

type EnvironmentChangeReceive struct {
	ObjectId       string
	PlayerInvolved bool
	PlayerId       string
	Description    string
}
