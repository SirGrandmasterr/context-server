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
	Log            *zap.SugaredLogger
	Val            *validator.Validate
	ServiceChannel chan *entities.WebSocketAnswer
	Storage        *storage.StorageReader
}

func NewAssistantService(log *zap.SugaredLogger, val *validator.Validate, serChan chan *entities.WebSocketAnswer, storage *storage.StorageReader) *Service {
	return &Service{
		Log:            log,
		Val:            val,
		ServiceChannel: serChan,
		Storage:        storage,
	}

}

var url_docker = "http://host.docker.internal:8080/completion"
var url = "http://host.docker.internal:8080/completion"
var method = "POST"

func (srv *Service) DetectAction(ctx context.Context, msg entities.WebSocketMessage, serviceChannel chan *entities.WebSocketAnswer) (entities.LlmActionResponse, string) {
	prompt, err := srv.assemblePrompt(ctx, msg, false, 0, "")
	srv.Log.Infoln("Beginning to detect Action from user Speech and Context")
	if err != nil {
		srv.Log.Panicln(err, "PromptAssembly failed")
	}
	srv.Log.Infoln(prompt)

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
		Grammar: ` action ::= ("\"standIdle\"" | "\"patrol\"" | "\"followPlayer\"" | "\"playMusic\"" | "\"stopMusic\"") space
action-kv ::= "\"action\"" space ":" space action
root ::= "{" space action-kv "}" space
space ::= | " " | "\n" [ \t]{0,20}`,
		NProbs:      0,
		MinKeep:     0,
		ImageData:   []interface{}{},
		CachePrompt: false,
		APIKey:      "",
		Prompt:      prompt,
	}
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		srv.Log.Errorln(err)
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}, ""
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referer", "http://localhost:8080/")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "http://localhost:8080")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")

	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}, ""
	}
	res, err := client.Do(req)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}, ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}, ""
	}
	var serverResponse entities.AssistantResponse
	var detectedAction entities.LlmActionResponse
	err = json.Unmarshal(body, &serverResponse)
	srv.Log.Infoln(serverResponse)
	if err != nil {
		srv.Log.Errorln(err)
	}
	err = json.Unmarshal([]byte(serverResponse.Content), &detectedAction)
	srv.Log.Infoln(detectedAction)
	if detectedAction.ActionName == "followPlayer" {
		srv.Log.Infoln("FollowPlayer detected.")
		ac, _ := srv.Storage.ReadActionOptionEntity(detectedAction.ActionName, ctx)
		srv.Log.Infoln("Action name follow player", ac.ActionName, ac.Description)
	}
	/*partial := entities.WebSocketAnswer{
		Type:       "partialAction",
		Text:       "",
		ActionName: detectedAction.ActionName,
	}
	serviceChannel <- &partial*/
	return detectedAction, ""

}

func (srv *Service) AskAssistant(ctx context.Context, msg entities.WebSocketMessage) (entities.AssistantAction, error) {
	method := "POST"

	prompt, err := srv.assemblePrompt(ctx, msg, false, 0, "")
	srv.Log.Infoln(prompt)
	if err != nil {
		srv.Log.Panicln(err, "PromptAssembly failed")
	}

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
		Prompt: `Imagine a museum. You work there as a guard. You are currently in a hall with six exhibits, three by the western and three by the eastern wall. You currently sit in a chair by the northern wall. You only talk to visitors when they ask you a question or when they violate the rules. There are two rules: No photographing with flash and no touching the exhibits. When the narrator informs you about the action of a visitor, you choose  one out of several actions that you will take, and if you need to talk to the visitor. 

		Narrator: A visitor looks at the statue of Julius Caesar, the second exhibit by the west wall. You have the following options: 
		0. Do nothing.
		1. Get up from the chair
		2. Walk aimlessly through the hall
		3. Walk towards the visitor
		Choose your action and elaborate why you have taken it.`,
		//Choose your action and if talking is necessary like this: {"action": 0, "speech": false}`,
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
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0")
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referer", "http://localhost:8080/")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "http://localhost:8080")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")

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

func (srv *Service) StreamAssistant(msg entities.WebSocketMessage) {

	method := "POST"

	payloadObject := entities.LlmRequest{
		Stream:      true,
		NPredict:    358,
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
			"USER"},
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
		Prompt:           "You work in a museum and it is your job to give lengthy answers to visitors who ask you questions. Currently, you stand idle as a visitor speaks to you:\n\n\n\nUSER: " + msg.Speech + " \nASSISTANT",
	}
	//payload := strings.NewReader("{" +
	//	"" + `"stream": true,` + "" + `"n_predict": 358,` + "" + `"temperature": 0.8,` + "" + `"stop": [` + "" + `"</s>",` + "" + `"<|end|>",` + "" + `"<|eot_id|>",` + "" + `"<|end_of_text|>",` + "" + `"<|im_end|>",` + "" + `"<|EOT|>",` + "" + `"<|END_OF_TURN_TOKEN|>",` + "" + `"<|end_of_turn|>",` + "" + `"<|endoftext|>",` + "" + `"ASSISTANT",` + "" + `"USER"` + "" + `],` + "" + `"repeat_last_n": 0,` + "" + `"repeat_penalty": 1,` + "" + `"penalize_nl": false,` + "" + `"top_k": 0,` + "" + `"top_p": 1,` + "" + `"min_p": 0.05,` + "" + `"tfs_z": 1,` + "" + `"typical_p": 1,` + "" + `"presence_penalty": 0,` + "" + `"frequency_penalty": 0,` + "" + `"mirostat": 0,` + "" + `"mirostat_tau": 5,` + "" + `"mirostat_eta": 0.1,` + "" + `"grammar": "",` + "" + `"n_probs": 0,` + "" + `"min_keep": 0,` + "" + `"image_data": [],` + "" + `"cache_prompt": true,` + "" + `"api_key": "",` + "" + `"prompt": "You work in a museum and it is your job to give lengthy answers to visitors who ask you questions. Currently, you stand idle as a visitor speaks to you:\n\n\n\nUSER: ` + txt + ` \nASSISTANT"` + "" + `}`)
	srv.Log.Infoln("Creating Client")
	client := &http.Client{}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payloadObject)
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
				srv.ServiceChannel <- &entities.WebSocketAnswer{
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
			srv.ServiceChannel <- &entities.WebSocketAnswer{
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

	srv.ServiceChannel <- &entities.WebSocketAnswer{
		Type:       "speech",
		Text:       str,
		ActionName: "stopSpeak",
	}

}

func (srv *Service) assemblePrompt(ctx context.Context, msg entities.WebSocketMessage, partialAction bool, partialIndex int, actionName string) (string, error) {
	var avActions string = strings.Join(msg.AssistantContext.AvailableActions, ", ")
	srv.Log.Infoln("Here are the available actions: ", avActions)
	srv.Log.Infoln(msg.AssistantContext)
	prompt := ""
	prompt += "You work in a museum as an assistant."                                                                                       //Find Setting
	prompt += "You converse with visitors and elaborate about the exhibition pieces, and visitors can request you to take certain actions." //Job Description
	prompt += "You are positioned on the lower level of the museum where all sculptures are located."                                       //Location
	prompt += "Currently, you stand idle as a visitor speaks to you."                                                                       //Get state from msg
	prompt += "Detect whether or not the visitor wants you to take one of the following actions: "
	prompt += avActions
	prompt += "\n\n\n\nUSER: " + msg.Speech
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

// This function transfers the available actions into an Enum, to make sure the lil' stupid Llama makes no spelling mistakes. :)
func (srv *Service) assembleActionGrammarEnum(ctx context.Context, msg entities.WebSocketMessage) string {
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
