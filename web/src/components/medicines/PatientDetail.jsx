import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import {
  Button,
  Card,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  FilterChips,
  IconButton,
  List,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  StatusPill,
  Switch,
  Table,
} from '../../ui'
import {
  PATIENT_KINDS,
  describeFrequency,
  doseState,
  doseTextOf,
  formatDateTime,
} from '../../utils/medicineSchedule'

const kindLabel = kind => PATIENT_KINDS.find(k => k.value === kind)?.label

const describeTreatment = treatment =>
  [
    [doseTextOf(treatment, treatment.dose_text), describeFrequency(treatment.schedule)].filter(Boolean).join(', '),
    treatment.ends_at && `hasta ${formatDateTime(treatment.ends_at)}`,
    treatment.total_doses && `${treatment.total_doses} dosis`,
  ]
    .filter(Boolean)
    .join(' · ')

const PatientDetail = () => {
  const { tenantId, patientId } = useParams()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const { showSuccess, showError } = useNotification()
  const tab = searchParams.get('tab') === 'history' ? 'history' : 'treatments'
  const listPath = `/portal/${tenantId}/medicines`

  const [patient, setPatient] = useState(null)
  const [treatments, setTreatments] = useState([])
  const [history, setHistory] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [busyId, setBusyId] = useState(null)
  const [confirm, setConfirm] = useState(null)

  const fetchData = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const [patientData, treatmentsData] = await Promise.all([
        medicinesApi.getPatient(patientId),
        medicinesApi.listTreatmentsByPatient(patientId),
      ])
      setPatient(patientData)
      const items = treatmentsData.data || []
      setTreatments(items)

      const doseLists = await Promise.all(items.map(t => medicinesApi.listDoses(t.id, 1, 50)))
      const merged = []
      doseLists.forEach((list, index) => {
        ;(list.data || []).forEach(dose => merged.push({ dose, treatment: items[index] }))
      })
      merged.sort((a, b) => new Date(b.dose.scheduled_at) - new Date(a.dose.scheduled_at))
      setHistory(merged)
    } catch {
      setLoadError(true)
    } finally {
      setLoading(false)
    }
  }, [patientId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleToggle = async treatment => {
    try {
      setBusyId(treatment.id)
      if (treatment.is_active) {
        await medicinesApi.deactivateTreatment(treatment.id)
        showSuccess(`Las próximas dosis de ${treatment.medicine_name} salen de la agenda hasta que lo reanudes.`, 'Tratamiento pausado')
      } else {
        await medicinesApi.activateTreatment(treatment.id)
        showSuccess(`${treatment.medicine_name} vuelve a la agenda.`, 'Tratamiento reanudado')
      }
      await fetchData()
    } catch {
      showError('Revisa tu conexión y vuelve a intentarlo.', 'No se pudo cambiar el tratamiento')
    } finally {
      setBusyId(null)
    }
  }

  const handleConfirm = async () => {
    try {
      setBusyId(confirm.id)
      if (confirm.type === 'patient') {
        await medicinesApi.deletePatient(patient.id)
        showSuccess(`${patient.name} ya no aparece en Medicinas.`, 'Paciente eliminado')
        navigate(listPath)
        return
      }
      await medicinesApi.deleteTreatment(confirm.id)
      showSuccess(`${confirm.name} ya no aparece en la agenda.`, 'Tratamiento eliminado')
      setConfirm(null)
      fetchData()
    } catch {
      showError(
        confirm.type === 'patient'
          ? 'Si tiene tratamientos activos, páusalos primero. Si no, revisa tu conexión y vuelve a intentarlo.'
          : 'Revisa tu conexión y vuelve a intentarlo.',
        'No se pudo eliminar',
      )
    } finally {
      setBusyId(null)
    }
  }

  if (loading && !patient) {
    return (
      <Page>
        <LoadingState>Cargando paciente…</LoadingState>
      </Page>
    )
  }

  if (loadError || !patient) {
    return (
      <Page>
        <PageHeader title="Paciente" back={{ to: listPath, label: 'Medicinas' }} />
        <ErrorState title="No se pudo cargar el paciente" onRetry={fetchData}>
          Revisa tu conexión y vuelve a intentarlo. Si lo eliminaron, vuelve a Medicinas.
        </ErrorState>
      </Page>
    )
  }

  const activeCount = treatments.filter(t => t.is_active).length
  const overdue = history.filter(({ dose }) => doseState(dose).key === 'overdue').length

  return (
    <Page>
      <PageHeader
        title={patient.name}
        description={[kindLabel(patient.kind), patient.notes].filter(Boolean).join(' · ')}
        back={{ to: listPath, label: 'Medicinas' }}
        actions={
          <>
            <Button
              variant="danger"
              icon={Trash2}
              onClick={() => setConfirm({ type: 'patient', id: patient.id, name: patient.name })}
            >
              Eliminar paciente
            </Button>
            <Button icon={Plus} to={`${listPath}/patients/${patientId}/treatments/new`}>
              Agregar tratamiento
            </Button>
          </>
        }
      />

      {overdue > 0 && (
        <Notice action="Abrir agenda" onAction={() => navigate(`${listPath}/agenda`)}>
          {overdue === 1 ? '1 dosis atrasada sin registrar.' : `${overdue} dosis atrasadas sin registrar.`}
        </Notice>
      )}

      <FilterChips
        label="Vista del paciente"
        value={tab}
        onChange={value => setSearchParams(value === 'history' ? { tab: 'history' } : {})}
        options={[
          { value: 'treatments', label: 'Tratamientos', count: treatments.length },
          { value: 'history', label: 'Historial', count: history.length },
        ]}
      />

      {tab === 'treatments' ? (
        treatments.length === 0 ? (
          <EmptyState title="Sin tratamientos">
            Agrega el primer tratamiento de {patient.name} para que sus dosis aparezcan en la agenda.
          </EmptyState>
        ) : (
          <List>
            {treatments.map(treatment => (
              <li key={treatment.id}>
                <Card>
                  <div className="flex flex-wrap items-center gap-4">
                    <div className="flex min-w-0 flex-1 basis-[240px] flex-col gap-1">
                      <span className="text-title">{treatment.medicine_name}</span>
                      <span className="text-body-sm text-muted">{describeTreatment(treatment)}</span>
                      {treatment.notes && <span className="text-body-sm text-muted">{treatment.notes}</span>}
                    </div>
                    <div className="flex items-center gap-2">
                      <StatusPill tone={treatment.is_active ? 'ok' : 'neutral'}>
                        {treatment.is_active ? 'Activo' : 'En pausa'}
                      </StatusPill>
                      <Switch
                        checked={treatment.is_active}
                        onChange={() => handleToggle(treatment)}
                        disabled={busyId === treatment.id}
                        label={`${treatment.medicine_name} activo`}
                      />
                      <IconButton
                        icon={Trash2}
                        variant="danger"
                        label={`Eliminar ${treatment.medicine_name}`}
                        onClick={() =>
                          setConfirm({ type: 'treatment', id: treatment.id, name: treatment.medicine_name })
                        }
                      />
                    </div>
                  </div>
                </Card>
              </li>
            ))}
          </List>
        )
      ) : history.length === 0 ? (
        <EmptyState title="Sin dosis todavía">
          Las dosis aparecen aquí a medida que se programan. Regístralas desde la agenda.
        </EmptyState>
      ) : (
        <Table caption={`Historial de dosis de ${patient.name}`}>
          <thead>
            <tr>
              <th scope="col">Programada</th>
              <th scope="col">Medicina</th>
              <th scope="col">Dosis</th>
              <th scope="col">Notas</th>
              <th scope="col">Estado</th>
            </tr>
          </thead>
          <tbody>
            {history.map(({ dose, treatment }) => {
              const state = doseState(dose)
              return (
                <tr key={dose.id}>
                  <td className="zs-num whitespace-nowrap">{formatDateTime(dose.scheduled_at)}</td>
                  <td>{treatment.medicine_name}</td>
                  <td className="whitespace-nowrap">{doseTextOf(dose, treatment.dose_text)}</td>
                  <td className="text-muted">{dose.notes || 'Sin notas'}</td>
                  <td>
                    <StatusPill tone={state.tone}>{state.label}</StatusPill>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </Table>
      )}

      <ConfirmDialog
        open={Boolean(confirm)}
        onClose={() => setConfirm(null)}
        onConfirm={handleConfirm}
        busy={Boolean(busyId)}
        destructive
        title={confirm?.type === 'patient' ? `¿Eliminar a ${confirm?.name}?` : `¿Eliminar ${confirm?.name ?? ''}?`}
        confirmLabel={confirm?.type === 'patient' ? 'Eliminar paciente' : 'Eliminar tratamiento'}
      >
        {confirm?.type !== 'patient'
          ? 'Sus próximas dosis dejan de aparecer en la agenda.'
          : activeCount > 0
            ? `Tiene ${activeCount === 1 ? '1 tratamiento activo' : `${activeCount} tratamientos activos`}: páusalos antes de eliminarlo.`
            : 'Deja de aparecer en Medicinas.'}
      </ConfirmDialog>
    </Page>
  )
}

export default PatientDetail
