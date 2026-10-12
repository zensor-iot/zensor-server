import { describe, it, expect } from 'vitest'
import { formatDate } from '../lib/zensor'
import {
  activityStatus,
  activityTypeLabel,
  buildFieldValues,
  buildSchedule,
  describeReminders,
  describeSchedule,
  dueLabel,
  executionStatus,
  formatFieldValue,
  initialFieldValues,
  missingRequiredFields,
  nextPendingExecution,
  toLocalInputValue,
  toSchedulePayload,
  unitLabel,
  upNext,
} from './maintenanceSchedule'

const now = new Date(2026, 9, 12, 10, 0)

describe('describeSchedule', () => {
  it('describes a single-unit schedule', () => {
    const start = '2026-01-15T12:00:00Z'
    expect(describeSchedule({ start_date: start, every: 1, unit: 'month' })).toBe(`Cada mes desde el ${formatDate(start)}`)
  })

  it('pluralizes multi-unit schedules', () => {
    const start = '2026-01-15T12:00:00Z'
    expect(describeSchedule({ start_date: start, every: 3, unit: 'month' })).toBe(`Cada 3 meses desde el ${formatDate(start)}`)
  })

  it('falls back when the schedule is incomplete', () => {
    expect(describeSchedule(null)).toBe('Sin calendario definido')
    expect(describeSchedule({ every: 1 })).toBe('Sin calendario definido')
  })
})

describe('unitLabel', () => {
  it('returns singular and plural forms', () => {
    expect(unitLabel('week', 1)).toBe('semana')
    expect(unitLabel('week', 2)).toBe('semanas')
    expect(unitLabel('unknown', 2)).toBe('unknown')
  })
})

describe('describeReminders', () => {
  it('handles no reminders', () => {
    expect(describeReminders([])).toBe('Sin avisos')
  })

  it('handles a single day', () => {
    expect(describeReminders([1])).toBe('1 día antes')
  })

  it('joins several days', () => {
    expect(describeReminders([7, 1, 3])).toBe('1, 3 y 7 días antes')
  })

  it('handles the same day', () => {
    expect(describeReminders([0])).toBe('El mismo día')
    expect(describeReminders([0, 2])).toBe('El mismo día y 2 días antes')
  })
})

describe('activityTypeLabel', () => {
  it('uses the custom type name for custom activities', () => {
    expect(activityTypeLabel({ type_name: 'custom', custom_type_name: 'Invernadero' })).toBe('Invernadero')
  })

  it('translates built-in types', () => {
    expect(activityTypeLabel({ type_name: 'car' })).toBe('Auto')
  })
})

describe('schedule payloads', () => {
  it('normalizes API dates for comparison', () => {
    expect(toSchedulePayload({ start_date: '2026-01-15T12:00:00Z', every: 1, unit: 'month' })).toEqual({
      start_date: '2026-01-15T12:00:00.000Z',
      every: 1,
      unit: 'month',
    })
  })

  it('round-trips through the local input value without shifting', () => {
    const original = toSchedulePayload({ start_date: '2026-01-15T12:00:00Z', every: 2, unit: 'week' })
    const rebuilt = buildSchedule({ startDate: toLocalInputValue(original.start_date), every: 2, unit: 'week' })
    expect(rebuilt).toEqual(original)
  })

  it('returns null without a start date', () => {
    expect(buildSchedule({ startDate: '', every: 1, unit: 'day' })).toBeNull()
    expect(toLocalInputValue('')).toBe('')
  })
})

describe('dueLabel', () => {
  it('describes today, tomorrow and yesterday', () => {
    expect(dueLabel(new Date(2026, 9, 12, 20, 0), now)).toBe('Vence hoy')
    expect(dueLabel(new Date(2026, 9, 13, 8, 0), now)).toBe('Vence mañana')
    expect(dueLabel(new Date(2026, 9, 11, 8, 0), now)).toBe('Venció ayer')
  })

  it('describes further dates', () => {
    expect(dueLabel(new Date(2026, 9, 17), now)).toBe('Vence en 5 días')
    expect(dueLabel(new Date(2026, 9, 2), now)).toBe('Venció hace 10 días')
  })
})

describe('execution status', () => {
  it('marks completed, overdue, today and pending executions', () => {
    expect(executionStatus({ completed_at: '2026-10-01T00:00:00Z' }, now).key).toBe('done')
    expect(executionStatus({ is_overdue: true, scheduled_date: new Date(2026, 9, 1) }, now)).toMatchObject({ tone: 'attention' })
    expect(executionStatus({ scheduled_date: new Date(2026, 9, 12, 18, 0) }, now).key).toBe('today')
    expect(executionStatus({ scheduled_date: new Date(2026, 9, 20) }, now)).toMatchObject({ key: 'pending', tone: 'neutral' })
  })

  it('reports paused and up-to-date activities', () => {
    expect(activityStatus({ is_active: false }, null, now).key).toBe('paused')
    expect(activityStatus({ is_active: true }, null, now).key).toBe('clear')
    expect(activityStatus({ is_active: true }, { is_overdue: true }, now).key).toBe('overdue')
  })
})

describe('upNext', () => {
  it('picks the earliest pending execution per activity, soonest first', () => {
    const items = [
      {
        activity: { id: 'a' },
        executions: [
          { id: 'a1', scheduled_date: '2026-10-01T00:00:00Z', completed_at: '2026-10-01T05:00:00Z' },
          { id: 'a3', scheduled_date: '2026-12-01T00:00:00Z' },
          { id: 'a2', scheduled_date: '2026-11-01T00:00:00Z' },
        ],
      },
      { activity: { id: 'b' }, executions: [{ id: 'b1', scheduled_date: '2026-10-20T00:00:00Z' }] },
      { activity: { id: 'c' }, executions: [] },
    ]
    expect(upNext(items).map(i => i.execution.id)).toEqual(['b1', 'a2'])
    expect(nextPendingExecution(items[2].executions)).toBeNull()
  })
})

describe('field values', () => {
  const fields = [
    { name: 'filter', type: 'text', is_required: true },
    { name: 'km', type: 'number', is_required: false },
    { name: 'ok', type: 'boolean', is_required: true, default_value: 'true' },
    { name: 'when', type: 'date', is_required: false, default_value: '2026-10-01' },
  ]

  it('builds initial values from defaults', () => {
    expect(initialFieldValues(fields)).toEqual({ filter: '', km: '', ok: true, when: '2026-10-01' })
  })

  it('finds missing required fields, ignoring booleans', () => {
    expect(missingRequiredFields(fields, { filter: ' ', ok: false })).toEqual(['filter'])
  })

  it('builds the payload, parsing numbers and dropping empty values', () => {
    expect(buildFieldValues(fields, { filter: 'HEPA', km: '120.5', ok: false, when: '' })).toEqual({
      filter: 'HEPA',
      km: 120.5,
      ok: false,
    })
  })

  it('formats recorded values for display', () => {
    expect(formatFieldValue('boolean', true)).toBe('Sí')
    expect(formatFieldValue('number', 120.5)).toBe('120,5')
    expect(formatFieldValue('date', '2026-10-01')).toBe(formatDate(new Date(2026, 9, 1)))
    expect(formatFieldValue('text', '')).toBe('—')
  })
})
