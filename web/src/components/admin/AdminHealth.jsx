import { useCallback, useEffect, useRef, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button, KeyValue, Notice, Page, PageHeader, Section, StatusPill, Table } from '../../ui'
import { formatTime, formatUnit } from '../../lib/zensor'

const TIMEOUT_MS = 5000

const COMPONENTS = [
  { key: 'api', label: 'Servidor' },
  { key: 'database', label: 'Base de datos' },
  { key: 'websocket', label: 'Mensajes en vivo' },
]

const STATUS = {
  checking: { label: 'Revisando', tone: 'neutral' },
  healthy: { label: 'Operativo', tone: 'ok' },
  unhealthy: { label: 'Sin respuesta', tone: 'attention' },
}

const initialHealth = () =>
  Object.fromEntries(COMPONENTS.map(({ key }) => [key, { status: 'checking', responseTime: null, lastCheck: null }]))

class CheckError extends Error {}

const timeoutSignal = () => (typeof AbortSignal !== 'undefined' && AbortSignal.timeout ? AbortSignal.timeout(TIMEOUT_MS) : undefined)

const describeError = error => {
  if (error instanceof CheckError) return error.message
  if (error?.name === 'TimeoutError' || error?.name === 'AbortError') return `No respondió en ${formatUnit(TIMEOUT_MS / 1000, 's')}`
  return 'No se pudo conectar'
}

async function checkHttp(path, parseBody) {
  const response = await fetch(path, { signal: timeoutSignal() })
  if (!response.ok) throw new CheckError(`Respondió con error ${response.status}`)
  if (!parseBody) return undefined
  try {
    return await response.json()
  } catch {
    throw new CheckError('Respondió con un formato inesperado')
  }
}

function checkWebSocket() {
  return new Promise((resolve, reject) => {
    const url = `${window.location.origin.replace(/^http/, 'ws')}/ws/device-messages`
    let ws
    try {
      ws = new WebSocket(url)
    } catch {
      reject(new CheckError('No se pudo abrir la conexión'))
      return
    }
    const timer = setTimeout(() => {
      ws.close()
      reject(new CheckError(`No respondió en ${formatUnit(TIMEOUT_MS / 1000, 's')}`))
    }, TIMEOUT_MS)
    ws.onopen = () => {
      clearTimeout(timer)
      ws.close()
      resolve()
    }
    ws.onerror = () => {
      clearTimeout(timer)
      reject(new CheckError('No se pudo abrir la conexión'))
    }
  })
}

const CHECKS = {
  api: () => checkHttp('/healthz', true),
  database: () => checkHttp('/v1/tenants', false),
  websocket: checkWebSocket,
}

export default function AdminHealth() {
  const [health, setHealth] = useState(initialHealth)
  const [checking, setChecking] = useState(false)
  const [lastRefresh, setLastRefresh] = useState(null)
  const mounted = useRef(true)

  const runChecks = useCallback(async () => {
    setChecking(true)
    setHealth(current =>
      Object.fromEntries(Object.entries(current).map(([key, value]) => [key, { ...value, status: 'checking' }])),
    )

    await Promise.all(
      COMPONENTS.map(async ({ key }) => {
        const started = Date.now()
        let result
        try {
          const details = await CHECKS[key]()
          result = { status: 'healthy', details }
        } catch (error) {
          result = { status: 'unhealthy', error: describeError(error) }
        }
        if (!mounted.current) return
        setHealth(current => ({
          ...current,
          [key]: { ...result, responseTime: Date.now() - started, lastCheck: new Date() },
        }))
      }),
    )

    if (!mounted.current) return
    setLastRefresh(new Date())
    setChecking(false)
  }, [])

  useEffect(() => {
    mounted.current = true
    runChecks()
    return () => {
      mounted.current = false
    }
  }, [runChecks])

  const down = COMPONENTS.filter(({ key }) => health[key].status === 'unhealthy')
  const apiDetails = health.api.details

  return (
    <Page>
      <PageHeader
        title="Salud del sistema"
        description={
          lastRefresh
            ? `Estado de los componentes del servidor. Última revisión a las ${formatTime(lastRefresh)}.`
            : 'Estado de los componentes del servidor.'
        }
        back={{ to: '/admin', label: 'Administración' }}
        actions={
          <Button variant="secondary" icon={RefreshCw} onClick={runChecks} disabled={checking}>
            {checking ? 'Revisando…' : 'Revisar de nuevo'}
          </Button>
        }
      />

      {!checking && down.length > 0 && (
        <Notice>
          {down.length === 1
            ? `${down[0].label} no responde. Revisa el servidor y vuelve a revisar.`
            : `${down.map(({ label }) => label).join(', ')} no responden. Revisa el servidor y vuelve a revisar.`}
        </Notice>
      )}

      <Table caption="Estado de los componentes">
        <thead>
          <tr>
            <th scope="col">Componente</th>
            <th scope="col">Estado</th>
            <th scope="col" className="text-right">Respuesta</th>
            <th scope="col">Revisado</th>
            <th scope="col">Detalle</th>
          </tr>
        </thead>
        <tbody>
          {COMPONENTS.map(({ key, label }) => {
            const item = health[key]
            const status = STATUS[item.status]
            return (
              <tr key={key}>
                <th scope="row" className="font-medium text-ink">
                  {label}
                </th>
                <td>
                  <StatusPill tone={status.tone}>{status.label}</StatusPill>
                </td>
                <td className="zs-num">
                  {item.responseTime === null || item.status === 'checking'
                    ? '—'
                    : formatUnit(item.responseTime, 'ms')}
                </td>
                <td className="whitespace-nowrap font-mono">{item.lastCheck ? formatTime(item.lastCheck) : '—'}</td>
                <td className={item.error ? 'text-attention' : 'text-muted'}>{item.error || '—'}</td>
              </tr>
            )
          })}
        </tbody>
      </Table>

      {apiDetails && (apiDetails.version || apiDetails.commit_hash) && (
        <Section title="Servidor">
          <div className="rounded-xl border border-line bg-surface p-5">
            <KeyValue
              items={[
                { label: 'Versión', value: apiDetails.version, mono: true },
                { label: 'Commit', value: apiDetails.commit_hash, mono: true },
              ]}
            />
          </div>
        </Section>
      )}
    </Page>
  )
}
