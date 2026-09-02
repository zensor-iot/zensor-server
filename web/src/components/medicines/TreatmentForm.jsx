import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ChevronLeft, Loader2 } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import {
    DOSE_UNITS,
    INTERVAL_UNITS,
    previewOccurrences,
    totalDosesForDuration,
    formatDateTime
} from '../../utils/medicineSchedule'
import './Medicines.css'

// The value an <input type="datetime-local"> expects, defaulted to the next
// whole hour so the common case needs no typing.
function defaultStartAt() {
    const next = new Date()
    next.setMinutes(0, 0, 0)
    next.setHours(next.getHours() + 1)
    const offset = next.getTimezoneOffset() * 60000
    return new Date(next.getTime() - offset).toISOString().slice(0, 16)
}

const TreatmentForm = () => {
    const { tenantId, patientId } = useParams()
    const navigate = useNavigate()
    const { showSuccess, showError } = useNotification()

    const [medicineName, setMedicineName] = useState('')
    const [quantity, setQuantity] = useState('1')
    const [unit, setUnit] = useState('pill')
    const [every, setEvery] = useState('8')
    const [intervalUnit, setIntervalUnit] = useState('hour')
    const [startAt, setStartAt] = useState(defaultStartAt)
    const [endMode, setEndMode] = useState('none')
    const [durationDays, setDurationDays] = useState('7')
    const [endDate, setEndDate] = useState('')
    const [notes, setNotes] = useState('')
    const [saving, setSaving] = useState(false)

    const schedule = {
        start_at: startAt ? new Date(startAt).toISOString() : '',
        every: Number(every),
        unit: intervalUnit
    }
    const preview = previewOccurrences(schedule, 3)
    const derivedTotalDoses = totalDosesForDuration(durationDays, every, intervalUnit)

    const handleSubmit = async (event) => {
        event.preventDefault()

        const payload = {
            tenant_id: tenantId,
            patient_id: patientId,
            medicine_name: medicineName.trim(),
            quantity: Number(quantity),
            unit,
            schedule,
            notes: notes.trim()
        }

        if (endMode === 'doses' && derivedTotalDoses) {
            payload.total_doses = derivedTotalDoses
        }
        if (endMode === 'date' && endDate) {
            payload.end_at = new Date(endDate).toISOString()
        }

        try {
            setSaving(true)
            await medicinesApi.createTreatment(payload)
            showSuccess(`Tratamiento de ${payload.medicine_name} creado`)
            navigate(`/portal/${tenantId}/medicines/patients/${patientId}`)
        } catch (err) {
            showError(err.message, 'No se pudo crear el tratamiento')
        } finally {
            setSaving(false)
        }
    }

    return (
        <div className="medicines-page">
            <Link to={`/portal/${tenantId}/medicines/patients/${patientId}`} className="medicines-back">
                <ChevronLeft size={16} /> Paciente
            </Link>
            <h1>Nuevo tratamiento</h1>

            <form className="medicines-form" onSubmit={handleSubmit}>
                <label className="medicines-field">
                    <span>Medicina</span>
                    <input
                        type="text"
                        value={medicineName}
                        onChange={e => setMedicineName(e.target.value)}
                        required
                        placeholder="Amoxicilina"
                    />
                </label>

                <div className="medicines-field-row">
                    <label className="medicines-field">
                        <span>Cantidad</span>
                        <input
                            type="number"
                            min="0"
                            step="0.5"
                            value={quantity}
                            onChange={e => setQuantity(e.target.value)}
                            required
                        />
                    </label>
                    <label className="medicines-field">
                        <span>Unidad</span>
                        <select value={unit} onChange={e => setUnit(e.target.value)}>
                            {DOSE_UNITS.map(u => (
                                <option key={u.value} value={u.value}>{u.label}</option>
                            ))}
                        </select>
                    </label>
                </div>

                <div className="medicines-field-row">
                    <label className="medicines-field">
                        <span>Cada</span>
                        <input
                            type="number"
                            min="1"
                            value={every}
                            onChange={e => setEvery(e.target.value)}
                            required
                        />
                    </label>
                    <label className="medicines-field">
                        <span>&nbsp;</span>
                        <select value={intervalUnit} onChange={e => setIntervalUnit(e.target.value)}>
                            {INTERVAL_UNITS.map(u => (
                                <option key={u.value} value={u.value}>{u.label}</option>
                            ))}
                        </select>
                    </label>
                </div>

                <label className="medicines-field">
                    <span>Primera dosis</span>
                    <input
                        type="datetime-local"
                        value={startAt}
                        onChange={e => setStartAt(e.target.value)}
                        required
                    />
                </label>

                <fieldset className="medicines-field">
                    <legend>Hasta cuándo</legend>
                    <div className="medicines-radio-group column">
                        <label>
                            <input
                                type="radio"
                                name="endMode"
                                value="none"
                                checked={endMode === 'none'}
                                onChange={e => setEndMode(e.target.value)}
                            />
                            Sin fin definido
                        </label>
                        <label>
                            <input
                                type="radio"
                                name="endMode"
                                value="doses"
                                checked={endMode === 'doses'}
                                onChange={e => setEndMode(e.target.value)}
                            />
                            Durante
                            <input
                                type="number"
                                min="1"
                                className="medicines-inline-input"
                                value={durationDays}
                                onChange={e => setDurationDays(e.target.value)}
                                disabled={endMode !== 'doses'}
                            />
                            días
                            {endMode === 'doses' && derivedTotalDoses && (
                                <span className="medicines-hint">({derivedTotalDoses} dosis)</span>
                            )}
                        </label>
                        <label>
                            <input
                                type="radio"
                                name="endMode"
                                value="date"
                                checked={endMode === 'date'}
                                onChange={e => setEndMode(e.target.value)}
                            />
                            Hasta el
                            <input
                                type="datetime-local"
                                className="medicines-inline-input wide"
                                value={endDate}
                                onChange={e => setEndDate(e.target.value)}
                                disabled={endMode !== 'date'}
                            />
                        </label>
                    </div>
                </fieldset>

                <label className="medicines-field">
                    <span>Notas</span>
                    <input
                        type="text"
                        value={notes}
                        onChange={e => setNotes(e.target.value)}
                        placeholder="con comida"
                    />
                </label>

                {preview.length > 0 && (
                    <div className="medicines-preview">
                        <span className="medicines-preview-title">Primeras dosis</span>
                        <ul>
                            {preview.map((occurrence, index) => (
                                <li key={index}>{formatDateTime(occurrence)}</li>
                            ))}
                        </ul>
                    </div>
                )}

                <button
                    type="submit"
                    className="medicines-button primary"
                    disabled={saving || !medicineName.trim()}
                >
                    {saving ? <><Loader2 size={16} className="spin" /> Guardando…</> : 'Crear tratamiento'}
                </button>
            </form>
        </div>
    )
}

export default TreatmentForm
