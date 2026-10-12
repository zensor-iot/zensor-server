import { clampDuration } from './irrigationSchedule'

export const COMMAND_STATUS = {
  failed: { label: 'Falló', tone: 'attention' },
  pending: { label: 'Pendiente', tone: 'neutral' },
  sent: { label: 'Enviado', tone: 'ok' },
}

export const EXECUTION_STATUS = {
  failed: { label: 'Con error', tone: 'attention' },
  pending: { label: 'Pendiente', tone: 'neutral' },
  sent: { label: 'Enviada', tone: 'ok' },
}

const FAILED = ['failed', 'error', 'timeout', 'rejected']
const SENT = ['sent', 'acked', 'ack', 'delivered', 'confirmed', 'done', 'completed', 'ok']

export function commandStatus(command) {
  const status = String(command?.status || '').toLowerCase()
  if (command?.failed_at || command?.error_message || FAILED.includes(status)) return 'failed'
  if (command?.acked_at || command?.sent || SENT.includes(status)) return 'sent'
  return 'pending'
}

export function executionStatus(execution) {
  const statuses = (execution?.commands || []).map(commandStatus)
  if (statuses.includes('failed')) return 'failed'
  if (statuses.length > 0 && statuses.every(status => status === 'sent')) return 'sent'
  return 'pending'
}

export function listData(response, legacyKey) {
  if (Array.isArray(response)) return response
  return response?.data ?? response?.[legacyKey] ?? []
}

const DURATION_PATTERN = /^(\d+(\.\d+)?(ns|us|µs|ms|s|m|h))+$/

export function isValidWaitFor(text) {
  return DURATION_PATTERN.test(String(text || '').trim())
}

export function commandSequencePayload({ index, value, priority, waitFor }) {
  const wait = String(waitFor || '').trim() || '0s'
  return {
    priority,
    sequence: [{ index: Number(index), value: Number(value), priority, wait_for: wait }],
  }
}

export function cronTaskPayload({ schedule, duration, isActive }) {
  return {
    schedule: schedule.trim(),
    is_active: isActive,
    commands: [
      { index: 1, value: 1, priority: 'NORMAL', wait_for: '0s' },
      { index: 1, value: 0, priority: 'NORMAL', wait_for: `${clampDuration(duration)}m` },
    ],
  }
}

export function cronOf(task) {
  if (typeof task?.schedule === 'string' && task.schedule) return task.schedule
  if (typeof task?.scheduling?.schedule === 'string') return task.scheduling.schedule
  return ''
}
