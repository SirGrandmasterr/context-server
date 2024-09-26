package entities

type InitJson struct {
	Actions     []Action         `json:"actions"`
	Objects     []RelevantObject `json:"objects"`
	Locations   []Location       `json:"locations"`
	BasePrompts []BasePrompt     `json:"baseprompts"`
	Players     []Player         `json:"players"`
}
