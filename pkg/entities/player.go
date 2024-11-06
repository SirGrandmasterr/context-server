package entities

type Player struct {
	ID           string   `json:"_id" bson:"_id"`
	Username     string   `json:"username" bson:"username"`
	Password     string   `json:"password" bson:"password"`
	HistoryArray []string `json:"historyArray" bson:"historyarray"`
}
