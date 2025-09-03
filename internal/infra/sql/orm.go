package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

//go:generate mockgen -source=orm.go -destination=../../../test/unit/doubles/infra/sql/orm_mock.go -package=sql -mock_names=ORM=MockORM

type ORM interface {
	AutoMigrate(dst ...any) error
	Count(count *int64) ORM
	Create(value any) ORM
	Delete(value any, conds ...any) ORM
	Find(dest any, conds ...any) ORM
	First(dest any, conds ...any) ORM
	Limit(limit int) ORM
	Model(value any) ORM
	Offset(offset int) ORM
	Order(value any) ORM
	Preload(query string, args ...any) ORM
	Save(value any) ORM
	Transaction(fc func(tx ORM) error, opts ...*sql.TxOptions) error
	Unscoped() ORM
	Where(query any, args ...any) ORM
	WithContext(ctx context.Context) ORM
	WithTimeout(ctx context.Context, timeout time.Duration) ORM
	Joins(value string, args ...any) ORM
	InnerJoins(value string, args ...any) ORM

	Error() error
}

type DB struct {
	*gorm.DB
	autoMigrationEnabled bool
	timeout              time.Duration
}

var (
	ErrRecordNotFound = errors.New("record not found")
)

func (d DB) Error() error {
	switch {
	case errors.Is(d.DB.Error, gorm.ErrRecordNotFound):
		return ErrRecordNotFound
	case d.DB.Error != nil:
		return fmt.Errorf("database error: %w", d.DB.Error)
	default:
		return nil
	}
}

var _ ORM = (*DB)(nil)

func (d DB) AutoMigrate(dst ...any) error {
	if d.autoMigrationEnabled {
		return d.DB.AutoMigrate(dst...)
	}

	return nil
}

func (d DB) Count(value *int64) ORM {
	tx := d.DB.Count(value)
	d.DB = tx
	return &d
}

func (d DB) Create(value any) ORM {
	d.setSpanAttributes("create")
	tx := d.DB.Create(value)
	d.DB = tx
	return &d
}

func (d DB) Delete(value any, conds ...any) ORM {
	tx := d.DB.Delete(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) Find(value any, conds ...any) ORM {
	d.setSpanAttributes("find")
	tx := d.DB.Find(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) First(value any, conds ...any) ORM {
	d.setSpanAttributes("first")
	tx := d.DB.First(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) Limit(value int) ORM {
	tx := d.DB.Limit(value)
	d.DB = tx
	return &d
}

func (d DB) Model(value any) ORM {
	tx := d.DB.Model(value)
	d.DB = tx
	return &d
}

func (d DB) Offset(value int) ORM {
	tx := d.DB.Offset(value)
	d.DB = tx
	return &d
}

func (d DB) Order(value any) ORM {
	tx := d.DB.Order(value)
	d.DB = tx
	return &d
}

func (d DB) Preload(value string, conds ...any) ORM {
	tx := d.DB.Preload(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) Save(value any) ORM {
	d.setSpanAttributes("save")
	tx := d.DB.Save(value)
	d.DB = tx
	return &d
}

func (d DB) Unscoped() ORM {
	tx := d.DB.Unscoped()
	d.DB = tx
	return &d
}

func (d DB) Where(value any, conds ...any) ORM {
	tx := d.DB.Where(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) WithContext(value context.Context) ORM {
	if d.timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(value, d.timeout)
		// Store the cancel function to be called when the context is done
		// This prevents context leaks while allowing the timeout to work properly
		go func() {
			<-timeoutCtx.Done()
			cancel()
		}()
		tx := d.DB.WithContext(timeoutCtx)
		d.DB = tx
		return &d
	}

	tx := d.DB.WithContext(value)
	d.DB = tx
	return &d
}

func (d DB) WithTimeout(ctx context.Context, timeout time.Duration) ORM {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	// Store the cancel function to be called when the context is done
	// This prevents context leaks while allowing the timeout to work properly
	go func() {
		<-timeoutCtx.Done()
		cancel()
	}()
	tx := d.DB.WithContext(timeoutCtx)
	d.DB = tx
	return &d
}

func (d DB) Transaction(f func(ORM) error, opts ...*sql.TxOptions) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return f(&DB{tx, d.autoMigrationEnabled, d.timeout})
	}, opts...)
}

func (d DB) Joins(value string, conds ...any) ORM {
	tx := d.DB.Joins(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) InnerJoins(value string, conds ...any) ORM {
	tx := d.DB.InnerJoins(value, conds...)
	d.DB = tx
	return &d
}

// setSpanAttributes sets OpenTelemetry span attributes for database operations
func (d DB) setSpanAttributes(operation string) {
	if ctx := d.DB.Statement.Context; ctx != nil {
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetAttributes(
				attribute.String("span.kind", "client"),
				attribute.String("component", "database"),
				attribute.String("db.system", "postgresql"),
				attribute.String("db.operation", operation),
			)
		}
	}
}
