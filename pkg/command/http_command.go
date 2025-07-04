package command

import (
	"Llamacommunicator/api/handler"
	"Llamacommunicator/api/router"
	"Llamacommunicator/pkg/storage"
	"context"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/urfave/cli"
)

type HttpCommand struct {
	BaseCommand
}

func NewHttpCommand(baseCommand BaseCommand) *HttpCommand {
	return &HttpCommand{
		BaseCommand: baseCommand,
	}
}

func (cmd *HttpCommand) Run(clictx *cli.Context) {
	db := cmd.BaseCommand.NewDatabaseConnection()
	defer db.Client().Disconnect(context.Background())
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	storageWriter := storage.NewStorageWriter(cmd.Log, db)
	storageReader := storage.NewStorageReader(cmd.Log, db)

	//authService := auth.NewAuthService(cmd.Config)
	app.Post("/login", handler.Login(storageReader, cmd.Config))
	app.Post("/create", handler.CreateUser(storageReader, storageWriter))

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cmd.Config.Secret)},
	}))

	api := app.Group("/api")

	//validator := validator.New()

	//Services
	//assistantService := assistant.NewAssistantService(cmd.Log, validator)

	//router.AssistantRouter(api, assistantService)
	router.DBRouter(api, storageWriter)

	cmd.BaseCommand.Log.Fatal(app.Listen(":8079"))
}
