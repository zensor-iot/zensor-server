import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import { scheduledTasksApi } from '../../config/api'
import {
  Button,
  Checkbox,
  ConfirmDialog,
  Drawer,
  EmptyState,
  ErrorState,
  Field,
  FilterChips,
  Form,
  IconButton,
  Input,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  StatusPill,
  Switch,
  Table,
} from '../../ui'
import { formatDateTime, formatUnit } from '../../lib/zensor'
import {
  DEFAULT_DURATION,
  MAX_DURATION,
  MIN_DURATION,
  clampDuration,
  describeSchedule,
  taskDuration,
  toggledTaskPayload,
} from '../../utils/irrigationSchedule'
import { cronOf, cronTaskPayload, listData } from '../../utils/adminOperations'

const EMPTY_FORM = { schedule: '', duration: DEFAULT_DURATION, isActive: true }
const PAGE_LIMIT = 100

const nameOf = entity => entity?.display_name || entity?.name || ''

export default function AdminScheduledTasks() {
  const { tenantId, deviceId } = useParams()
  const { showSuccess, showError } = useNotification()
  const [tasks, setTasks] = useState([])
  const [device, setDevice] = useState(null)
  const [tenant, setTenant] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [filter, setFilter] = useState('all')
  const [editor, setEditor] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [togglingId, setTogglingId] = useState(null)
  const [deleting, setDeleting] = useState(null)
  const [deleteBusy, setDeleteBusy] = useState(false)

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [tenantResponse, deviceResponse, tasksResponse] = await Promise.all([
        fetch(`/v1/tenants/${tenantId}`),
        fetch(`/v1/devices/${deviceId}`),
        scheduledTasksApi.getScheduledTasks(tenantId, deviceId, 1, PAGE_LIMIT),
      ])
      if (tenantResponse.ok) setTenant(await tenantResponse.json())
      if (deviceResponse.ok) setDevice(await deviceResponse.json())
      setTasks(listData(tasksResponse, 'tasks'))
    } catch {
      setError('Revisa tu conexión y vuelve a intentarlo.')
    } finally {
      setLoading(false)
    }
  }, [tenantId, deviceId])

  useEffect(() => {
    if (tenantId && deviceId) fetchData()
  }, [tenantId, deviceId, fetchData])

  const openCreate = () => {
    setForm(EMPTY_FORM)
    setEditor({ mode: 'create' })
  }

  const openEdit = task => {
    setForm({ schedule: cronOf(task), duration: taskDuration(task), isActive: task.is_active ?? true })
    setEditor({ mode: 'edit', task })
  }

  const closeEditor = () => {
    if (!saving) setEditor(null)
  }

  const handleSubmit = async event => {
    event.preventDefault()
    const payload = cronTaskPayload(form)
    const editing = editor?.mode === 'edit'
    setSaving(true)
    try {
      if (editing) {
        await scheduledTasksApi.updateScheduledTask(tenantId, deviceId, editor.task.id, payload)
        showSuccess('Programa guardado')
      } else {
        await scheduledTasksApi.createScheduledTask(tenantId, deviceId, payload)
        showSuccess('Programa creado')
      }
      setEditor(null)
      fetchData()
    } catch {
      showError(
        editing
          ? 'No se pudo guardar el programa. Revisa la expresión cron y vuelve a intentarlo.'
          : 'No se pudo crear el programa. Revisa la expresión cron y vuelve a intentarlo.',
      )
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async task => {
    setTogglingId(task.id)
    try {
      await scheduledTasksApi.updateScheduledTask(tenantId, deviceId, task.id, toggledTaskPayload(task))
      showSuccess(task.is_active ? 'Programa pausado' : 'Programa activado')
      fetchData()
    } catch {
      showError(
        task.is_active
          ? 'No se pudo pausar el programa. Vuelve a intentarlo.'
          : 'No se pudo activar el programa. Vuelve a intentarlo.',
      )
    } finally {
      setTogglingId(null)
    }
  }

  const handleDelete = async () => {
    if (!deleting) return
    setDeleteBusy(true)
    try {
      await scheduledTasksApi.deleteScheduledTask(tenantId, deviceId, deleting.id)
      showSuccess('Programa eliminado')
      setDeleting(null)
      fetchData()
    } catch {
      showError('No se pudo eliminar el programa. Vuelve a intentarlo.')
    } finally {
      setDeleteBusy(false)
    }
  }

  const activeCount = tasks.filter(task => task.is_active).length
  const pausedCount = tasks.length - activeCount
  const visibleTasks = tasks.filter(task => {
    if (filter === 'active') return task.is_active
    if (filter === 'paused') return !task.is_active
    return true
  })

  const context = [nameOf(device), nameOf(tenant)].filter(Boolean).join(' · ')
  const editingInterval = editor?.mode === 'edit' && !cronOf(editor.task)

  return (
    <Page>
      <PageHeader
        title="Programas"
        description={context ? `Riegos programados de ${context}.` : 'Riegos programados de este dispositivo.'}
        back={{ to: `/admin/tenants/${tenantId}/devices`, label: 'Dispositivos' }}
        actions={
          <Button icon={Plus} onClick={openCreate}>
            Nuevo programa
          </Button>
        }
      />

      {loading && tasks.length === 0 ? (
        <LoadingState>Cargando programas…</LoadingState>
      ) : error ? (
        <ErrorState title="No se pudieron cargar los programas" onRetry={fetchData}>
          {error}
        </ErrorState>
      ) : tasks.length === 0 ? (
        <EmptyState title="Aún no hay programas">
          Crea uno con “Nuevo programa” para que este dispositivo riegue solo.
        </EmptyState>
      ) : (
        <div className="flex flex-col gap-4">
          <FilterChips
            label="Filtro de programas"
            value={filter}
            onChange={setFilter}
            options={[
              { value: 'all', label: 'Todos', count: tasks.length },
              { value: 'active', label: 'Activos', count: activeCount },
              { value: 'paused', label: 'Pausados', count: pausedCount },
            ]}
          />
          {visibleTasks.length === 0 ? (
            <EmptyState>
              {filter === 'active'
                ? 'No hay programas activos. Activa uno desde “Todos”.'
                : 'No hay programas pausados.'}
            </EmptyState>
          ) : (
            <Table caption="Programas del dispositivo">
              <thead>
                <tr>
                  <th scope="col">Horario</th>
                  <th scope="col" className="text-right">Duración</th>
                  <th scope="col">Próximo riego</th>
                  <th scope="col">Estado</th>
                  <th scope="col">
                    <span className="sr-only">Acciones</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {visibleTasks.map(task => {
                  const label = describeSchedule(task)
                  const next = task.scheduling?.next_execution
                  return (
                    <tr key={task.id}>
                      <td>
                        <div className="flex flex-col gap-1">
                          <span className="font-medium">{label}</span>
                          {cronOf(task) && <span className="font-mono text-caption text-muted">{cronOf(task)}</span>}
                        </div>
                      </td>
                      <td className="zs-num">{formatUnit(taskDuration(task), 'min')}</td>
                      <td className="whitespace-nowrap">
                        {next && task.is_active ? (
                          <span className="font-mono">{formatDateTime(next)}</span>
                        ) : (
                          <span className="text-muted">{task.is_active ? 'Sin calcular' : 'En pausa'}</span>
                        )}
                      </td>
                      <td>
                        <div className="flex items-center gap-2">
                          <Switch
                            checked={!!task.is_active}
                            onChange={() => handleToggle(task)}
                            disabled={togglingId === task.id}
                            label={task.is_active ? `Pausar programa ${label}` : `Activar programa ${label}`}
                          />
                          <StatusPill tone={task.is_active ? 'ok' : 'neutral'}>
                            {task.is_active ? 'Activo' : 'En pausa'}
                          </StatusPill>
                        </div>
                      </td>
                      <td>
                        <div className="flex items-center justify-end gap-1">
                          <Link
                            to={`/admin/tenants/${tenantId}/devices/${deviceId}/scheduled-tasks/${task.id}/executions`}
                            className="inline-flex min-h-touch items-center rounded-pill px-3 text-label text-ink hover:bg-sunken"
                          >
                            Ejecuciones
                          </Link>
                          <IconButton icon={Pencil} label={`Editar programa ${label}`} onClick={() => openEdit(task)} />
                          <IconButton
                            icon={Trash2}
                            variant="danger"
                            label={`Eliminar programa ${label}`}
                            onClick={() => setDeleting(task)}
                          />
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </Table>
          )}
        </div>
      )}

      <Drawer
        open={!!editor}
        onClose={closeEditor}
        title={editor?.mode === 'edit' ? 'Editar programa' : 'Nuevo programa'}
        subtitle={context || undefined}
        footer={
          <>
            <Button type="submit" form="admin-program-form" disabled={saving}>
              {saving ? 'Guardando…' : editor?.mode === 'edit' ? 'Guardar programa' : 'Crear programa'}
            </Button>
            <Button variant="ghost" onClick={closeEditor} disabled={saving}>
              Cancelar
            </Button>
          </>
        }
      >
        {editingInterval && (
          <Notice tone="neutral">
            Este programa usa un horario por intervalo ({describeSchedule(editor.task)}). Si guardas una expresión cron,
            reemplaza ese horario.
          </Notice>
        )}
        <Form id="admin-program-form" onSubmit={handleSubmit}>
          <Field
            label="Horario (cron)"
            required
            hint="Seis campos: segundo minuto hora día mes día-de-semana. Ejemplo: 0 0 6 * * * riega todos los días a las 06:00."
          >
            {props => (
              <Input
                {...props}
                className="font-mono"
                value={form.schedule}
                onChange={event => setForm({ ...form, schedule: event.target.value })}
                placeholder="0 0 6 * * *"
                autoComplete="off"
                spellCheck={false}
              />
            )}
          </Field>
          <Field label="Duración del riego (min)" hint={`Entre ${MIN_DURATION} y ${MAX_DURATION} minutos.`}>
            {props => (
              <Input
                {...props}
                type="number"
                inputMode="numeric"
                min={MIN_DURATION}
                max={MAX_DURATION}
                className="font-mono"
                value={form.duration}
                onChange={event => setForm({ ...form, duration: event.target.value })}
                onBlur={() => setForm(current => ({ ...current, duration: clampDuration(current.duration) }))}
              />
            )}
          </Field>
          <Checkbox
            label="Programa activo"
            description="Si lo desmarcas, queda guardado en pausa."
            checked={form.isActive}
            onChange={event => setForm({ ...form, isActive: event.target.checked })}
          />
        </Form>
      </Drawer>

      <ConfirmDialog
        open={!!deleting}
        onClose={() => !deleteBusy && setDeleting(null)}
        onConfirm={handleDelete}
        title="¿Eliminar este programa?"
        confirmLabel={deleteBusy ? 'Eliminando…' : 'Eliminar programa'}
        destructive
        busy={deleteBusy}
      >
        {deleting && `${describeSchedule(deleting)} dejará de regar. Esta acción no se puede deshacer.`}
      </ConfirmDialog>
    </Page>
  )
}
