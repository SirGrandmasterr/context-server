package entities

type RelevantObject struct {
	ObjectName  string   `json:"objectname"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
}
