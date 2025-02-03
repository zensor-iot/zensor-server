package sql

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgreDatabase struct {
	url  string
	Conn *pgxpool.Pool
	DB   *gorm.DB
}

func NewPosgreORM(dsn string) (*DB, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{
		DB: gormDB,
	}, nil
}

func NewPosgreDatabase(url string) *PostgreDatabase {
	return &PostgreDatabase{
		url: url,
	}
}

func (d *PostgreDatabase) Open() error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		conn, err1 := pgxpool.New(context.Background(), d.url)

		if err1 != nil {
			time.Sleep(5 * time.Second)
		} else {
			d.Conn = conn
			return nil
		}
	}

	return fmt.Errorf("imposible to connect to database after %d retries", maxRetries)
}

func (d *PostgreDatabase) Close() {
	d.Conn.Close()
}

func (d *PostgreDatabase) Command(sql string) error {
	_, err := d.Conn.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("run failed: %w", err)
	}

	return nil
}

func (d *PostgreDatabase) Query(ctx context.Context, sql string, args ...any) ([][]byte, error) {
	rows, err := d.Conn.Query(context.Background(), sql, args)
	if err != nil {
		return nil, fmt.Errorf("postgre query: %w", err)
	}

	defer rows.Close()
	values := make([][]byte, 0)
	for rows.Next() {
		values = append(values, rows.RawValues()[0])
	}
	return values, nil
}

func (d *PostgreDatabase) Up(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if strings.Contains(file.Name(), upSuffix) {
			content, err := os.ReadFile(path + "/" + file.Name())
			if err != nil {
				panic(err)
			}

			slog.Info("applying migration", slog.String("file", file.Name()))
			statement := string(content)

			err = d.Command(statement)
			if err != nil {
				panic(err)
			}
		}
	}
}
