import { useState } from 'react'
import { CheckCircle } from 'lucide-react'
import { maintenanceApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import './Maintenance.css'

// Completion dialog: completed_by is auto-filled from the logged-in user and
// the activity's custom field definitions become typed inputs captured into
// field_values.
const CompleteExecutionDialog = ({ activity, execution, userEmail, onClose, onCompleted }) => {
    const { showSuccess, showError } = useNotification()
    const fields = activity?.fields || []
    const [completedBy, setCompletedBy] = useState(userEmail || '')
    const [values, setValues] = useState(() => {
        const initial = {}
        fields.forEach(f => {
            if (f.default_value != null) {
                initial[f.name] = f.type === 'boolean' ? f.default_value === 'true' : f.default_value
            } else if (f.type === 'boolean') {
                initial[f.name] = false
            } else {
                initial[f.name] = ''
            }
        })
        return initial
    })
    const [submitting, setSubmitting] = useState(false)

    const setValue = (name, value) => setValues(prev => ({ ...prev, [name]: value }))

    const handleSubmit = async (e) => {
        e.preventDefault()

        if (!completedBy.trim()) {
            showError('Completed by is required')
            return
        }
        for (const field of fields) {
            if (field.is_required && field.type !== 'boolean' && String(values[field.name] ?? '').trim() === '') {
                showError(`${field.display_name} is required`)
                return
            }
        }

        const fieldValues = {}
        fields.forEach(f => {
            const raw = values[f.name]
            if (f.type === 'boolean') {
                fieldValues[f.name] = Boolean(raw)
                return
            }
            if (String(raw ?? '').trim() === '') {
                return
            }
            fieldValues[f.name] = f.type === 'number' ? Number(raw) : raw
        })

        setSubmitting(true)
        try {
            await maintenanceApi.completeExecution(execution.id, completedBy, fieldValues)
            showSuccess('Execution marked as completed')
            onCompleted()
        } catch (err) {
            showError(err.message, 'Failed to complete execution')
        } finally {
            setSubmitting(false)
        }
    }

    const renderFieldInput = (field) => {
        switch (field.type) {
            case 'boolean':
                return (
                    <label style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                        <input
                            type="checkbox"
                            checked={Boolean(values[field.name])}
                            onChange={(e) => setValue(field.name, e.target.checked)}
                        />
                        Yes
                    </label>
                )
            case 'number':
                return (
                    <input
                        type="number"
                        value={values[field.name]}
                        onChange={(e) => setValue(field.name, e.target.value)}
                    />
                )
            case 'date':
                return (
                    <input
                        type="date"
                        value={values[field.name]}
                        onChange={(e) => setValue(field.name, e.target.value)}
                    />
                )
            default:
                return (
                    <input
                        type="text"
                        value={values[field.name]}
                        onChange={(e) => setValue(field.name, e.target.value)}
                    />
                )
        }
    }

    return (
        <div className="maintenance-dialog-backdrop" onClick={onClose}>
            <div className="maintenance-dialog" onClick={(e) => e.stopPropagation()}>
                <h3>
                    <CheckCircle size={20} />
                    Complete “{activity?.name}”
                </h3>

                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                    <div className="maintenance-form-field">
                        <label htmlFor="completed-by">Completed by</label>
                        <input
                            id="completed-by"
                            type="text"
                            value={completedBy}
                            onChange={(e) => setCompletedBy(e.target.value)}
                        />
                    </div>

                    {fields.map(field => (
                        <div key={field.name} className="maintenance-form-field">
                            <label>
                                {field.display_name}
                                {field.is_required && ' *'}
                            </label>
                            {renderFieldInput(field)}
                        </div>
                    ))}

                    <div className="maintenance-form-actions">
                        <button type="button" className="maintenance-btn" onClick={onClose} disabled={submitting}>
                            Cancel
                        </button>
                        <button type="submit" className="maintenance-btn primary" disabled={submitting}>
                            {submitting ? 'Completing...' : 'Complete'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}

export default CompleteExecutionDialog
