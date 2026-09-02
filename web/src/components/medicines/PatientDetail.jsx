import { useState, useEffect, useCallback } from 'react'
import { useParams, Link, useSearchParams } from 'react-router-dom'
import { ChevronLeft, Plus, Loader2, Pause, Play, Trash2, PawPrint, User, Check, SkipForward, Clock } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { describeFrequency, formatDateTime } from '../../utils/medicineSchedule'
import './Medicines.css'

const STATUS_LABELS = {
    administered: 'Administrada',
    skipped: 'Saltada',
    pending: 'Pendiente'
}

const STATUS_ICONS = {
    administered: Check,
    skipped: SkipForward,
    pending: Clock
}

const PatientDetail = () => {
    const { tenantId, patientId } = useParams()
    const [searchParams, setSearchParams] = useSearchParams()
    const { showSuccess, showError } = useNotification()
    const tab = searchParams.get('tab') === 'history' ? 'history' : 'treatments'

    const [patient, setPatient] = useState(null)
    const [treatments, setTreatments] = useState([])
    const [history, setHistory] = useState([])
    const [loading, setLoading] = useState(true)

    const fetchData = useCallback(async () => {
        try {
            setLoading(true)
            const [patientData, treatmentsData] = await Promise.all([
                medicinesApi.getPatient(patientId),
                medicinesApi.listTreatmentsByPatient(patientId)
            ])
            setPatient(patientData)
            const items = treatmentsData.data || []
            setTreatments(items)

            const doseLists = await Promise.all(items.map(t => medicinesApi.listDoses(t.id, 1, 50)))
            const merged = []
            doseLists.forEach((list, index) => {
                (list.data || []).forEach(dose => {
                    merged.push({ dose, treatment: items[index] })
                })
            })
            merged.sort((a, b) => new Date(b.dose.scheduled_at) - new Date(a.dose.scheduled_at))
            setHistory(merged)
        } catch (err) {
            showError(err.message, 'No se pudo cargar el paciente')
        } finally {
            setLoading(false)
        }
    }, [patientId]) // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        fetchData()
    }, [fetchData])

    const handleToggle = async (treatment) => {
        try {
            if (treatment.is_active) {
                await medicinesApi.deactivateTreatment(treatment.id)
                showSuccess(`${treatment.medicine_name} pausado`)
            } else {
                await medicinesApi.activateTreatment(treatment.id)
                showSuccess(`${treatment.medicine_name} reanudado`)
            }
            fetchData()
        } catch (err) {
            showError(err.message, 'No se pudo actualizar el tratamiento')
        }
    }

    const handleDelete = async (treatment) => {
        if (!window.confirm(`¿Eliminar el tratamiento de ${treatment.medicine_name}?`)) {
            return
        }
        try {
            await medicinesApi.deleteTreatment(treatment.id)
            showSuccess('Tratamiento eliminado')
            fetchData()
        } catch (err) {
            showError(err.message, 'No se pudo eliminar')
        }
    }

    if (loading) {
        return (
            <div className="medicines-page">
                <div className="medicines-loading"><Loader2 size={20} className="spin" /> Cargando…</div>
            </div>
        )
    }

    return (
        <div className="medicines-page">
            <Link to={`/portal/${tenantId}/medicines`} className="medicines-back">
                <ChevronLeft size={16} /> Pacientes
            </Link>

            <header className="medicines-header">
                <div>
                    <h1>
                        {patient?.kind === 'animal' ? <PawPrint size={22} /> : <User size={22} />}
                        {' '}{patient?.name}
                    </h1>
                    {patient?.notes && <p className="medicines-subtitle">{patient.notes}</p>}
                </div>
                <Link
                    to={`/portal/${tenantId}/medicines/patients/${patientId}/treatments/new`}
                    className="medicines-button primary"
                >
                    <Plus size={16} /> Nuevo tratamiento
                </Link>
            </header>

            <div className="medicines-tabs">
                <button
                    type="button"
                    className={`medicines-tab${tab === 'treatments' ? ' active' : ''}`}
                    onClick={() => setSearchParams({})}
                >
                    Tratamientos
                </button>
                <button
                    type="button"
                    className={`medicines-tab${tab === 'history' ? ' active' : ''}`}
                    onClick={() => setSearchParams({ tab: 'history' })}
                >
                    Historial
                </button>
            </div>

            {tab === 'treatments' ? (
                treatments.length === 0 ? (
                    <div className="medicines-empty">
                        <p>Sin tratamientos todavía.</p>
                    </div>
                ) : (
                    <ul className="medicines-list">
                        {treatments.map(treatment => (
                            <li key={treatment.id} className={`medicines-card${treatment.is_active ? '' : ' inactive'}`}>
                                <span className="medicines-card-text">
                                    <strong>{treatment.medicine_name}</strong>
                                    <span className="medicines-card-notes">
                                        {treatment.dose_text}, {describeFrequency(treatment.schedule)}
                                        {treatment.ends_at && ` · hasta ${formatDateTime(treatment.ends_at)}`}
                                        {treatment.total_doses && ` · ${treatment.total_doses} dosis`}
                                    </span>
                                </span>
                                <button
                                    type="button"
                                    className="medicines-icon-button"
                                    onClick={() => handleToggle(treatment)}
                                    aria-label={treatment.is_active ? 'Pausar' : 'Reanudar'}
                                >
                                    {treatment.is_active ? <Pause size={16} /> : <Play size={16} />}
                                </button>
                                <button
                                    type="button"
                                    className="medicines-icon-button danger"
                                    onClick={() => handleDelete(treatment)}
                                    aria-label="Eliminar"
                                >
                                    <Trash2 size={16} />
                                </button>
                            </li>
                        ))}
                    </ul>
                )
            ) : history.length === 0 ? (
                <div className="medicines-empty">
                    <p>Sin dosis registradas todavía.</p>
                </div>
            ) : (
                <ul className="medicines-list">
                    {history.map(({ dose, treatment }) => {
                        const StatusIcon = STATUS_ICONS[dose.status]
                        return (
                            <li key={dose.id} className={`medicines-card history ${dose.is_overdue ? 'overdue' : ''}`}>
                                <span className="medicines-agenda-time">{formatDateTime(dose.scheduled_at)}</span>
                                <span className="medicines-card-text">
                                    <strong>{treatment.medicine_name}</strong>
                                    <span className="medicines-card-notes">
                                        {dose.notes || treatment.dose_text}
                                    </span>
                                </span>
                                <span className={`medicines-status medicines-status-${dose.status}`}>
                                    {StatusIcon && <StatusIcon size={14} />}
                                    {STATUS_LABELS[dose.status]}
                                </span>
                            </li>
                        )
                    })}
                </ul>
            )}
        </div>
    )
}

export default PatientDetail
