package assistant

import (
	"Llamacommunicator/pkg/entities"

	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

type AssistantProcess struct {
	clients         map[*websocket.Conn]bool
	Log             *zap.SugaredLogger
	responseChannel chan *entities.WebSocketAnswer
}

func NewAssistantProcess(log *zap.SugaredLogger, resChan chan *entities.WebSocketAnswer) *AssistantProcess {
	return &AssistantProcess{
		clients:         make(map[*websocket.Conn]bool),
		Log:             log,
		responseChannel: resChan,
	}
}

func (ap *AssistantProcess) Awake() {
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
