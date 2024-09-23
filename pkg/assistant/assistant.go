package assistant

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"Llamacommunicator/pkg/storage"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

type AssistantProcess struct {
	clients         map[*websocket.Conn]bool
	Log             *zap.SugaredLogger
	responseChannel chan *entities.WebSocketAnswer
	serviceChannel  chan *entities.WebSocketAnswer
	aserv           assistant.Service
}

func NewAssistantProcess(log *zap.SugaredLogger, clientResponseChan chan *entities.WebSocketAnswer, stor *storage.StorageReader) *AssistantProcess {
	return &AssistantProcess{
		clients:         make(map[*websocket.Conn]bool),
		Log:             log,
		responseChannel: clientResponseChan,
		serviceChannel:  make(chan *entities.WebSocketAnswer),
		aserv:           *assistant.NewAssistantService(log, validator.New(), clientResponseChan, stor),
	}
}

func (ap *AssistantProcess) Analyze(msg entities.WebSocketMessage) {
	switch msg.MessageType {
	case "speech":
		action, _ := ap.aserv.DetectAction(context.Background(), msg, ap.serviceChannel)
		//ap.Log.Infoln(action.ActionName)

		testmsg := entities.WebSocketAnswer{
			Type:       "action",
			Text:       "followPlayer",
			ActionName: action.ActionName,
		}
		ap.responseChannel <- &testmsg

		/*for {
			msg := <-ap.serviceChannel
			if msg.Type == "partial" {

			}
		}*/

		ap.aserv.StreamAssistant(msg)

	case "update":

	}

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

}
