package entities

type WebSocketMessage struct {
	MessageType             string //AssistantUpdate, PlayerAction
	Speech                  string
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
	Type string `json:"type"`
	Text string `json:"text"`
}
