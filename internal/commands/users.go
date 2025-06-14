package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/babanini95/gatorcli/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("login command expected an argument")
	}

	userName := cmd.arguments[0]

	if !isUserExist(s, userName) {
		os.Exit(1)
	}

	err := s.cfg.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("register command expected an argument")
	}
	userName := cmd.arguments[0]

	if isUserExist(s, userName) {
		os.Exit(1)
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: sqlCurrentTime(),
		UpdatedAt: sqlCurrentTime(),
		Name:      userName,
	}
	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		return fmt.Errorf("create user failed: %v", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User %s has been created. User data:\n%v\n", userName, user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUser(context.Background())
	if err != nil {
		os.Exit(1)
		return err
	}
	fmt.Println("users database reset successfully")
	os.Exit(0)
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get all users: %v", err)
	}

	for _, user := range users {
		currentUserMark := ""
		if s.cfg.CurrentUserName == user.Name {
			currentUserMark = "(current)"
		}
		fmt.Printf("* %s %s\n", user.Name, currentUserMark)
	}

	return nil
}

func isUserExist(s *state, userName string) bool {
	u, _ := s.db.GetUser(context.Background(), userName)
	return u.ID != uuid.Nil
}
