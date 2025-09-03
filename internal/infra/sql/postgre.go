package sql

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	_defaultQueryTimeout = 30 * time.Second
	_maxRetries          = 5
)

type PostgreDatabase struct {
	url  string
	Conn *pgxpool.Pool
	DB   *gorm.DB
}

// Singleton pattern for PostgreSQL database
var (
	postgreInstance *PostgreDatabase
	postgreOnce     sync.Once
	postgreMutex    sync.RWMutex
)

func NewPosgreORM(dsn string) (*DB, error) {
	return NewPosgreORMWithTimeout(dsn, _defaultQueryTimeout)
}

func NewPosgreORMWithTimeout(dsn string, timeout time.Duration) (*DB, error) {
	pass, ok := os.LookupEnv("ZENSOR_SERVER_POSTGRES_PASSWORD")
	if ok {
		dsn = fmt.Sprintf("%s password=%s", dsn, pass)
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{
		DB:                   gormDB,
		autoMigrationEnabled: true,
		timeout:              timeout,
	}, nil
}

func NewPosgreDatabase(url string) *PostgreDatabase {
	postgreMutex.Lock()
	defer postgreMutex.Unlock()

	postgreOnce.Do(func() {
		postgreInstance = &PostgreDatabase{
			url: url,
		}
	})

	return postgreInstance
}

func (d *PostgreDatabase) Open() error {
	for range _maxRetries {
		conn, err1 := pgxpool.New(context.Background(), d.url)

		if err1 != nil {
			time.Sleep(5 * time.Second)
		} else {
			d.Conn = conn
			return nil
		}
	}

	return fmt.Errorf("imposible to connect to database after %d retries", _maxRetries)
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
	queryCtx, cancelFn := context.WithTimeout(ctx, _defaultQueryTimeout)
	defer cancelFn()

	rows, err := d.Conn.Query(queryCtx, sql, args)
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
