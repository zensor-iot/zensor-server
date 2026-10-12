import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Check } from 'lucide-react'
import { Button, Card, EmptyState, ErrorState, List, LoadingState, Notice, Page, PageHeader, StatusPill } from '../../ui'
import { formatDateTime } from '../../lib/zensor'
import { useAuth } from '../../contexts/AuthContext'
import { dueLabel, executionStatus, isDue, upNext } from '../../utils/maintenanceSchedule'
import { useMaintenanceOverview } from './useMaintenanceOverview'
import CompleteExecutionDialog from './CompleteExecutionDialog'

/** Up next: the earliest pending execution of every activity, soonest first. */
const MaintenanceUpNext = () => {
  const { tenantId } = useParams()
  const { user } = useAuth()
  const { items, loading, error, reload } = useMaintenanceOverview(tenantId)
  const [completing, setCompleting] = useState(null)

  const basePath = `/portal/${tenantId}/maintenance`
  const rows = useMemo(() => upNext(items || []), [items])
  const overdue = rows.filter(r => r.execution.is_overdue)

  let content
  if (!items && loading) {
    content = <LoadingState>Cargando próximas mantenciones…</LoadingState>
  } else if (!items && error) {
    content = (
      <ErrorState title="No se pudieron cargar las próximas mantenciones" onRetry={reload}>
        Revisa tu conexión y vuelve a intentarlo.
      </ErrorState>
    )
  } else if (rows.length === 0) {
    content = (
      <EmptyState title="Todo al día">
        No hay mantenciones pendientes. Cuando se acerque la próxima fecha de una actividad, aparecerá aquí.
      </EmptyState>
    )
  } else {
    content = (
      <>
        {overdue.length > 0 && (
          <Notice>
            {overdue.length === 1
              ? `“${overdue[0].activity.name}” está atrasada.`
              : `${overdue.length} mantenciones atrasadas. La más antigua es “${overdue[0].activity.name}”.`}
          </Notice>
        )}
        <List>
          {rows.map(({ activity, execution }) => {
            const status = executionStatus(execution)
            return (
              <li key={execution.id}>
                <Card>
                  <div className="flex flex-wrap items-center justify-between gap-4">
                    <div className="flex min-w-0 flex-1 flex-col gap-1">
                      <div className="flex flex-wrap items-center gap-3">
                        <Link
                          to={`${basePath}/activities/${activity.id}`}
                          className="inline-flex min-h-touch items-center break-words text-title text-ink underline-offset-4 hover:underline"
                        >
                          {activity.name}
                        </Link>
                        <StatusPill tone={status.tone}>
                          {status.key === 'overdue' && execution.overdue_days > 0
                            ? `${status.label} · ${execution.overdue_days} d`
                            : status.label}
                        </StatusPill>
                      </div>
                      <span className={status.key === 'overdue' ? 'text-body-sm text-attention' : 'text-body-sm text-muted'}>
                        {dueLabel(execution.scheduled_date)} · {formatDateTime(execution.scheduled_date)}
                      </span>
                    </div>
                    {isDue(execution) && (
                      <Button variant="secondary" icon={Check} onClick={() => setCompleting({ activity, execution })}>
                        Marcar como hecha
                      </Button>
                    )}
                  </div>
                </Card>
              </li>
            )
          })}
        </List>
      </>
    )
  }

  return (
    <Page>
      <PageHeader
        title="Próximas mantenciones"
        description="Lo siguiente que toca en cada actividad, de lo más urgente a lo más lejano."
        back={{ to: basePath, label: 'Mantención' }}
      />

      {content}

      <CompleteExecutionDialog
        target={completing}
        userEmail={user?.email}
        onClose={() => setCompleting(null)}
        onCompleted={() => {
          setCompleting(null)
          reload()
        }}
      />
    </Page>
  )
}

export default MaintenanceUpNext
