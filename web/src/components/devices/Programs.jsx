import { useState } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { scheduledTasksApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { formatDateTime } from '../../lib/zensor'
import {
  MAX_DURATION,
  MIN_DURATION,
  describeSchedule,
  emptyProgram,
  nextRun,
  programToTaskPayload,
  taskDuration,
  taskNextRun,
  taskToProgram,
  toggledTaskPayload,
} from '../../utils/irrigationSchedule'
import {
  Button,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Field,
  Form,
  FormActions,
  IconButton,
  Input,
  LoadingState,
  Section,
  Switch,
} from '../../ui'

const describeNext = date => (date ? `Próximo: ${formatDateTime(date)}` : null)

function ProgramRow({ task, busy, onToggle, onEdit, onDelete }) {
  const next = task.is_active ? describeNext(taskNextRun(task)) : 'Pausado'
  return (
    <li className="flex flex-wrap items-center gap-3 rounded-md border border-line bg-surface py-3 pl-4 pr-2">
      <div className="flex min-w-0 flex-[1_1_180px] flex-col">
        <span className="font-bold">{describeSchedule(task)}</span>
        <span className="text-caption text-muted">
          {taskDuration(task)}&nbsp;min{next && ` · ${next}`}
        </span>
      </div>
      <div className="flex items-center">
        <Switch
          checked={!!task.is_active}
          disabled={busy}
          onChange={() => onToggle(task)}
          label={task.is_active ? 'Pausar programa' : 'Activar programa'}
        />
        <IconButton icon={Pencil} label="Editar programa" onClick={() => onEdit(task)} />
        <IconButton icon={Trash2} label="Eliminar programa" variant="danger" onClick={() => onDelete(task)} />
      </div>
    </li>
  )
}

export function ProgramList({ tenantId, device, entry, onChange, onAdd, onEdit, onRetry }) {
  const { showSuccess, showError } = useNotification()
  const [busyId, setBusyId] = useState(null)
  const [toDelete, setToDelete] = useState(null)

  const toggle = async task => {
    setBusyId(task.id)
    try {
      await scheduledTasksApi.updateScheduledTask(tenantId, device.id, task.id, toggledTaskPayload(task))
      onChange(tasks => tasks.map(t => (t.id === task.id ? { ...t, is_active: !task.is_active } : t)))
      showSuccess(task.is_active ? 'Programa pausado' : 'Programa activado')
    } catch {
      showError('No se pudo cambiar el programa. Vuelve a intentarlo.')
    } finally {
      setBusyId(null)
    }
  }

  const remove = async () => {
    const task = toDelete
    setBusyId(task.id)
    try {
      await scheduledTasksApi.deleteScheduledTask(tenantId, device.id, task.id)
      onChange(tasks => tasks.filter(t => t.id !== task.id))
      showSuccess('Programa eliminado')
      setToDelete(null)
    } catch {
      showError('No se pudo eliminar el programa. Vuelve a intentarlo.')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <Section
      title="Programas"
      actions={
        <Button variant="secondary" size="sm" icon={Plus} onClick={onAdd}>
          Agregar
        </Button>
      }
    >
      {!entry ? (
        <LoadingState>Cargando programas…</LoadingState>
      ) : entry.error && entry.tasks.length === 0 ? (
        <ErrorState title="No se pudieron cargar los programas" onRetry={onRetry} />
      ) : entry.tasks.length === 0 ? (
        <EmptyState>Aún no hay programas. Crea uno para que {device.display_name || device.name} riegue solo.</EmptyState>
      ) : (
        <ul className="m-0 flex list-none flex-col gap-3 p-0">
          {entry.tasks.map(task => (
            <ProgramRow
              key={task.id}
              task={task}
              busy={busyId === task.id}
              onToggle={toggle}
              onEdit={onEdit}
              onDelete={setToDelete}
            />
          ))}
        </ul>
      )}

      <ConfirmDialog
        open={!!toDelete}
        onClose={() => setToDelete(null)}
        onConfirm={remove}
        busy={!!toDelete && busyId === toDelete.id}
        destructive
        title="¿Eliminar este programa?"
        confirmLabel="Eliminar programa"
      >
        {toDelete && `${describeSchedule(toDelete)} · ${taskDuration(toDelete)} min. Esta acción no se puede deshacer.`}
      </ConfirmDialog>
    </Section>
  )
}

export function ProgramForm({ tenantId, device, task, onSaved, onCancel }) {
  const { showSuccess, showError } = useNotification()
  const [program, setProgram] = useState(() => (task ? taskToProgram(task) : emptyProgram()))
  const [saving, setSaving] = useState(false)

  const update = field => event => {
    const { value, type } = event.target
    setProgram(prev => ({ ...prev, [field]: type === 'number' ? Number(value) : value }))
  }

  const next = nextRun(program)

  const submit = async event => {
    event.preventDefault()
    setSaving(true)
    const payload = programToTaskPayload(program)
    try {
      if (task) {
        await scheduledTasksApi.updateScheduledTask(tenantId, device.id, task.id, payload)
        onSaved(tasks => tasks.map(t => (t.id === task.id ? { ...t, ...payload, id: task.id } : t)))
      } else {
        const created = await scheduledTasksApi.createScheduledTask(tenantId, device.id, payload)
        onSaved(tasks => [...tasks, { ...payload, id: created?.id || created?.uuid }])
      }
      showSuccess('Programa guardado')
    } catch {
      showError('No se pudo guardar el programa. Revisa los datos y vuelve a intentarlo.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Form onSubmit={submit}>
      <p className="m-0 rounded-md bg-sunken px-4 py-3 text-body-sm">
        {program.dayInterval > 1 ? `Cada ${program.dayInterval} días` : 'Todos los días'} a las {program.time} durante{' '}
        {program.duration}&nbsp;min.{next && ` Próximo riego: ${formatDateTime(next)}.`}
      </p>

      <Field label="Hora" required>
        {props => <Input {...props} type="time" value={program.time} onChange={update('time')} />}
      </Field>

      <Field label="Duración (minutos)" hint={`Entre ${MIN_DURATION} y ${MAX_DURATION} minutos.`} required>
        {props => (
          <Input
            {...props}
            type="number"
            inputMode="numeric"
            min={MIN_DURATION}
            max={MAX_DURATION}
            value={program.duration}
            onChange={update('duration')}
          />
        )}
      </Field>

      <Field label="Repetir cada (días)" hint="1 = todos los días. Hasta 15 días." required>
        {props => (
          <Input
            {...props}
            type="number"
            inputMode="numeric"
            min={1}
            max={15}
            value={program.dayInterval}
            onChange={update('dayInterval')}
          />
        )}
      </Field>

      <Field label="Desde" required>
        {props => <Input {...props} type="date" value={program.initialDay} onChange={update('initialDay')} />}
      </Field>

      <div className="flex items-center justify-between gap-3">
        <span className="text-label">Programa activo</span>
        <Switch
          checked={program.isActive}
          onChange={isActive => setProgram(prev => ({ ...prev, isActive }))}
          label="Programa activo"
        />
      </div>

      <FormActions>
        <Button type="submit" disabled={saving}>
          {saving ? 'Guardando…' : 'Guardar programa'}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={saving}>
          Cancelar
        </Button>
      </FormActions>
    </Form>
  )
}
