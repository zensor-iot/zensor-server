import { formatRelative } from '../../lib/zensor'
import { isValveOpen } from '../../utils/sensorReadings'
import { describeSchedule, taskDuration } from '../../utils/irrigationSchedule'

export const deviceName = device => device.display_name || device.name

export const isOnline = device => device.status === 'online'

export function deviceStatus(device, sensorData) {
  const seen = formatRelative(device.lastSeen)
  if (!isOnline(device)) {
    return { tone: 'attention', label: 'Sin conexión', detail: `Sin conexión · ${seen}` }
  }
  if (isValveOpen(sensorData)) {
    return { tone: 'live', label: 'Regando', detail: `En línea · ${seen}` }
  }
  return { tone: 'ok', label: 'En línea', detail: `En línea · ${seen}` }
}

export function programSummary(entry) {
  if (!entry) return null
  const active = entry.tasks.filter(task => task.is_active)
  if (entry.tasks.length === 0) return 'Sin programas'
  if (active.length === 0) return entry.tasks.length === 1 ? '1 programa pausado' : `${entry.tasks.length} programas pausados`
  if (active.length === 1) return `${describeSchedule(active[0])} · ${taskDuration(active[0])}\u00A0min`
  return `${active.length} programas activos`
}
