/**
 * Maps the useWebSocket connection fields to the tone and Spanish label of a StatusPill.
 */
export function connectionState({ isConnected, connectionStatus }) {
  if (isConnected) return { tone: 'ok', label: 'En línea' }
  if (/connecting/i.test(connectionStatus ?? '')) return { tone: 'neutral', label: 'Conectando…' }
  return { tone: 'attention', label: 'Sin conexión' }
}
