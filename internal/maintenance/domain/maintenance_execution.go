package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
	sharedDomain "zensor-server/internal/shared_kernel/domain"
)

type MaintenanceExecution struct {
	ID            sharedDomain.ID
	Version       sharedDomain.Version
	ActivityID    sharedDomain.ID
	ScheduledDate utils.Time
	CompletedAt   *utils.Time
	CompletedBy   *CompletedBy
	OverdueDays   OverdueDays
	FieldValues   map[string]any
	CreatedAt     utils.Time
	UpdatedAt     utils.Time
	DeletedAt     *utils.Time
}

func (me *MaintenanceExecution) IsDeleted() bool {
	return me.DeletedAt != nil
}

func (me *MaintenanceExecution) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	me.DeletedAt = &now
	me.UpdatedAt = now
}

func (me *MaintenanceExecution) MarkCompleted(completedBy string) {
	now := utils.Time{Time: time.Now()}
	me.CompletedAt = &now
	completedByVO := CompletedBy(completedBy)
	me.CompletedBy = &completedByVO
	me.OverdueDays = 0
	me.UpdatedAt = now
}

func (me *MaintenanceExecution) IsCompleted() bool {
	return me.CompletedAt != nil
}

func (me *MaintenanceExecution) IsOverdue() bool {
	if me.CompletedAt != nil {
		return false
	}
	now := time.Now()
	return now.After(me.ScheduledDate.Time)
}

func NewMaintenanceExecutionBuilder() *maintenanceExecutionBuilder {
	return &maintenanceExecutionBuilder{}
}

type maintenanceExecutionBuilder struct {
	actions []maintenanceExecutionHandler
}

type maintenanceExecutionHandler func(v *MaintenanceExecution) error

func (b *maintenanceExecutionBuilder) WithActivityID(value sharedDomain.ID) *maintenanceExecutionBuilder {
	b.actions = append(b.actions, func(d *MaintenanceExecution) error {
		d.ActivityID = value
		return nil
	})
	return b
}

func (b *maintenanceExecutionBuilder) WithScheduledDate(value time.Time) *maintenanceExecutionBuilder {
	b.actions = append(b.actions, func(d *MaintenanceExecution) error {
		d.ScheduledDate = utils.Time{Time: value}
		return nil
	})
	return b
}

func (b *maintenanceExecutionBuilder) WithFieldValues(value map[string]any) *maintenanceExecutionBuilder {
	b.actions = append(b.actions, func(d *MaintenanceExecution) error {
		d.FieldValues = value
		return nil
	})
	return b
}

func (b *maintenanceExecutionBuilder) Build() (MaintenanceExecution, error) {
	now := utils.Time{Time: time.Now()}
	result := MaintenanceExecution{
		ID:          sharedDomain.ID(utils.GenerateUUID()),
		Version:     1,
		FieldValues: make(map[string]any),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return MaintenanceExecution{}, err
		}
	}

	if result.ActivityID == "" {
		return MaintenanceExecution{}, ErrActivityIDRequired
	}

	if result.ScheduledDate.IsZero() {
		return MaintenanceExecution{}, ErrScheduledDateRequired
	}

	return result, nil
}
