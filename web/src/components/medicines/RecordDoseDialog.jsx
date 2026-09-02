import { useState } from 'react'
import { Check, SkipForward, X, Loader2 } from 'lucide-react'
import { describeDose, formatDateTime } from '../../utils/medicineSchedule'
import './Medicines.css'

// RecordDoseDialog confirms giving or skipping a dose, and takes an optional
// note about what actually happened.
const RecordDoseDialog = ({ entry, onClose, onResolved }) => {
    const [notes, setNotes] = useState('')
    const [saving, setSaving] = useState(false)

    if (!entry) {
        return null
    }

    const handle = async (action) => {
        try {
            setSaving(true)
            await onResolved(action, notes.trim() || null)
        } finally {
            setSaving(false)
        }
    }

    return (
        <div className="medicines-dialog-backdrop" role="dialog" aria-modal="true">
            <div className="medicines-dialog">
                <header className="medicines-dialog-header">
                    <h2>{entry.patient_name}</h2>
                    <button type="button" className="medicines-icon-button" onClick={onClose} aria-label="Cerrar">
                        <X size={18} />
                    </button>
                </header>

                <p className="medicines-dialog-dose">
                    <strong>{entry.dose_text || describeDose(entry.dose.quantity, entry.dose.unit)}</strong>
                    {' de '}{entry.medicine_name}
                </p>
                <p className="medicines-dialog-time">{formatDateTime(entry.dose.scheduled_at)}</p>

                <label className="medicines-field">
                    <span>Notas</span>
                    <input
                        type="text"
                        value={notes}
                        onChange={e => setNotes(e.target.value)}
                        placeholder="opcional"
                    />
                </label>

                <div className="medicines-dialog-actions">
                    <button
                        type="button"
                        className="medicines-button secondary"
                        onClick={() => handle('skip')}
                        disabled={saving}
                    >
                        <SkipForward size={16} /> Saltar
                    </button>
                    <button
                        type="button"
                        className="medicines-button primary"
                        onClick={() => handle('administer')}
                        disabled={saving}
                    >
                        {saving ? <Loader2 size={16} className="spin" /> : <Check size={16} />} Administrada
                    </button>
                </div>
            </div>
        </div>
    )
}

export default RecordDoseDialog
