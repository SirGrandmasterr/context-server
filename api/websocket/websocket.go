package websocketServer

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"context"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

type WebSocketServer struct {
	clients          map[*websocket.Conn]bool
	Log              *zap.SugaredLogger
	AssistantService assistant.Service
	val              *validator.Validate
}

func NewWebSocket(log *zap.SugaredLogger, val *validator.Validate) *WebSocketServer {
	return &WebSocketServer{
		clients:          make(map[*websocket.Conn]bool),
		Log:              log,
		AssistantService: *assistant.NewAssistantService(log, val),
	}
}

func (s *WebSocketServer) HandleWebSocket(conn *websocket.Conn) {

	// Register a new Client
	s.clients[conn] = true
	defer func() {
		delete(s.clients, conn)
		conn.Close()
	}()
	//var assistantChannel chan *entities.WebSocketAnswer = make(chan *entities.WebSocketAnswer)
	//var assistant = assistant.NewAssistantProcess(s.Log, assistantChannel)
	//go assistant.Awake()
	//go s.LoopForAssistantChannel(conn, assistantChannel)
	for {
		_, msg, err := conn.ReadMessage()
		s.Log.Debugln("Received Message: ")
		if err != nil {
			s.Log.Errorln("Read Error:", err)
			break
		}
		var message entities.WebSocketMessage
		if err := json.Unmarshal(msg, &message); err != nil {
			s.Log.Fatalf("Error Unmarshalling")
		} else {
			s.Log.Infoln(message)
		}

		//testmsg := <-assistantChannel
		s.AssistantService.AskAssistant(context.Background(), &entities.RequestAssistantReaction{
			ActionName: "TestAction",
		})
		testanswer := entities.WebSocketAnswer{
			Type: "speech",
			Text: message.Speech,
		}
		testJson, err := json.Marshal(testanswer)
		err = conn.WriteMessage(2, testJson)
		if err != nil {
			s.Log.Errorln("Writing Json didn't work.")
		}
	}
}

func (s *WebSocketServer) LoopForAssistantChannel(conn *websocket.Conn, ch chan *entities.WebSocketAnswer) {
	for {
		msg := <-ch

		answer := entities.WebSocketAnswer{Type: "speech", Text: msg.Text}
		err := conn.WriteJSON(answer)
		if err != nil {
			s.Log.Errorln("Writing Json didn't work.")
		}
	}
}
