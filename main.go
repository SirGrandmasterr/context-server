package main

import (
	"Llamacommunicator/pkg/command"
	"Llamacommunicator/pkg/config"
	"Llamacommunicator/pkg/logger"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "llm-whisperer"
	app.Usage = "Mediates between Unity and Llama"
	app.UsageText = fmt.Sprintf("%s command [arguments]", "llm-whisperer")

	log := logger.NewLogger()
	err := godotenv.Load()
	if err != nil {
		log.DPanic(err)
	}
	//ctx := context.Background()
	conf := config.NewSpecification()
	fmt.Print(conf)

	baseCommand := command.NewBaseCommand(conf, log)
	httpCommand := command.NewHttpCommand(baseCommand)
	migrateCommand := command.NewMigrateCommand(baseCommand)
	websocketCommand := command.NewWebsocketCommand(baseCommand)
	benchmarkCommand := "TODO"
	log.Infoln(benchmarkCommand + ": gotta implement BenchmarkCommand")

	app.Commands = []cli.Command{
		{
			Name:   "http", // ./llm-whisperer http
			Usage:  "Start REST API service",
			Action: httpCommand.Run,
		},
		{
			Name:   "migrate", // ./llm-whisperer migrate
			Usage:  "Upload init_db.json to mongodb",
			Action: migrateCommand.Run,
		},
		{
			Name:   "websocket", // ./llm-whisperer websocket
			Usage:  "Start Websocket Server",
			Action: websocketCommand.Run,
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
