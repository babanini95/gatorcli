package commands

import (
	"database/sql"
	"time"
)

func sqlCurrentTime() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}
