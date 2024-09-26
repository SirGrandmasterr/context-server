package assistant

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"Llamacommunicator/pkg/storage"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AssistantProcess struct {
	clients         map[*websocket.Conn]bool
	Log             *zap.SugaredLogger
	responseChannel chan *entities.WebSocketAnswer
	serviceChannel  chan *entities.WebSocketAnswer
	aserv           assistant.Service
}

func NewAssistantProcess(log *zap.SugaredLogger, clientResponseChan chan *entities.WebSocketAnswer, stor *storage.StorageReader, storwr *storage.StorageWriter) *AssistantProcess {
	return &AssistantProcess{
		clients:         make(map[*websocket.Conn]bool),
		Log:             log,
		responseChannel: clientResponseChan,
		serviceChannel:  make(chan *entities.WebSocketAnswer),
		aserv:           *assistant.NewAssistantService(log, validator.New(), clientResponseChan, stor, storwr),
	}
}

func (ap *AssistantProcess) Analyze(msg entities.WebSocketMessage) {
	switch msg.MessageType {
	case "speech":

		player, err := ap.aserv.Storage.ReadPlayer(msg.PlayerContext.PlayerUsername, context.Background())
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}
		ap.Log.Infoln("Updating Player History")
		player.History = player.History + "VISITOR: " + msg.Speech + "\n"
		err = ap.aserv.StorageWriter.UpdatePlayerHistory(player.Username, player.History)
		if err != nil {
			ap.Log.Errorln(err)
		}

		action, _ := ap.aserv.DetectAction(context.Background(), msg, ap.serviceChannel)
		ap.Log.Infoln("Detected Action", action)
		action_db, err := ap.aserv.Storage.ReadActionOptionEntity(action.ActionName, context.Background())
		ap.Log.Infoln("Found Action in Database:", action_db)
		if err != nil {
			ap.Log.Errorln(err)
			return
		}
		ap.Log.Infoln("Creating ActionToken")
		tok := entities.ActionToken{
			ID:          primitive.NewObjectID(),
			Name:        action_db.ActionName,
			Description: action_db.Description,
		}
		ap.Log.Infoln("Saving ActionToken, ", tok)
		_, err = ap.aserv.StorageWriter.SaveActionToken(tok, context.Background())
		if err != nil {
			ap.Log.Errorln("Error during saving of action token: ", err)
			return
		}
		ap.Log.Infoln("CheckingSingleStage")
		// In this case there is only an actionselection stage and we can return immediately.
		/*if action_db.Stages == 1 {
			ap.Log.Infoln("Sending back answer due to single-stage")
			answer := entities.WebSocketAnswer{
				Type:       "action",
				Text:       "",
				ActionName: action_db.ActionName,
				Token:      primitive.ObjectID{},
			}
			ap.responseChannel <- &answer
			return
		}*/
		instructionCounter := 0 //Needed to validate if all instructions are done
		ap.InstructionsLoop(instructionCounter, action_db, tok, msg, false)

	case "playerHistoryUpdate":
		player, err := ap.aserv.Storage.ReadPlayer(msg.PlayerContext.PlayerUsername, context.Background())
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}
		player.History = player.History + msg.Speech + "\n"
		err = ap.aserv.StorageWriter.UpdatePlayerHistory(player.Username, player.History)
		if err != nil {
			ap.Log.Errorln(err)
		}
	case "actionUpdate":
		/*_, err := ap.aserv.Storage.ReadPlayer(msg.PlayerContext.PlayerUsername, context.Background())
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}*/
		tok, err := ap.aserv.Storage.ReadActionToken(context.Background(), msg.Token)
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}
		action, err := ap.aserv.Storage.ReadActionOptionEntity(tok.Name, context.Background())
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}
		ap.InstructionsLoop(tok.CurrentStage, action, tok, msg, true)
	}

}

func (ap *AssistantProcess) InstructionsLoop(stageCounter int, action_db entities.Action, tok entities.ActionToken, msg entities.WebSocketMessage, actionUpdate bool) {
	for stageCounter <= action_db.Stages-1 {
		inst := action_db.Instructions[stageCounter]
		ap.Log.Infoln("Entering Instructions-Loop")
		switch inst.Type {
		case "actionselection":
			if inst.Conditional {
				return
			}
			answer := entities.WebSocketAnswer{
				Type:       "action",
				Text:       "",
				ActionName: action_db.ActionName,
				Token:      tok.ID,
				Stage:      inst.Stage,
			}
			ap.responseChannel <- &answer

		case "actionquery":
			if inst.Conditional && !msg.ActionContext.Permission {
				return
			}
			ap.Log.Infoln("Instructionloop: actionquery.", "Stage: ", rune(inst.Stage))
		case "objectselection":
			if inst.Conditional && !msg.ActionContext.Permission {
				return
			}
			ap.Log.Infoln("Instructionloop: objectselection.", "Stage: ", rune(inst.Stage))
		case "playerSpeechAnalysis":
			if inst.Conditional && !msg.ActionContext.Permission {

			}
			//In this type the LLM needs to filter out relevant info from the users words.
			ap.Log.Infoln("Instructionloop: objectselection.", "Stage: ", rune(inst.Stage))
			result, err := ap.aserv.PlayerSpeechAnalysis(msg, inst, action_db.ActionName)
			if err != nil {
				ap.Log.Errorln(err)
			}
			ap.Log.Infoln(result)
			result.Token = tok.ID
			result.Stage = inst.Stage
			ap.responseChannel <- &result

		case "speech":
			if inst.Conditional && !msg.ActionContext.Permission {
				ap.aserv.StorageWriter.UpdateActionTokenStage(tok.ID, inst.Stage, context.Background())
				return
			}
			ap.aserv.StreamAssistant(msg, inst)

			if stageCounter == inst.Stage-1 {
				//Action completed.
				ap.Log.Infoln("Action " + action_db.ActionName + " completed at stage " + string(inst.Stage))
				ap.aserv.StorageWriter.DeleteActionToken(tok.ID, context.Background())
				return
			}
			break
		}
		stageCounter++
		//This is only ever true in actionUpdate, and is supposed to be used for one instruction only.
		msg.ActionContext.Permission = false
		//Update gives Permission fo
		actionUpdate = false
	}
}

//ap.Log.Infoln(action.ActionName)

//

/*for {
	msg := <-ap.serviceChannel
	if msg.Type == "partial" {

	}
}*/

//ap.aserv.StreamAssistant(msg)

/*for i := range 10 {
	ap.Log.Infoln("Awake.", i)
	testmsg := entities.WebSocketAnswer{
		Type: "speech",
		Text: "A very particular piece of text",
		//Action:     false,
	}

	ap.responseChannel <- &testmsg
	ap.Log.Infoln("Sent a msg.")
}*/
