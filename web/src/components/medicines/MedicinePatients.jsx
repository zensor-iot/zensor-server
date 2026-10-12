import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { CalendarClock, ChevronRight, Plus } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { Button, Card, EmptyState, ErrorState, List, LoadingState, Page, PageHeader } from '../../ui'
import { PATIENT_KINDS } from '../../utils/medicineSchedule'
import PatientKindIcon from './PatientKindIcon'

const kindLabel = kind => PATIENT_KINDS.find(k => k.value === kind)?.label

const MedicinePatients = () => {
  const { tenantId } = useParams()
  const [patients, setPatients] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)

  const fetchPatients = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const data = await medicinesApi.listPatients(tenantId)
      setPatients(data.data || [])
    } catch {
      setLoadError(true)
    } finally {
      setLoading(false)
    }
  }, [tenantId])

  useEffect(() => {
    fetchPatients()
  }, [fetchPatients])

  const renderBody = () => {
    if (loading) return <LoadingState>Cargando pacientes…</LoadingState>
    if (loadError) {
      return (
        <ErrorState title="No se pudieron cargar los pacientes" onRetry={fetchPatients}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )
    }
    if (patients.length === 0) {
      return (
        <EmptyState title="Aún no hay pacientes">
          Agrega a la primera persona o animal que toma medicamentos. Después sumas sus tratamientos y Zensor te avisa
          cuándo dar cada dosis.
        </EmptyState>
      )
    }
    return (
      <List>
        {patients.map(patient => (
          <li key={patient.id}>
            <Card
              as={Link}
              interactive
              to={`/portal/${tenantId}/medicines/patients/${patient.id}`}
              className="no-underline"
            >
              <span className="flex items-center gap-4">
                <PatientKindIcon kind={patient.kind} />
                <span className="flex min-w-0 flex-1 flex-col gap-1">
                  <span className="text-title">{patient.name}</span>
                  <span className="text-body-sm text-muted">
                    {[kindLabel(patient.kind), patient.notes].filter(Boolean).join(' · ')}
                  </span>
                </span>
                <span className="flex shrink-0 items-center gap-1 text-label text-muted">
                  Detalle
                  <ChevronRight size={20} strokeWidth={1.8} aria-hidden="true" />
                </span>
              </span>
            </Card>
          </li>
        ))}
      </List>
    )
  }

  return (
    <Page>
      <PageHeader
        title="Medicinas"
        description="Quién toma qué y cada cuánto."
        actions={
          <>
            <Button variant="secondary" icon={CalendarClock} to={`/portal/${tenantId}/medicines/agenda`}>
              Ver agenda
            </Button>
            <Button icon={Plus} to={`/portal/${tenantId}/medicines/patients/new`}>
              Agregar paciente
            </Button>
          </>
        }
      />
      {renderBody()}
    </Page>
  )
}

export default MedicinePatients
