import { useEffect, useMemo, useRef, useState } from 'react'
import { Pause, Play, Trash2 } from 'lucide-react'
import useWebSocket from '../hooks/useWebSocket'
import { getWebSocketUrl } from '../config/api'
import { connectionState } from '../utils/connectionState'
import { sensorReadings } from '../utils/sensorReadings'
import { Button, EmptyState, Field, FilterChips, List, Notice, Page, PageHeader, Select, StatusPill } from '../ui'
import { formatNumber, formatRelative } from '../lib/zensor'

const MAX_OPTIONS = [25, 50, 100, 200]
const MAX_CHIP_DEVICES = 5
const NBSP = '\u00a0'

const clock = new Intl.DateTimeFormat('es-CL', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })

function readingText(reading) {
  if (reading.kind === 'relay') return Number(reading.value) === 1 ? 'Abierta' : 'Cerrada'
  const value = typeof reading.value === 'number' ? formatNumber(reading.value, 1) : String(reading.value)
  return reading.unit ? `${value}${NBSP}${reading.unit}` : value
}

function useNow(intervalMs) {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])
  return now
}

function MessageRow({ message, now }) {
  const readings = sensorReadings(message)
  const parsed = new Date(message.timestamp)
  const sentAt = message.timestamp && !Number.isNaN(parsed.getTime()) ? parsed : message.receivedAt
  return (
    <li className="flex flex-col gap-2 rounded-md border border-line bg-surface px-4 py-3">
      <div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
        <span className="min-w-0 break-all text-label">{message.device_id}</span>
        <time
          dateTime={sentAt.toISOString()}
          title={`Recibido ${clock.format(message.receivedAt)}`}
          className="text-caption text-muted"
        >
          <span className="font-mono">{clock.format(sentAt)}</span> · {formatRelative(sentAt, now)}
        </time>
      </div>
      {readings.length > 0 ? (
        <pre className="m-0 overflow-x-auto whitespace-pre font-mono text-sm text-ink">
          {readings.map(reading => `${reading.label}: ${readingText(reading)}`).join('\n')}
        </pre>
      ) : (
        <p className="m-0 text-body-sm text-muted">El mensaje no trae lecturas.</p>
      )}
      <details className="text-body-sm">
        <summary className="inline-flex min-h-touch cursor-pointer items-center text-muted hover:text-ink">
          Ver mensaje completo
        </summary>
        <pre className="m-0 mt-2 overflow-x-auto rounded-sm bg-sunken p-3 font-mono text-xs text-ink">
          {JSON.stringify({ device_id: message.device_id, timestamp: message.timestamp, data: message.data }, null, 2)}
        </pre>
      </details>
    </li>
  )
}

const DeviceMessagesLive = () => {
  const [messages, setMessages] = useState([])
  const [pending, setPending] = useState([])
  const [paused, setPaused] = useState(false)
  const [selectedDevice, setSelectedDevice] = useState('all')
  const [maxMessages, setMaxMessages] = useState(50)
  const maxRef = useRef(maxMessages)
  const pausedRef = useRef(paused)
  const now = useNow(15000)

  const wsUrl = getWebSocketUrl('/ws/device-messages')
  const { isConnected, lastMessage, connectionError, connectionStatus, retry } = useWebSocket(wsUrl)
  const connection = connectionState({ isConnected, connectionStatus })

  useEffect(() => {
    maxRef.current = maxMessages
    pausedRef.current = paused
  }, [maxMessages, paused])

  useEffect(() => {
    if (!lastMessage || lastMessage.type !== 'device_state') return
    const entry = {
      ...lastMessage,
      id: `${lastMessage.device_id}-${lastMessage.timestamp}-${Date.now()}`,
      receivedAt: new Date(),
    }
    const add = prev => [entry, ...prev].slice(0, maxRef.current)
    if (pausedRef.current) setPending(add)
    else setMessages(add)
  }, [lastMessage])

  const deviceCounts = useMemo(() => {
    const counts = new Map()
    messages.forEach(msg => counts.set(msg.device_id, (counts.get(msg.device_id) ?? 0) + 1))
    return [...counts.entries()].sort(([a], [b]) => String(a).localeCompare(String(b)))
  }, [messages])

  const activeDevice = deviceCounts.some(([id]) => id === selectedDevice) ? selectedDevice : 'all'
  const filteredMessages =
    activeDevice === 'all' ? messages : messages.filter(msg => msg.device_id === activeDevice)

  const changeMax = value => {
    setMaxMessages(value)
    setMessages(prev => prev.slice(0, value))
    setPending(prev => prev.slice(0, value))
  }

  const resume = () => {
    setMessages(prev => [...pending, ...prev].slice(0, maxMessages))
    setPending([])
    setPaused(false)
  }

  const clear = () => {
    setMessages([])
    setPending([])
    setSelectedDevice('all')
  }

  const deviceOptions = [
    { value: 'all', label: 'Todos', count: messages.length },
    ...deviceCounts.map(([id, count]) => ({ value: id, label: id, count })),
  ]

  return (
    <Page>
      <PageHeader
        title="Mensajes en vivo"
        description="Las lecturas que envían tus dispositivos, a medida que llegan."
        actions={
          <>
            <StatusPill tone={connection.tone}>{connection.label}</StatusPill>
            {paused ? (
              <Button variant="secondary" icon={Play} onClick={resume}>
                Reanudar
              </Button>
            ) : (
              <Button variant="secondary" icon={Pause} onClick={() => setPaused(true)}>
                Pausar
              </Button>
            )}
            <Button variant="ghost" icon={Trash2} onClick={clear} disabled={messages.length === 0 && pending.length === 0}>
              Limpiar mensajes
            </Button>
          </>
        }
      />

      {connectionError && !isConnected && (
        <Notice action="Reintentar conexión" onAction={retry}>
          Se perdió la conexión y no llegan mensajes nuevos. Revisa tu red y vuelve a intentarlo.
        </Notice>
      )}

      {paused && (
        <Notice tone="neutral" action="Reanudar" onAction={resume}>
          {pending.length === 0
            ? 'En pausa. Los mensajes nuevos esperarán hasta que reanudes.'
            : `En pausa. ${pending.length} ${pending.length === 1 ? 'mensaje nuevo espera' : 'mensajes nuevos esperan'} hasta que reanudes.`}
        </Notice>
      )}

      <div className="flex flex-wrap items-end justify-between gap-4">
        {deviceCounts.length > 0 && deviceCounts.length <= MAX_CHIP_DEVICES ? (
          <FilterChips
            label="Filtrar por dispositivo"
            options={deviceOptions}
            value={activeDevice}
            onChange={setSelectedDevice}
          />
        ) : deviceCounts.length > MAX_CHIP_DEVICES ? (
          <Field label="Dispositivo" className="min-w-[240px]">
            {props => (
              <Select {...props} value={activeDevice} onChange={e => setSelectedDevice(e.target.value)}>
                {deviceOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label} ({option.count})
                  </option>
                ))}
              </Select>
            )}
          </Field>
        ) : (
          <span />
        )}
        <Field label="Mostrar hasta" className="min-w-[160px]">
          {props => (
            <Select {...props} value={maxMessages} onChange={e => changeMax(Number(e.target.value))}>
              {MAX_OPTIONS.map(value => (
                <option key={value} value={value}>
                  {value} mensajes
                </option>
              ))}
            </Select>
          )}
        </Field>
      </div>

      {filteredMessages.length === 0 ? (
        <EmptyState title={isConnected ? 'Esperando mensajes' : 'Sin mensajes'}>
          {isConnected
            ? 'Cada vez que un dispositivo envíe lecturas de humedad, temperatura, caudal o el estado de su válvula, aparecerán aquí con la hora de envío.'
            : 'Cuando se restablezca la conexión, aquí aparecerán las lecturas que envían tus dispositivos.'}
        </EmptyState>
      ) : (
        <List>
          {filteredMessages.map(message => (
            <MessageRow key={message.id} message={message} now={now} />
          ))}
        </List>
      )}
    </Page>
  )
}

export default DeviceMessagesLive
