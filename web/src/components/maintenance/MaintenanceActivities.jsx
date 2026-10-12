import { useMemo, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { CalendarClock, ChevronRight, Plus } from 'lucide-react'
import {
  Button,
  Card,
  CardGrid,
  Drawer,
  EmptyState,
  ErrorState,
  FilterChips,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  StatusPill,
} from '../../ui'
import { formatDateTime } from '../../lib/zensor'
import { useAuth } from '../../contexts/AuthContext'
import { activityStatus, activityTypeLabel, describeSchedule, dueLabel } from '../../utils/maintenanceSchedule'
import { useMaintenanceOverview } from './useMaintenanceOverview'
import ActivityDrawer from './ActivityDrawer'
import ActivityForm from './ActivityForm'

const URGENCY = { overdue: 0, today: 1, pending: 2, clear: 3, paused: 4 }

const plural = (count, singular, pluralForm) => `${count} ${count === 1 ? singular : pluralForm}`

function ActivityCard({ activity, next, status, onOpen }) {
  return (
    <Card as="button" type="button" interactive onClick={onOpen}>
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1">
          <span className="break-words text-title">{activity.name}</span>
          <span className="text-body-sm text-muted">
            {activityTypeLabel(activity)} · {describeSchedule(activity.schedule)}
          </span>
        </div>
        <StatusPill tone={status.tone}>{status.label}</StatusPill>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-3 text-body-sm">
        <span className={status.key === 'overdue' ? 'text-attention' : 'text-muted'}>
          {!activity.is_active
            ? 'Sin avisos mientras esté en pausa'
            : next
              ? `${dueLabel(next.scheduled_date)} · ${formatDateTime(next.scheduled_date)}`
              : 'Sin fechas pendientes'}
        </span>
        <span className="inline-flex items-center gap-1 text-label text-ink">
          Detalle
          <ChevronRight size={20} strokeWidth={1.8} aria-hidden="true" />
        </span>
      </div>
    </Card>
  )
}

function summaryText(overdue, today) {
  const parts = []
  if (overdue > 0) parts.push(plural(overdue, 'actividad atrasada', 'actividades atrasadas'))
  if (today > 0) parts.push(`${plural(today, 'actividad que toca', 'actividades que tocan')} hoy`)
  return `${parts.join(' y ')}.`
}

/**
 * Maintenance activities of a tenant. The same component serves the activity detail
 * and "new activity" routes, rendering those as drawers over the list.
 */
const MaintenanceActivities = () => {
  const { tenantId, activityId } = useParams()
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { items, loading, error, reload } = useMaintenanceOverview(tenantId)
  const [filter, setFilter] = useState('all')
  const [lastActivityId, setLastActivityId] = useState(activityId ?? null)

  if (activityId && activityId !== lastActivityId) setLastActivityId(activityId)

  const basePath = `/portal/${tenantId}/maintenance`
  const creating = pathname.replace(/\/$/, '').endsWith('/maintenance/new')
  const closeDrawer = () => navigate(basePath)

  const rows = useMemo(
    () =>
      (items || [])
        .map(item => ({ ...item, status: activityStatus(item.activity, item.next) }))
        .sort((a, b) => {
          const byUrgency = URGENCY[a.status.key] - URGENCY[b.status.key]
          if (byUrgency !== 0) return byUrgency
          if (a.next && b.next) return new Date(a.next.scheduled_date) - new Date(b.next.scheduled_date)
          return a.activity.name.localeCompare(b.activity.name, 'es')
        }),
    [items],
  )

  const counts = useMemo(
    () => ({
      overdue: rows.filter(r => r.status.key === 'overdue').length,
      today: rows.filter(r => r.status.key === 'today').length,
      paused: rows.filter(r => r.status.key === 'paused').length,
    }),
    [rows],
  )

  const visible = rows.filter(r => filter === 'all' || r.status.key === filter)
  const selected = activityId ? rows.find(r => r.activity.id === activityId) ?? null : null

  const chips = [
    { value: 'all', label: 'Todas', count: rows.length },
    { value: 'overdue', label: 'Atrasadas', count: counts.overdue },
    { value: 'paused', label: 'En pausa', count: counts.paused },
  ]

  let content
  if (!items && loading) {
    content = <LoadingState>Cargando actividades…</LoadingState>
  } else if (!items && error) {
    content = (
      <ErrorState title="No se pudieron cargar las actividades" onRetry={reload}>
        Revisa tu conexión y vuelve a intentarlo.
      </ErrorState>
    )
  } else if (rows.length === 0) {
    content = (
      <EmptyState title="Aún no hay actividades">
        Crea una con “Nueva actividad” para que Zensor te avise cuando toque cambiar un filtro, revisar el auto o
        desparasitar a las mascotas.
      </EmptyState>
    )
  } else {
    content = (
      <>
        {(counts.overdue > 0 || counts.today > 0) && (
          <Notice
            tone={counts.overdue > 0 ? 'attention' : 'live'}
            action={counts.overdue > 0 && filter !== 'overdue' ? 'Ver atrasadas' : undefined}
            onAction={() => setFilter('overdue')}
          >
            {summaryText(counts.overdue, counts.today)}
          </Notice>
        )}
        <FilterChips label="Filtrar actividades" options={chips} value={filter} onChange={setFilter} />
        {visible.length === 0 ? (
          <EmptyState>
            {filter === 'overdue' ? 'No hay actividades atrasadas.' : 'No hay actividades en pausa.'}
          </EmptyState>
        ) : (
          <CardGrid>
            {visible.map(row => (
              <ActivityCard
                key={row.activity.id}
                activity={row.activity}
                next={row.next}
                status={row.status}
                onOpen={() => navigate(`${basePath}/activities/${row.activity.id}`)}
              />
            ))}
          </CardGrid>
        )}
      </>
    )
  }

  return (
    <Page>
      <PageHeader
        title="Mantención"
        description="Tareas que se repiten en tu casa y tu campo. Zensor te avisa antes de que toquen."
        actions={
          <>
            <Button variant="secondary" icon={CalendarClock} to={`${basePath}/up-next`}>
              Próximas
            </Button>
            <Button icon={Plus} to={`${basePath}/new`}>
              Nueva actividad
            </Button>
          </>
        }
      />

      {content}

      <ActivityDrawer
        key={lastActivityId ?? 'none'}
        open={Boolean(activityId)}
        item={selected}
        loading={loading}
        tenantId={tenantId}
        userEmail={user?.email}
        onClose={closeDrawer}
        onChanged={reload}
        onDeleted={() => {
          closeDrawer()
          reload()
        }}
      />

      <Drawer open={creating} onClose={closeDrawer} title="Nueva actividad">
        {creating && (
          <ActivityForm
            tenantId={tenantId}
            onSaved={() => {
              closeDrawer()
              reload()
            }}
            onCancel={closeDrawer}
          />
        )}
      </Drawer>
    </Page>
  )
}

export default MaintenanceActivities
