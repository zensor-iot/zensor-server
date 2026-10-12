import { describe, expect, it } from 'vitest'
import { jsonOrThrow, listOf, taskStatus, totalOf } from './adminData'

describe('listOf', () => {
  it('returns a bare array as is', () => {
    expect(listOf([{ id: 1 }])).toEqual([{ id: 1 }])
  })

  it('unwraps the paginated envelope', () => {
    expect(listOf({ data: [{ id: 1 }], pagination: { total: 1 } })).toEqual([{ id: 1 }])
  })

  it('returns an empty list for anything else', () => {
    expect(listOf(null)).toEqual([])
    expect(listOf({ data: null })).toEqual([])
  })
})

describe('totalOf', () => {
  it('prefers the pagination total over the page length', () => {
    expect(totalOf({ data: [{ id: 1 }], pagination: { total: 42 } })).toBe(42)
  })

  it('falls back to the number of items', () => {
    expect(totalOf([{ id: 1 }, { id: 2 }])).toBe(2)
  })
})

describe('jsonOrThrow', () => {
  it('parses the body of a successful response', async () => {
    const response = { ok: true, status: 200, text: async () => '{"id":"a"}' }
    await expect(jsonOrThrow(response, 'fallo')).resolves.toEqual({ id: 'a' })
  })

  it('returns null for empty bodies', async () => {
    await expect(jsonOrThrow({ ok: true, status: 204 }, 'fallo')).resolves.toBeNull()
    await expect(jsonOrThrow({ ok: true, status: 200, text: async () => '' }, 'fallo')).resolves.toBeNull()
  })

  it('throws the given message when the request failed', async () => {
    await expect(jsonOrThrow({ ok: false, status: 500 }, 'No se pudo')).rejects.toThrow('No se pudo')
  })
})

describe('taskStatus', () => {
  const task = (...statuses) => ({ commands: statuses.map(status => ({ status })) })

  it('flags tasks with any failed command', () => {
    expect(taskStatus(task('ack', 'failed'))).toEqual({ label: 'Con fallas', tone: 'attention' })
  })

  it('confirms tasks whose commands were all acknowledged', () => {
    expect(taskStatus(task('ack', 'ack'))).toEqual({ label: 'Confirmada', tone: 'ok' })
  })

  it('marks tasks in progress while commands are on their way', () => {
    expect(taskStatus(task('ack', 'sent'))).toEqual({ label: 'En curso', tone: 'live' })
    expect(taskStatus(task('queued', 'pending'))).toEqual({ label: 'En curso', tone: 'live' })
  })

  it('keeps tasks pending until a command leaves the server', () => {
    expect(taskStatus(task('pending'))).toEqual({ label: 'Pendiente', tone: 'neutral' })
  })

  it('handles tasks without commands', () => {
    expect(taskStatus({})).toEqual({ label: 'Sin comandos', tone: 'neutral' })
  })
})
