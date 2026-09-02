import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Pill, Plus, CalendarClock, Trash2, Loader2, User, PawPrint } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import './Medicines.css'

const MedicinePatients = () => {
    const { tenantId } = useParams()
    const { showSuccess, showError } = useNotification()
    const [patients, setPatients] = useState([])
    const [loading, setLoading] = useState(true)

    const fetchPatients = useCallback(async () => {
        try {
            setLoading(true)
            const data = await medicinesApi.listPatients(tenantId)
            setPatients(data.data || [])
        } catch (err) {
            showError(err.message, 'No se pudieron cargar los pacientes')
        } finally {
            setLoading(false)
        }
    }, [tenantId]) // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        fetchPatients()
    }, [fetchPatients])

    const handleDelete = async (patient) => {
        if (!window.confirm(`¿Eliminar a ${patient.name}?`)) {
            return
        }
        try {
            await medicinesApi.deletePatient(patient.id)
            showSuccess(`${patient.name} eliminado`)
            fetchPatients()
        } catch (err) {
            showError(err.message, 'No se pudo eliminar')
        }
    }

    return (
        <div className="medicines-page">
            <header className="medicines-header">
                <div>
                    <h1><Pill size={24} /> Medicinas</h1>
                    <p className="medicines-subtitle">Quién toma qué, y cada cuánto</p>
                </div>
                <div className="medicines-header-actions">
                    <Link to={`/portal/${tenantId}/medicines/agenda`} className="medicines-button secondary">
                        <CalendarClock size={16} /> Agenda
                    </Link>
                    <Link to={`/portal/${tenantId}/medicines/patients/new`} className="medicines-button primary">
                        <Plus size={16} /> Nuevo paciente
                    </Link>
                </div>
            </header>

            {loading ? (
                <div className="medicines-loading"><Loader2 size={20} className="spin" /> Cargando…</div>
            ) : patients.length === 0 ? (
                <div className="medicines-empty">
                    <Pill size={32} />
                    <p>Todavía no hay pacientes.</p>
                    <Link to={`/portal/${tenantId}/medicines/patients/new`} className="medicines-button primary">
                        <Plus size={16} /> Agregar el primero
                    </Link>
                </div>
            ) : (
                <ul className="medicines-list">
                    {patients.map(patient => (
                        <li key={patient.id} className="medicines-card">
                            <Link
                                to={`/portal/${tenantId}/medicines/patients/${patient.id}`}
                                className="medicines-card-main"
                            >
                                <span className={`medicines-kind medicines-kind-${patient.kind}`}>
                                    {patient.kind === 'animal' ? <PawPrint size={18} /> : <User size={18} />}
                                </span>
                                <span className="medicines-card-text">
                                    <strong>{patient.name}</strong>
                                    {patient.notes && <span className="medicines-card-notes">{patient.notes}</span>}
                                </span>
                            </Link>
                            <button
                                type="button"
                                className="medicines-icon-button danger"
                                onClick={() => handleDelete(patient)}
                                aria-label={`Eliminar a ${patient.name}`}
                            >
                                <Trash2 size={16} />
                            </button>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}

export default MedicinePatients
