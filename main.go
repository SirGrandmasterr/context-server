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
	benchmarkCommand := "TODO"
	log.Infoln(benchmarkCommand + ": gotta implement BenchmarkCommand")

	app.Commands = []cli.Command{
		{
			Name:   "http",
			Usage:  "Start REST API service",
			Action: httpCommand.Run,
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}