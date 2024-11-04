package evaluation

import (
	"Llamacommunicator/pkg/config"
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"Llamacommunicator/pkg/storage"
	"context"

	"go.uber.org/zap"
)

type EvalService struct {
	Log              *zap.SugaredLogger
	Conf             *config.Specification
	AssistantService *assistant.Service
}

func NewEvalService(log *zap.SugaredLogger, conf *config.Specification, serv *assistant.Service) *EvalService {
	return &EvalService{
		Log:              log,
		Conf:             conf,
		AssistantService: serv,
	}

}

var arr_actionSelection_music = []string{
	"Hey, play some tunes.",
	"Could you put on some music, please?",
	"I'd like to listen to some music now.",
	"Play some music, any genre.",
	"Can you play some chill music?",
	"I could use some upbeat music.",
	"Let's get this party started with some music.",
	"I'm in the mood for some classical music.",
	"Play some rock music, please.",
	"I'd like to listen to some jazz.",
	"Can you play some country music?",
	"I'm craving some hip-hop.",
	"Could you play some electronic music?",
	"Let's hear some indie music.",
	"I'm feeling nostalgic, play some old school music.",
	"I need some focus music.",
	"Play some background music, please.",
	"I'd like to listen to a specific playlist.",
	"Can you play a random song?",
	"Surprise me with some music.",
}

var arr_actionSelection_followPlayer = []string{
	"Hey, follow me.",
	"Could you follow me, please?",
	"I need you to follow me.",
	"Let's go, follow my lead.",
	"Follow me, I'm going this way.",
	"Can you follow me, please?",
	"I'm going to the store, come with me.",
	"Let's go for a walk, join me.",
	"I'm going to the park, accompany me.",
	"Follow me, I know a shortcut.",
	"I need you to keep up, stay with me.",
	"Let's go on an adventure, come along.",
	"Can you keep up? Stay close.",
	"I'm going to show you something, come see.",
	"Follow me, I'll lead the way.",
	"Let's explore, come with me.",
	"I'm going to the coffee shop, join me.",
	"I need your help, come with me.",
	"Let's go, follow my lead.",
	"Follow me, I know where we're going.",
}

var arr_artInformation = []string{
	"Who created this piece?",
	"What can be seen in this piece?",
	"What is in the lower right corner of the piece?",
	"Is there a dinosaur in this picture?",
	"Where is the dinosaur in this picture?",
}

func (srv *EvalService) TestActionSelectionPrecision() string {
	//create WebSocketMessage for simulationPurposes
	ac := entities.AssistantContext{
		Location:      "lower gallery",
		PlayerVisible: true,
		PlayerAudible: true,
		AssetsInView:  []string{},
		AvailableActions: []string{
			"walkToVisitor",
			"walkToObject",
			"patrol",
			"standIdle",
			"admireArt",
			"followVisitor",
			"stopFollowingVisitor",
			"investigate",
			"repair",
			"ignore",
			"playMusic",
			"stopMusic",
			"explainWhatYouCanDo",
			"provideArtInformation",
			"continueConversation",
		},
		WalkingState:       "idle",
		FocusedAsset:       "",
		SelectedBasePrompt: "museumAssistant",
	}
	pc := entities.PlayerContext{
		PlayerUsername: "Sir Grandmasterr",
		Location:       "lower_gallery",
		InConversation: false,
		AssetsInView:   []string{},
	}
	msg := entities.WebSocketMessage{
		MessageType:      "speech",
		Speech:           "",
		AssistantContext: ac,
		PlayerContext:    pc,
		ActionContext:    entities.ActionContext{},
		EventContext:     entities.EventContext{},
	}
	ret := ""
	for _, entry := range arr_actionSelection_music {
		msg.Speech = entry
		resp := srv.AssistantService.DetectAction(context.Background(), msg, srv.AssistantService.ClientResponseChannel, 0.8)
		ret += "" + msg.Speech + " & " + resp.ActionName + " & " + "true" + " \\\\ \n"
		ret += "\\hline"
	}
	return ret
}

func (srv *EvalService) TestActionSelectionPrecisionFollowPlayer() string {
	//create WebSocketMessage for simulationPurposes
	ac := entities.AssistantContext{
		Location:      "lower gallery",
		PlayerVisible: true,
		PlayerAudible: true,
		AssetsInView:  []string{},
		AvailableActions: []string{
			"walkToVisitor",
			"walkToObject",
			"patrol",
			"standIdle",
			"admireArt",
			"followVisitor",
			"stopFollowingVisitor",
			"investigate",
			"repair",
			"ignore",
			"playMusic",
			"stopMusic",
			"explainWhatYouCanDo",
			"provideArtInformation",
			"continueConversation",
		},
		WalkingState:       "idle",
		FocusedAsset:       "",
		SelectedBasePrompt: "museumAssistant",
	}
	pc := entities.PlayerContext{
		PlayerUsername: "Sir Grandmasterr",
		Location:       "lower_gallery",
		InConversation: false,
		AssetsInView:   []string{},
	}
	msg := entities.WebSocketMessage{
		MessageType:      "speech",
		Speech:           "",
		AssistantContext: ac,
		PlayerContext:    pc,
		ActionContext:    entities.ActionContext{},
		EventContext:     entities.EventContext{},
	}
	ret := ""
	for _, entry := range arr_actionSelection_followPlayer {
		msg.Speech = entry
		resp := srv.AssistantService.DetectAction(context.Background(), msg, srv.AssistantService.ClientResponseChannel, 0.8)
		ret += "" + msg.Speech + " & " + resp.ActionName + " & " + "true" + " \\\\ \n"
		ret += "\\hline"
	}
	return ret
}

func (srv *EvalService) TestActionSelectionPrecisionFollowPlayerNoWalk() string {
	//create WebSocketMessage for simulationPurposes
	ac := entities.AssistantContext{
		Location:      "lower gallery",
		PlayerVisible: true,
		PlayerAudible: true,
		AssetsInView:  []string{},
		AvailableActions: []string{
			//"walkToVisitor",
			"walkToObject",
			"patrol",
			"standIdle",
			"admireArt",
			"followVisitor",
			"stopFollowingVisitor",
			"investigate",
			"repair",
			"ignore",
			"playMusic",
			"stopMusic",
			"explainWhatYouCanDo",
			"provideArtInformation",
			"continueConversation",
		},
		WalkingState:       "idle",
		FocusedAsset:       "",
		SelectedBasePrompt: "museumAssistant",
	}
	pc := entities.PlayerContext{
		PlayerUsername: "Sir Grandmasterr",
		Location:       "lower_gallery",
		InConversation: false,
		AssetsInView:   []string{},
	}
	msg := entities.WebSocketMessage{
		MessageType:      "speech",
		Speech:           "",
		AssistantContext: ac,
		PlayerContext:    pc,
		ActionContext:    entities.ActionContext{},
		EventContext:     entities.EventContext{},
	}
	ret := ""
	for _, entry := range arr_actionSelection_followPlayer {
		msg.Speech = entry
		resp := srv.AssistantService.DetectAction(context.Background(), msg, srv.AssistantService.ClientResponseChannel, 0.8)
		ret += "" + msg.Speech + " & " + resp.ActionName + " & " + "true" + " \\\\ \n"
		ret += "\\hline"
	}
	return ret
}

func (srv *EvalService) CreateArtInformationNeedleHaystackPrompt(r *storage.StorageReader, w *storage.StorageWriter) []string {
	ac := entities.AssistantContext{
		Location:      "lower gallery",
		PlayerVisible: true,
		PlayerAudible: true,
		AssetsInView:  []string{},
		AvailableActions: []string{
			"walkToVisitor",
			"walkToObject",
			"patrol",
			"standIdle",
			"admireArt",
			"followVisitor",
			"stopFollowingVisitor",
			"investigate",
			"repair",
			"ignore",
			"playMusic",
			"stopMusic",
			"explainWhatYouCanDo",
			"provideArtInformation",
			"continueConversation",
		},
		WalkingState:       "idle",
		FocusedAsset:       "Magical Evaluation",
		SelectedBasePrompt: "museumAssistant",
	}
	pc := entities.PlayerContext{
		PlayerUsername: "Sir Grandmasterr",
		Location:       "lower_gallery",
		InConversation: true,
		AssetsInView:   []string{},
	}
	msg := entities.WebSocketMessage{
		MessageType:      "speech",
		Speech:           "",
		AssistantContext: ac,
		PlayerContext:    pc,
		ActionContext:    entities.ActionContext{},
		EventContext:     entities.EventContext{},
	}
	ret := []string{}
	for i := 0; i < 10; i++ {

		for _, entry := range arr_artInformation {
			w.ResetPlayerHistory("Sir Grandmasterr")
			w.UpdatePlayerHistory("Sir Grandmasterr", "VISITOR: "+entry)
			msg.Speech = entry
			str := srv.AssistantService.StreamAssistantTest(msg, entities.Instructions{
				Stage:              2,
				StageInstructions:  "Read the last entry in the player history to understand what the visitor wants to know about the piece of Art that is in both of your focus. Answer the visitor directly, providing the requested information if it is found in the pieces' description. Extrapolate, if necessary, but point it out if you do. If the information is not contained in the description, apologize for not knowing.",
				Type:               "speech",
				Material:           []string{"focus"},
				ResultVar:          "",
				Limit:              0,
				PermissionRequired: false,
				BasePrompt:         "museumAssistant",
				LlmSize:            "big",
			}, 1.2, 8)
			ret = append(ret, str)
		}
		ret = append(ret, "=============================================================")
	}
	return ret
}
