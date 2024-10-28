package prompting

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"context"
	"strings"

	"go.uber.org/zap"
)

type PromptService struct {
	Log           *zap.SugaredLogger
	Storage       *storage.StorageReader
	StorageWriter *storage.StorageWriter
}

func NewPromptService(log *zap.SugaredLogger, storage *storage.StorageReader, storagewriter *storage.StorageWriter) *PromptService {
	return &PromptService{
		Log:           log,
		Storage:       storage,
		StorageWriter: storagewriter,
	}
}

func (srv *PromptService) AssemblePrompt(msg entities.WebSocketMessage) (string, error) {
	//Options are necessary, so material array is forced.
	matArray := []string{"options"}
	mats, err := srv.Storage.ReadMaterials(matArray, msg.AssistantContext, context.Background())
	if err != nil {
		srv.Log.Errorln()
	}

	//Initialize empty Prompt
	prompt := "<|begin_of_text|>"

	baseprompt, err := srv.Storage.ReadBasePrompt("languageInterpreter", context.Background())

	player, err := srv.Storage.ReadPlayer(msg.PlayerContext.PlayerUsername, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading BasePrompt from DB")
	}

	location, err := srv.Storage.ReadLocation(msg.AssistantContext.Location, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Location from DB")
	} //Find Setting

	//Declare begin of system prompt
	prompt += "<|start_header_id|>system<|end_header_id|>"
	prompt += baseprompt.Prompt
	//prompt += "You are positioned at the " + location.LocationName + ": " + location.Description //Location
	prompt += srv.getActivityState(msg)
	if msg.MessageType != "innerThoughtEvent" { //Get state from msg
		prompt += "You are provided a "
	}
	prompt += "[\n"
	for _, avac := range mats {
		prompt += `{"action: ` + avac.Name + `", "description: ` + avac.Description + `}` + ",\n"
	}
	prompt = strings.TrimRight(prompt, ",")
	prompt += "] \n"
	prompt += "The following list of strings denotes the chronological chain of events leading up to this point. Take it into consideration."
	prompt += "[\n"
	prompt += player.History
	prompt += "] \n"
	prompt += "You are positioned at " + location.LocationName + ": " + location.Description + ".\n"
	if msg.MessageType != "innerThoughtEvent" {
		prompt += "Choose one of the listed actions by comparing their intention to the action descriptions. Choose the action that best describes the users intention. Use the provided chain of events to solidify your understanding of the users intention."
		prompt += "<|eot_id|>"
		prompt += "<|start_header_id|>user<|end_header_id|>" + msg.Speech + "<|eot_id|>"
	} else {
		prompt += "Choose your next action from the set of provided actions. Let the provided chain of events influence your decision."
		prompt += "<|eot_id|>"
	}
	prompt += "<|start_header_id|>assistant<|end_header_id|>"
	srv.Log.Infoln("Prompt: ", prompt)
	return prompt, nil
}

func (srv *PromptService) AssembleEnvEventPrompt(msg entities.WebSocketMessage) (string, error) {
	prompt := "<|begin_of_text|>"
	baseprompt, err := srv.Storage.ReadBasePrompt(msg.AssistantContext.SelectedBasePrompt, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Baseprompt from DB")
	}

	assistantLocation, err := srv.Storage.ReadLocation(msg.AssistantContext.Location, context.Background())
	prompt += "<|start_header_id|>system<|end_header_id|>"
	prompt += baseprompt.Prompt
	prompt += "You are positioned at " + assistantLocation.LocationName + ": " + assistantLocation.Description + ".\n"
	prompt += srv.getActivityState(msg) + "\n"
	prompt += "<|eot_id|>"
	prompt += "<|start_header_id|>user<|end_header_id|>"
	prompt += "Something happened: \n"
	prompt += msg.Speech + "\n"
	prompt += "Here is a list of possible reactions you can take. " + "\n"
	prompt += "[\n"
	for _, opt := range msg.AssistantContext.AvailableActions {
		action, err := srv.Storage.ReadActionOptionEntity(opt, context.Background())
		if err != nil {
			return prompt, err
		}
		prompt += `{"action: ` + action.ActionName + `", "description: ` + action.Description + `}` + ",\n"
	}
	prompt = strings.TrimRight(prompt, ",")
	prompt += "] \n"
	prompt += "Choose the most appropriate response by comparing the descriptions of the available reactions and output the most suitable reaction in form of a json, like so: {\"action:\" : \"chosen action\"}"
	prompt += "<|eot_id|>"
	prompt += "<|start_header_id|>assistant<|end_header_id|>"
	srv.Log.Infoln("EnvEvent Prompt: ", prompt)
	return prompt, nil
}

func (srv *PromptService) AssembleInstructionsPrompt(msg entities.WebSocketMessage, inst entities.Instructions, basepromptstr string) (string, error) {
	prompt := "<|begin_of_text|>"
	baseprompt, err := srv.Storage.ReadBasePrompt(msg.AssistantContext.SelectedBasePrompt, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Baseprompt from DB")
	}

	player, err := srv.Storage.ReadPlayer(msg.PlayerContext.PlayerUsername, context.Background())
	if err != nil {
		srv.Log.Errorln(err)
	}
	prompt += "<|start_header_id|>system<|end_header_id|>"
	prompt += baseprompt.Prompt
	material, err := srv.Storage.ReadMaterials(inst.Material, msg.AssistantContext, context.Background())
	if err != nil {
		srv.Log.Errorln(err)
	}

	switch inst.Type {
	case "playerSpeechAnalysis": // Will be sent to small LLM
		prompt += inst.StageInstructions + "<|eot_id|>"
		prompt += "<|start_header_id|>user<|end_header_id|>" + msg.Speech + "<|eot_id|>"
		break
	case "speech": // Will be sent to big LLM
		prompt += "Here is what happened so far:"
		prompt += "[\n"
		prompt += player.History
		prompt += "] \n"
		prompt += inst.StageInstructions + "\n"
		if len(inst.Material) > 0 {
			prompt += "You are presented with a selection of materials that you can talk about. It is a list of json containing name and description of each material."

			prompt += "[\n"
			for _, avac := range material {
				prompt += `{"action: ` + avac.Name + `", "description: ` + avac.Description + `}` + ",\n"
			}
			prompt = strings.TrimRight(prompt, ",")
			prompt += "] \n"
		}
		prompt += "<|eot_id|>"
		break
	case "actionquery":
		prompt += inst.StageInstructions + "<|eot_id|>"
		prompt += "You are presented with a selection of materials. It is a list of json containing name and description of each material."
		var focus entities.Material
		hasFocus := false
		counter := 1
		prompt += "[\n"
		for _, mat := range material {
			if mat.Type != "focus" {
				prompt += `{"name": "` + mat.Name + `",` + `"description":"` + mat.Description + `"}` + "\n"
				counter++
			} else {
				focus = mat
				hasFocus = true
			}
		}
		prompt = strings.TrimRight(prompt, ",")
		prompt += "] \n"
		if hasFocus { //Save the focus for last, for relevancy
			prompt += "Currently the focus lies on this material:"
			prompt += `{"name": "` + focus.Name + `",` + `"description":"` + focus.Description + `"}`
		}
		prompt += "<|start_header_id|>user<|end_header_id|>" + msg.Speech + "<|eot_id|>"
	}
	// prompt += location? actionselection, speech, actionquery, speechAnalysis
	// prompt += playerState, etc.?
	return prompt, nil
}

func (srv *PromptService) getActivityState(msg entities.WebSocketMessage) string {
	text := ""
	switch msg.AssistantContext.WalkingState {
	case "idle":
		text += "You currently stand idle. "
	case "patrolling":
		text += "You are currently patrolling the area, searching for things out of place. "
	case "followPlayer":
		text += "You are currently following a user around."
	case "moving":
		text += "You are currently moving towards your destination."
	}
	if msg.AssistantContext.PlayerVisible {
		text += "A user is in your field of vision. "
	}
	if msg.PlayerContext.InConversation {
		text += "You are currently in conversation with a user."
	}
	if msg.MessageType == "innerThoughtEvent" {
		text += "You had nothing to do for a while, and you are bored. What do you want to do?"
	}
	return text

}
