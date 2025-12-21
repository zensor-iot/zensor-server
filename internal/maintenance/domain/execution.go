package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type Execution struct {
	ID            shareddomain.ID
	Version       shareddomain.Version
	ActivityID    shareddomain.ID
	ScheduledDate utils.Time
	CompletedAt   *utils.Time
	CompletedBy   *CompletedBy
	OverdueDays   OverdueDays
	FieldValues   map[string]any
	CreatedAt     utils.Time
	UpdatedAt     utils.Time
	DeletedAt     *utils.Time
}

func (me *Execution) IsDeleted() bool {
	return me.DeletedAt != nil
}

func (me *Execution) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	me.DeletedAt = &now
	me.UpdatedAt = now
}

func (me *Execution) MarkCompleted(completedBy string) {
	now := utils.Time{Time: time.Now()}
	me.CompletedAt = &now
	completedByVO := CompletedBy(completedBy)
	me.CompletedBy = &completedByVO
	me.OverdueDays = 0
	me.UpdatedAt = now
}

func (me *Execution) IsCompleted() bool {
	return me.CompletedAt != nil
}

func (me *Execution) IsOverdue() bool {
	if me.CompletedAt != nil {
		return false
	}
	now := time.Now()
	return now.After(me.ScheduledDate.Time)
}

func NewExecutionBuilder() *executionBuilder {
	return &executionBuilder{}
}

type executionBuilder struct {
	actions []executionHandler
}

type executionHandler func(v *Execution) error

func (b *executionBuilder) WithActivityID(value shareddomain.ID) *executionBuilder {
	b.actions = append(b.actions, func(d *Execution) error {
		d.ActivityID = value
		return nil
	})
	return b
}

func (b *executionBuilder) WithScheduledDate(value time.Time) *executionBuilder {
	b.actions = append(b.actions, func(d *Execution) error {
		d.ScheduledDate = utils.Time{Time: value}
		return nil
	})
	return b
}

func (b *executionBuilder) WithFieldValues(value map[string]any) *executionBuilder {
	b.actions = append(b.actions, func(d *Execution) error {
		d.FieldValues = value
		return nil
	})
	return b
}

func (b *executionBuilder) Build() (Execution, error) {
	now := utils.Time{Time: time.Now()}
	result := Execution{
		ID:          shareddomain.ID(utils.GenerateUUID()),
		Version:     1,
		FieldValues: make(map[string]any),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Execution{}, err
		}
	}

	if result.ActivityID == "" {
		return Execution{}, ErrActivityIDRequired
	}

	if result.ScheduledDate.IsZero() {
		return Execution{}, ErrScheduledDateRequired
	}

	return result, nil
}
