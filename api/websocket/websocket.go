package websocketServer

import (
	"Llamacommunicator/pkg/assistant"
	"Llamacommunicator/pkg/entities"
	"encoding/json"

	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

type WebSocketServer struct {
	clients map[*websocket.Conn]bool
	Log     *zap.SugaredLogger
}

func NewWebSocket(log *zap.SugaredLogger) *WebSocketServer {
	return &WebSocketServer{
		clients: make(map[*websocket.Conn]bool),
		Log:     log,
	}
}

func (s *WebSocketServer) HandleWebSocket(conn *websocket.Conn) {

	// Register a new Client
	s.clients[conn] = true
	defer func() {
		delete(s.clients, conn)
		conn.Close()
	}()
	var assistantChannel chan *entities.WebSocketAnswer = make(chan *entities.WebSocketAnswer)
	var assistant = assistant.NewAssistantProcess(s.Log, assistantChannel)
	go assistant.Awake()
	go s.LoopForAssistantChannel(conn, assistantChannel)
	for {
		s.Log.Infoln("Brother are we good?")
		_, msg, err := conn.ReadMessage()
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

		testmsg := <-assistantChannel
		conn.WriteJSON(testmsg)
	}
}

func (s *WebSocketServer) LoopForAssistantChannel(conn *websocket.Conn, ch chan *entities.WebSocketAnswer) {
	for {
		msg := <-ch
		conn.WriteJSON(msg)
	}
}
