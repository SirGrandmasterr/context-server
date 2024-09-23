package entities

type WebSocketMessage struct {
	MessageType      string `json:"messageType"`
	PlayerActionType string `json:"playerActionType"`
	Speech           string `json:"speech"`

	AssistantContext AssistantContext `json:"assistantContext"`
	PlayerContext    PlayerContext    `json:"playerContext"`
}

type WebSocketAnswer struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	ActionName string `json:"actionName"`
}
