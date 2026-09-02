import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { CalendarClock, ChevronLeft, Loader2, PawPrint, User, AlertCircle, Check, SkipForward } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { formatTime, formatDateTime } from '../../utils/medicineSchedule'
import RecordDoseDialog from './RecordDoseDialog'
import './Medicines.css'

const STATUS_ICONS = {
    administered: Check,
    skipped: SkipForward
}

// MedicineAgenda shows what the family has to give next. It reads the whole
// window from the server in one request rather than fetching per treatment.
const MedicineAgenda = () => {
    const { tenantId } = useParams()
    const { showSuccess, showError } = useNotification()
    const [entries, setEntries] = useState([])
    const [loading, setLoading] = useState(true)
    const [selected, setSelected] = useState(null)

    const fetchAgenda = useCallback(async () => {
        try {
            setLoading(true)
            // A little into the past so an overdue dose is still actionable.
            const from = new Date(Date.now() - 12 * 3600 * 1000).toISOString()
            const to = new Date(Date.now() + 24 * 3600 * 1000).toISOString()
            const data = await medicinesApi.listAgenda(tenantId, from, to)
            setEntries(data.data || [])
        } catch (err) {
            showError(err.message, 'No se pudo cargar la agenda')
        } finally {
            setLoading(false)
        }
    }, [tenantId]) // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        fetchAgenda()
    }, [fetchAgenda])

    const handleResolve = async (action, notes) => {
        try {
            if (action === 'administer') {
                await medicinesApi.administerDose(selected.dose.id, notes)
                showSuccess(`${selected.dose_text} de ${selected.medicine_name} registrada`)
            } else {
                await medicinesApi.skipDose(selected.dose.id, notes)
                showSuccess('Dosis saltada')
            }
            setSelected(null)
            fetchAgenda()
        } catch (err) {
            showError(err.message, 'No se pudo registrar la dosis')
        }
    }

    return (
        <div className="medicines-page">
            <Link to={`/portal/${tenantId}/medicines`} className="medicines-back">
                <ChevronLeft size={16} /> Pacientes
            </Link>
            <h1><CalendarClock size={24} /> Agenda</h1>
            <p className="medicines-subtitle">Lo que hay que dar hoy</p>

            {loading ? (
                <div className="medicines-loading"><Loader2 size={20} className="spin" /> Cargando…</div>
            ) : entries.length === 0 ? (
                <div className="medicines-empty">
                    <CalendarClock size={32} />
                    <p>No hay dosis en las próximas 24 horas.</p>
                </div>
            ) : (
                <ul className="medicines-list">
                    {entries.map(entry => {
                        const StatusIcon = STATUS_ICONS[entry.dose.status]
                        const pending = entry.dose.status === 'pending'
                        return (
                            <li
                                key={entry.dose.id}
                                className={`medicines-card agenda ${entry.dose.is_overdue ? 'overdue' : ''}`}
                            >
                                <span className="medicines-agenda-time">
                                    {formatTime(entry.dose.scheduled_at)}
                                    {entry.dose.is_overdue && (
                                        <span className="medicines-overdue-badge">
                                            <AlertCircle size={12} /> vencida
                                        </span>
                                    )}
                                </span>

                                <span className="medicines-card-text">
                                    <strong>
                                        {entry.patient_kind === 'animal' ? <PawPrint size={14} /> : <User size={14} />}
                                        {' '}{entry.patient_name}
                                    </strong>
                                    <span className="medicines-card-notes">
                                        {entry.dose_text} de {entry.medicine_name}
                                    </span>
                                </span>

                                {pending ? (
                                    <button
                                        type="button"
                                        className="medicines-button primary small"
                                        onClick={() => setSelected(entry)}
                                    >
                                        Registrar
                                    </button>
                                ) : (
                                    <span className={`medicines-status medicines-status-${entry.dose.status}`}>
                                        {StatusIcon && <StatusIcon size={14} />}
                                        {entry.dose.status === 'administered' ? 'Administrada' : 'Saltada'}
                                    </span>
                                )}
                            </li>
                        )
                    })}
                </ul>
            )}

            <RecordDoseDialog
                entry={selected}
                onClose={() => setSelected(null)}
                onResolved={handleResolve}
            />

            {entries.length > 0 && (
                <p className="medicines-footnote">
                    Ventana: {formatDateTime(new Date(Date.now() - 12 * 3600 * 1000))} — {formatDateTime(new Date(Date.now() + 24 * 3600 * 1000))}
                </p>
            )}
        </div>
    )
}

export default MedicineAgenda
