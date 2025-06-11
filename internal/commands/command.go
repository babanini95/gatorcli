package commands

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/babanini95/gatorcli/internal/config"
	"github.com/babanini95/gatorcli/internal/database"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) generateCommands() {
	cmds := map[string]func(*state, command) error{
		"login":     handlerLogin,
		"register":  handlerRegister,
		"reset":     handlerReset,
		"users":     handlerUsers,
		"agg":       middlewareLoggedIn(handlerAgg),
		"addfeed":   middlewareLoggedIn(handlerAddFeed),
		"feeds":     handlerFeeds,
		"follow":    middlewareLoggedIn(handlerFollow),
		"following": middlewareLoggedIn(handlerFollowing),
		"unfollow":  middlewareLoggedIn(handlerUnfollow),
		"browse":    middlewareLoggedIn(handlerBrowse),
	}

	for name, fn := range cmds {
		c.register(name, fn)
	}
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

func (s *state) CreateQueries() error {
	db, err := sql.Open("postgres", s.cfg.DbURL)
	if err != nil {
		return fmt.Errorf("queries can not be created: %v", err)
	}

	s.db = database.New(db)
	return nil
}

func CreateNewState(c *config.Config) (*state, error) {
	if c == nil {
		return &state{}, fmt.Errorf("config is empty")
	}
	return &state{cfg: c}, nil
}

func InitCommands() *commands {
	cmds := &commands{
		cmds: make(map[string]func(*state, command) error),
	}

	cmds.generateCommands()
	return cmds
}
