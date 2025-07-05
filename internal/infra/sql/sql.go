package sql

import "context"

const (
	upSuffix   = "up.sql"
	downSuffix = "down.sql"

	maxRetries int = 10
)

type Database interface {
	Open() error
	Close()
	Command(string) error
	Query(context.Context, string, ...any) ([][]byte, error)
}
