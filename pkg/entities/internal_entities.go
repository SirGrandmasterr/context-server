package entities

type BasePrompt struct {
	Prompt     string `json:"prompt" bson:"prompt"`
	PromptName string `json:"promptName" bson:"prompt_name"`
}

type Location struct {
	Description  string `json:"description" bson:"description"`
	LocationName string `json:"locationName" bson:"location_name"`
}

type Action struct {
	ActionName  string `json:"actionname" bson:"action_name"`
	Description string `json:"description" bson:"description"`
	Application string `json:"application" bson:"application"`
}

type LlmActionResponse struct {
	ActionName string `json:"action"`
}
