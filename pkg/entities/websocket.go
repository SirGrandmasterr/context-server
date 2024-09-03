package entities

type WebSocketMessage struct {
	MessageType             string //AssistantUpdate, PlayerAction
	AssistantUpdateType     string //LocationChange, FoV Object change, Music Change, Light change
	PlayerActionType        string
	PlayerActionAssets      []string
	PlayerActionLocation    string
	AssistantUpdateLocation string
	AssistantUpdateAssets   []string
	HasPlayerUpdate         bool
	HasAssistantUpdate      bool
}

type WebSocketAnswer struct {
	Speech     bool
	Speechtext string
	Action     bool
}
