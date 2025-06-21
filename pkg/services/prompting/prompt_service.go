package prompting

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"context"
	"encoding/json"
	"strconv"
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

// AssemblePrompt creates a prompt for the LLM to select an action based on user speech or internal thought.
func (srv *PromptService) AssemblePrompt(msg entities.WebSocketMessage) (string, error) {
	mats, err := srv.Storage.ReadMaterials([]string{"options"}, msg.AssistantContext, msg.PlayerContext, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading materials for AssemblePrompt:", err)
		// Return a default prompt or error indication if material loading fails
		// For now, we'll proceed but log the error.
	}

	prompt := "<|begin_of_text|>"
	selectedBaseprompt := msg.AssistantContext.SelectedBasePrompt
	if selectedBaseprompt == "" {
		selectedBaseprompt = "galleryGuideInterpreter"
	}
	// Use a specific base prompt for action selection, emphasizing the gallery guide persona.
	baseprompt, err := srv.Storage.ReadBasePrompt(selectedBaseprompt, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading BasePrompt 'galleryGuideInterpreter':", err)
		// Fallback to a generic instruction if the specific base prompt is missing
		baseprompt = entities.BasePrompt{Prompt: "You are an AI assistant. Your task is to select an action."}
	}
	hist := srv.Storage.ReadPlayerHistory(5, msg.PlayerContext.PlayerUsername) // Limiting history to last 5 exchanges for brevity

	location, err := srv.Storage.ReadLocation(msg.AssistantContext.Location, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Location from DB:", err)
		location.LocationName = "an unspecified area"
		location.Description = "The current location is unknown."
	}

	prompt += "<|start_header_id|>system<|end_header_id|>\n"
	prompt += baseprompt.Prompt + "\n\n" // Ensure newline separation for clarity

	prompt += "CURRENT SITUATION:\n"
	prompt += srv.getActivityState(msg) + "\n"
	prompt += "You are currently in: " + location.LocationName + " (" + location.Description + ").\n\n"

	prompt += "AVAILABLE ACTIONS:\n"
	prompt += "Here is a list of actions you can choose from. Each action has a 'name' and a 'description' explaining when to use it.\n"
	prompt += "[\n"
	for i, avac := range mats {
		prompt += `  {"action": "` + avac.Name + `", "description": "` + strings.ReplaceAll(avac.Description, "\"", "'") + `"}` // Escape quotes in description
		if i < len(mats)-1 {
			prompt += ",\n"
		} else {
			prompt += "\n"
		}
	}
	prompt += "]\n\n"

	prompt += "CONVERSATION HISTORY (most recent first):\n"
	if len(hist) == 0 {
		prompt += "No recent conversation history.\n"
	} else {
		for _, entry := range hist {
			prompt += entry + "\n" // Assumes history entries are already formatted like "VISITOR: ..." or "ASSISTANT: ..."
		}
	}
	prompt += "\n"

	if msg.MessageType != "innerThoughtEvent" {
		prompt += "TASK:\nBased on the visitor's last statement (see below), the conversation history, and the current situation, identify which of the AVAILABLE ACTIONS best matches their request or your assessment of the situation. \n"
		prompt += "Respond *only* with a JSON object containing the 'action' name you have chosen. For example: `{\"action\": \"action_name\"}`.\n"
		prompt += "<|eot_id|>\n"
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "Visitor's last statement: \"" + msg.Speech + "\"\n"
		prompt += "<|eot_id|>\n"
	} else {
		prompt += "TASK:\nBased on the current situation and conversation history, decide your next action from the AVAILABLE ACTIONS. Consider what would be most helpful or appropriate for a gallery assistant to do now. If there's an urgent task implied by the situation (e.g., something needs repair), prioritize that.\n"
		prompt += "Respond *only* with a JSON object containing the 'action' name you have chosen. For example: `{\"action\": \"action_name\"}`.\n"
		prompt += "<|eot_id|>\n"
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "What is your next autonomous action?\n"
		prompt += "<|eot_id|>\n"
	}
	prompt += "<|start_header_id|>assistant<|end_header_id|>\n" // LLM response starts here
	srv.Log.Infoln("Assembled Action Prompt: ", prompt)
	return prompt, nil
}

// AssembleEnvEventPrompt creates a prompt for the LLM to react to an environmental event.
func (srv *PromptService) AssembleEnvEventPrompt(msg entities.WebSocketMessage) (string, error) {
	prompt := "<|begin_of_text|>"
	baseprompt, err := srv.Storage.ReadBasePrompt(msg.AssistantContext.SelectedBasePrompt, context.Background()) // Should be "museumAssistant"
	if err != nil {
		srv.Log.Errorln("Error reading BasePrompt from DB for EnvEvent:", err)
		baseprompt = entities.BasePrompt{Prompt: "You are a gallery assistant."}
	}

	assistantLocation, err := srv.Storage.ReadLocation(msg.AssistantContext.Location, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading assistant location for EnvEvent:", err)
		assistantLocation.LocationName = "an unspecified area"
		assistantLocation.Description = "The current location is unknown."
	}

	prompt += "<|start_header_id|>system<|end_header_id|>\n"
	prompt += baseprompt.Prompt + "\n\n"
	prompt += "CURRENT SITUATION:\n"
	prompt += "You are in: " + assistantLocation.LocationName + " (" + assistantLocation.Description + ").\n"
	prompt += srv.getActivityState(msg) + "\n\n"
	prompt += "An event has occurred in the gallery that requires your attention.\n"
	prompt += "Event Details: \"" + msg.Speech + "\"\n\n" // msg.Speech here describes the event

	prompt += "AVAILABLE REACTIONS (Actions):\n"
	prompt += "Here is a list of possible reactions. Choose the most appropriate one.\n"
	prompt += "[\n"
	for i, opt := range msg.AssistantContext.AvailableActions {
		action, err := srv.Storage.ReadActionOptionEntity(opt, context.Background())
		if err != nil {
			srv.Log.Warnln("Could not read action entity for env event prompt:", opt, err)
			continue
		}
		prompt += `  {"action": "` + action.ActionName + `", "description": "` + strings.ReplaceAll(action.Description, "\"", "'") + `"}`
		if i < len(msg.AssistantContext.AvailableActions)-1 {
			prompt += ",\n"
		} else {
			prompt += "\n"
		}
	}
	prompt += "]\n\n"
	prompt += "TASK:\nBased on the event details and the descriptions of the available reactions, choose the most appropriate reaction. \n"
	prompt += "Respond *only* with a JSON object containing the 'action' name of your chosen reaction. For example: `{\"action\": \"chosen_action_name\"}`.\n"
	prompt += "<|eot_id|>\n"
	prompt += "<|start_header_id|>user<|end_header_id|>\n"
	prompt += "How do you react to this event: \"" + msg.Speech + "\"?\n"
	prompt += "<|eot_id|>\n"
	prompt += "<|start_header_id|>assistant<|end_header_id|>\n"
	srv.Log.Infoln("Assembled EnvEvent Prompt: ", prompt)
	return prompt, nil
}

// AssembleInstructionsPrompt creates a prompt for executing a specific stage of an action.
func (srv *PromptService) AssembleInstructionsPrompt(msg entities.WebSocketMessage, inst entities.Instructions, basepromptstr string) (string, error) {
	prompt := "<|begin_of_text|>"
	// Use the baseprompt name specified in the instruction, or default to museumAssistant
	finalBasePromptName := inst.BasePrompt
	srv.Log.Infoln("Inst.BasePrompt: ", finalBasePromptName)
	// If none are in the instructions, check if the webSocketMessage sent something
	if finalBasePromptName == "" {
		srv.Log.Infoln("SelectedBasePrompt: ", finalBasePromptName)
		finalBasePromptName = msg.AssistantContext.SelectedBasePrompt
	}
	// TODO: Create an Override system in the instructions to decide in the action instructions whether the msg BasePrompt should be used or a special one.
	if finalBasePromptName == "" {
		finalBasePromptName = "galleryGuideInterpreter" // Default to museumAssistant if not specified
	}

	baseprompt, err := srv.Storage.ReadBasePrompt(finalBasePromptName, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading BasePrompt '"+finalBasePromptName+"' from DB:", err)
		// Fallback to a generic instruction if the specific base prompt is missing
		baseprompt = entities.BasePrompt{Prompt: "You are a helpful gallery assistant."}
	}
	hist := srv.Storage.ReadPlayerHistory(40, msg.PlayerContext.PlayerUsername) // Shorter history for instruction-specific context

	prompt += "<|start_header_id|>system<|end_header_id|>\n"
	prompt += baseprompt.Prompt + "\n\n" // Base persona prompt

	// Provide materials if any
	if len(inst.Material) > 0 {
		material, err := srv.Storage.ReadMaterials(inst.Material, msg.AssistantContext, msg.PlayerContext, context.Background())
		if err != nil {
			srv.Log.Errorln("Error reading materials for instruction prompt:", err)
		} else if len(material) > 0 {
			prompt += "RELEVANT INFORMATION (Materials):\n"
			for _, mat := range material {
				prompt += `- ` + mat.Name + `: ` + mat.Description + `\n`
			}
			prompt += "\n"
		}
	}

	// Conversation history
	prompt += "RECENT CONVERSATION HISTORY (if relevant):\n"
	if len(hist) == 0 {
		prompt += "No recent conversation history available for this specific task.\n"
	} else {
		for _, entry := range hist {
			prompt += entry + "\n"
		}
	}
	prompt += "\n"

	// Emotional state
	prompt += "YOUR EMOTIONAL STATE \n"
	prompt += "Let your emotional state and its triggers strongly your answer and its wording."
	prompt += "EMOTIONAL VALUES: \n"
	for em, _ := range msg.AssistantContext.EmotionalState.Emotions {
		prompt += `- ` + em + `: ` + strconv.Itoa(msg.AssistantContext.EmotionalState.Emotions[em]) + `\n`
	}
	prompt += "EMOTIONAL TRIGGERS: \n"
	for _, em := range msg.AssistantContext.EmotionalState.Triggers {
		prompt += `- ` + `This caused ` + strconv.Itoa(em.Intensity) + "/100" + em.TargetEmotion + `: ` + em.Description + `\n`
	}

	// Specific instruction for the current stage
	prompt += "YOUR CURRENT TASK:\n"
	prompt += inst.StageInstructions + "\n\n" // This is the core instruction for the LLM for this stage

	switch inst.Type {
	case "playerSpeechAnalysis":
		prompt += "The visitor has just said: \"" + msg.Speech + "\"\n"
		prompt += "Based on their statement and your task, provide the result *only* in JSON format, like so: `{\"result\": \"extracted_information\"}`.\n"
		prompt += "<|eot_id|>\n"
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "Visitor's statement: \"" + msg.Speech + "\"\nWhat is the result of your analysis based on the task?\n" // User part reiterates the input
		prompt += "<|eot_id|>\n"
	case "speech":
		// The main instruction is already in `inst.StageInstructions`.
		// We need to guide the LLM to generate a speech response.
		prompt += "Formulate a concise response to the visitor based on this task. Speak directly to them. Let your emotional state strongly influence your answer and its wording.\n"
		prompt += "Keep answers short and interactive unless a technical answer requires a small monologue. Let your emotional triggers influence the flow of the conversation. DO NOT LITERALLY STATE YOUR FEELINGS UNLESS ASKED DIRECTLY."
		prompt += "Use the following tags at appropriate places to augment the emotional impact of your answer: <laugh>, <chuckle>, <sigh>, <cough>, <sniffle>, <groan>, <yawn>, <gasp>. THESE TAGS ARE LITERAL AND NOT TO BE IMPROVISED UPON.\n"
		prompt += "Do not use any kind of actiontags and instead describe noises using either the specific tags above, or by utilizing onomatopoeia. Example: Instead of *yells*, write out the yell like \"AAAAAAH!\". Instead of \"(shakes head)\" use a fitting tag like <chuckle> or leave it out entirely. Instead of *thinks strongly*, write \"Hmmmm.\"."
		prompt += "<|eot_id|>\n"
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "Interlocutor's last relevant statement (if any, otherwise consider the general context): \"" + msg.Speech + "\"\nWhat do you say?\n"
		prompt += "<|eot_id|>\n"
	case "actionquery":
		// This type expects the LLM to choose from the provided materials.
		prompt += "The visitor's relevant statement is: \"" + msg.Speech + "\"\n"
		prompt += "Based on their statement, your task, and the RELEVANT INFORMATION provided above, select the most appropriate item. \n"
		prompt += "Respond *only* with a JSON object containing the 'name' of the selected item, like so: `{\"result\": \"name_of_selected_item\"}`.\n"
		prompt += "<|eot_id|>\n"
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "Visitor's statement: \"" + msg.Speech + "\"\nWhich item do you select based on your task?\n"
		prompt += "<|eot_id|>\n"
	case "reactiveEmotionalStateAnalysis":
		currentEmotionsJSON, _ := json.Marshal(msg.AssistantContext.EmotionalState)
		currentEmotionsJSONString := string(currentEmotionsJSON)
		// Additions to the system prompt for emotional analysis:
		prompt += "ADDITIONAL INSTRUCTIONS FOR EMOTIONAL ANALYSIS:\n"
		prompt += "Your current internal emotional state is represented by the following JSON object: " + currentEmotionsJSONString + "\n"
		prompt += "The emotion values range from 0 (emotion not present) to 100 (emotion very intense).\n"
		prompt += "Your task is to analyze the RECENT CONVERSATION HISTORY and the VISITOR'S LATEST STATEMENT (below) and update your emotional state accordingly.\n"
		prompt += "Consider how specific events or phrases might influence your emotions. For example:\n"
		prompt += "  - A sincere apology for a mistake might significantly lower 'anger' and slightly raise 'neutral' or 'relief'.\n"
		prompt += "  - A direct insult could sharply increase 'anger' or 'sadness'.\n"
		prompt += "  - A compliment or positive feedback might increase 'joy'.\n"
		prompt += "  - Unexpected news or events could increase 'surprise'.\n"
		prompt += "After your analysis, you MUST output your *new, updated* emotional state.\n"
		prompt += "The output MUST be a JSON object containing all the original emotion keys with their new values. Respond *only* with this JSON object.\n"
		// Example of the expected output format (though GBNF will enforce this):
		prompt += "Example Output Format: `{\"anger\": 5, \"joy\": 60, \"sadness\": 0, ...etc.}`\n\n"
		prompt += "<|eot_id|>\n" // End of system message

		// User message for this instruction type
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "Visitor's latest statement: \"" + msg.Speech + "\"\n"
		prompt += "Given your instructions, the conversation history, your current emotional state, and my latest statement, what is your updated emotional state? Provide ONLY the JSON object.\n"
		prompt += "<|eot_id|>\n"

	default:
		// Generic fallback if type is not specifically handled for user prompt part
		prompt += "<|eot_id|>\n"
		prompt += "<|start_header_id|>user<|end_header_id|>\n"
		prompt += "Visitor's last statement: \"" + msg.Speech + "\"\nProceed with the task.\n"
		prompt += "<|eot_id|>\n"
	}

	prompt += "<|start_header_id|>assistant<|end_header_id|>\n"
	srv.Log.Infoln("Assembled Instruction Prompt ("+inst.Type+"): ", prompt)
	return prompt, nil
}

// getActivityState provides a textual description of the assistant's current state.
func (srv *PromptService) getActivityState(msg entities.WebSocketMessage) string {
	text := ""
	switch msg.AssistantContext.WalkingState {
	case "idle":
		text += "You are currently available to assist visitors. "
	case "patrolling":
		text += "You are currently patrolling the gallery. "
	case "followPlayer":
		text += "You are currently accompanying a visitor. "
	case "moving":
		text += "You are currently moving to a destination. "
	default:
		text += "Your current walking state is " + msg.AssistantContext.WalkingState + ". "
	}

	if msg.AssistantContext.PlayerVisible {
		text += "A visitor is nearby. "
	} else {
		text += "There are no visitors immediately visible. "
	}

	if msg.PlayerContext.InConversation {
		text += "You are currently speaking with a visitor. "
	} else {
		text += "You are not currently in a direct conversation. "
	}

	if msg.MessageType == "innerThoughtEvent" {
		text += "It's been quiet for a moment. You are considering what might be useful or appropriate to do next in the gallery. "
	}
	return strings.TrimSpace(text)
}

// AssembleActionGrammarEnum creates a GBNF grammar for selecting an action.
func (srv *PromptService) AssembleActionGrammarEnum(msg entities.WebSocketMessage) string {
	if len(msg.AssistantContext.AvailableActions) == 0 {
		// Fallback grammar if no actions are available, though this case should ideally be handled earlier.
		return `root ::= "{\"action\": \"ignore\"}"`
	}
	var actionLiterals []string
	for _, actionName := range msg.AssistantContext.AvailableActions {
		actionLiterals = append(actionLiterals, `"\"`+actionName+`\""`) //  `"\"action_name\""`
	}
	// GBNF for {"action": "action_name"}
	// Example: action ::= "\"walkToVisitor\"" | "\"playMusic\""
	grammar := `
space ::= " "?
value ::= string | object
string ::= "\"" ([^"\\] | "\\" (["\\/bfnrt] | "u" [0-9a-fA-F]{4}))* "\"" space
object ::= "{" space pair ("," space pair)* "}" space
pair ::= string ":" space value
action-enum ::= ` + strings.Join(actionLiterals, " | ") + `
root ::= "{" space "\"action\"" space ":" space action-enum "}" space
`
	srv.Log.Debugln("Generated Action Grammar: ", grammar)
	return grammar
}

// AssembleMaterialChoiceGrammar creates a GBNF grammar for selecting a material.
func (srv *PromptService) AssembleMaterialChoiceGrammar(msg entities.WebSocketMessage, inst entities.Instructions) string {
	material, err := srv.Storage.ReadMaterials(inst.Material, msg.AssistantContext, msg.PlayerContext, context.Background())
	if err != nil || len(material) == 0 {
		srv.Log.Errorln("Error reading materials for grammar or no materials found:", err)
		// Fallback grammar if no materials, LLM might output a freeform string if this is used.
		// A better fallback might be to allow any string for "result".
		return `root ::= "{\"result\": \"unknown\"}"`
	}

	var materialLiterals []string
	for _, mat := range material {
		// Ensure material names are properly escaped for GBNF string literals if they contain special characters.
		// For simplicity, assuming names are simple strings here.
		materialLiterals = append(materialLiterals, `"\"`+mat.Name+`\""`)
	}

	// GBNF for {"result": "material_name"}
	grammar := `
space ::= " "?
value ::= string | object
string ::= "\"" ([^"\\] | "\\" (["\\/bfnrt] | "u" [0-9a-fA-F]{4}))* "\"" space
object ::= "{" space pair ("," space pair)* "}" space
pair ::= string ":" space value
material-enum ::= ` + strings.Join(materialLiterals, " | ") + `
root ::= "{" space "\"result\"" space ":" space material-enum "}" space
`
	srv.Log.Debugln("Generated Material Choice Grammar: ", grammar)
	return grammar
}

// AssembleGrammarString creates a GBNF grammar for extracting a generic string result.
func (srv *PromptService) AssembleGrammarString() string {
	// GBNF for {"result": "any string value"}
	grammar := `
space ::= " "?
value ::= string
string ::= "\"" ([^"\\] | "\\" (["\\/bfnrt] | "u" [0-9a-fA-F]{4}))* "\"" space
root ::= "{" space "\"result\"" space ":" space string "}" space
`
	srv.Log.Debugln("Generated String Result Grammar: ", grammar)
	return grammar
}

func (srv *PromptService) AssembleEmotionalGrammar() string {
	schema := `

`
	srv.Log.Debugln("Generated Emotional Grammar: ", schema)
	return schema
}

func (srv *PromptService) AssembleEmpty() string {
	return ""
}
