import { useCallback, useEffect, useState } from 'react'
import { Send } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import { deviceCommandsApi } from '../../config/api'
import {
  Button,
  Drawer,
  EmptyState,
  ErrorState,
  Field,
  FilterChips,
  Form,
  Input,
  KeyValue,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  Section,
  Select,
  StatusPill,
  Table,
} from '../../ui'
import { formatDateTime, formatNumber, formatRelative } from '../../lib/zensor'
import { commandSequencePayload, isValidWaitFor, listData } from '../../utils/adminOperations'

const DEVICE_LIMIT = 100
const EMPTY_FORM = { deviceId: '', index: 1, value: 0, priority: 'NORMAL', waitFor: '0s' }
const PRIORITIES = [
  { value: 'LOW', label: 'Baja' },
  { value: 'NORMAL', label: 'Normal' },
  { value: 'HIGH', label: 'Alta' },
  { value: 'URGENT', label: 'Urgente' },
]
const LOG_STATUS = {
  queued: { label: 'En cola', tone: 'live' },
  failed: { label: 'Falló', tone: 'attention' },
}

const nameOf = device => device?.display_name || device?.name || device?.id || ''
const priorityLabel = value => PRIORITIES.find(priority => priority.value === value)?.label || value

export default function AdminCommands() {
  const { showSuccess, showError } = useNotification()
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [formOpen, setFormOpen] = useState(false)
  const [form, setForm] = useState(EMPTY_FORM)
  const [formError, setFormError] = useState(null)
  const [sending, setSending] = useState(false)
  const [log, setLog] = useState([])
  const [filter, setFilter] = useState('all')
  const [selected, setSelected] = useState(null)

  const fetchDevices = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch(`/v1/devices?limit=${DEVICE_LIMIT}`)
      if (!response.ok) throw new Error(String(response.status))
      setDevices(listData(await response.json(), 'devices'))
    } catch {
      setError('Revisa tu conexión y vuelve a intentarlo.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchDevices()
  }, [fetchDevices])

  const openForm = () => {
    setForm(EMPTY_FORM)
    setFormError(null)
    setFormOpen(true)
  }

  const handleSend = async event => {
    event.preventDefault()
    if (!form.deviceId) {
      setFormError({ deviceId: 'Elige el dispositivo que recibirá el comando.' })
      return
    }
    if (!isValidWaitFor(form.waitFor)) {
      setFormError({ waitFor: 'Usa un formato como 0s, 30s, 5m o 1h.' })
      return
    }
    setFormError(null)

    const device = devices.find(item => item.id === form.deviceId)
    const payload = commandSequencePayload(form)
    const entry = {
      id: `${Date.now()}-${log.length}`,
      deviceId: form.deviceId,
      deviceName: nameOf(device),
      createdAt: new Date().toISOString(),
      payload,
    }

    setSending(true)
    try {
      await deviceCommandsApi.sendCommand(form.deviceId, payload)
      setLog(current => [{ ...entry, status: 'queued' }, ...current])
      showSuccess(`Comando en cola para ${entry.deviceName}`)
      setFormOpen(false)
    } catch (sendError) {
      setLog(current => [{ ...entry, status: 'failed', error: sendError.message }, ...current])
      showError(`No se pudo enviar el comando a ${entry.deviceName}. Revisa los datos y vuelve a intentarlo.`)
    } finally {
      setSending(false)
    }
  }

  const failedCount = log.filter(entry => entry.status === 'failed').length
  const queuedCount = log.length - failedCount
  const visibleLog = filter === 'all' ? log : log.filter(entry => entry.status === filter)

  return (
    <Page>
      <PageHeader
        title="Comandos"
        description="Envía un comando a cualquier dispositivo del sistema. Úsalo para pruebas y soporte."
        back={{ to: '/admin', label: 'Administración' }}
        actions={
          <Button icon={Send} onClick={openForm} disabled={loading || devices.length === 0}>
            Enviar comando
          </Button>
        }
      />

      {loading && devices.length === 0 ? (
        <LoadingState>Cargando dispositivos…</LoadingState>
      ) : error ? (
        <ErrorState title="No se pudieron cargar los dispositivos" onRetry={fetchDevices}>
          {error}
        </ErrorState>
      ) : devices.length === 0 ? (
        <EmptyState title="Aún no hay dispositivos">
          Registra o adopta un dispositivo para poder enviarle comandos.
        </EmptyState>
      ) : (
        <Section
          title="Enviados en esta sesión"
          description="El servidor aún no guarda un historial de comandos: aquí ves solo los que enviaste desde esta página, hasta que la recargues."
        >
          {log.length === 0 ? (
            <EmptyState>Envía un comando con “Enviar comando” y verás aquí su resultado.</EmptyState>
          ) : (
            <>
              {failedCount > 0 && filter !== 'failed' && (
                <Notice action="Ver" onAction={() => setFilter('failed')}>
                  {failedCount === 1
                    ? '1 comando no llegó a la cola.'
                    : `${formatNumber(failedCount)} comandos no llegaron a la cola.`}
                </Notice>
              )}
              <FilterChips
                label="Filtro de comandos"
                value={filter}
                onChange={setFilter}
                options={[
                  { value: 'all', label: 'Todos', count: log.length },
                  { value: 'queued', label: 'En cola', count: queuedCount },
                  { value: 'failed', label: 'Fallidos', count: failedCount },
                ]}
              />
              {visibleLog.length === 0 ? (
                <EmptyState>No hay comandos con este estado. Elige “Todos” para ver la lista completa.</EmptyState>
              ) : (
                <Table caption="Comandos enviados en esta sesión">
                  <thead>
                    <tr>
                      <th scope="col">Enviado</th>
                      <th scope="col">Dispositivo</th>
                      <th scope="col" className="text-right">Relé</th>
                      <th scope="col" className="text-right">Valor</th>
                      <th scope="col">Estado</th>
                    </tr>
                  </thead>
                  <tbody>
                    {visibleLog.map(entry => {
                      const [command] = entry.payload.sequence
                      const status = LOG_STATUS[entry.status]
                      return (
                        <tr key={entry.id}>
                          <td>
                            <button
                              type="button"
                              onClick={() => setSelected(entry)}
                              className="flex min-h-touch flex-col items-start justify-center bg-transparent p-0 text-left text-ink underline-offset-4 hover:underline"
                            >
                              <span className="font-mono">{formatDateTime(entry.createdAt)}</span>
                              <span className="text-caption text-muted">{formatRelative(entry.createdAt)}</span>
                            </button>
                          </td>
                          <td>{entry.deviceName}</td>
                          <td className="zs-num">{formatNumber(command.index)}</td>
                          <td className="zs-num">{formatNumber(command.value)}</td>
                          <td>
                            <StatusPill tone={status.tone}>{status.label}</StatusPill>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </Table>
              )}
            </>
          )}
        </Section>
      )}

      <Drawer
        open={formOpen}
        onClose={() => !sending && setFormOpen(false)}
        title="Enviar comando"
        subtitle="El comando entra a la cola y se despacha al dispositivo."
        footer={
          <>
            <Button type="submit" form="admin-command-form" icon={Send} disabled={sending}>
              {sending ? 'Enviando…' : 'Enviar comando'}
            </Button>
            <Button variant="ghost" onClick={() => setFormOpen(false)} disabled={sending}>
              Cancelar
            </Button>
          </>
        }
      >
        <Form id="admin-command-form" onSubmit={handleSend} noValidate>
          <Field label="Dispositivo" required error={formError?.deviceId}>
            {props => (
              <Select {...props} value={form.deviceId} onChange={event => setForm({ ...form, deviceId: event.target.value })}>
                <option value="">Elige un dispositivo</option>
                {devices.map(device => (
                  <option key={device.id} value={device.id}>
                    {nameOf(device)}
                    {device.tenant_id ? '' : ' · sin cuenta'}
                  </option>
                ))}
              </Select>
            )}
          </Field>
          <div className="grid gap-5 sm:grid-cols-2">
            <Field label="Relé (índice)" hint="Del 1 al 10.">
              {props => (
                <Input
                  {...props}
                  type="number"
                  inputMode="numeric"
                  min="1"
                  max="10"
                  className="font-mono"
                  value={form.index}
                  onChange={event => setForm({ ...form, index: parseInt(event.target.value, 10) || 1 })}
                />
              )}
            </Field>
            <Field label="Valor" hint="1 abre, 0 cierra.">
              {props => (
                <Input
                  {...props}
                  type="number"
                  inputMode="numeric"
                  min="0"
                  max="100"
                  className="font-mono"
                  value={form.value}
                  onChange={event => setForm({ ...form, value: parseInt(event.target.value, 10) || 0 })}
                />
              )}
            </Field>
            <Field label="Prioridad">
              {props => (
                <Select {...props} value={form.priority} onChange={event => setForm({ ...form, priority: event.target.value })}>
                  {PRIORITIES.map(priority => (
                    <option key={priority.value} value={priority.value}>
                      {priority.label}
                    </option>
                  ))}
                </Select>
              )}
            </Field>
            <Field label="Esperar antes de enviar" hint="Ejemplos: 0s, 30s, 5m, 1h." error={formError?.waitFor}>
              {props => (
                <Input
                  {...props}
                  className="font-mono"
                  value={form.waitFor}
                  onChange={event => setForm({ ...form, waitFor: event.target.value })}
                  placeholder="0s"
                  autoComplete="off"
                  spellCheck={false}
                />
              )}
            </Field>
          </div>
        </Form>
      </Drawer>

      <Drawer
        open={!!selected}
        onClose={() => setSelected(null)}
        title="Comando"
        subtitle={
          selected && (
            <>
              <span>{selected.deviceName}</span>
              <StatusPill tone={LOG_STATUS[selected.status].tone}>{LOG_STATUS[selected.status].label}</StatusPill>
            </>
          )
        }
      >
        {selected && (
          <>
            <KeyValue
              items={[
                { label: 'Enviado', value: formatDateTime(selected.createdAt), mono: true },
                { label: 'Dispositivo', value: selected.deviceName },
                { label: 'ID del dispositivo', value: selected.deviceId, mono: true },
                { label: 'Prioridad', value: priorityLabel(selected.payload.priority) },
                { label: 'Error', value: selected.error },
              ]}
            />
            <Section title="Datos enviados">
              <pre className="m-0 overflow-x-auto rounded-md border border-line bg-sunken p-4 font-mono text-sm">
                {JSON.stringify(selected.payload, null, 2)}
              </pre>
            </Section>
          </>
        )}
      </Drawer>
    </Page>
  )
}
