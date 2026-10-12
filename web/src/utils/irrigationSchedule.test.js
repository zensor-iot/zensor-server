import { describe, expect, it } from 'vitest'
import {
  clampDuration,
  describeSchedule,
  emptyProgram,
  irrigateNowPayload,
  nextRun,
  programToTaskPayload,
  taskDuration,
  taskToProgram,
  toggledTaskPayload,
} from './irrigationSchedule'

const intervalTask = (overrides = {}) => ({
  id: 'task-1',
  is_active: true,
  commands: [
    { index: 1, value: 1, wait_for: '0s' },
    { index: 1, value: 0, wait_for: '20m' },
  ],
  scheduling: { type: 'interval', initial_day: '2026-10-01T00:00:00Z', day_interval: 1, execution_time: '01:00' },
  ...overrides,
})

describe('clampDuration', () => {
  it('keeps values within 1 and 60 minutes', () => {
    expect(clampDuration(0)).toBe(1)
    expect(clampDuration(90)).toBe(60)
    expect(clampDuration(12)).toBe(12)
  })
})

describe('taskDuration', () => {
  it('reads minutes from the deactivation command', () => {
    expect(taskDuration(intervalTask())).toBe(20)
  })

  it('converts seconds and hours to minutes', () => {
    expect(taskDuration({ commands: [{ value: 0, wait_for: '90s' }] })).toBe(2)
    expect(taskDuration({ commands: [{ value: 0, wait_for: '1h' }] })).toBe(60)
    expect(taskDuration({ commands: [{ value: 0, wait_for: 300 }] })).toBe(5)
  })

  it('falls back to the default duration', () => {
    expect(taskDuration({})).toBe(5)
  })
})

describe('describeSchedule', () => {
  it('describes daily interval programs', () => {
    expect(describeSchedule(intervalTask())).toBe('Todos los días · 01:00')
  })

  it('describes multi-day interval programs', () => {
    expect(describeSchedule(intervalTask({ scheduling: { type: 'interval', day_interval: 3, execution_time: '19:30' } }))).toBe(
      'Cada 3 días · 19:30',
    )
  })

  it('describes legacy cron programs', () => {
    expect(describeSchedule({ schedule: '0 30 6 * * 1' })).toBe('Cada lunes · 06:30')
    expect(describeSchedule({ schedule: '15 7 */2 * *' })).toBe('Cada 2 días · 07:15')
  })

  it('flags invalid schedules', () => {
    expect(describeSchedule({ schedule: 'nope' })).toBe('Horario no válido')
    expect(describeSchedule({})).toBe('Sin horario')
  })
})

describe('nextRun', () => {
  it('returns the initial day when it is still ahead', () => {
    const now = new Date(2026, 9, 1, 0, 30)
    expect(nextRun({ initialDay: '2026-10-01', time: '01:00', dayInterval: 1 }, now)).toEqual(new Date(2026, 9, 1, 1, 0))
  })

  it('advances by whole intervals once the initial day is past', () => {
    const now = new Date(2026, 9, 12, 10, 0)
    expect(nextRun({ initialDay: '2026-10-01', time: '01:00', dayInterval: 3 }, now)).toEqual(new Date(2026, 9, 13, 1, 0))
  })
})

describe('taskToProgram', () => {
  it('maps interval tasks to form values', () => {
    expect(taskToProgram(intervalTask({ is_active: false }))).toMatchObject({
      time: '01:00',
      initialDay: '2026-10-01',
      dayInterval: 1,
      duration: 20,
      isActive: false,
    })
  })
})

describe('programToTaskPayload', () => {
  it('builds an interval task that opens and closes the valve', () => {
    expect(
      programToTaskPayload(
        { time: '01:00', initialDay: '2026-10-01', dayInterval: 2, duration: 75, isActive: true },
        new Date(2026, 8, 30, 12, 0),
      ),
    ).toEqual({
      commands: [
        { index: 1, value: 1, priority: 'NORMAL', wait_for: '0s' },
        { index: 1, value: 0, priority: 'NORMAL', wait_for: '60m' },
      ],
      scheduling: { type: 'interval', initial_day: '2026-10-01T00:00:00Z', day_interval: 2, execution_time: '01:00' },
      is_active: true,
    })
  })
})

describe('programToTaskPayload start date', () => {
  it('moves a past start to the next run so the server accepts it', () => {
    const payload = programToTaskPayload(
      { time: '06:00', initialDay: '2026-10-12', dayInterval: 1, duration: 5, isActive: true },
      new Date(2026, 9, 12, 11, 40),
    )
    expect(payload.scheduling.initial_day).toBe('2026-10-13T00:00:00Z')
  })
})

describe('emptyProgram', () => {
  it('starts tomorrow once 06:00 has passed', () => {
    expect(emptyProgram(new Date(2026, 9, 12, 11, 40)).initialDay).toBe('2026-10-13')
    expect(emptyProgram(new Date(2026, 9, 12, 5, 0)).initialDay).toBe('2026-10-12')
  })
})

describe('toggledTaskPayload', () => {
  it('flips the active flag and normalizes numeric waits', () => {
    const payload = toggledTaskPayload(intervalTask({ commands: [{ index: 1, value: 0, wait_for: 1200 }] }))
    expect(payload.is_active).toBe(false)
    expect(payload.commands[0].wait_for).toBe('1200s')
  })
})

describe('irrigateNowPayload', () => {
  it('opens the valve now and closes it after the chosen minutes', () => {
    expect(irrigateNowPayload(5).commands.map(command => [command.value, command.wait_for])).toEqual([
      [1, '0s'],
      [0, '5m'],
    ])
  })
})
