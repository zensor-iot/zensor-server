import { formatDateTime as formatDateTimeEsCl, formatTime as formatTimeEsCl } from '../lib/zensor'

const UNIT_LABELS = {
  drop: ['gota', 'gotas'],
  pill: ['pastilla', 'pastillas'],
  spoonful: ['cucharada', 'cucharadas'],
}

export const DOSE_UNITS = [
  { value: 'drop', label: 'Gotas' },
  { value: 'pill', label: 'Pastillas' },
  { value: 'spoonful', label: 'Cucharadas' },
]

export const INTERVAL_UNITS = [
  { value: 'hour', label: 'horas' },
  { value: 'day', label: 'días' },
]

export const PATIENT_KINDS = [
  { value: 'human', label: 'Persona' },
  { value: 'animal', label: 'Animal' },
]

const quantityFormat = new Intl.NumberFormat('es-CL', { maximumFractionDigits: 2 })

const formatQuantity = quantity => {
  const n = Number(quantity)
  return Number.isFinite(n) ? quantityFormat.format(n) : String(quantity)
}

/** Renders a quantity and unit the way a person reads it ("1 pastilla", "0,5 pastilla"). */
export function describeDose(quantity, unit) {
  const labels = UNIT_LABELS[unit]
  const amount = formatQuantity(quantity)
  if (!labels) return `${amount} ${unit}`
  const singular = Math.abs(Number(quantity)) <= 1
  return `${amount} ${singular ? labels[0] : labels[1]}`
}

/** Dose text for anything carrying quantity + unit (a dose or a treatment), falling back to the server's text. */
export function doseTextOf(item, fallback = '') {
  if (!item || item.quantity === undefined || item.quantity === null || !item.unit) return fallback
  return describeDose(item.quantity, item.unit)
}

/** Turns a schedule into "cada 8 horas" or "cada día". */
export function describeFrequency(schedule) {
  if (!schedule || !schedule.every || !schedule.unit) return ''
  const unit = INTERVAL_UNITS.find(u => u.value === schedule.unit)
  if (!unit) return ''
  if (schedule.every === 1) return schedule.unit === 'hour' ? 'cada hora' : 'cada día'
  return `cada ${schedule.every} ${unit.label}`
}

/** Converts a duration in days into the number of doses a schedule produces (the API's end condition). */
export function totalDosesForDuration(days, every, unit) {
  const numericDays = Number(days)
  const numericEvery = Number(every)
  if (!Number.isFinite(numericDays) || !Number.isFinite(numericEvery) || numericDays <= 0 || numericEvery <= 0) {
    return null
  }
  const hours = unit === 'hour' ? numericEvery : numericEvery * 24
  return Math.max(1, Math.floor((numericDays * 24) / hours))
}

/** Lists the first dose times a schedule produces, so the form can preview what it creates. */
export function previewOccurrences(schedule, count = 3) {
  if (!schedule || !schedule.start_at || !schedule.every || !schedule.unit) return []
  const start = new Date(schedule.start_at)
  if (Number.isNaN(start.getTime())) return []

  const occurrences = []
  for (let n = 0; n < count; n++) {
    const next = new Date(start)
    if (schedule.unit === 'hour') next.setHours(next.getHours() + n * schedule.every)
    else next.setDate(next.getDate() + n * schedule.every)
    occurrences.push(next)
  }
  return occurrences
}

const toDate = value => (value instanceof Date ? value : new Date(value))
const isValid = date => !Number.isNaN(date.getTime())

/** es-CL day + 24 h time ("12 oct, 08:00"); empty for an invalid date. */
export function formatDateTime(value) {
  const date = toDate(value)
  return isValid(date) ? formatDateTimeEsCl(date) : ''
}

/** es-CL 24 h time ("08:00"); empty for an invalid date. */
export function formatTime(value) {
  const date = toDate(value)
  return isValid(date) ? formatTimeEsCl(date) : ''
}

/**
 * State of a dose as the family reads it: a pending dose past its time is late,
 * administered is ok, everything else is neutral.
 */
export function doseState(dose) {
  if (dose.status === 'administered') return { key: 'administered', label: 'Administrada', tone: 'ok' }
  if (dose.status === 'skipped') return { key: 'skipped', label: 'Saltada', tone: 'neutral' }
  if (dose.is_overdue) return { key: 'overdue', label: 'Atrasada', tone: 'attention' }
  return { key: 'pending', label: 'Pendiente', tone: 'neutral' }
}

const startOfDay = date => new Date(date.getFullYear(), date.getMonth(), date.getDate())

/** Names a day relative to now: "Hoy", "Mañana", "Ayer", or an es-CL date ("lunes 13 oct"). */
export function dayLabel(value, now = new Date()) {
  const date = toDate(value)
  if (!isValid(date)) return ''
  const diff = Math.round((startOfDay(date) - startOfDay(now)) / 86400000)
  if (diff === 0) return 'Hoy'
  if (diff === 1) return 'Mañana'
  if (diff === -1) return 'Ayer'
  const label = new Intl.DateTimeFormat('es-CL', { weekday: 'long', day: 'numeric', month: 'short' }).format(date)
  return label.charAt(0).toUpperCase() + label.slice(1)
}

/** Groups agenda entries by calendar day, keeping their order. */
export function groupByDay(entries, now = new Date()) {
  const groups = []
  for (const entry of entries) {
    const date = new Date(entry.dose.scheduled_at)
    const key = startOfDay(date).getTime()
    const last = groups[groups.length - 1]
    if (last && last.key === key) last.entries.push(entry)
    else groups.push({ key, label: dayLabel(date, now), entries: [entry] })
  }
  return groups
}
