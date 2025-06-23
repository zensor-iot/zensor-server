package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

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
	Preload(query string, args ...any) ORM
	Save(value any) ORM
	Transaction(fc func(tx ORM) error, opts ...*sql.TxOptions) error
	Unscoped() ORM
	Where(query any, args ...any) ORM
	WithContext(ctx context.Context) ORM
	Joins(value string, args ...any) ORM
	InnerJoins(value string, args ...any) ORM

	Error() error
}

type DB struct {
	*gorm.DB
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

const (
	_connectionFormat       = "%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true"
	_defaultStringSize uint = 256
)

var (
	_defaultDatetimePrecision int = 2
)

var _ ORM = (*DB)(nil)

func (d DB) Count(value *int64) ORM {
	tx := d.DB.Count(value)
	d.DB = tx
	return &d
}

func (d DB) Create(value any) ORM {
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
	tx := d.DB.Find(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) First(value any, conds ...any) ORM {
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

func (d DB) Preload(value string, conds ...any) ORM {
	tx := d.DB.Preload(value, conds...)
	d.DB = tx
	return &d
}

func (d DB) Save(value any) ORM {
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
	tx := d.DB.WithContext(value)
	d.DB = tx
	return &d
}

func (d DB) Transaction(f func(ORM) error, opts ...*sql.TxOptions) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return f(&DB{tx})
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
