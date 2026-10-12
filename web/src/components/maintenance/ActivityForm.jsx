import { useState } from 'react'
import { Plus, Trash2, X } from 'lucide-react'
import { Button, Checkbox, Field, Form, FormActions, Input, Section, Select, Textarea } from '../../ui'
import {
  ACTIVITY_TYPES,
  FIELD_TYPES,
  UNITS,
  buildSchedule,
  toLocalInputValue,
  toSchedulePayload,
  unitLabel,
} from '../../utils/maintenanceSchedule'
import { maintenanceApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'

const toFieldPayload = f => ({
  name: f.name,
  display_name: f.display_name,
  type: f.type,
  is_required: f.is_required,
  ...(f.default_value != null && f.default_value !== '' ? { default_value: String(f.default_value) } : {}),
})

/**
 * Create/edit form for a maintenance activity. In edit mode the type is read-only and
 * only changed fields are sent, matching the server contract.
 */
const ActivityForm = ({ tenantId, activity = null, onSaved, onCancel }) => {
  const isEdit = Boolean(activity)
  const { showSuccess, showError } = useNotification()

  const [name, setName] = useState(activity?.name || '')
  const [description, setDescription] = useState(activity?.description || '')
  const [typeName, setTypeName] = useState(activity?.type_name || 'water_system')
  const [customTypeName, setCustomTypeName] = useState(activity?.custom_type_name || '')
  const [schedule, setSchedule] = useState({
    startDate: toLocalInputValue(activity?.schedule?.start_date),
    every: activity?.schedule?.every || 1,
    unit: activity?.schedule?.unit || 'month',
  })
  const [reminderDays, setReminderDays] = useState(activity?.notification_days_before || [1])
  const [reminderDayInput, setReminderDayInput] = useState('')
  const [fields, setFields] = useState(
    (activity?.fields || []).map(f => ({
      name: f.name,
      display_name: f.display_name,
      type: f.type,
      is_required: f.is_required,
      default_value: f.default_value ?? '',
    })),
  )
  const [errors, setErrors] = useState({})
  const [saving, setSaving] = useState(false)

  const updateSchedule = patch => setSchedule(prev => ({ ...prev, ...patch }))

  const addReminderDay = () => {
    const day = parseInt(reminderDayInput, 10)
    if (Number.isNaN(day) || day < 0) {
      setErrors(prev => ({ ...prev, reminder: 'Escribe un número de días, 0 o más.' }))
      return
    }
    setErrors(prev => ({ ...prev, reminder: undefined }))
    if (!reminderDays.includes(day)) setReminderDays([...reminderDays, day].sort((a, b) => a - b))
    setReminderDayInput('')
  }

  const addField = () =>
    setFields([...fields, { name: '', display_name: '', type: 'text', is_required: false, default_value: '' }])
  const updateField = (index, patch) => setFields(fields.map((f, i) => (i === index ? { ...f, ...patch } : f)))
  const removeField = index => setFields(fields.filter((_, i) => i !== index))

  const validate = schedulePayload => {
    const next = {}
    if (!name.trim()) next.name = 'Escribe un nombre para la actividad.'
    if (!isEdit && typeName === 'custom' && !customTypeName.trim()) next.customTypeName = 'Escribe el nombre del tipo.'
    if (!schedulePayload) next.startDate = 'Elige cuándo toca la primera vez.'
    fields.forEach((f, i) => {
      if (!f.name.trim()) next[`field-${i}-name`] = 'Escribe una clave.'
      if (!f.display_name.trim()) next[`field-${i}-display`] = 'Escribe cómo se muestra.'
    })
    return next
  }

  const handleSubmit = async e => {
    e.preventDefault()
    const schedulePayload = buildSchedule(schedule)
    const nextErrors = validate(schedulePayload)
    setErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) return

    const fieldsPayload = fields.map(toFieldPayload)
    setSaving(true)
    try {
      if (isEdit) {
        const updates = {}
        if (name !== activity.name) updates.name = name
        if (description !== activity.description) updates.description = description
        if (JSON.stringify(schedulePayload) !== JSON.stringify(toSchedulePayload(activity.schedule))) {
          updates.schedule = schedulePayload
        }
        if (JSON.stringify(reminderDays) !== JSON.stringify(activity.notification_days_before || [])) {
          updates.notification_days_before = reminderDays
        }
        if (JSON.stringify(fieldsPayload) !== JSON.stringify((activity.fields || []).map(toFieldPayload))) {
          updates.fields = fieldsPayload
        }

        if (Object.keys(updates).length === 0) {
          showSuccess('No había cambios que guardar')
        } else {
          await maintenanceApi.updateActivity(activity.id, updates)
          showSuccess('Cambios guardados')
        }
      } else {
        const payload = {
          tenant_id: tenantId,
          type_name: typeName,
          name,
          description,
          schedule: schedulePayload,
          notification_days_before: reminderDays,
          fields: fieldsPayload,
        }
        if (typeName === 'custom') payload.custom_type_name = customTypeName
        await maintenanceApi.createActivity(payload)
        showSuccess('Actividad creada')
      }
      onSaved()
    } catch {
      showError(
        'Revisa tu conexión y vuelve a intentarlo.',
        isEdit ? 'No se pudieron guardar los cambios' : 'No se pudo crear la actividad',
      )
    } finally {
      setSaving(false)
    }
  }

  return (
    <Form onSubmit={handleSubmit} noValidate>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Tipo" hint={isEdit ? 'El tipo no se puede cambiar.' : undefined}>
          {props => (
            <Select {...props} value={typeName} onChange={e => setTypeName(e.target.value)} disabled={isEdit}>
              {ACTIVITY_TYPES.map(t => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </Select>
          )}
        </Field>
        {typeName === 'custom' && (
          <Field label="Nombre del tipo" required={!isEdit} error={errors.customTypeName}>
            {props => (
              <Input
                {...props}
                type="text"
                value={customTypeName}
                onChange={e => setCustomTypeName(e.target.value)}
                disabled={isEdit}
                placeholder="Por ejemplo, Invernadero"
              />
            )}
          </Field>
        )}
      </div>

      <Field label="Nombre" required error={errors.name}>
        {props => (
          <Input
            {...props}
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Por ejemplo, Cambio de filtro de agua"
          />
        )}
      </Field>

      <Field label="Descripción" hint="Opcional. Qué hay que hacer y con qué.">
        {props => <Textarea {...props} rows={3} value={description} onChange={e => setDescription(e.target.value)} />}
      </Field>

      <Section title="Calendario">
        <Field label="Primera vez" required error={errors.startDate}>
          {props => (
            <Input
              {...props}
              type="datetime-local"
              value={schedule.startDate}
              onChange={e => updateSchedule({ startDate: e.target.value })}
            />
          )}
        </Field>
        <div className="grid grid-cols-2 gap-4">
          <Field label="Repetir cada">
            {props => (
              <Input
                {...props}
                type="number"
                inputMode="numeric"
                min="1"
                value={schedule.every}
                onChange={e => updateSchedule({ every: Math.max(1, Number(e.target.value) || 1) })}
                className="font-mono"
              />
            )}
          </Field>
          <Field label="Unidad">
            {props => (
              <Select {...props} value={schedule.unit} onChange={e => updateSchedule({ unit: e.target.value })}>
                {UNITS.map(u => (
                  <option key={u.value} value={u.value}>
                    {unitLabel(u.value, schedule.every)}
                  </option>
                ))}
              </Select>
            )}
          </Field>
        </div>
      </Section>

      <Section
        title="Avisos"
        description="Zensor te avisa estos días antes de que toque. Usa 0 para el mismo día."
      >
        {reminderDays.length > 0 ? (
          <ul className="m-0 flex list-none flex-wrap gap-2 p-0">
            {reminderDays.map(day => (
              <li key={day} className="inline-flex items-center gap-1 rounded-pill bg-sunken pl-4 text-label">
                {day === 0 ? 'El mismo día' : `${day} ${day === 1 ? 'día' : 'días'} antes`}
                <button
                  type="button"
                  onClick={() => setReminderDays(reminderDays.filter(d => d !== day))}
                  aria-label={day === 0 ? 'Quitar aviso del mismo día' : `Quitar aviso de ${day} ${day === 1 ? 'día' : 'días'} antes`}
                  className="inline-flex min-h-touch min-w-touch items-center justify-center rounded-pill text-muted hover:text-ink"
                >
                  <X size={20} strokeWidth={1.8} aria-hidden="true" />
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="m-0 text-body-sm text-muted">Sin avisos. Agrega al menos uno para recibir recordatorios.</p>
        )}
        <div className="flex flex-wrap items-end gap-3">
          <Field label="Días antes" error={errors.reminder} className="w-[140px]">
            {props => (
              <Input
                {...props}
                type="number"
                inputMode="numeric"
                min="0"
                value={reminderDayInput}
                onChange={e => setReminderDayInput(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addReminderDay()
                  }
                }}
                className="font-mono"
              />
            )}
          </Field>
          <Button variant="secondary" icon={Plus} onClick={addReminderDay}>
            Agregar aviso
          </Button>
        </div>
      </Section>

      <Section
        title="Datos al completar"
        description="Lo que quieres anotar cada vez que la marques como hecha, como el tipo de filtro o el kilometraje."
      >
        {fields.map((field, index) => (
          <div key={index} className="flex flex-col gap-4 rounded-md border border-line bg-surface p-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Nombre visible" required error={errors[`field-${index}-display`]}>
                {props => (
                  <Input
                    {...props}
                    type="text"
                    value={field.display_name}
                    onChange={e => updateField(index, { display_name: e.target.value })}
                    placeholder="Tipo de filtro"
                  />
                )}
              </Field>
              <Field label="Clave" required hint="Sin espacios ni tildes." error={errors[`field-${index}-name`]}>
                {props => (
                  <Input
                    {...props}
                    type="text"
                    value={field.name}
                    onChange={e => updateField(index, { name: e.target.value })}
                    placeholder="tipo_filtro"
                    className="font-mono"
                  />
                )}
              </Field>
              <Field label="Tipo de dato">
                {props => (
                  <Select {...props} value={field.type} onChange={e => updateField(index, { type: e.target.value })}>
                    {FIELD_TYPES.map(t => (
                      <option key={t.value} value={t.value}>
                        {t.label}
                      </option>
                    ))}
                  </Select>
                )}
              </Field>
              <Field label="Valor por defecto" hint="Opcional.">
                {props => (
                  <Input
                    {...props}
                    type="text"
                    value={field.default_value}
                    onChange={e => updateField(index, { default_value: e.target.value })}
                  />
                )}
              </Field>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Checkbox
                label="Obligatorio"
                checked={field.is_required}
                onChange={e => updateField(index, { is_required: e.target.checked })}
              />
              <Button variant="danger" size="sm" icon={Trash2} onClick={() => removeField(index)}>
                Quitar dato
              </Button>
            </div>
          </div>
        ))}
        <div>
          <Button variant="secondary" icon={Plus} onClick={addField}>
            Agregar dato
          </Button>
        </div>
      </Section>

      <FormActions>
        <Button type="submit" disabled={saving}>
          {saving ? 'Guardando…' : isEdit ? 'Guardar cambios' : 'Crear actividad'}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={saving}>
          Cancelar
        </Button>
      </FormActions>
    </Form>
  )
}

export default ActivityForm
