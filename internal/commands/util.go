package commands

import (
	"context"
	"database/sql"
	"time"

	"github.com/babanini95/gatorcli/internal/database"
)

func sqlCurrentTime() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

func (s *state) currentUser() database.User {
	currentUser, _ := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	return currentUser
}
