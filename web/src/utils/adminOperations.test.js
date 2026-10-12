import { describe, expect, it } from 'vitest'
import {
  commandSequencePayload,
  commandStatus,
  cronOf,
  cronTaskPayload,
  executionStatus,
  isValidWaitFor,
  listData,
} from './adminOperations'

describe('commandStatus', () => {
  it('treats an error message as failed', () => {
    expect(commandStatus({ sent: true, error_message: 'timeout' })).toBe('failed')
  })

  it('treats a sent command as sent', () => {
    expect(commandStatus({ sent: true })).toBe('sent')
  })

  it('treats an unsent command as pending', () => {
    expect(commandStatus({ sent: false, ready: true })).toBe('pending')
  })
})

describe('executionStatus', () => {
  it('is failed when any command failed', () => {
    expect(executionStatus({ commands: [{ sent: true }, { status: 'failed' }] })).toBe('failed')
  })

  it('is sent when every command was sent', () => {
    expect(executionStatus({ commands: [{ sent: true }, { sent: true }] })).toBe('sent')
  })

  it('is pending when some command is waiting', () => {
    expect(executionStatus({ commands: [{ sent: true }, { sent: false }] })).toBe('pending')
  })

  it('is pending without commands', () => {
    expect(executionStatus({})).toBe('pending')
  })
})

describe('listData', () => {
  it('reads paginated data', () => {
    expect(listData({ data: [1], pagination: {} }, 'tasks')).toEqual([1])
  })

  it('falls back to the legacy key', () => {
    expect(listData({ tasks: [2] }, 'tasks')).toEqual([2])
  })

  it('returns an empty list for missing responses', () => {
    expect(listData(null, 'tasks')).toEqual([])
  })
})

describe('isValidWaitFor', () => {
  it('accepts Go durations', () => {
    expect(isValidWaitFor('0s')).toBe(true)
    expect(isValidWaitFor('1h30m')).toBe(true)
  })

  it('rejects free text', () => {
    expect(isValidWaitFor('5 minutos')).toBe(false)
    expect(isValidWaitFor('')).toBe(false)
  })
})

describe('commandSequencePayload', () => {
  it('wraps a single command in a sequence', () => {
    expect(commandSequencePayload({ index: '2', value: '1', priority: 'HIGH', waitFor: ' 5m ' })).toEqual({
      priority: 'HIGH',
      sequence: [{ index: 2, value: 1, priority: 'HIGH', wait_for: '5m' }],
    })
  })
})

describe('cronTaskPayload', () => {
  it('opens and closes the valve after the duration', () => {
    expect(cronTaskPayload({ schedule: ' 0 0 6 * * * ', duration: 90, isActive: false })).toEqual({
      schedule: '0 0 6 * * *',
      is_active: false,
      commands: [
        { index: 1, value: 1, priority: 'NORMAL', wait_for: '0s' },
        { index: 1, value: 0, priority: 'NORMAL', wait_for: '60m' },
      ],
    })
  })
})

describe('cronOf', () => {
  it('reads the legacy schedule first', () => {
    expect(cronOf({ schedule: '0 0 6 * * *' })).toBe('0 0 6 * * *')
  })

  it('reads a cron scheduling configuration', () => {
    expect(cronOf({ scheduling: { type: 'cron', schedule: '0 30 7 * * *' } })).toBe('0 30 7 * * *')
  })

  it('is empty for interval programs', () => {
    expect(cronOf({ scheduling: { type: 'interval', day_interval: 2 } })).toBe('')
  })
})
