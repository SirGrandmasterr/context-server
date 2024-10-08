package assistant

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Service struct {
	Log                   *zap.SugaredLogger
	Val                   *validator.Validate
	ClientResponseChannel chan *entities.WebSocketAnswer
	Storage               *storage.StorageReader
	StorageWriter         *storage.StorageWriter
}

func NewAssistantService(log *zap.SugaredLogger, val *validator.Validate, serChan chan *entities.WebSocketAnswer, storage *storage.StorageReader, storagewriter *storage.StorageWriter) *Service {
	return &Service{
		Log:                   log,
		Val:                   val,
		ClientResponseChannel: serChan,
		Storage:               storage,
		StorageWriter:         storagewriter,
	}

}

var url_stream = "http://host.docker.internal:8081/completion"
var url = "http://host.docker.internal:8080/completion"
var method = "POST"

func (srv *Service) DetectAction(ctx context.Context, msg entities.WebSocketMessage, serviceChannel chan *entities.WebSocketAnswer) entities.LlmActionResponse {
	prompt, err := srv.assemblePrompt(msg)
	if err != nil {
		srv.Log.Panicln(err, "PromptAssembly failed")
	}

	var payload_struct = srv.AssemblePayload(200, false, 1.2, prompt, srv.assembleActionGrammarEnum(msg))
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		srv.Log.Errorln(err)
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}
	res, err := client.Do(req)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}
	var serverResponse entities.AssistantResponse
	var detectedAction entities.LlmActionResponse
	err = json.Unmarshal(body, &serverResponse)
	if err != nil {
		srv.Log.Errorln(err)
	}
	err = json.Unmarshal([]byte(serverResponse.Content), &detectedAction)
	if err != nil {
		srv.Log.Errorln(err)
	}

	return detectedAction

}

func (srv *Service) DecideReaction(ctx context.Context, msg entities.WebSocketMessage, serviceChannel chan *entities.WebSocketAnswer) entities.LlmActionResponse {
	prompt, err := srv.assembleEnvEventPrompt(msg)
	if err != nil {
		srv.Log.Panicln(err, "PromptAssembly failed")
	}
	var payload_struct = srv.AssemblePayload(200, false, 1.2, prompt, srv.assembleActionGrammarEnum(msg))
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		srv.Log.Errorln(err)
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}
	res, err := client.Do(req)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}
	var serverResponse entities.AssistantResponse
	var detectedAction entities.LlmActionResponse
	err = json.Unmarshal(body, &serverResponse)
	if err != nil {
		srv.Log.Errorln(err)
	}
	err = json.Unmarshal([]byte(serverResponse.Content), &detectedAction)
	if err != nil {
		srv.Log.Errorln(err)
	}

	return detectedAction
}

func (srv *Service) AskAssistant(ctx context.Context, msg entities.WebSocketMessage) (entities.AssistantAction, error) {
	method := "POST"

	var payload_struct = entities.LlmRequest{
		Stream:      false,
		NPredict:    400,
		Temperature: 1.2,
		Stop: []string{"</s>",
			"Attendant:",
			"Overlord:"},
		RepeatLastN:      256,
		RepeatPenalty:    1.18,
		TopK:             40,
		TopP:             0.95,
		MinP:             0.05,
		TfsZ:             1,
		TypicalP:         1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Mirostat:         0,
		MirostatTau:      5,
		MirostatEta:      0.1,
		Grammar: `action-kv ::= "\"action\"" space ":" space integer
		boolean ::= ("true" | "false") space
		integer ::= ("-"? ([0-9] | [1-9] [0-9]*)) space
		root ::= "{" space action-kv "," space speech-kv "}" space
		space ::= " "?
		speech-kv ::= "\"speech\"" space ":" space boolean
		`,
		NProbs:      0,
		MinKeep:     0,
		ImageData:   []interface{}{},
		CachePrompt: true,
		APIKey:      "",
		Prompt:      "",
	}
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		fmt.Println(err)
		return entities.AssistantAction{}, nil
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))

	if err != nil {
		fmt.Println(err)
		return entities.AssistantAction{}, nil
	}
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return entities.AssistantAction{}, nil
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return entities.AssistantAction{}, nil
	}
	fmt.Println(string(body))
	return entities.AssistantAction{}, nil
}

func (srv *Service) StreamAssistant(msg entities.WebSocketMessage, inst entities.Instructions) {

	method := "POST"
	prompt, err := srv.assembleInstructionsPrompt(msg, inst, "museumAssistant")
	if err != nil {
		srv.Log.Errorln(err)
	}

	payloadObject := entities.LlmRequest{
		Stream:      true,
		NPredict:    500,
		Temperature: 0.8,
		Stop: []string{
			"</s>",
			"<|end|>",
			"<|eot_id|>",
			"<|end_of_text|>",
			"<|im_end|>",
			"<|EOT|>",
			"<|END_OF_TURN_TOKEN|>",
			"<|end_of_turn|>",
			"<|endoftext|>",
			"ASSISTANT",
			"NARRATOR",
			"VISITOR"},
		RepeatLastN:      0,
		RepeatPenalty:    1,
		TopK:             0,
		TopP:             1,
		MinP:             0.05,
		TfsZ:             1,
		TypicalP:         1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Mirostat:         0,
		MirostatTau:      5,
		MirostatEta:      0.1,
		Grammar:          "",
		NProbs:           0,
		MinKeep:          0,
		ImageData:        []interface{}{},
		CachePrompt:      false,
		APIKey:           "",
		Prompt:           prompt,
	}
	client := &http.Client{}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(payloadObject)
	if err != nil {
		srv.Log.Infoln("Error marshaling Payload")
	}

	req, err := http.NewRequest(method, url, &buf)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	srv.Log.Infoln("Doing Request")
	res, err := client.Do(req)
	if err != nil {
		srv.Log.Errorln("Error during request, ", err)
	}
	reader := bufio.NewReader(res.Body)
	str := ""
	for {
		line, _ := reader.ReadBytes('\n')
		linestr := string(line)
		linelen := len(linestr)
		if linelen == 1 {
			srv.Log.Debugln("continue")
			continue
		} else if linelen == 0 {
			break
		}

		linestr = linestr[6 : linelen-1]
		var rsp entities.StreamResponse
		err = json.Unmarshal(
			[]byte(linestr),
			&rsp,
		)
		if err != nil {
			srv.Log.Infoln("MarshalError", err)
		}
		str = str + rsp.Content
		if err != nil {
			break
		}
		switch {
		case strings.Contains(str, "\n"):
			srv.Log.Infoln("string contains newline, str is:", str)
			if strings.TrimSpace(str) != "\n" {
				srv.Log.Infoln("str with newline sent.")
				srv.ClientResponseChannel <- &entities.WebSocketAnswer{
					Type: "speech",
					Text: str,
				}
				str = ""
				break
			} else {
				srv.Log.Infoln("string contains newline, str is:", str)
				str = ""
				srv.Log.Infoln("Resetted string to:", str)
				break
			}
		case strings.Contains(str, ",") || strings.Contains(str, ".") || strings.Contains(str, "!") || strings.Contains(str, "?"):
			srv.Log.Infoln("string contains , . ! ? => str is:", str)
			srv.ClientResponseChannel <- &entities.WebSocketAnswer{
				Type:       "speech",
				Text:       str,
				ActionName: "speak",
			}
			str = ""
			srv.Log.Infoln("Resetted string to:", str)
		default:
			srv.Log.Infoln("Reached Default", str)
		}

	}

	srv.ClientResponseChannel <- &entities.WebSocketAnswer{
		Type:       "speech",
		Text:       str,
		ActionName: "stopSpeak",
	}

}

func (srv *Service) PlayerSpeechAnalysis(msg entities.WebSocketMessage, inst entities.Instructions, actionName string) (entities.WebSocketAnswer, error) {
	prompt, err := srv.assembleInstructionsPrompt(msg, inst, "analysisMachine")
	if err != nil {
		srv.Log.Errorln(err)
	}
	grammar := srv.assembleGrammarString()
	payloadobj := srv.AssemblePayload(250, false, 0.5, prompt, grammar)
	payload, err := json.Marshal(payloadobj)
	if err != nil {
		fmt.Println(err)
		return entities.WebSocketAnswer{}, nil
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, nil
	}
	srv.Log.Infoln("Payload for PlayerSpeechAnalysis: ", payloadobj)
	res, err := client.Do(req)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, nil
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, nil
	}
	var serverResponse entities.AssistantResponse
	var result entities.LlmAnalysisResult
	err = json.Unmarshal(body, &serverResponse)
	if err != nil {
		srv.Log.Errorln(err)
	}
	srv.Log.Infoln("Serverresponse for PlayerSpeechAnalysis: ", serverResponse)
	err = json.Unmarshal([]byte(serverResponse.Content), &result)
	if err != nil {
		srv.Log.Errorln(err)
	}
	srv.Log.Infoln("Result: ", result.Result)
	return entities.WebSocketAnswer{
		Type:       inst.Type,
		Text:       result.Result,
		ActionName: actionName,
	}, nil
}

func (srv *Service) ActionQuery(msg entities.WebSocketMessage, inst entities.Instructions, actionName string) (entities.WebSocketAnswer, error) {
	prompt, err := srv.assembleInstructionsPrompt(msg, inst, "analysisMachine")
	if err != nil {
		srv.Log.Errorln(err)
	}
	var payload_struct = srv.AssemblePayload(200, false, 1.2, prompt, srv.assembleMaterialChoiceGrammar(msg, inst))
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		srv.Log.Errorln(err)
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Infoln("Payload for ActionQuery: ", payload_struct)
	res, err := client.Do(req)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.WebSocketAnswer{}, err
	}
	var serverResponse entities.AssistantResponse
	var result entities.LlmAnalysisResult
	err = json.Unmarshal(body, &serverResponse)
	if err != nil {
		srv.Log.Errorln(err)
	}
	srv.Log.Infoln("Serverresponse for PlayerSpeechAnalysis: ", serverResponse)
	err = json.Unmarshal([]byte(serverResponse.Content), &result)
	if err != nil {
		srv.Log.Errorln(err)
	}
	srv.Log.Infoln("Result: ", result.Result)
	return entities.WebSocketAnswer{
		Type:       inst.Type,
		Text:       result.Result,
		ActionName: actionName,
	}, nil
}

func (srv *Service) AssemblePayload(npredict int, stream bool, temperature float32, prompt string, grammar string) entities.LlmRequest {
	return entities.LlmRequest{
		Stream:      stream,
		NPredict:    npredict,
		Temperature: float64(temperature),
		Stop: []string{
			"</s>",
			"<|end|>",
			"<|eot_id|>",
			"<|end_of_text|>",
			"<|im_end|>",
			"<|EOT|>",
			"<|END_OF_TURN_TOKEN|>",
			"<|end_of_turn|>",
			"<|endoftext|>",
			"ASSISTANT",
			"NARRATOR",
			"VISITOR"},
		RepeatLastN:      0,
		RepeatPenalty:    1,
		TopK:             0,
		TopP:             1,
		MinP:             0.05,
		TfsZ:             1,
		TypicalP:         1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Mirostat:         0,
		MirostatTau:      5,
		MirostatEta:      0.1,
		Grammar:          grammar,
		NProbs:           0,
		MinKeep:          0,
		ImageData:        []interface{}{},
		CachePrompt:      false,
		APIKey:           "",
		Prompt:           prompt,
	}
}

func (srv *Service) assemblePrompt(msg entities.WebSocketMessage) (string, error) {
	matArray := []string{"options"}
	mats, err := srv.Storage.ReadMaterials(matArray, msg.AssistantContext, context.Background())
	if err != nil {
		srv.Log.Errorln()
	}
	prompt := ""
	baseprompt, err := srv.Storage.ReadBasePrompt("museumAssistant", context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Baseprompt from DB")
	} //Find Setting
	prompt += baseprompt.Prompt
	prompt += "You are positioned on the lower level of the museum where all sculptures are located." //Location
	prompt += "Currently, you stand idle as a visitor speaks to you."                                 //Get state from msg
	prompt += "Detect whether or not the visitor wants you to take one of the following actions: \n"
	for _, avac := range mats {
		prompt += `{"` + avac.Name + `": "` + avac.Description + `}` + "\n"
	}
	prompt += "\n\n\n\nVISITOR: " + msg.Speech
	prompt += "\nASSISTANT:"

	//srv.Storage.ReadActionOptionEntity()
	//GetBasePrompt
	//GetLocation
	//Add UserContext
	//SearchAvailableActions
	//Combine
	srv.Log.Infoln("Generated Prompt: ", prompt)
	return prompt, nil
}

func (srv *Service) assembleEnvEventPrompt(msg entities.WebSocketMessage) (string, error) {
	prompt := ""
	baseprompt, err := srv.Storage.ReadBasePrompt(msg.AssistantContext.SelectedBasePrompt, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Baseprompt from DB")
	}

	assistantLocation, err := srv.Storage.ReadLocation(msg.AssistantContext.Location, context.Background())
	prompt += baseprompt.Prompt + "\n"
	prompt += "You are currently located at the " + assistantLocation.LocationName + ". \n"
	prompt += assistantLocation.Description + "\n"
	prompt += srv.getActivityState(msg) + "\n"
	prompt += "Something happened: \n"
	prompt += msg.Speech + "\n"
	prompt += "You have " + string(len(msg.AssistantContext.AvailableActions)) + " options, what do you want to do?" + "\n"
	for _, opt := range msg.AssistantContext.AvailableActions {
		action, err := srv.Storage.ReadActionOptionEntity(opt, context.Background())
		if err != nil {
			return prompt, err
		}
		prompt += `{"` + action.ActionName + `": "` + action.Description + `}` + "\n"
	}
	return prompt, nil
}

func (srv *Service) assembleInstructionsPrompt(msg entities.WebSocketMessage, inst entities.Instructions, basepromptstr string) (string, error) {
	prompt := ""
	baseprompt, err := srv.Storage.ReadBasePrompt(basepromptstr, context.Background())
	if err != nil {
		srv.Log.Errorln("Error reading Baseprompt from DB")
	}
	player, err := srv.Storage.ReadPlayer(msg.PlayerContext.PlayerUsername, context.Background())
	if err != nil {
		srv.Log.Errorln(err)
	}
	prompt += baseprompt.Prompt + "\n"
	material, err := srv.Storage.ReadMaterials(inst.Material, msg.AssistantContext, context.Background())
	if err != nil {
		srv.Log.Errorln(err)
	}

	switch inst.Type {
	case "playerSpeechAnalysis": // Will be sent to small LLM
		prompt += inst.StageInstructions + "\n"
		prompt += "INPUT: " + msg.Speech + "\n"
		break
	case "speech": // Will be sent to big LLM
		prompt += "Here is what happened so far:"
		prompt += player.History + "\n"
		prompt += "Currently" + " "
		prompt += inst.StageInstructions + "\n"
		for _, avac := range material {
			prompt += `{"name": "` + avac.Name + `",` + `"description":"` + avac.Description + `"}` + "\n"
		}
		prompt += "ASSISTANT: "
		break
	case "actionquery":
		prompt += inst.StageInstructions + "\n"
		prompt += "INPUT: " + msg.Speech + "\n"
		prompt += "MATERIAL: \n"
		var focus entities.Material
		hasFocus := false
		counter := 1
		for _, mat := range material {
			if mat.Type != "focus" {
				prompt += `{"name": "` + mat.Name + `",` + `"description":"` + mat.Description + `"}` + "\n"
				counter++
			} else {
				focus = mat
				hasFocus = true
			}
		}
		if hasFocus { //Save the focus for last, for relevancy
			prompt += "Both you and the Visitor are intently staring at: \n"
			prompt += focus.Name + ", " + focus.Description
		}
	}
	// prompt += location? actionselection, speech, actionquery, speechAnalysis
	// prompt += playerState, etc.?

	prompt += "ASSISTANT: "
	return prompt, nil
}

// This function transfers the available actions into an Enum, to make sure the lil' stupid Llama makes no spelling mistakes. :)
func (srv *Service) assembleActionGrammarEnum(msg entities.WebSocketMessage) string {
	len := len(msg.AssistantContext.AvailableActions)
	result := "("
	for i, st := range msg.AssistantContext.AvailableActions {
		result += `"\"`
		result += st
		result += `\""`
		if i != len-1 {
			result += ` | `
		}
	}
	result += ")"

	grammar := `action ::= ` + result + ` space
action-kv ::= "\"action\"" space ":" space action
root ::= "{" space action-kv "}" space
space ::= | " " | "\n" [ \t]{0,20}`
	print("grammar: ", grammar)
	return grammar
}

func (srv *Service) assembleMaterialChoiceGrammar(msg entities.WebSocketMessage, inst entities.Instructions) string {
	material, err := srv.Storage.ReadMaterials(inst.Material, msg.AssistantContext, context.Background())
	if err != nil {
		srv.Log.Errorln(err)
	}
	len := len(material)
	result := "("
	for i, st := range material {
		result += `"\"`
		result += st.Name
		result += `\""`
		if i != len-1 {
			result += ` | `
		}
	}
	result += ")"
	grammar := `result ::= ` + result + ` space
result-kv ::= "\"result\"" space ":" space result
root ::= "{" space result-kv "}" space
space ::= | " " | "\n" [ \t]{0,20}`
	srv.Log.Infoln("grammar: ", grammar)
	return grammar
}

func (srv *Service) assembleGrammarString() string {
	//makes the llm return something like "{"query": "somestring"}"
	grammar := `char ::= [^"\\\x7F\x00-\x1F] | [\\] (["\\bfnrt] | "u" [0-9a-fA-F]{4})
result-kv ::= "\"result\"" space ":" space string
root ::= "{" space result-kv "}" space
space ::= | " " | "\n" [ \t]{0,20}
string ::= "\"" char* "\"" space`
	return grammar
}

func (srv *Service) getActivityState(msg entities.WebSocketMessage) string {
	text := ""
	switch msg.AssistantContext.WalkingState {
	case "idle":
		text += "You currently stand idle. "
	case "patrolling":
		text += "You are currently patrolling the area, searching for things out of place. "
	case "followPlayer":
		text += "You are currently following a visitor around."
	case "moving":
		text += "You are currently moving towards your destination."
	}
	if msg.AssistantContext.PlayerVisible {
		text += "A Visitor is in your field of vision. "
	}
	if msg.PlayerContext.InConversation {
		text += "You are currently in conversation with a visitor."
	}
	return text

}
