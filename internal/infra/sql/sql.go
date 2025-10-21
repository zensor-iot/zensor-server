package sql

import "context"

type Database interface {
	Open() error
	Close()
	Command(string) error
	Query(context.Context, string, ...any) ([][]byte, error)
}
