package assistant

import (
	"Llamacommunicator/pkg/config"
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/prompting"
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
	Conf                  *config.Specification
	Pr                    *prompting.PromptService
}

func NewAssistantService(log *zap.SugaredLogger, val *validator.Validate, serChan chan *entities.WebSocketAnswer, storage *storage.StorageReader, storagewriter *storage.StorageWriter, conf *config.Specification, pr *prompting.PromptService) *Service {
	return &Service{
		Log:                   log,
		Val:                   val,
		ClientResponseChannel: serChan,
		Storage:               storage,
		StorageWriter:         storagewriter,
		Conf:                  conf,
		Pr:                    pr,
	}

}

var method = "POST"

func (srv *Service) DetectAction(ctx context.Context, msg entities.WebSocketMessage, serviceChannel chan *entities.WebSocketAnswer, temp float32) entities.LlmActionResponse {
	prompt, err := srv.Pr.AssemblePrompt(msg)
	print("srv.Conf.LlmSmall ", srv.Conf.LlmSmall)
	if err != nil {
		srv.Log.Panicln(err, "PromptAssembly failed")
	}
	var payload_struct entities.LlmRequest
	if msg.MessageType == "innerThoughtEvent" {
		payload_struct = srv.AssemblePayload(200, false, 0.8, prompt, srv.assembleActionGrammarEnum(msg), 10)
	} else {
		payload_struct = srv.AssemblePayload(200, false, 0.8, prompt, srv.assembleActionGrammarEnum(msg), 5)
	}
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		srv.Log.Errorln(err)
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, srv.Conf.LlmSmall, bytes.NewBuffer(payload))
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
	prompt, err := srv.Pr.AssembleEnvEventPrompt(msg)
	if err != nil {
		srv.Log.Panicln(err, "PromptAssembly failed")
	}
	var payload_struct = srv.AssemblePayload(200, false, 1.2, prompt, srv.assembleActionGrammarEnum(msg), 5)
	payload, err := json.Marshal(payload_struct)
	if err != nil {
		srv.Log.Errorln(err)
		srv.Log.Infoln(err)
		return entities.LlmActionResponse{}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, srv.Conf.LlmSmall, bytes.NewBuffer(payload))
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

func (srv *Service) StreamAssistant(msg entities.WebSocketMessage, inst entities.Instructions) {
	url := ""
	if inst.LlmSize == "small" {
		url = srv.Conf.LlmSmall
	} else {
		url = srv.Conf.LlmBig
	}
	method := "POST"
	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, "museumAssistant")
	if err != nil {
		srv.Log.Errorln(err)
	}

	payloadObject := entities.LlmRequest{
		Stream:      true,
		NPredict:    500,
		Temperature: 0.7,
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
		},
		RepeatLastN:      0,
		RepeatPenalty:    1.18,
		TopK:             40,
		TopP:             0.95,
		MinP:             0.05,
		TfsZ:             1,
		TypicalP:         1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Mirostat:         2,
		MirostatTau:      5,
		MirostatEta:      0.1,
		Grammar:          "",
		NProbs:           0,
		MinKeep:          0,
		ImageData:        []interface{}{},
		CachePrompt:      true,
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
	url := ""
	if inst.LlmSize == "small" {
		url = srv.Conf.LlmSmall
	} else {
		url = srv.Conf.LlmBig
	}
	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, "analysisMachine")
	if err != nil {
		srv.Log.Errorln(err)
	}
	grammar := srv.assembleGrammarString()
	payloadobj := srv.AssemblePayload(250, false, 0.5, prompt, grammar, 5)
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
	url := ""
	if inst.LlmSize == "big" {
		url = srv.Conf.LlmBig
	} else {
		url = srv.Conf.LlmSmall
	}
	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, "analysisMachine")
	if err != nil {
		srv.Log.Errorln(err)
	}
	var payload_struct = srv.AssemblePayload(200, false, 1.2, prompt, srv.assembleMaterialChoiceGrammar(msg, inst), 5)
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

func (srv *Service) AssemblePayload(npredict int, stream bool, temperature float32, prompt string, grammar string, mirostat int) entities.LlmRequest {
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
		},
		RepeatLastN:      0,
		RepeatPenalty:    1.18,
		TopK:             40,
		TopP:             1,
		MinP:             0.05,
		TfsZ:             1,
		TypicalP:         1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Mirostat:         2,
		MirostatTau:      mirostat,
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
