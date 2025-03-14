package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type AssistantContext struct {
	Location           string   `json:"location"`
	PlayerVisible      bool     `json:"playerVisible"`
	PlayerAudible      bool     `json:"playerAudible"`
	AssetsInView       []string `json:"assetsInView"`
	AvailableActions   []string `json:"availableActions"`
	WalkingState       string   `json:"walkingState"` //
	FocusedAsset       string   `json:"focusedAsset"` // If following Player and looking together at artwork
	SelectedBasePrompt string   `json:"selectedBasePrompt"`
}

type PlayerContext struct {
	PlayerUsername string   `json:"playerUsername"`
	Location       string   `json:"location"`
	InConversation bool     `json:"inConversation"`
	AssetsInView   []string `json:"assetsInView"`
}
type ActionContext struct {
	ActionName string             `json:"actionName" bson:"actionName"`
	Token      primitive.ObjectID `json:"token" bson:"token"`
	Stage      int                `json:"stage" bson:"stage"`
	Permission bool               `json:"permission" bson:"permission"`
}

type EventContext struct {
	RelevantObjects []string `json:"relevantObjects" bson:"relevantObjects"`
	EventLocation   string   `json:"eventLocation" bson:"eventLocation"`
}
