package websocketServer

import (
	assistantProcess "Llamacommunicator/pkg/assistant"
	"Llamacommunicator/pkg/config"
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

type WebSocketServer struct {
	clients   map[*websocket.Conn]bool
	Log       *zap.SugaredLogger
	Val       *validator.Validate
	Storage   storage.StorageReader
	Storagewr storage.StorageWriter
	Config    *config.Specification
}

func NewWebSocket(log *zap.SugaredLogger, val *validator.Validate, storage storage.StorageReader, storagewr storage.StorageWriter, conf *config.Specification) *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[*websocket.Conn]bool),
		Log:       log,
		Val:       val,
		Storage:   storage,
		Storagewr: storagewr,
		Config:    conf,
	}
}

func (s *WebSocketServer) HandleWebSocket(conn *websocket.Conn) {
	s.Log.Infoln("Handling incoming connection")
	// Register a new Client
	s.clients[conn] = true
	defer func() {
		delete(s.clients, conn)
		conn.Close()
	}()
	var clientResponseChannel chan *entities.WebSocketAnswer = make(chan *entities.WebSocketAnswer)
	var assistant = assistantProcess.NewAssistantProcess(s.Log, clientResponseChannel, &s.Storage, &s.Storagewr, s.Config)
	go s.LoopForClientResponseChannel(conn, clientResponseChannel)
	for {
		_, msg, err := conn.ReadMessage()
		s.Log.Infoln("Received Message: ")
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
		assistant.Analyze(message)
	}
}

func (s *WebSocketServer) LoopForClientResponseChannel(conn *websocket.Conn, ch chan *entities.WebSocketAnswer) {
	for {
		msg := <-ch

		sendToCon, err := json.Marshal(msg)
		if err != nil {
			s.Log.Errorln("Writing Json didn't work.")
		}
		err = conn.WriteMessage(2, sendToCon)
		if err != nil {
			s.Log.Errorln("Writing message to Conn didn't work.")
		}
	}
}
