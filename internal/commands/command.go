package commands

import (
	"fmt"
	"os"

	"github.com/babanini95/gatorcli/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.cmds[cmd.name]
	if !ok {
		return fmt.Errorf("command unavailable")
	}

	err := f(s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func (c *commands) Register() {
	c.register("login", handlerLogin)
}

func (c *commands) Run(s *state, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "not enough arguments")
		os.Exit(1)
	}

	cmd := command{
		name: args[1],
	}
	if len(args) >= 3 {
		cmd.arguments = args[2:]
	}

	err := c.run(s, cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (s *state) SaveConfig(c *config.Config) {
	s.cfg = c
}

func CreateNewState(c *config.Config) (*state, error) {
	if c == nil {
		return &state{}, fmt.Errorf("config is empty")
	}
	return &state{cfg: c}, nil
}

func InitCommands() *commands {
	return &commands{
		cmds: make(map[string]func(*state, command) error),
	}
}
