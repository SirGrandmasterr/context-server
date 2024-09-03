package entities

type InitJson struct {
	Actions   []Action         `json:"actions"`
	Objects   []RelevantObject `json:"objects"`
	Locations []Player         `json:"locations"`
}
