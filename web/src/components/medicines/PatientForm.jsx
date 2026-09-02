import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ChevronLeft, Loader2 } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import './Medicines.css'

const PatientForm = () => {
    const { tenantId } = useParams()
    const navigate = useNavigate()
    const { showSuccess, showError } = useNotification()
    const [name, setName] = useState('')
    const [kind, setKind] = useState('human')
    const [notes, setNotes] = useState('')
    const [saving, setSaving] = useState(false)

    const handleSubmit = async (event) => {
        event.preventDefault()
        try {
            setSaving(true)
            const patient = await medicinesApi.createPatient({
                tenant_id: tenantId,
                name: name.trim(),
                kind,
                notes: notes.trim()
            })
            showSuccess(`${patient.name} agregado`)
            navigate(`/portal/${tenantId}/medicines/patients/${patient.id}`)
        } catch (err) {
            showError(err.message, 'No se pudo crear el paciente')
        } finally {
            setSaving(false)
        }
    }

    return (
        <div className="medicines-page">
            <Link to={`/portal/${tenantId}/medicines`} className="medicines-back">
                <ChevronLeft size={16} /> Pacientes
            </Link>
            <h1>Nuevo paciente</h1>

            <form className="medicines-form" onSubmit={handleSubmit}>
                <label className="medicines-field">
                    <span>Nombre</span>
                    <input
                        type="text"
                        value={name}
                        onChange={e => setName(e.target.value)}
                        required
                        placeholder="Luna"
                    />
                </label>

                <fieldset className="medicines-field">
                    <legend>Tipo</legend>
                    <div className="medicines-radio-group">
                        <label>
                            <input
                                type="radio"
                                name="kind"
                                value="human"
                                checked={kind === 'human'}
                                onChange={e => setKind(e.target.value)}
                            />
                            Persona
                        </label>
                        <label>
                            <input
                                type="radio"
                                name="kind"
                                value="animal"
                                checked={kind === 'animal'}
                                onChange={e => setKind(e.target.value)}
                            />
                            Animal
                        </label>
                    </div>
                </fieldset>

                <label className="medicines-field">
                    <span>Notas</span>
                    <input
                        type="text"
                        value={notes}
                        onChange={e => setNotes(e.target.value)}
                        placeholder="perra, 12 kg"
                    />
                </label>

                <button type="submit" className="medicines-button primary" disabled={saving || !name.trim()}>
                    {saving ? <><Loader2 size={16} className="spin" /> Guardando…</> : 'Guardar'}
                </button>
            </form>
        </div>
    )
}

export default PatientForm
