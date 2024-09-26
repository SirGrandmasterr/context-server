package entities

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

type RelevantObject struct {
	ObjectName     string   `json:"objectname" bson:"object_name"`
	ObjectType     string   `json:"objecttype" bson:"object_type"`
	ObjectLocation string   `json:"objectlocation" bson:"object_location"`
	Description    string   `json:"description" bson:"description"`
	Actions        []string `json:"actions" bson:"actions"`
}

type ActionContext struct {
	ActionName string `json:"actionName" bson:"actionName"`
	Token      string `json:"token" bson:"token"`
	Stage      int    `json:"stage" bson:"stage"`
	Permission bool   `json:"permission" bson:"permission"`
}

/*

{
"messageType":"speech",
"playerActionType":"speech",
"speech":" Please be so kind and follow me.",
"assistantContext":
	{
	"location":"",
	"playerVisible":false,
	"PlayerAudible":false,
	"AssetsInView":["Pixelated_Woman",
	"Desert_Wanderer",
	"Psychedelic_Dog",
	"Aquarell_City",
	"True_Art"],
	"AvailableActions":
		[
		"stand_idle",
		"patrol",
		"followPlayer",
		"play_music",
		"stop_music",
		"warn_player",
		"talk_and_follow"
		],
	"WalkingState":"idle",
	"FocusedAsset":""
	},
"playerContext":
	{
	"Location":"unknown",
	"AssetsInView":[""],
	"InConversation":false,
	"PlayerId":""
	}
}

*/
