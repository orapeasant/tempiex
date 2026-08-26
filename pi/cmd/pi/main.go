package main

import (
	"context"
	"flag"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/tempiex/pi/internal/agent"
	"github.com/tempiex/pi/internal/config"
	"github.com/tempiex/pi/internal/tool"
	"github.com/tempiex/pi/tools/file"
	"github.com/tempiex/pi/tools/httptool"
	"github.com/tempiex/pi/tools/shell"
)

func main() {
	configPath := flag.String("config", "config/development.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	registry := tool.NewRegistry()
	registry.Register(shell.New())
	registry.Register(httptool.New())
	registry.Register(file.NewReadTool())
	registry.Register(file.NewWriteTool())

	a := agent.New(agent.Options{
		Config:   cfg,
		Registry: registry,
	})

	if err := a.Run(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("agent run")
		os.Exit(1)
	}
}
