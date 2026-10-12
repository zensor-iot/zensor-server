import { useCallback, useEffect, useMemo, useState } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import {
  Button,
  EmptyState,
  ErrorState,
  FilterChips,
  List,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  Section,
  StatusPill,
} from '../../ui'
import { doseState, doseTextOf, formatTime, groupByDay } from '../../utils/medicineSchedule'
import PatientKindIcon from './PatientKindIcon'
import RecordDoseDialog from './RecordDoseDialog'

const PAST_HOURS = 12
const FUTURE_HOURS = 24

const FILTERS = {
  all: () => true,
  overdue: entry => doseState(entry.dose).key === 'overdue',
  pending: entry => entry.dose.status === 'pending',
  resolved: entry => entry.dose.status !== 'pending',
}

const EMPTY_FILTER_TEXT = {
  overdue: 'No hay dosis atrasadas. Todo lo que tocaba ya está registrado.',
  pending: 'No quedan dosis por dar en esta ventana.',
  resolved: 'Aún no registras dosis en esta ventana. Registra la próxima cuando la des.',
}

const MedicineAgenda = () => {
  const { tenantId } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const { showSuccess, showError } = useNotification()
  const [entries, setEntries] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [filter, setFilter] = useState('all')
  const [selected, setSelected] = useState(null)

  const fetchAgenda = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const from = new Date(Date.now() - PAST_HOURS * 3600 * 1000).toISOString()
      const to = new Date(Date.now() + FUTURE_HOURS * 3600 * 1000).toISOString()
      const data = await medicinesApi.listAgenda(tenantId, from, to)
      setEntries(data.data || [])
    } catch {
      setLoadError(true)
    } finally {
      setLoading(false)
    }
  }, [tenantId])

  useEffect(() => {
    fetchAgenda()
  }, [fetchAgenda])

  const deeplinkDoseId = searchParams.get('dose')
  useEffect(() => {
    if (!deeplinkDoseId || loading) return
    const entry = entries.find(e => e.dose.id === deeplinkDoseId)
    if (entry && entry.dose.status === 'pending') setSelected(entry)
    setSearchParams(
      params => {
        const next = new URLSearchParams(params)
        next.delete('dose')
        return next
      },
      { replace: true },
    )
  }, [deeplinkDoseId, loading, entries, setSearchParams])

  const counts = useMemo(
    () => Object.fromEntries(Object.entries(FILTERS).map(([key, test]) => [key, entries.filter(test).length])),
    [entries],
  )
  const groups = useMemo(() => groupByDay(entries.filter(FILTERS[filter])), [entries, filter])

  const handleResolve = async (action, notes) => {
    try {
      if (action === 'administer') {
        await medicinesApi.administerDose(selected.dose.id, notes)
        showSuccess(
          `${doseTextOf(selected.dose, selected.dose_text)} de ${selected.medicine_name} para ${selected.patient_name}`,
          'Dosis registrada',
        )
      } else {
        await medicinesApi.skipDose(selected.dose.id, notes)
        showSuccess(`${selected.medicine_name} para ${selected.patient_name}`, 'Dosis saltada')
      }
      setSelected(null)
      fetchAgenda()
    } catch {
      showError('Puede que ya esté registrada. Actualiza la agenda y vuelve a intentarlo.', 'No se pudo registrar la dosis')
    }
  }

  const renderBody = () => {
    if (loading) return <LoadingState>Cargando la agenda…</LoadingState>
    if (loadError) {
      return (
        <ErrorState title="No se pudo cargar la agenda" onRetry={fetchAgenda}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )
    }
    if (entries.length === 0) {
      return (
        <EmptyState title="Nada que dar por ahora">
          No hay dosis en las últimas {PAST_HOURS} horas ni en las próximas {FUTURE_HOURS}. Agrega un tratamiento desde
          Pacientes para verlo aquí.
        </EmptyState>
      )
    }
    return (
      <>
        {counts.overdue > 0 && (
          <Notice action={filter === 'overdue' ? undefined : 'Ver atrasadas'} onAction={() => setFilter('overdue')}>
            {counts.overdue === 1
              ? '1 dosis atrasada sin registrar.'
              : `${counts.overdue} dosis atrasadas sin registrar.`}
          </Notice>
        )}
        <FilterChips
          label="Filtrar dosis"
          value={filter}
          onChange={setFilter}
          options={[
            { value: 'all', label: 'Todas', count: counts.all },
            { value: 'overdue', label: 'Atrasadas', count: counts.overdue },
            { value: 'pending', label: 'Pendientes', count: counts.pending },
            { value: 'resolved', label: 'Registradas', count: counts.resolved },
          ]}
        />
        {groups.length === 0 ? (
          <EmptyState>{EMPTY_FILTER_TEXT[filter]}</EmptyState>
        ) : (
          groups.map(group => (
            <Section key={group.key} title={group.label}>
              <List>
                {group.entries.map(entry => (
                  <AgendaRow key={entry.dose.id} entry={entry} onRecord={() => setSelected(entry)} />
                ))}
              </List>
            </Section>
          ))
        )}
      </>
    )
  }

  return (
    <Page>
      <PageHeader
        title="Agenda"
        description={`Dosis de las últimas ${PAST_HOURS} horas y de las próximas ${FUTURE_HOURS}.`}
        back={{ to: `/portal/${tenantId}/medicines`, label: 'Medicinas' }}
      />
      {renderBody()}
      <RecordDoseDialog entry={selected} onClose={() => setSelected(null)} onResolved={handleResolve} />
    </Page>
  )
}

function AgendaRow({ entry, onRecord }) {
  const state = doseState(entry.dose)
  const doseText = doseTextOf(entry.dose, entry.dose_text)
  return (
    <li className="flex flex-wrap items-center gap-4 rounded-xl border border-line bg-surface p-4">
      <span className="w-12 shrink-0 text-data-sm">{formatTime(entry.dose.scheduled_at)}</span>
      <PatientKindIcon kind={entry.patient_kind} />
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <span className="text-label">{entry.patient_name}</span>
        <span className="text-body-sm text-muted">
          {doseText} de {entry.medicine_name}
        </span>
      </div>
      <StatusPill tone={state.tone}>{state.label}</StatusPill>
      {entry.dose.status === 'pending' && (
        <Button
          variant="secondary"
          size="sm"
          onClick={onRecord}
          aria-label={`Registrar dosis de ${entry.medicine_name} para ${entry.patient_name}`}
        >
          Registrar
        </Button>
      )}
    </li>
  )
}

export default MedicineAgenda
