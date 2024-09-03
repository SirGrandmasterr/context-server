package entities

type Action struct {
	ActionName  string `json:"actionname", bson:"action_name"`
	Description string `json:"description", bson:"description"`
	Application string `json:"application", bson:"application"`
}
