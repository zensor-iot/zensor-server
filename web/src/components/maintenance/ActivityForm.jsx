import { useState } from 'react'
import { Plus, X, Trash2 } from 'lucide-react'
import {
    FREQUENCIES,
    WEEKDAYS,
    MONTHS,
    buildCronExpression,
    parseCronExpression
} from '../../utils/maintenanceSchedule'
import { maintenanceApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import './Maintenance.css'

const ACTIVITY_TYPES = [
    { value: 'water_system', label: 'Water System' },
    { value: 'car', label: 'Car' },
    { value: 'pets', label: 'Pets' },
    { value: 'custom', label: 'Custom' }
]

const FIELD_TYPES = ['text', 'number', 'date', 'boolean']

const capitalize = (value) => value.charAt(0).toUpperCase() + value.slice(1)

// Create/edit form for a maintenance activity. In edit mode the type is
// read-only and only changed fields are sent, matching the server contract.
const ActivityForm = ({ tenantId, activity = null, onSaved, onCancel }) => {
    const isEdit = Boolean(activity)
    const { showSuccess, showError } = useNotification()

    const parsedSchedule = isEdit ? parseCronExpression(activity.schedule) : null
    const scheduleUnparsable = isEdit && !parsedSchedule

    const [name, setName] = useState(activity?.name || '')
    const [description, setDescription] = useState(activity?.description || '')
    const [typeName, setTypeName] = useState(activity?.type_name || 'water_system')
    const [customTypeName, setCustomTypeName] = useState(activity?.custom_type_name || '')
    const [useRawCron, setUseRawCron] = useState(scheduleUnparsable)
    const [rawCron, setRawCron] = useState(activity?.schedule || '')
    const [scheduleForm, setScheduleForm] = useState(parsedSchedule || {
        frequency: 'weekly',
        dayOfWeek: 1,
        dayOfMonth: 1,
        month: 1
    })
    const [reminderDays, setReminderDays] = useState(activity?.notification_days_before || [1])
    const [reminderDayInput, setReminderDayInput] = useState('')
    const [fields, setFields] = useState(
        (activity?.fields || []).map(f => ({
            name: f.name,
            display_name: f.display_name,
            type: f.type,
            is_required: f.is_required,
            default_value: f.default_value ?? ''
        }))
    )
    const [saving, setSaving] = useState(false)

    const schedule = useRawCron ? rawCron : buildCronExpression(scheduleForm)

    const updateScheduleForm = (patch) => setScheduleForm(prev => ({ ...prev, ...patch }))

    const addReminderDay = () => {
        const day = parseInt(reminderDayInput, 10)
        if (Number.isNaN(day) || day < 0 || reminderDays.includes(day)) {
            return
        }
        setReminderDays([...reminderDays, day].sort((a, b) => a - b))
        setReminderDayInput('')
    }

    const removeReminderDay = (day) => {
        setReminderDays(reminderDays.filter(d => d !== day))
    }

    const addField = () => {
        setFields([...fields, { name: '', display_name: '', type: 'text', is_required: false, default_value: '' }])
    }

    const updateField = (index, patch) => {
        setFields(fields.map((f, i) => (i === index ? { ...f, ...patch } : f)))
    }

    const removeField = (index) => {
        setFields(fields.filter((_, i) => i !== index))
    }

    const buildFieldsPayload = () =>
        fields.map(f => {
            const field = {
                name: f.name,
                display_name: f.display_name,
                type: f.type,
                is_required: f.is_required
            }
            if (f.default_value !== '') {
                field.default_value = String(f.default_value)
            }
            return field
        })

    const handleSubmit = async (e) => {
        e.preventDefault()

        if (!name.trim()) {
            showError('Name is required')
            return
        }
        if (!isEdit && typeName === 'custom' && !customTypeName.trim()) {
            showError('Custom type name is required')
            return
        }
        if (fields.some(f => !f.name.trim() || !f.display_name.trim())) {
            showError('Every custom field needs a name and a display name')
            return
        }

        setSaving(true)
        try {
            if (isEdit) {
                const updates = {}
                if (name !== activity.name) updates.name = name
                if (description !== activity.description) updates.description = description
                if (schedule !== activity.schedule) updates.schedule = schedule
                const originalReminders = activity.notification_days_before || []
                if (JSON.stringify(reminderDays) !== JSON.stringify(originalReminders)) {
                    updates.notification_days_before = reminderDays
                }
                const fieldsPayload = buildFieldsPayload()
                const originalFields = (activity.fields || []).map(f => ({
                    name: f.name,
                    display_name: f.display_name,
                    type: f.type,
                    is_required: f.is_required,
                    ...(f.default_value != null ? { default_value: f.default_value } : {})
                }))
                if (JSON.stringify(fieldsPayload) !== JSON.stringify(originalFields)) {
                    updates.fields = fieldsPayload
                }

                if (Object.keys(updates).length === 0) {
                    showSuccess('Nothing to save')
                } else {
                    await maintenanceApi.updateActivity(activity.id, updates)
                    showSuccess('Activity updated')
                }
            } else {
                const payload = {
                    tenant_id: tenantId,
                    type_name: typeName,
                    name,
                    description,
                    schedule,
                    notification_days_before: reminderDays,
                    fields: buildFieldsPayload()
                }
                if (typeName === 'custom') {
                    payload.custom_type_name = customTypeName
                }
                await maintenanceApi.createActivity(payload)
                showSuccess('Activity created')
            }
            onSaved()
        } catch (err) {
            showError(err.message, isEdit ? 'Failed to update activity' : 'Failed to create activity')
        } finally {
            setSaving(false)
        }
    }

    return (
        <form className="maintenance-form" onSubmit={handleSubmit}>
            <div className="maintenance-form-row">
                <div className="maintenance-form-field">
                    <label htmlFor="activity-type">Type</label>
                    <select
                        id="activity-type"
                        value={typeName}
                        onChange={(e) => setTypeName(e.target.value)}
                        disabled={isEdit}
                    >
                        {ACTIVITY_TYPES.map(t => (
                            <option key={t.value} value={t.value}>{t.label}</option>
                        ))}
                    </select>
                </div>
                {typeName === 'custom' && (
                    <div className="maintenance-form-field">
                        <label htmlFor="custom-type-name">Custom type name</label>
                        <input
                            id="custom-type-name"
                            type="text"
                            value={customTypeName}
                            onChange={(e) => setCustomTypeName(e.target.value)}
                            disabled={isEdit}
                            placeholder="e.g., Greenhouse"
                        />
                    </div>
                )}
            </div>

            <div className="maintenance-form-field">
                <label htmlFor="activity-name">Name</label>
                <input
                    id="activity-name"
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g., Water Filter Replacement"
                />
            </div>

            <div className="maintenance-form-field">
                <label htmlFor="activity-description">Description</label>
                <textarea
                    id="activity-description"
                    rows={3}
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="What does this maintenance involve?"
                />
            </div>

            <div className="maintenance-form-field">
                <label>Schedule</label>
                {scheduleUnparsable && !useRawCron && (
                    <div className="maintenance-form-warning">
                        The existing schedule could not be converted to the friendly editor; saving will replace it.
                    </div>
                )}
                {useRawCron ? (
                    <>
                        {scheduleUnparsable && (
                            <div className="maintenance-form-warning">
                                This schedule uses a cron expression the friendly editor cannot represent. Edit it as raw cron below.
                            </div>
                        )}
                        <input
                            type="text"
                            value={rawCron}
                            onChange={(e) => setRawCron(e.target.value)}
                            placeholder="0 9 * * 1"
                        />
                        <button type="button" className="maintenance-btn" onClick={() => setUseRawCron(false)}>
                            Use friendly editor
                        </button>
                    </>
                ) : (
                    <div className="maintenance-form-row">
                        <div className="maintenance-form-field">
                            <label htmlFor="schedule-frequency">Frequency</label>
                            <select
                                id="schedule-frequency"
                                value={scheduleForm.frequency}
                                onChange={(e) => updateScheduleForm({ frequency: e.target.value })}
                            >
                                {FREQUENCIES.map(f => (
                                    <option key={f} value={f}>{capitalize(f)}</option>
                                ))}
                            </select>
                        </div>
                        {scheduleForm.frequency === 'weekly' && (
                            <div className="maintenance-form-field">
                                <label htmlFor="schedule-weekday">Day of week</label>
                                <select
                                    id="schedule-weekday"
                                    value={scheduleForm.dayOfWeek}
                                    onChange={(e) => updateScheduleForm({ dayOfWeek: Number(e.target.value) })}
                                >
                                    {WEEKDAYS.map(d => (
                                        <option key={d.value} value={d.value}>{d.label}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {scheduleForm.frequency !== 'weekly' && (
                            <div className="maintenance-form-field">
                                <label htmlFor="schedule-day">Day of month</label>
                                <select
                                    id="schedule-day"
                                    value={scheduleForm.dayOfMonth}
                                    onChange={(e) => updateScheduleForm({ dayOfMonth: Number(e.target.value) })}
                                >
                                    {Array.from({ length: 28 }, (_, i) => i + 1).map(d => (
                                        <option key={d} value={d}>{d}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {scheduleForm.frequency === 'yearly' && (
                            <div className="maintenance-form-field">
                                <label htmlFor="schedule-month">Month</label>
                                <select
                                    id="schedule-month"
                                    value={scheduleForm.month}
                                    onChange={(e) => updateScheduleForm({ month: Number(e.target.value) })}
                                >
                                    {MONTHS.map(m => (
                                        <option key={m.value} value={m.value}>{m.label}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                    </div>
                )}
            </div>

            <div className="maintenance-form-field">
                <label>Remind me (days before)</label>
                <div className="maintenance-chip-list">
                    {reminderDays.map(day => (
                        <span key={day} className="maintenance-chip">
                            {day} {day === 1 ? 'day' : 'days'}
                            <button type="button" onClick={() => removeReminderDay(day)} aria-label={`Remove ${day} days`}>
                                <X size={14} />
                            </button>
                        </span>
                    ))}
                    <input
                        type="number"
                        min="0"
                        style={{ width: '80px' }}
                        value={reminderDayInput}
                        onChange={(e) => setReminderDayInput(e.target.value)}
                        placeholder="days"
                    />
                    <button type="button" className="maintenance-btn" onClick={addReminderDay}>
                        <Plus size={14} />
                        Add
                    </button>
                </div>
            </div>

            <div className="maintenance-form-field">
                <label>Custom fields (captured on completion)</label>
                {fields.map((field, index) => (
                    <div key={index} className="maintenance-field-def">
                        <div className="maintenance-form-row">
                            <div className="maintenance-form-field">
                                <label>Name</label>
                                <input
                                    type="text"
                                    value={field.name}
                                    onChange={(e) => updateField(index, { name: e.target.value })}
                                    placeholder="filter_type"
                                />
                            </div>
                            <div className="maintenance-form-field">
                                <label>Display name</label>
                                <input
                                    type="text"
                                    value={field.display_name}
                                    onChange={(e) => updateField(index, { display_name: e.target.value })}
                                    placeholder="Filter Type"
                                />
                            </div>
                            <div className="maintenance-form-field">
                                <label>Type</label>
                                <select
                                    value={field.type}
                                    onChange={(e) => updateField(index, { type: e.target.value })}
                                >
                                    {FIELD_TYPES.map(t => (
                                        <option key={t} value={t}>{capitalize(t)}</option>
                                    ))}
                                </select>
                            </div>
                            <div className="maintenance-form-field">
                                <label>Default value</label>
                                <input
                                    type="text"
                                    value={field.default_value}
                                    onChange={(e) => updateField(index, { default_value: e.target.value })}
                                />
                            </div>
                        </div>
                        <div className="maintenance-form-row" style={{ alignItems: 'center' }}>
                            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                                <input
                                    type="checkbox"
                                    checked={field.is_required}
                                    onChange={(e) => updateField(index, { is_required: e.target.checked })}
                                />
                                Required
                            </label>
                            <button type="button" className="maintenance-btn danger" onClick={() => removeField(index)}>
                                <Trash2 size={14} />
                                Remove field
                            </button>
                        </div>
                    </div>
                ))}
                <div>
                    <button type="button" className="maintenance-btn" onClick={addField}>
                        <Plus size={14} />
                        Add field
                    </button>
                </div>
            </div>

            <div className="maintenance-form-actions">
                <button type="button" className="maintenance-btn" onClick={onCancel} disabled={saving}>
                    Cancel
                </button>
                <button type="submit" className="maintenance-btn primary" disabled={saving}>
                    {saving ? 'Saving...' : isEdit ? 'Save changes' : 'Create activity'}
                </button>
            </div>
        </form>
    )
}

export default ActivityForm
