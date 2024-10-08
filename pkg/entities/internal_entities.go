package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type BasePrompt struct {
	Prompt     string `json:"prompt" bson:"prompt"`
	PromptName string `json:"promptName" bson:"prompt_name"`
}

type Location struct {
	Description  string `json:"description" bson:"description"`
	LocationName string `json:"locationName" bson:"location_name"`
}

type Action struct {
	ActionName   string         `json:"actionname" bson:"action_name"`
	Description  string         `json:"description" bson:"description"`
	Stages       int            `json:"stages" bson:"stages"`
	Instructions []Instructions `json:"instructions" bson:"instructions"`
}

type Instructions struct {
	Stage              int      `json:"stage" bson:"stage"`
	StageInstructions  string   `json:"stage_instructions" bson:"stage_instructions"`
	Type               string   `json:"type" bson:"type"` //actionselection, speech, actionquery, speechAnalysis
	Material           []string `json:"material" bson:"material"`
	ResultVar          string   `json:"resultVar" bson:"resultVar"`
	Limit              int      `json:"limit" bson:"limit"` //Word limit in speech analysis type
	PermissionRequired bool     `json:"permissionRequired" bson:"permissionRequired"`
}

type LlmActionResponse struct {
	ActionName string `json:"action"`
}

type LlmAnalysisResult struct {
	Result string `json:"result"`
}

type Material struct {
	Type        string `json:"type" bson:"type"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

// Used as a temporary file to recognize a
type ActionToken struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Description  string             `json:"description" bson:"description"`
	CurrentStage int                `json:"currentStage" bson:"currentStage"`
}

/*
 "stages": 4,
            "instructions": [
                {"stage": 1, "instructions": "This action makes you show the visitor the way to a different location.", "type": "actionselection"},
                {"stage": 2, "instructions": "Inform the visitor what location you'll be leading him to.", "type": "speech"},
                {"stage": 3, "instructions": "Analyze the words of the visitor to determine which of these locations you should approach: ", "type": "actionquery"},
                {"stage": 4, "instructions": "Inform the visitor that you have arrived at the location.", "type": "speech"}
            ]
*/
