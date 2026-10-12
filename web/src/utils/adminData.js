/**
 * Returns the items of a list response, accepting both a bare array and the
 * paginated `{ data, pagination }` envelope.
 */
export function listOf(body) {
  if (Array.isArray(body)) return body
  return Array.isArray(body?.data) ? body.data : []
}

/**
 * Returns the total number of items of a list response, preferring the
 * server-side pagination total over the length of the current page.
 */
export function totalOf(body) {
  const total = body?.pagination?.total
  if (Number.isFinite(total)) return total
  return listOf(body).length
}

/**
 * Reads a JSON response and throws an Error with the given message when the
 * request failed, so callers can surface a readable reason.
 */
export async function jsonOrThrow(response, message) {
  if (!response.ok) throw new Error(message)
  if (response.status === 204) return null
  const text = await response.text()
  return text ? JSON.parse(text) : null
}

/**
 * Summarizes the delivery state of a task from the status of its commands.
 * Returns a label and a StatusPill tone.
 */
export function taskStatus(task) {
  const statuses = (task?.commands || []).map(command => command.status)
  if (statuses.length === 0) return { label: 'Sin comandos', tone: 'neutral' }
  if (statuses.includes('failed')) return { label: 'Con fallas', tone: 'attention' }
  if (statuses.every(status => status === 'ack')) return { label: 'Confirmada', tone: 'ok' }
  if (statuses.some(status => status === 'sent' || status === 'queued' || status === 'ack')) {
    return { label: 'En curso', tone: 'live' }
  }
  return { label: 'Pendiente', tone: 'neutral' }
}
