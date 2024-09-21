package assistant

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"Llamacommunicator/pkg/storage"

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

func NewAssistantProcess(log *zap.SugaredLogger, resChan chan *entities.WebSocketAnswer, stor *storage.StorageReader) *AssistantProcess {
	return &AssistantProcess{
		clients:         make(map[*websocket.Conn]bool),
		Log:             log,
		responseChannel: resChan,
		serviceChannel:  make(chan *entities.WebSocketAnswer),
		aserv:           *assistant.NewAssistantService(log, validator.New(), resChan, stor),
	}
}

func (ap *AssistantProcess) Analyze(msg entities.WebSocketMessage) {
	switch msg.MessageType {
	case "speech":
		ap.aserv.StreamAssistant(msg.Speech)
		break
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
