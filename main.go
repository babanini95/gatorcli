package main

import (
	"fmt"
	"os"

	"github.com/babanini95/gatorcli/internal/commands"
	"github.com/babanini95/gatorcli/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	appState, err := commands.CreateNewState(cfg)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	appState.SaveConfig(cfg)
	cmds := commands.InitCommands()
	cmds.Register()
	cmds.Run(appState, os.Args)
}
