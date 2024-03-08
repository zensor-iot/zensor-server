package persistence

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"zensor-server/internal/logger"

	"github.com/jackc/pgx/v4"
)

const (
	upSuffix   = "up.sql"
	downSuffix = "down.sql"

	maxRetries int = 10
)

type Database interface {
	Open() error
	Close()
	Run(string) error
	Query(string, ...interface{}) [][]byte
	Up(string)
}

type DatabaseWrapper struct {
	URL  string
	Conn *pgx.Conn
}

func NewDatabase(url string) Database {
	return &DatabaseWrapper{url, nil}
}

func (d *DatabaseWrapper) Open() error {
	for try := 0; try < maxRetries; try++ {
		conn, err := pgx.Connect(context.Background(), d.URL)

		if err != nil {
			logger.Info("error connecting to database", logger.Err(err))
			logger.Info("retrying... ")
			time.Sleep(5 * time.Second)
		} else {
			d.Conn = conn
			return nil
		}
	}

	return fmt.Errorf("imposible to connect to database after %d retries", maxRetries)
}

func (d *DatabaseWrapper) Close() {
	d.Conn.Close(context.Background())
}

func (d *DatabaseWrapper) Run(sql string) error {
	_, err := d.Conn.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("run failed: %w", err)
	}

	return nil
}

func (d *DatabaseWrapper) Query(sql string, args ...interface{}) [][]byte {
	rows, err := d.Conn.Query(context.Background(), sql, args)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	values := make([][]byte, 0)
	for rows.Next() {
		values = append(values, rows.RawValues()[0])
	}
	return values
}

func (d *DatabaseWrapper) Up(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.Contains(file.Name(), upSuffix) {
			content, err := ioutil.ReadFile(path + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}

			statement := string(content)

			d.Run(statement)
		}
	}

}
