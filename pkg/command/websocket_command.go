package command

import (
	"Llamacommunicator/api/handler"
	websocketServer "Llamacommunicator/api/websocket"
	"Llamacommunicator/pkg/storage"
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/urfave/cli"
)

type WebsocketCommand struct {
	BaseCommand
}

func NewWebsocketCommand(baseCommand BaseCommand) *WebsocketCommand {
	return &WebsocketCommand{
		BaseCommand: baseCommand,
	}
}

type MyCustomClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (cmd *WebsocketCommand) Run(clictx *cli.Context) {
	db := cmd.BaseCommand.NewDatabaseConnection()
	defer db.Client().Disconnect(context.Background())
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	storager := storage.NewStorageReader(cmd.Log, cmd.Db)
	storagewr := storage.NewStorageWriter(cmd.Log, cmd.Db)
	server := websocketServer.NewWebSocket(cmd.Log, validator.New(), *storager, *storagewr, cmd.BaseCommand.Config)
	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {

		if !cmd.verifyToken(c.Params("id")) {
			log.Println("ERROR")
			server.KillWebSocket(c)
			return
		}
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		server.HandleWebSocket(c)

	}))
	app.Get("/ping", handler.Pong())
	app.Post("/login", handler.Login(storager, cmd.BaseCommand.Config))
	app.Post("/create", handler.CreateUser(storager, storagewr))

	log.Fatal(app.Listen(":3000"))
}

func (cmd *WebsocketCommand) verifyToken(tokenString string) bool {
	// Parse the token with the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cmd.Config.Secret), nil
	})

	// Check for verification errors
	if err != nil {
		cmd.Log.Infoln("Err Token", err)
		return false
	}

	// Check if the token is valid
	if !token.Valid {
		return false
	}
	return true
}
