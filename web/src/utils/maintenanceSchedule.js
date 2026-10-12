import { formatDate, formatNumber } from '../lib/zensor'

// Calendar helpers for maintenance activity schedules and executions.
// Schedules are { start_date: ISO8601, every: number, unit: 'day'|'week'|'month'|'quarter'|'year' }.

export const UNITS = [
  { value: 'day', singular: 'día', plural: 'días' },
  { value: 'week', singular: 'semana', plural: 'semanas' },
  { value: 'month', singular: 'mes', plural: 'meses' },
  { value: 'quarter', singular: 'trimestre', plural: 'trimestres' },
  { value: 'year', singular: 'año', plural: 'años' },
]

export const ACTIVITY_TYPES = [
  { value: 'water_system', label: 'Sistema de agua' },
  { value: 'car', label: 'Auto' },
  { value: 'pets', label: 'Mascotas' },
  { value: 'custom', label: 'Otro' },
]

export const FIELD_TYPES = [
  { value: 'text', label: 'Texto' },
  { value: 'number', label: 'Número' },
  { value: 'date', label: 'Fecha' },
  { value: 'boolean', label: 'Sí o no' },
]

const DAY_MS = 86400000

export const capitalize = value => (value ? value.charAt(0).toUpperCase() + value.slice(1) : value)

export function unitLabel(unit, count = 1) {
  const found = UNITS.find(u => u.value === unit)
  if (!found) return unit
  return count === 1 ? found.singular : found.plural
}

export function activityTypeLabel(activity) {
  if (!activity) return ''
  if (activity.type_name === 'custom' && activity.custom_type_name) return activity.custom_type_name
  return ACTIVITY_TYPES.find(t => t.value === activity.type_name)?.label ?? activity.type_name
}

export function describeSchedule(schedule) {
  if (!schedule || typeof schedule !== 'object') return 'Sin calendario definido'
  const { start_date: startDate, every, unit } = schedule
  if (!startDate || !every || !unit) return 'Sin calendario definido'
  const since = `desde el ${formatDate(startDate)}`
  if (every === 1) return `Cada ${unitLabel(unit, 1)} ${since}`
  return `Cada ${every} ${unitLabel(unit, every)} ${since}`
}

export function describeReminders(days) {
  const sorted = [...(days || [])].sort((a, b) => a - b)
  if (sorted.length === 0) return 'Sin avisos'
  const sameDay = sorted.includes(0)
  const before = sorted.filter(d => d > 0)
  const parts = []
  if (sameDay) parts.push('el mismo día')
  if (before.length > 0) {
    const numbers = joinWithY(before.map(String))
    const last = before[before.length - 1]
    parts.push(`${numbers} ${last === 1 && before.length === 1 ? 'día' : 'días'} antes`)
  }
  return capitalize(joinWithY(parts))
}

function joinWithY(items) {
  if (items.length <= 1) return items.join('')
  return `${items.slice(0, -1).join(', ')} y ${items[items.length - 1]}`
}

/**
 * Converts a schedule object to the payload shape the API expects, normalizing the
 * start date to an ISO string so it can be compared with form output.
 */
export function toSchedulePayload(schedule) {
  if (!schedule || typeof schedule !== 'object') return null
  const { start_date: startDate, every, unit } = schedule
  if (!startDate || !every || !unit) return null
  return { start_date: new Date(startDate).toISOString(), every, unit }
}

/**
 * Builds a schedule payload from form state { startDate, every, unit }, where
 * startDate is a local datetime-local value.
 */
export function buildSchedule({ startDate, every, unit }) {
  if (!startDate || !every || !unit) return null
  return toSchedulePayload({ start_date: new Date(startDate).toISOString(), every, unit })
}

/**
 * Formats a date as the local "YYYY-MM-DDTHH:mm" value a datetime-local input expects.
 */
export function toLocalInputValue(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = n => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const startOfDay = d => new Date(d.getFullYear(), d.getMonth(), d.getDate())

export function daysUntil(dateString, now = new Date()) {
  return Math.round((startOfDay(new Date(dateString)) - startOfDay(now)) / DAY_MS)
}

/**
 * Relative due label for an execution's scheduled date ("Vence hoy", "Venció hace 3 días").
 */
export function dueLabel(dateString, now = new Date()) {
  const diff = daysUntil(dateString, now)
  if (diff === 0) return 'Vence hoy'
  if (diff === 1) return 'Vence mañana'
  if (diff === -1) return 'Venció ayer'
  if (diff > 1) return `Vence en ${diff} días`
  return `Venció hace ${-diff} días`
}

export function isDue(execution, now = new Date()) {
  return new Date(execution.scheduled_date) <= now
}

export function nextPendingExecution(executions) {
  const pending = (executions || [])
    .filter(e => !e.completed_at)
    .sort((a, b) => new Date(a.scheduled_date) - new Date(b.scheduled_date))
  return pending[0] ?? null
}

/**
 * Status of a single execution as { key, tone, label } for a StatusPill.
 */
export function executionStatus(execution, now = new Date()) {
  if (execution.completed_at) return { key: 'done', tone: 'ok', label: 'Hecha' }
  if (execution.is_overdue) return { key: 'overdue', tone: 'attention', label: 'Atrasada' }
  if (daysUntil(execution.scheduled_date, now) <= 0) return { key: 'today', tone: 'live', label: 'Toca hoy' }
  return { key: 'pending', tone: 'neutral', label: 'Pendiente' }
}

/**
 * Status of an activity as a whole, driven by its next pending execution.
 */
export function activityStatus(activity, next, now = new Date()) {
  if (!activity.is_active) return { key: 'paused', tone: 'neutral', label: 'En pausa' }
  if (!next) return { key: 'clear', tone: 'ok', label: 'Al día' }
  return executionStatus(next, now)
}

export function sortByScheduledDesc(executions) {
  return [...(executions || [])].sort((a, b) => new Date(b.scheduled_date) - new Date(a.scheduled_date))
}

/**
 * Picks the earliest pending execution of every activity, soonest first.
 */
export function upNext(items) {
  return items
    .map(({ activity, executions }) => {
      const execution = nextPendingExecution(executions)
      return execution ? { activity, execution } : null
    })
    .filter(Boolean)
    .sort((a, b) => new Date(a.execution.scheduled_date) - new Date(b.execution.scheduled_date))
}

export function initialFieldValues(fields) {
  const initial = {}
  ;(fields || []).forEach(f => {
    if (f.default_value != null) initial[f.name] = f.type === 'boolean' ? f.default_value === 'true' : f.default_value
    else initial[f.name] = f.type === 'boolean' ? false : ''
  })
  return initial
}

export function missingRequiredFields(fields, values) {
  return (fields || [])
    .filter(f => f.is_required && f.type !== 'boolean' && String(values[f.name] ?? '').trim() === '')
    .map(f => f.name)
}

/**
 * Converts form values into the field_values payload: booleans always sent,
 * empty values dropped, numbers parsed.
 */
export function buildFieldValues(fields, values) {
  const result = {}
  ;(fields || []).forEach(f => {
    const raw = values[f.name]
    if (f.type === 'boolean') {
      result[f.name] = Boolean(raw)
      return
    }
    if (String(raw ?? '').trim() === '') return
    result[f.name] = f.type === 'number' ? Number(raw) : raw
  })
  return result
}

export function formatFieldValue(type, value) {
  if (value === null || value === undefined || value === '') return '—'
  if (type === 'boolean' || typeof value === 'boolean') return value === true || value === 'true' ? 'Sí' : 'No'
  if (type === 'number' || typeof value === 'number') {
    const n = Number(value)
    if (Number.isNaN(n)) return String(value)
    const decimals = Math.min(2, (String(value).split('.')[1] || '').length)
    return formatNumber(n, decimals)
  }
  if (type === 'date') {
    const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(String(value))
    if (match) return formatDate(new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3])))
  }
  return String(value)
}
