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
	"time"

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

var method = "POST" // HTTP method for LLM requests

// callLLM is a helper function to make HTTP requests to the LLM.
// It handles request creation, sending, and basic error handling.
func (srv *Service) callLLM(url string, payload []byte, stream bool) (*http.Response, error) {
	client := &http.Client{
		Timeout: 60 * time.Second, // Added timeout for LLM requests
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		srv.Log.Errorw("Failed to create LLM request", "url", url, "error", err)
		return nil, err
	}
	if stream {
		req.Header.Add("Accept", "text/event-stream")
	} else {
		req.Header.Add("Accept", "application/json")
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		srv.Log.Errorw("LLM request failed", "url", url, "error", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		res.Body.Close() // Close body immediately after reading
		srv.Log.Errorw("LLM returned non-OK status", "url", url, "status_code", res.StatusCode, "response_body", string(bodyBytes))
		return nil, fmt.Errorf("LLM request failed with status %d", res.StatusCode)
	}
	return res, nil
}

// DetectAction determines the action an assistant should take based on the message context.
func (srv *Service) DetectAction(ctx context.Context, msg entities.WebSocketMessage, temp float32) entities.LlmActionResponse {

	// Bypass asking small LLM for action selection if only a single action is available anyways. Saves a lot of time.
	if len(msg.AssistantContext.AvailableActions) == 1 {
		return entities.LlmActionResponse{ActionName: msg.AssistantContext.AvailableActions[0]}
	}

	prompt, err := srv.Pr.AssemblePrompt(msg)
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for DetectAction", "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"} // Fallback action
	}

	llmURL := srv.Conf.LlmSmall
	if msg.MessageType == "innerThoughtEvent" { // Use bigger LLM for more nuanced internal thought processing
		llmURL = srv.Conf.LlmBig
	}

	// Temperature for action detection should be lower for more deterministic results.
	// Let's use a slightly lower temp than speech generation, e.g. 0.5, unless a higher one is passed.
	effectiveTemp := temp
	if temp > 0.7 { // Cap temp if it's for action detection unless explicitly higher
		effectiveTemp = 0.5
	}

	payloadStruct := srv.AssemblePayload(50, false, effectiveTemp, prompt, srv.Pr.AssembleActionGrammarEnum(msg), srv.Pr.AssembleEmpty(), 5) // Reduced nPredict for action selection
	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for DetectAction", "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}

	srv.Log.Debugw("DetectAction LLM Request", "url", llmURL, "payload", string(payload))

	res, err := srv.callLLM(llmURL, payload, false)
	if err != nil {
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Errorw("Failed to read LLM response body for DetectAction", "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	srv.Log.Debugw("DetectAction LLM Raw Response", "body", string(body))

	var serverResponse entities.AssistantResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		srv.Log.Errorw("Failed to unmarshal LLM server response for DetectAction", "body", string(body), "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}

	var detectedAction entities.LlmActionResponse
	// The content from LLM for action selection should be a JSON string like {"action": "actionName"}
	if err := json.Unmarshal([]byte(serverResponse.Content), &detectedAction); err != nil {
		srv.Log.Errorw("Failed to unmarshal detected action from LLM content", "content", serverResponse.Content, "error", err)
		// Attempt to extract action name if JSON parsing fails but content might be just the action name
		if serverResponse.Content != "" && !strings.Contains(serverResponse.Content, "{") { // Basic check
			srv.Log.Infow("LLM content for action was not JSON, attempting direct use", "content", serverResponse.Content)

			return entities.LlmActionResponse{ActionName: "ignore"} // Stricter: if not JSON, ignore.
		}
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	if detectedAction.ActionName == "" {
		srv.Log.Warnw("Detected action name is empty", "content", serverResponse.Content)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	srv.Log.Infow("Action Detected", "action", detectedAction.ActionName)
	return detectedAction
}

// DecideReaction determines how the assistant should react to an environmental event.
func (srv *Service) DecideReaction(ctx context.Context, msg entities.WebSocketMessage) entities.LlmActionResponse {
	prompt, err := srv.Pr.AssembleEnvEventPrompt(msg)
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for DecideReaction", "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}

	// Environmental reactions might need more nuanced understanding, so LlmBig is appropriate.
	// Temperature can be moderate to allow for some flexibility but not too random.
	payloadStruct := srv.AssemblePayload(50, false, 0.6, prompt, srv.Pr.AssembleActionGrammarEnum(msg), srv.Pr.AssembleEmpty(), 5) // Reduced nPredict
	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for DecideReaction", "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	srv.Log.Debugw("DecideReaction LLM Request", "url", srv.Conf.LlmBig, "payload", string(payload))

	res, err := srv.callLLM(srv.Conf.LlmBig, payload, false)
	if err != nil {
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Errorw("Failed to read LLM response body for DecideReaction", "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	srv.Log.Debugw("DecideReaction LLM Raw Response", "body", string(body))

	var serverResponse entities.AssistantResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		srv.Log.Errorw("Failed to unmarshal LLM server response for DecideReaction", "body", string(body), "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}

	var detectedAction entities.LlmActionResponse
	if err := json.Unmarshal([]byte(serverResponse.Content), &detectedAction); err != nil {
		srv.Log.Errorw("Failed to unmarshal detected reaction from LLM content", "content", serverResponse.Content, "error", err)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	if detectedAction.ActionName == "" {
		srv.Log.Warnw("Detected reaction name is empty", "content", serverResponse.Content)
		return entities.LlmActionResponse{ActionName: "ignore"}
	}
	srv.Log.Infow("Reaction Decided", "action", detectedAction.ActionName)
	return detectedAction
}

// StreamAssistant streams responses from the LLM for speech.
func (srv *Service) StreamAssistant(msg entities.WebSocketMessage, inst entities.Instructions, tok entities.ActionToken) {
	url := srv.Conf.LlmSmall
	if inst.LlmSize == "big" {
		url = srv.Conf.LlmBig
	}

	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, inst.BasePrompt) // BasePrompt from instruction
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for StreamAssistant", "instruction_type", inst.Type, "error", err)
		srv.ClientResponseChannel <- &entities.WebSocketAnswer{Type: "error", Text: "Sorry, I encountered an issue.", Token: tok.ID}
		return
	}
	// Temperature for speech can be higher for more natural responses.
	payloadStruct := srv.AssemblePayload(inst.Limit, true, 0.8, prompt, srv.Pr.AssembleEmpty(), srv.Pr.AssembleEmpty(), 5) // inst.Limit for nPredict if set, else default
	if inst.Limit == 0 {                                                                                                   // Ensure nPredict is reasonable if Limit is 0
		payloadStruct.NPredict = 256 // Default for speech
	}

	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for StreamAssistant", "error", err)
		srv.ClientResponseChannel <- &entities.WebSocketAnswer{Type: "error", Text: "Sorry, I had a problem preparing my response.", Token: tok.ID}
		return
	}
	srv.Log.Debugw("StreamAssistant LLM Request", "url", url, "payload", string(payload))

	res, err := srv.callLLM(url, payload, true)
	if err != nil {
		srv.ClientResponseChannel <- &entities.WebSocketAnswer{Type: "error", Text: "I'm having trouble connecting right now.", Token: tok.ID}
		return
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)
	var accumulatedResponse strings.Builder
	var lastSentResponse string

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			srv.Log.Errorw("Error reading stream from LLM", "error", err)
			break
		}

		lineStr := string(line)
		if strings.HasPrefix(lineStr, "data: ") {
			jsonData := strings.TrimPrefix(lineStr, "data: ")
			jsonData = strings.TrimSpace(jsonData) // Trim whitespace

			if jsonData == "" { // Skip empty data lines
				continue
			}

			var streamResp entities.StreamResponse
			if err := json.Unmarshal([]byte(jsonData), &streamResp); err != nil {
				srv.Log.Warnw("Failed to unmarshal stream data chunk", "data", jsonData, "error", err)
				continue
			}

			accumulatedResponse.WriteString(streamResp.Content)

			// Send partial responses based on sentence-ending punctuation or significant length
			currentFullText := accumulatedResponse.String()
			if strings.ContainsAny(streamResp.Content, ".!?\n") || (len(currentFullText)-len(lastSentResponse) > 80) { // Heuristic for sending partials
				if strings.TrimSpace(currentFullText) != "" && currentFullText != lastSentResponse {
					srv.ClientResponseChannel <- &entities.WebSocketAnswer{
						Type:       "speech",
						Text:       strings.TrimSpace(currentFullText), // Send the accumulated part
						ActionName: inst.Type,                          // Or a specific action name like "speak"
						Token:      tok.ID,
						Stage:      inst.Stage,
					}
					lastSentResponse = currentFullText
					accumulatedResponse.Reset() //  reset accumulatedResponse here, otherwise TTS will repeat itself
				}
			}
			if streamResp.Stop {
				break
			}
		}
	}

	finalText := strings.TrimSpace(accumulatedResponse.String())
	if finalText != "" && finalText != strings.TrimSpace(lastSentResponse) { // Send any remaining part
		srv.ClientResponseChannel <- &entities.WebSocketAnswer{
			Type:       "speech",
			Text:       finalText,
			ActionName: inst.Type,
			Token:      tok.ID,
			Stage:      inst.Stage,
		}
	}
	// Signal end of speech for this stage
	srv.ClientResponseChannel <- &entities.WebSocketAnswer{
		Type:       "speech",    // Or a dedicated type like "speechEnd"
		Text:       "",          // No more text
		ActionName: "stopSpeak", // Client uses this to know the turn is over
		Token:      tok.ID,
		Stage:      inst.Stage,
	}
	srv.Log.Infow("Finished streaming assistant speech", "action_token", tok.ID, "stage", inst.Stage)
}

// PlayerSpeechAnalysis extracts information from player's speech based on instructions.
func (srv *Service) PlayerSpeechAnalysis(msg entities.WebSocketMessage, inst entities.Instructions, actionName string) (entities.WebSocketAnswer, error) {
	url := srv.Conf.LlmSmall
	if inst.LlmSize == "big" {
		url = srv.Conf.LlmBig
	}

	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, inst.BasePrompt)
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for PlayerSpeechAnalysis", "error", err)
		return entities.WebSocketAnswer{}, err
	}

	// Temperature for analysis/extraction should be low.
	payloadStruct := srv.AssemblePayload(inst.Limit, false, 0.3, prompt, srv.Pr.AssembleGrammarString(), srv.Pr.AssembleEmpty(), 5) // inst.Limit for nPredict
	if inst.Limit == 0 {
		payloadStruct.NPredict = 100 // Default for analysis
	}

	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for PlayerSpeechAnalysis", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Debugw("PlayerSpeechAnalysis LLM Request", "url", url, "payload", string(payload))

	res, err := srv.callLLM(url, payload, false)
	if err != nil {
		return entities.WebSocketAnswer{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Errorw("Failed to read LLM response body for PlayerSpeechAnalysis", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Debugw("PlayerSpeechAnalysis LLM Raw Response", "body", string(body))

	var serverResponse entities.AssistantResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		srv.Log.Errorw("Failed to unmarshal LLM server response for PlayerSpeechAnalysis", "body", string(body), "error", err)
		return entities.WebSocketAnswer{}, err
	}

	var result entities.LlmAnalysisResult
	if err := json.Unmarshal([]byte(serverResponse.Content), &result); err != nil {
		srv.Log.Errorw("Failed to unmarshal analysis result from LLM content", "content", serverResponse.Content, "error", err)
		// If LLM fails to produce valid JSON, we might return the raw content or an error
		// For now, returning an error if JSON is expected but not received.
		return entities.WebSocketAnswer{}, fmt.Errorf("LLM returned non-JSON content for analysis: %s", serverResponse.Content)
	}

	srv.Log.Infow("Player Speech Analysis Result", "result", result.Result)
	return entities.WebSocketAnswer{
		Type:       inst.Type, // e.g., "playerSpeechAnalysis"
		Text:       result.Result,
		ActionName: actionName, // The parent action
		Stage:      inst.Stage,
	}, nil
}

func (srv *Service) EmotionUpdate(msg entities.WebSocketMessage, inst entities.Instructions, actionName string) (entities.WebSocketAnswer, error) {
	url := srv.Conf.LlmSmall
	if inst.LlmSize == "big" {
		url = srv.Conf.LlmBig
	}
	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, inst.BasePrompt)
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for ActionQuery", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	// Temperature for selection from a list should be low.
	payloadStruct := srv.AssemblePayload(500, false, 0.7, prompt, srv.Pr.AssembleEmpty(), srv.Pr.AssembleEmotionalGrammar(), 5) // Reduced nPredict
	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for ActionQuery", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Debugw("EmotionUpdate LLM Request", "url", url, "payload", string(payload))

	res, err := srv.callLLM(url, payload, false)
	if err != nil {
		return entities.WebSocketAnswer{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Errorw("Failed to read LLM response body for ActionQuery", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Debugw("EmotionUpdate LLM Raw Response", "body", string(body))

	var serverResponse entities.AssistantResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		srv.Log.Errorw("Failed to unmarshal LLM server response for ActionQuery", "body", string(body), "error", err)
		return entities.WebSocketAnswer{}, err
	}

	var result entities.EmotionalState // Expecting {"result": "chosen_material_name"}
	if err := json.Unmarshal([]byte(serverResponse.Content), &result); err != nil {
		srv.Log.Errorw("Failed to unmarshal action query result from LLM content", "content", serverResponse.Content, "error", err)
		return entities.WebSocketAnswer{}, fmt.Errorf("LLM returned non-JSON content for action query: %s", serverResponse.Content)
	}

	srv.Log.Infow("EmotionUpdate Result", result)
	return entities.WebSocketAnswer{
		Type:       inst.Type,
		Text:       serverResponse.Content,
		ActionName: actionName,
		Stage:      inst.Stage,
	}, nil

}

// ActionQuery lets the LLM choose from a list of materials based on instructions and speech.
func (srv *Service) ActionQuery(msg entities.WebSocketMessage, inst entities.Instructions, actionName string) (entities.WebSocketAnswer, error) {
	url := srv.Conf.LlmSmall
	if inst.LlmSize == "big" {
		url = srv.Conf.LlmBig
	}

	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, inst.BasePrompt)
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for ActionQuery", "error", err)
		return entities.WebSocketAnswer{}, err
	}

	// Temperature for selection from a list should be low.
	payloadStruct := srv.AssemblePayload(50, false, 0.3, prompt, srv.Pr.AssembleMaterialChoiceGrammar(msg, inst), "", 5) // Reduced nPredict
	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for ActionQuery", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Debugw("ActionQuery LLM Request", "url", url, "payload", string(payload))

	res, err := srv.callLLM(url, payload, false)
	if err != nil {
		return entities.WebSocketAnswer{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Errorw("Failed to read LLM response body for ActionQuery", "error", err)
		return entities.WebSocketAnswer{}, err
	}
	srv.Log.Debugw("ActionQuery LLM Raw Response", "body", string(body))

	var serverResponse entities.AssistantResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		srv.Log.Errorw("Failed to unmarshal LLM server response for ActionQuery", "body", string(body), "error", err)
		return entities.WebSocketAnswer{}, err
	}

	var result entities.LlmAnalysisResult // Expecting {"result": "chosen_material_name"}
	if err := json.Unmarshal([]byte(serverResponse.Content), &result); err != nil {
		srv.Log.Errorw("Failed to unmarshal action query result from LLM content", "content", serverResponse.Content, "error", err)
		return entities.WebSocketAnswer{}, fmt.Errorf("LLM returned non-JSON content for action query: %s", serverResponse.Content)
	}

	srv.Log.Infow("Action Query Result", "selected_material", result.Result)
	return entities.WebSocketAnswer{
		Type:       inst.Type,     // e.g., "actionQuery"
		Text:       result.Result, // This will be the name of the chosen material
		ActionName: actionName,    // The parent action
		Stage:      inst.Stage,
	}, nil
}

// AssemblePayload creates the request payload for the LLM.
func (srv *Service) AssemblePayload(nPredict int, stream bool, temperature float32, prompt string, grammar string, schema string, mirostatTau int) entities.LlmRequest {
	if nPredict <= 0 {
		nPredict = 256 // Default prediction length
	}
	if nPredict > 2048 { // Cap max prediction length
		nPredict = 2048
	}

	return entities.LlmRequest{
		Stream:      stream,
		NPredict:    nPredict,
		Temperature: float64(temperature),
		Stop: []string{ // Common stop tokens
			"</s>",
			"<|end|>",
			"<|eot_id|>",
			"<|end_of_text|>",
			"<|im_end|>",
			"<|EOT|>",
			"<|END_OF_TURN_TOKEN|>",
			"<|end_of_turn|>",
			"<|endoftext|>",
			"VISITOR:", // Stop if assistant starts hallucinating visitor turns
			"USER:",
			"\n\n\n", // Multiple newlines can indicate end of thought
		},
		RepeatLastN:      256, // Consider context window size
		RepeatPenalty:    1.1,
		TopK:             40,
		TopP:             0.9, // Standard TopP
		MinP:             0.05,
		TfsZ:             1,
		TypicalP:         1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Mirostat:         2,           // Enable Mirostat
		MirostatTau:      mirostatTau, // Convert int to float64
		MirostatEta:      0.1,
		Grammar:          grammar, // GBNF grammar for structured output
		NProbs:           0,
		MinKeep:          0,
		ImageData:        []interface{}{},
		CachePrompt:      true, // Enable prompt caching if LLM supports it well
		APIKey:           "",   // API key if required by the LLM server
		Prompt:           prompt,
		JsonSchema:       schema,
	}
}

// StreamAssistantTest is a non-streaming version for testing speech generation.
// Kept for compatibility with evaluation service, but ideally, eval would use streaming or a dedicated non-streaming endpoint.
func (srv *Service) StreamAssistantTest(msg entities.WebSocketMessage, inst entities.Instructions, temp float32, miro int) string {
	url := srv.Conf.LlmSmall
	if inst.LlmSize == "big" {
		url = srv.Conf.LlmBig
	}

	prompt, err := srv.Pr.AssembleInstructionsPrompt(msg, inst, inst.BasePrompt)
	if err != nil {
		srv.Log.Errorw("Prompt assembly failed for StreamAssistantTest", "error", err)
		return "[Error assembling prompt]"
	}

	nPredict := inst.Limit
	if nPredict == 0 {
		nPredict = 256 // Default for speech
	}

	payloadStruct := srv.AssemblePayload(nPredict, false, temp, prompt, srv.Pr.AssembleEmpty(), srv.Pr.AssembleEmpty(), miro) // stream is false
	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		srv.Log.Errorw("Failed to marshal payload for StreamAssistantTest", "error", err)
		return "[Error marshalling payload]"
	}

	res, err := srv.callLLM(url, payload, false)
	if err != nil {
		return "[Error calling LLM]"
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		srv.Log.Errorw("Failed to read LLM response body for StreamAssistantTest", "error", err)
		return "[Error reading LLM response]"
	}

	var serverResponse entities.AssistantResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		srv.Log.Errorw("Failed to unmarshal LLM server response for StreamAssistantTest", "body", string(body), "error", err)
		return "[Error unmarshalling LLM response: " + string(body) + "]"
	}

	return serverResponse.Content
}
