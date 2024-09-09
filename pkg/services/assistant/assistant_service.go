package assistant

import (
	"Llamacommunicator/pkg/entities"
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
	Log *zap.SugaredLogger
	Val *validator.Validate
}

func NewAssistantService(log *zap.SugaredLogger, val *validator.Validate) *Service {
	return &Service{
		Log: log,
		Val: val,
	}

}

func (srv *Service) AskAssistant(ctx context.Context, r *entities.RequestAssistantReaction) (entities.AssistantAction, error) {
	url := "http://localhost:8080/completion"
	method := "POST"

	prompt, err := srv.assemblePrompt(ctx, r)
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

func (srv *Service) StreamAssistant(txt string) {

	url := "http://localhost:8080/completion"
	method := "POST"

	payload := strings.NewReader("{" +
		"" + `"stream": true,` + "" + `"n_predict": 358,` + "" + `"temperature": 0.8,` + "" + `"stop": [` + "" + `"</s>",` + "" + `"<|end|>",` + "" + `"<|eot_id|>",` + "" + `"<|end_of_text|>",` + "" + `"<|im_end|>",` + "" + `"<|EOT|>",` + "" + `"<|END_OF_TURN_TOKEN|>",` + "" + `"<|end_of_turn|>",` + "" + `"<|endoftext|>",` + "" + `"ASSISTANT",` + "" + `"USER"` + "" + `],` + "" + `"repeat_last_n": 0,` + "" + `"repeat_penalty": 1,` + "" + `"penalize_nl": false,` + "" + `"top_k": 0,` + "" + `"top_p": 1,` + "" + `"min_p": 0.05,` + "" + `"tfs_z": 1,` + "" + `"typical_p": 1,` + "" + `"presence_penalty": 0,` + "" + `"frequency_penalty": 0,` + "" + `"mirostat": 0,` + "" + `"mirostat_tau": 5,` + "" + `"mirostat_eta": 0.1,` + "" + `"grammar": "",` + "" + `"n_probs": 0,` + "" + `"min_keep": 0,` + "" + `"image_data": [],` + "" + `"cache_prompt": true,` + "" + `"api_key": "",` + "" + `"prompt": "You work in a museum and it is your job to give lengthy answers to visitors who ask you questions. Currently, you stand idle as a visitor speaks to you:\n\n\n\nUSER: What do you think makes a woman most beautiful? \nASSISTANT"` + "" + `}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, _ := client.Do(req)
	reader := bufio.NewReader(res.Body)
	str := ""
	for {
		line, _ := reader.ReadBytes('\n')
		print("read")
		linestr := string(line)
		print("string")
		linelen := len(linestr)
		print(linelen)
		if linelen == 1 {
			print("continue")
			continue
		} else if linelen == 0 {
			break
		}

		print("slice")
		linestr = linestr[6 : linelen-1]
		print("Response")
		var rsp entities.StreamResponse
		err = json.Unmarshal(
			[]byte(linestr),
			&rsp,
		)
		if err != nil {
			print("MarshalError", err)
		}
		str = str + rsp.Content
		if err != nil {
			break
		}
		print(string(line))
	}
	print(str)

	//TODO:
	// => Connect to AssistantProcess => Offer a channel => search for ., ;, !, ? and then send text.

}

func (srv *Service) assemblePrompt(ctx context.Context, r *entities.RequestAssistantReaction) (string, error) {
	return "", nil
}
