package command

import (
	websocketServer "Llamacommunicator/api/websocket"
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
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

	server := websocketServer.NewWebSocket(cmd.Log, validator.New())
	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		server.HandleWebSocket(c)

	}))

	log.Fatal(app.Listen(":3000"))
}
