import { useState } from 'react'
import { Check, Pause, Pencil, Play, Trash2 } from 'lucide-react'
import { Button, ConfirmDialog, Drawer, EmptyState, KeyValue, List, LoadingState, Section, StatusPill } from '../../ui'
import { formatDate, formatDateTime } from '../../lib/zensor'
import { maintenanceApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import {
  activityStatus,
  activityTypeLabel,
  describeReminders,
  describeSchedule,
  dueLabel,
  executionStatus,
  formatFieldValue,
  isDue,
} from '../../utils/maintenanceSchedule'
import ActivityForm from './ActivityForm'
import { CompleteExecutionForm } from './CompleteExecutionDialog'

const HISTORY_PREVIEW = 10

function HistoryItem({ execution, fields }) {
  const status = executionStatus(execution)
  const values = Object.entries(execution.field_values || {})
  return (
    <li className="flex flex-col gap-3 rounded-md border border-line bg-surface p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-mono text-md">{formatDate(execution.scheduled_date)}</span>
        <StatusPill tone={status.tone}>
          {status.key === 'overdue' && execution.overdue_days > 0
            ? `${status.label} · ${execution.overdue_days} d`
            : status.label}
        </StatusPill>
      </div>
      {execution.completed_at && (
        <p className="m-0 text-body-sm text-muted">
          Hecha el {formatDateTime(execution.completed_at)}
          {execution.completed_by ? ` por ${execution.completed_by}` : ''}
        </p>
      )}
      {values.length > 0 && (
        <KeyValue
          className="text-body-sm"
          items={values.map(([key, value]) => {
            const field = fields.find(f => f.name === key)
            return { label: field?.display_name ?? key, value: formatFieldValue(field?.type, value) }
          })}
        />
      )}
    </li>
  )
}

function ActivityDetail({ item, onComplete }) {
  const { activity, executions, next } = item
  const [showAll, setShowAll] = useState(false)
  const fields = activity.fields || []
  const history = showAll ? executions : executions.slice(0, HISTORY_PREVIEW)
  const nextStatus = next ? executionStatus(next) : null

  return (
    <>
      {activity.description && <p className="zs-prose m-0">{activity.description}</p>}

      <Section title="Próxima vez">
        {!activity.is_active ? (
          <p className="m-0 text-body-sm text-muted">
            Actividad en pausa: mientras esté así no se programan nuevas fechas ni te llegan avisos. Reanúdala para volver
            a recibirlos.
          </p>
        ) : next ? (
          <div className="flex flex-col gap-4 rounded-md border border-line bg-surface p-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <span className="font-mono text-data-sm">{formatDateTime(next.scheduled_date)}</span>
              <StatusPill tone={nextStatus.tone}>{nextStatus.label}</StatusPill>
            </div>
            <p className="m-0 text-body-sm text-muted">{dueLabel(next.scheduled_date)}</p>
            {isDue(next) ? (
              <Button icon={Check} onClick={() => onComplete(next)} className="self-start">
                Marcar como hecha
              </Button>
            ) : (
              <p className="m-0 text-body-sm text-muted">Podrás marcarla como hecha desde el momento en que toque.</p>
            )}
          </div>
        ) : (
          <p className="m-0 text-body-sm text-muted">No hay fechas pendientes. Zensor creará la próxima según el calendario.</p>
        )}
      </Section>

      <Section title="Configuración">
        <KeyValue
          items={[
            { label: 'Tipo', value: activityTypeLabel(activity) },
            { label: 'Calendario', value: describeSchedule(activity.schedule) },
            { label: 'Avisos', value: describeReminders(activity.notification_days_before) },
            {
              label: 'Datos al completar',
              value: fields.length > 0 ? fields.map(f => f.display_name).join(', ') : 'Ninguno',
            },
          ]}
        />
      </Section>

      <Section title="Historial">
        {executions.length === 0 ? (
          <p className="m-0 text-body-sm text-muted">Aún no hay fechas registradas para esta actividad.</p>
        ) : (
          <>
            <List>
              {history.map(execution => (
                <HistoryItem key={execution.id} execution={execution} fields={fields} />
              ))}
            </List>
            {executions.length > HISTORY_PREVIEW && (
              <Button variant="ghost" onClick={() => setShowAll(!showAll)} className="self-start">
                {showAll ? 'Mostrar menos' : `Mostrar todo el historial (${executions.length})`}
              </Button>
            )}
          </>
        )}
      </Section>
    </>
  )
}

/**
 * Right drawer with an activity's next date, configuration and history. Editing and
 * marking an execution as done happen as steps inside the same drawer.
 */
const ActivityDrawer = ({ open, item, loading, tenantId, userEmail, onClose, onChanged, onDeleted }) => {
  const { showSuccess, showError } = useNotification()
  const [shown, setShown] = useState(item)
  const [mode, setMode] = useState('detail')
  const [completing, setCompleting] = useState(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [busy, setBusy] = useState(false)

  const [wasOpen, setWasOpen] = useState(open)

  if (item && item !== shown) setShown(item)
  if (open !== wasOpen) {
    setWasOpen(open)
    if (open) {
      setMode('detail')
      setCompleting(null)
    }
  }

  const activity = shown?.activity
  const status = shown ? activityStatus(activity, shown.next) : null

  const toggleActive = async () => {
    setBusy(true)
    try {
      if (activity.is_active) {
        await maintenanceApi.deactivateActivity(activity.id)
        showSuccess('Actividad pausada')
      } else {
        await maintenanceApi.activateActivity(activity.id)
        showSuccess('Actividad reanudada')
      }
      onChanged()
    } catch {
      showError(
        'Revisa tu conexión y vuelve a intentarlo.',
        activity.is_active ? 'No se pudo pausar la actividad' : 'No se pudo reanudar la actividad',
      )
    } finally {
      setBusy(false)
    }
  }

  const deleteActivity = async () => {
    setBusy(true)
    try {
      await maintenanceApi.deleteActivity(activity.id)
      showSuccess(`“${activity.name}” eliminada`)
      setConfirmDelete(false)
      onDeleted()
    } catch {
      showError('Revisa tu conexión y vuelve a intentarlo.', 'No se pudo eliminar la actividad')
    } finally {
      setBusy(false)
    }
  }

  const backToDetail = () => {
    setMode('detail')
    setCompleting(null)
  }

  const title = !activity
    ? 'Actividad'
    : mode === 'edit'
      ? 'Editar actividad'
      : mode === 'complete'
        ? 'Marcar como hecha'
        : activity.name

  const subtitle = activity && (
    <>
      <StatusPill tone={status.tone}>{status.label}</StatusPill>
      <span>{mode === 'detail' ? activityTypeLabel(activity) : activity.name}</span>
    </>
  )

  const footer = activity && mode === 'detail' && (
    <>
      <Button variant="secondary" icon={Pencil} onClick={() => setMode('edit')} disabled={busy}>
        Editar
      </Button>
      <Button variant="ghost" icon={activity.is_active ? Pause : Play} onClick={toggleActive} disabled={busy}>
        {activity.is_active ? 'Pausar' : 'Reanudar'}
      </Button>
      <Button variant="danger" icon={Trash2} onClick={() => setConfirmDelete(true)} disabled={busy}>
        Eliminar
      </Button>
    </>
  )

  let body
  if (!shown) {
    body = loading ? (
      <LoadingState>Cargando actividad…</LoadingState>
    ) : (
      <EmptyState title="No encontramos esta actividad">
        Puede que se haya eliminado. Cierra este panel para ver las actividades que siguen activas.
      </EmptyState>
    )
  } else if (mode === 'edit') {
    body = (
      <ActivityForm
        tenantId={tenantId}
        activity={activity}
        onSaved={() => {
          setMode('detail')
          onChanged()
        }}
        onCancel={backToDetail}
      />
    )
  } else if (mode === 'complete' && completing) {
    body = (
      <>
        <p className="m-0 text-body-sm text-muted">
          Programada para el <span className="font-mono">{formatDateTime(completing.scheduled_date)}</span>.
        </p>
        <CompleteExecutionForm
          activity={activity}
          execution={completing}
          userEmail={userEmail}
          onCancel={backToDetail}
          onCompleted={() => {
            backToDetail()
            onChanged()
          }}
        />
      </>
    )
  } else {
    body = (
      <ActivityDetail
        item={shown}
        onComplete={execution => {
          setCompleting(execution)
          setMode('complete')
        }}
      />
    )
  }

  return (
    <>
      <Drawer open={open} onClose={confirmDelete ? () => {} : onClose} title={title} subtitle={subtitle} footer={footer || undefined}>
        {body}
      </Drawer>
      <ConfirmDialog
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        onConfirm={deleteActivity}
        title={activity ? `¿Eliminar “${activity.name}”?` : '¿Eliminar actividad?'}
        confirmLabel="Eliminar actividad"
        destructive
        busy={busy}
      >
        Dejarás de recibir sus avisos y no podrás registrar nuevas fechas. Esta acción no se puede deshacer.
      </ConfirmDialog>
    </>
  )
}

export default ActivityDrawer
