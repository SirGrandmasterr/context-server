package assistant

import (
	"Llamacommunicator/pkg/config"
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

func NewAssistantProcess(log *zap.SugaredLogger, clientResponseChan chan *entities.WebSocketAnswer, stor *storage.StorageReader, storwr *storage.StorageWriter, conf *config.Specification) *AssistantProcess {
	return &AssistantProcess{
		clients:         make(map[*websocket.Conn]bool),
		Log:             log,
		responseChannel: clientResponseChan,
		serviceChannel:  make(chan *entities.WebSocketAnswer),
		aserv:           *assistant.NewAssistantService(log, validator.New(), clientResponseChan, stor, storwr, conf),
	}
}

func (ap *AssistantProcess) Analyze(msg entities.WebSocketMessage) {
	switch msg.MessageType {
	case "initializePlayer":
		//Wipe Player History, we don't have enough context window for extended sessions.
		ap.aserv.StorageWriter.ResetPlayerHistory(msg.PlayerContext.PlayerUsername)
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

		action := ap.aserv.DetectAction(context.Background(), msg, ap.serviceChannel, 1.2)
		ap.Log.Infoln("Detected Action", action)
		action_db, err := ap.aserv.Storage.ReadActionOptionEntity(action.ActionName, context.Background())
		ap.Log.Infoln("Found Action in Database:", action_db)
		if err != nil {
			ap.Log.Errorln(err)
			return
		}
		ap.Log.Infoln("Creating ActionToken")
		tok := entities.ActionToken{
			ID:           primitive.NewObjectID(),
			Name:         action_db.ActionName,
			Description:  action_db.Description,
			CurrentStage: 0,
		}
		ap.Log.Infoln("Saving ActionToken, ", tok)
		_, err = ap.aserv.StorageWriter.SaveActionToken(tok, context.Background())
		if err != nil {
			ap.Log.Errorln("Error during saving of action token: ", err)
			return
		}

		//Needed to validate if all instructions are done
		ap.InstructionsLoop(action_db, tok, msg, false)

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

		tok, err := ap.aserv.Storage.ReadActionToken(context.Background(), msg.ActionContext.Token)
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}

		action, err := ap.aserv.Storage.ReadActionOptionEntity(tok.Name, context.Background())
		if err != nil {
			ap.Log.Errorln("Error retrieving player", err)
		}
		ap.Log.Infoln("About to begin instructionsloop with ", tok)
		tok.CurrentStage = tok.CurrentStage - 1 // We are repeating the stage that we aborted the loop at
		ap.InstructionsLoop(action, tok, msg, true)

	case "envEvent":
		actionResponse := ap.aserv.DecideReaction(context.Background(), msg, ap.serviceChannel)

		if actionResponse.ActionName == "ignore" {
			ap.Log.Infoln("Event was ignored.")
			answer := entities.WebSocketAnswer{
				Type:       "action",
				Text:       "Event was ignored.",
				ActionName: "ignore",
				Token:      primitive.NilObjectID,
				Stage:      0,
			}
			ap.responseChannel <- &answer
		} else {
			action_db, err := ap.aserv.Storage.ReadActionOptionEntity(actionResponse.ActionName, context.Background())
			ap.Log.Infoln("Found Action in Database:", action_db)
			if err != nil {
				ap.Log.Errorln(err)
				return
			}
			ap.Log.Infoln("Creating ActionToken")
			tok := entities.ActionToken{
				ID:           primitive.NewObjectID(),
				Name:         action_db.ActionName,
				Description:  action_db.Description,
				CurrentStage: 0,
			}
			ap.Log.Infoln("Saving ActionToken, ", tok)
			_, err = ap.aserv.StorageWriter.SaveActionToken(tok, context.Background())
			if err != nil {
				ap.Log.Errorln("Error during saving of action token: ", err)
				return
			}

			//Needed to validate if all instructions are done
			ap.InstructionsLoop(action_db, tok, msg, false)
		}
	case "innerThoughtEvent":
		action := ap.aserv.DetectAction(context.Background(), msg, ap.serviceChannel, 2)
		ap.Log.Infoln("Detected Action", action)
		action_db, err := ap.aserv.Storage.ReadActionOptionEntity(action.ActionName, context.Background())
		ap.Log.Infoln("Found Action in Database:", action_db)
		if err != nil {
			ap.Log.Errorln(err)
			return
		}
		ap.Log.Infoln("Creating ActionToken")
		tok := entities.ActionToken{
			ID:           primitive.NewObjectID(),
			Name:         action_db.ActionName,
			Description:  action_db.Description,
			CurrentStage: 0,
		}
		ap.Log.Infoln("Saving ActionToken, ", tok)
		_, err = ap.aserv.StorageWriter.SaveActionToken(tok, context.Background())
		if err != nil {
			ap.Log.Errorln("Error during saving of action token: ", err)
			return
		}

		//Needed to validate if all instructions are done
		ap.InstructionsLoop(action_db, tok, msg, false)
	}

}

func (ap *AssistantProcess) InstructionsLoop(action_db entities.Action, tok entities.ActionToken, msg entities.WebSocketMessage, actionUpdate bool) {
	for tok.CurrentStage <= action_db.Stages-1 {
		inst := action_db.Instructions[tok.CurrentStage]
		tok.CurrentStage = tok.CurrentStage + 1
		err := ap.updateToken(tok.CurrentStage, tok)
		if err != nil {
			ap.Log.Errorln(err)
		}
		ap.Log.Infoln("Entering Instructions-Loop")
		switch inst.Type {
		case "actionselection":
			if inst.PermissionRequired && !msg.ActionContext.Permission {
				if actionUpdate {
					deleted, _ := ap.CheckDeleteToken(action_db.Stages, tok)
					if deleted {
						return
					}
					continue
				}
				return
			}
			if inst.Stage == 1 {
				ap.Log.Infoln("PreparingWebSocketAnswer in Loop")
				answer := entities.WebSocketAnswer{
					Type:       "action",
					Text:       "",
					ActionName: action_db.ActionName,
					Token:      tok.ID,
					Stage:      inst.Stage,
				}
				ap.responseChannel <- &answer
				_, _ = ap.CheckDeleteToken(action_db.Stages, tok)
			} else {
				ac := ap.aserv.DetectAction(context.Background(), msg, ap.serviceChannel, 1.2)
				secondaryAction, err := ap.aserv.Storage.ReadActionOptionEntity(ac.ActionName, context.Background())
				if err != nil {
					ap.Log.Errorln(err)
				}
				answer := entities.WebSocketAnswer{
					Type:       "actionSelection",
					Text:       secondaryAction.ActionName,
					ActionName: action_db.ActionName,
					Token:      tok.ID,
					Stage:      inst.Stage,
				}
				if secondaryAction.Stages != 1 {
					answer.Text = "ignore"
				}
				ap.responseChannel <- &answer
				_, _ = ap.CheckDeleteToken(action_db.Stages, tok)
			}

		case "actionquery":
			if inst.PermissionRequired && !msg.ActionContext.Permission {
				if actionUpdate {
					deleted, _ := ap.CheckDeleteToken(action_db.Stages, tok)
					if deleted {
						return
					}
					continue
				}
				return
			}
			ap.Log.Infoln("Instructionloop: actionquery.", "Stage: ", rune(inst.Stage))
			result, err := ap.aserv.ActionQuery(msg, inst, action_db.ActionName)
			if err != nil {
				ap.Log.Errorln(err)
			}
			result.Token = tok.ID
			result.Stage = inst.Stage
			ap.responseChannel <- &result
			_, _ = ap.CheckDeleteToken(action_db.Stages, tok)
		case "playerSpeechAnalysis":
			if inst.PermissionRequired && !msg.ActionContext.Permission {
				if actionUpdate {
					deleted, _ := ap.CheckDeleteToken(action_db.Stages, tok)
					if deleted {
						return
					}
					continue
				}
				return
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
			_, _ = ap.CheckDeleteToken(action_db.Stages, tok)
		case "speech":
			if inst.PermissionRequired && !msg.ActionContext.Permission {
				//This means that, while looping, we made it to a stage that requires the 3D-Client to tell us if this stage is necessary.
				//
				if actionUpdate {
					ap.Log.Infoln("CurrentStage: ", tok.CurrentStage, " perm req, none give, action update given")
					deleted, _ := ap.CheckDeleteToken(action_db.Stages, tok)
					if deleted {
						ap.Log.Infoln("CurrentStage: ", tok.CurrentStage, " perm req, none give, action update given, deleted true")
						return
					}
					ap.Log.Infoln("CurrentStage: ", tok.CurrentStage, " perm req, none give, action update given, deleted false")
					continue
				}
				return
			}

			ap.aserv.StreamAssistant(msg, inst)

			_, _ = ap.CheckDeleteToken(action_db.Stages, tok)
			break
		}
		//This is only ever true in actionUpdate, and is supposed to be used for one instruction only.
		msg.ActionContext.Permission = false
		//Update gives Permission fo
		actionUpdate = false
	}
}

func (ap *AssistantProcess) CheckDeleteToken(numStages int, tok entities.ActionToken) (bool, error) {
	if tok.CurrentStage == numStages {
		//Action completed.
		ap.Log.Infoln("Action " + tok.Name + " completed at stage " + string(tok.CurrentStage))
		err := ap.aserv.StorageWriter.DeleteActionToken(tok.ID, context.Background())
		if err != nil {
			ap.Log.Errorln(err)
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func (ap *AssistantProcess) updateToken(currentStage int, tok entities.ActionToken) error {
	err := ap.aserv.StorageWriter.UpdateActionTokenStage(tok.ID, currentStage, context.Background())
	if err != nil {
		ap.Log.Errorln(err)
		return err
	}
	return nil
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
