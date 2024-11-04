package command

import (
	"Llamacommunicator/api/handler"
	"Llamacommunicator/api/router"
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"Llamacommunicator/pkg/services/evaluation"
	"Llamacommunicator/pkg/services/prompting"
	"Llamacommunicator/pkg/storage"
	"context"

	"github.com/go-playground/validator/v10"
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
	assistantserv := assistant.NewAssistantService(
		cmd.Log,
		validator.New(),
		make(chan *entities.WebSocketAnswer),
		storageReader,
		storageWriter,
		cmd.Config,
		prompting.NewPromptService(cmd.Log, storageReader, storageWriter),
	)
	evalserv := evaluation.NewEvalService(cmd.Log, cmd.Config, assistantserv)
	//authService := auth.NewAuthService(cmd.Config)
	app.Post("/login", handler.Login(storageReader))
	app.Post("/create", handler.CreateUser(storageReader, storageWriter))

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte("secret")},
	}))

	api := app.Group("/api")
	testApi := app.Group("/eval")
	//validator := validator.New()

	//Services
	//assistantService := assistant.NewAssistantService(cmd.Log, validator)

	//router.AssistantRouter(api, assistantService)
	router.DBRouter(api, storageWriter)
	router.EvalRouter(testApi, storageWriter, storageReader, evalserv)

	cmd.BaseCommand.Log.Fatal(app.Listen(":8079"))
}
