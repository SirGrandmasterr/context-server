package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type WebSocketMessage struct {
	MessageType string `json:"messageType"`
	//PlayerActionType string `json:"playerActionType"`
	Speech string `json:"speech"`
	//Token            primitive.ObjectID `json:"token"`

	AssistantContext AssistantContext `json:"assistantContext"`
	PlayerContext    PlayerContext    `json:"playerContext"`
	ActionContext    ActionContext    `json:"actionContext"`
	//EventContext     EventContext     `json:"eventContext"`
}

type WebSocketAnswer struct {
	Type       string             `json:"type"`
	Text       string             `json:"text"`
	ActionName string             `json:"actionName"`
	Token      primitive.ObjectID `json:"token"`
	Stage      int                `json:"stage"`
}
