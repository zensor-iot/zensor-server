import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { scheduledTasksApi } from '../../config/api'
import {
  Drawer,
  EmptyState,
  ErrorState,
  FilterChips,
  KeyValue,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  Section,
  StatusPill,
  Table,
} from '../../ui'
import { formatDateTime, formatNumber, formatRelative, formatUnit } from '../../lib/zensor'
import { describeSchedule, taskDuration } from '../../utils/irrigationSchedule'
import { COMMAND_STATUS, EXECUTION_STATUS, commandStatus, executionStatus, listData } from '../../utils/adminOperations'

const EXECUTION_LIMIT = 50

const nameOf = entity => entity?.display_name || entity?.name || ''
const valveLabel = value => (Number(value) === 1 ? 'Abrir' : Number(value) === 0 ? 'Cerrar' : formatNumber(value))

export default function AdminTaskExecutions() {
  const { tenantId, deviceId, taskId } = useParams()
  const [executions, setExecutions] = useState([])
  const [scheduledTask, setScheduledTask] = useState(null)
  const [device, setDevice] = useState(null)
  const [tenant, setTenant] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [filter, setFilter] = useState('all')
  const [selected, setSelected] = useState(null)

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [tenantResponse, deviceResponse, taskResponse, executionsResponse] = await Promise.all([
        fetch(`/v1/tenants/${tenantId}`),
        fetch(`/v1/devices/${deviceId}`),
        scheduledTasksApi.getScheduledTask(tenantId, deviceId, taskId),
        scheduledTasksApi.getTaskExecutions(tenantId, deviceId, taskId, EXECUTION_LIMIT),
      ])
      if (tenantResponse.ok) setTenant(await tenantResponse.json())
      if (deviceResponse.ok) setDevice(await deviceResponse.json())
      setScheduledTask(taskResponse || null)
      setExecutions(listData(executionsResponse, 'tasks'))
    } catch {
      setError('Revisa tu conexión y vuelve a intentarlo.')
    } finally {
      setLoading(false)
    }
  }, [tenantId, deviceId, taskId])

  useEffect(() => {
    if (tenantId && deviceId && taskId) fetchData()
  }, [tenantId, deviceId, taskId, fetchData])

  const counts = executions.reduce((acc, execution) => {
    const status = executionStatus(execution)
    return { ...acc, [status]: (acc[status] || 0) + 1 }
  }, {})
  const visible = filter === 'all' ? executions : executions.filter(execution => executionStatus(execution) === filter)
  const failedCount = counts.failed || 0

  const context = [nameOf(device), nameOf(tenant)].filter(Boolean).join(' · ')
  const scheduleLabel = scheduledTask ? describeSchedule(scheduledTask) : null

  return (
    <Page>
      <PageHeader
        title="Ejecuciones"
        description={
          scheduleLabel
            ? `Historial del programa ${scheduleLabel}${context ? ` de ${context}` : ''}.`
            : 'Historial de ejecuciones de este programa.'
        }
        back={{ to: `/admin/tenants/${tenantId}/devices/${deviceId}/scheduled-tasks`, label: 'Programas' }}
      />

      {scheduledTask && (
        <Section title="Programa">
          <div className="rounded-xl border border-line bg-surface p-5">
            <KeyValue
              items={[
                { label: 'Horario', value: scheduleLabel },
                {
                  label: 'Estado',
                  value: (
                    <StatusPill tone={scheduledTask.is_active ? 'ok' : 'neutral'}>
                      {scheduledTask.is_active ? 'Activo' : 'En pausa'}
                    </StatusPill>
                  ),
                },
                { label: 'Duración', value: formatUnit(taskDuration(scheduledTask), 'min'), mono: true },
                { label: 'Comandos', value: formatNumber(scheduledTask.commands?.length || 0), mono: true },
              ]}
            />
          </div>
        </Section>
      )}

      <Section
        title="Historial"
        description={`Últimas ${formatNumber(EXECUTION_LIMIT)} ejecuciones, de la más reciente a la más antigua.`}
      >
        {loading && executions.length === 0 ? (
          <LoadingState>Cargando ejecuciones…</LoadingState>
        ) : error ? (
          <ErrorState title="No se pudieron cargar las ejecuciones" onRetry={fetchData}>
            {error}
          </ErrorState>
        ) : executions.length === 0 ? (
          <EmptyState title="Aún no hay ejecuciones">
            Aparecerán aquí cuando el programa se active en su próximo horario.
          </EmptyState>
        ) : (
          <>
            {failedCount > 0 && filter !== 'failed' && (
              <Notice action="Ver" onAction={() => setFilter('failed')}>
                {failedCount === 1
                  ? '1 ejecución tuvo errores al enviar comandos.'
                  : `${formatNumber(failedCount)} ejecuciones tuvieron errores al enviar comandos.`}
              </Notice>
            )}
            <FilterChips
              label="Filtro de ejecuciones"
              value={filter}
              onChange={setFilter}
              options={[
                { value: 'all', label: 'Todas', count: executions.length },
                { value: 'pending', label: 'Pendientes', count: counts.pending || 0 },
                { value: 'sent', label: 'Enviadas', count: counts.sent || 0 },
                { value: 'failed', label: 'Con error', count: failedCount },
              ]}
            />
            {visible.length === 0 ? (
              <EmptyState>No hay ejecuciones con este estado. Elige “Todas” para ver el historial completo.</EmptyState>
            ) : (
              <Table caption="Ejecuciones del programa">
                <thead>
                  <tr>
                    <th scope="col">Creada</th>
                    <th scope="col" className="text-right">Comandos enviados</th>
                    <th scope="col">Estado</th>
                  </tr>
                </thead>
                <tbody>
                  {visible.map(execution => {
                    const commands = execution.commands || []
                    const sent = commands.filter(command => commandStatus(command) === 'sent').length
                    const status = EXECUTION_STATUS[executionStatus(execution)]
                    return (
                      <tr key={execution.id}>
                        <td>
                          <button
                            type="button"
                            onClick={() => setSelected(execution)}
                            className="flex min-h-touch flex-col items-start justify-center bg-transparent p-0 text-left text-ink underline-offset-4 hover:underline"
                          >
                            <span className="font-mono">{formatDateTime(execution.created_at)}</span>
                            <span className="text-caption text-muted">{formatRelative(execution.created_at)}</span>
                          </button>
                        </td>
                        <td className="zs-num">
                          {formatNumber(sent)} de {formatNumber(commands.length)}
                        </td>
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

      <ExecutionDrawer execution={selected} onClose={() => setSelected(null)} />
    </Page>
  )
}

function ExecutionDrawer({ execution, onClose }) {
  const status = execution ? EXECUTION_STATUS[executionStatus(execution)] : null
  const commands = execution?.commands || []

  return (
    <Drawer
      open={!!execution}
      onClose={onClose}
      title="Ejecución"
      subtitle={
        execution && (
          <>
            <span className="font-mono">{formatDateTime(execution.created_at)}</span>
            <StatusPill tone={status.tone}>{status.label}</StatusPill>
          </>
        )
      }
    >
      {execution && (
        <>
          <Section title="Comandos">
            {commands.length === 0 ? (
              <EmptyState>Esta ejecución no generó comandos.</EmptyState>
            ) : (
              <ul className="m-0 flex list-none flex-col gap-3 p-0">
                {commands.map((command, position) => {
                  const commandState = COMMAND_STATUS[commandStatus(command)]
                  return (
                    <li key={command.id || position} className="flex flex-col gap-3 rounded-md border border-line bg-surface p-4">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className="font-medium">
                          {valveLabel(command.value)} relé <span className="font-mono">{formatNumber(command.index)}</span>
                        </span>
                        <StatusPill tone={commandState.tone}>{commandState.label}</StatusPill>
                      </div>
                      <KeyValue
                        className="text-body-sm"
                        items={[
                          { label: 'Prioridad', value: command.priority, mono: true },
                          {
                            label: 'Enviar desde',
                            value: command.dispatch_after && formatDateTime(command.dispatch_after),
                            mono: true,
                          },
                          { label: 'Enviado', value: command.sent_at && formatDateTime(command.sent_at), mono: true },
                          { label: 'Confirmado', value: command.acked_at && formatDateTime(command.acked_at), mono: true },
                          { label: 'Falló', value: command.failed_at && formatDateTime(command.failed_at), mono: true },
                          { label: 'Error', value: command.error_message },
                        ]}
                      />
                    </li>
                  )
                })}
              </ul>
            )}
          </Section>
          <Section title="Datos completos">
            <pre className="m-0 overflow-x-auto rounded-md border border-line bg-sunken p-4 font-mono text-sm">
              {JSON.stringify(execution, null, 2)}
            </pre>
          </Section>
        </>
      )}
    </Drawer>
  )
}
