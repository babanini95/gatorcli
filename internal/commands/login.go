package commands

import "fmt"

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("login command expected an argument")
	}

	err := s.cfg.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}
