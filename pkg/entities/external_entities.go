package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmotionalTrigger struct {
	ID            int    `json:"id"`
	Description   string `json:"description"`
	TargetEmotion string `json:"targetEmotion"`
	Intensity     int    `json:"intensity"`
}

type EmotionalState struct {
	Emotions map[string]int     `json:"emotions"`
	Triggers []EmotionalTrigger `json:"triggers"`
}

/*
{
    "Emotions": {
        "Joy": 50,
        "Trust": 60,
        "Fear": 10,
        "Surprise": 0,
        "Sadness": 0,
        "Disgust": 0,
        "Anger": 0,
        "Anticipation": 30
    },
    "Triggers": [
        {
            "id": 1,
            "description": "Very angry",
            "targetEmotion": "delectus aut autem",
            "intensity": 100
        },
        {
            "id": 2,
            "description": "Reason why I'm angry",
            "targetEmotion": "delectus aut autem",
            "intensity": 100
        }
    ]
}
	Schema:

	{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Generated schema for Root",
  "type": "object",
  "properties": {
    "Emotions": {
      "type": "object",
      "properties": {
        "Joy": {
          "type": "number"
        },
        "Trust": {
          "type": "number"
        },
        "Fear": {
          "type": "number"
        },
        "Surprise": {
          "type": "number"
        },
        "Sadness": {
          "type": "number"
        },
        "Disgust": {
          "type": "number"
        },
        "Anger": {
          "type": "number"
        },
        "Anticipation": {
          "type": "number"
        }
      },
      "required": [
        "Joy",
        "Trust",
        "Fear",
        "Surprise",
        "Sadness",
        "Disgust",
        "Anger",
        "Anticipation"
      ]
    },
    "Triggers": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {
            "type": "number"
          },
          "description": {
            "type": "string"
          },
          "targetEmotion": {
            "type": "string",
            "enum": [
              "Joy",
              "Trust",
              "Fear",
              "Surprise",
              "Sadness",
              "Disgust",
              "Anger",
              "Anticipation"
            ]
          },
          "intensity": {
            "type": "number"
          }
        },
        "required": [
          "id",
          "description",
          "targetEmotion",
          "intensity"
        ]
      }
    }
  },
  "required": [
    "Emotions",
    "Triggers"
  ]
}
*/
// AssistantContext represents the overall context sent from Unity.
type AssistantContext struct {
	Location           string         `json:"location"`
	PlayerVisible      bool           `json:"playerVisible"`
	PlayerAudible      bool           `json:"playerAudible"`
	AssetsInView       []string       `json:"assetsInView"`
	AvailableActions   []string       `json:"availableActions"`
	WalkingState       string         `json:"walkingState"` //
	FocusedAsset       string         `json:"focusedAsset"` // If following Player and looking together at artwork
	SelectedBasePrompt string         `json:"selectedBasePrompt"`
	EmotionalState     EmotionalState `json:"emotionalState"` // New field for the emotional state
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
