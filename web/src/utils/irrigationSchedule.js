export const MIN_DURATION = 1
export const MAX_DURATION = 60
export const DEFAULT_DURATION = 5

const WEEKDAYS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado']
const DAY_MS = 24 * 60 * 60 * 1000

export const clampDuration = minutes => {
  const value = Number(minutes)
  if (!Number.isFinite(value)) return DEFAULT_DURATION
  return Math.max(MIN_DURATION, Math.min(MAX_DURATION, Math.round(value)))
}

const todayIso = (now = new Date()) => {
  const local = new Date(now.getTime() - now.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 10)
}

export const emptyProgram = (now = new Date()) => ({
  time: '06:00',
  initialDay: now.getHours() < 6 ? todayIso(now) : todayIso(new Date(now.getTime() + DAY_MS)),
  dayInterval: 1,
  duration: DEFAULT_DURATION,
  isActive: true,
})

export function taskDuration(task) {
  const off = task?.commands?.find(command => command.value === 0)
  if (!off?.wait_for) return DEFAULT_DURATION
  const waitFor = typeof off.wait_for === 'number' ? `${off.wait_for}s` : String(off.wait_for)
  const hours = waitFor.match(/(\d+)h/)
  const minutes = waitFor.match(/(\d+)m(?!s)/)
  const seconds = waitFor.match(/(\d+)s/)
  if (minutes) return clampDuration(Number(minutes[1]))
  if (hours) return clampDuration(Number(hours[1]) * 60)
  if (seconds) return clampDuration(Math.ceil(Number(seconds[1]) / 60))
  return DEFAULT_DURATION
}

const parseCron = expression => {
  const parts = String(expression || '').trim().split(/\s+/)
  if (parts.length !== 5 && parts.length !== 6) return null
  const [minute, hour, day, , weekday] = parts.length === 6 ? parts.slice(1) : parts
  if (!/^\d+$/.test(minute) || !/^\d+$/.test(hour)) return null
  return {
    time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`,
    weekday: weekday !== '*' ? Number(weekday) % 7 : null,
    dayInterval: day.startsWith('*/') ? Number(day.slice(2)) : 1,
  }
}

const schedulingOf = task => task?.scheduling || task?.schedule || null

export function describeSchedule(task) {
  const scheduling = schedulingOf(task)
  if (!scheduling) return 'Sin horario'

  if (typeof scheduling === 'object' && scheduling.type === 'interval') {
    const time = scheduling.execution_time || '06:00'
    const interval = scheduling.day_interval || 1
    return interval === 1 ? `Todos los días · ${time}` : `Cada ${interval} días · ${time}`
  }

  const cron = parseCron(typeof scheduling === 'string' ? scheduling : scheduling.schedule)
  if (!cron) return 'Horario no válido'
  if (cron.weekday !== null) return `Cada ${WEEKDAYS[cron.weekday] ?? 'semana'} · ${cron.time}`
  if (cron.dayInterval > 1) return `Cada ${cron.dayInterval} días · ${cron.time}`
  return `Todos los días · ${cron.time}`
}

export function nextRun(program, now = new Date()) {
  if (!program?.initialDay || !program?.time) return null
  const [year, month, day] = program.initialDay.split('-').map(Number)
  const [hours, minutes] = program.time.split(':').map(Number)
  if ([year, month, day, hours, minutes].some(Number.isNaN)) return null
  const interval = Math.max(1, Number(program.dayInterval) || 1)
  const next = new Date(year, month - 1, day, hours, minutes, 0, 0)
  if (next <= now) {
    const periods = Math.floor((now - next) / (interval * DAY_MS)) + 1
    next.setDate(next.getDate() + periods * interval)
    while (next <= now) next.setDate(next.getDate() + interval)
  }
  return next
}

export function taskToProgram(task, now = new Date()) {
  const scheduling = schedulingOf(task)
  const base = { ...emptyProgram(now), duration: taskDuration(task), isActive: task?.is_active ?? true }
  if (scheduling && typeof scheduling === 'object' && scheduling.type === 'interval') {
    return {
      ...base,
      time: scheduling.execution_time || base.time,
      initialDay: scheduling.initial_day ? scheduling.initial_day.slice(0, 10) : base.initialDay,
      dayInterval: scheduling.day_interval || 1,
    }
  }
  const cron = parseCron(typeof scheduling === 'string' ? scheduling : scheduling?.schedule)
  return cron ? { ...base, time: cron.time, dayInterval: cron.weekday !== null ? 7 : cron.dayInterval } : base
}

export function taskNextRun(task, now = new Date()) {
  const scheduling = schedulingOf(task)
  if (!scheduling || typeof scheduling !== 'object' || scheduling.type !== 'interval') return null
  return nextRun(taskToProgram(task, now), now)
}

const isoDate = date =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

export function programToTaskPayload(program, now = new Date()) {
  const duration = clampDuration(program.duration)
  const first = nextRun(program, now)
  const initialDay = first ? isoDate(first) : program.initialDay
  return {
    commands: [
      { index: 1, value: 1, priority: 'NORMAL', wait_for: '0s' },
      { index: 1, value: 0, priority: 'NORMAL', wait_for: `${duration}m` },
    ],
    scheduling: {
      type: 'interval',
      initial_day: `${initialDay}T00:00:00Z`,
      day_interval: Math.max(1, Number(program.dayInterval) || 1),
      execution_time: program.time,
    },
    is_active: program.isActive,
  }
}

export function toggledTaskPayload(task) {
  return {
    commands: (task.commands || []).map(command => ({
      ...command,
      wait_for: typeof command.wait_for === 'number' ? `${command.wait_for}s` : String(command.wait_for),
    })),
    ...(task.scheduling ? { scheduling: task.scheduling } : { schedule: task.schedule }),
    is_active: !task.is_active,
  }
}

export function irrigateNowPayload(minutes) {
  const duration = clampDuration(minutes)
  return {
    commands: [
      { wait_for: '0s', priority: 'HIGH', index: 1, value: 1 },
      { wait_for: `${duration}m`, priority: 'HIGH', index: 1, value: 0 },
    ],
  }
}
