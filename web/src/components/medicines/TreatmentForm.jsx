import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { Button, Field, Form, FormActions, Input, Page, PageHeader, Section, Select } from '../../ui'
import {
  DOSE_UNITS,
  INTERVAL_UNITS,
  dayLabel,
  formatTime,
  previewOccurrences,
  totalDosesForDuration,
} from '../../utils/medicineSchedule'

const END_MODES = [
  { value: 'none', label: 'Sin fecha de término' },
  { value: 'doses', label: 'Durante unos días' },
  { value: 'date', label: 'Hasta una fecha' },
]

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
  const patientPath = `/portal/${tenantId}/medicines/patients/${patientId}`

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
    unit: intervalUnit,
  }
  const preview = previewOccurrences(schedule, 3)
  const derivedTotalDoses = totalDosesForDuration(durationDays, every, intervalUnit)

  const handleSubmit = async event => {
    event.preventDefault()

    const payload = {
      tenant_id: tenantId,
      patient_id: patientId,
      medicine_name: medicineName.trim(),
      quantity: Number(quantity),
      unit,
      schedule,
      notes: notes.trim(),
    }
    if (endMode === 'doses' && derivedTotalDoses) payload.total_doses = derivedTotalDoses
    if (endMode === 'date' && endDate) payload.end_at = new Date(endDate).toISOString()

    try {
      setSaving(true)
      await medicinesApi.createTreatment(payload)
      showSuccess(`Las dosis de ${payload.medicine_name} ya están en la agenda.`, 'Tratamiento guardado')
      navigate(patientPath)
    } catch {
      showError('Revisa los datos y tu conexión, y vuelve a intentarlo.', 'No se pudo guardar el tratamiento')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Page>
      <PageHeader
        title="Nuevo tratamiento"
        description="Qué medicina, cuánto y cada cuánto."
        back={{ to: patientPath, label: 'Paciente' }}
      />
      <Form onSubmit={handleSubmit} className="w-full max-w-[36rem]">
        <Field label="Medicina" required>
          {props => (
            <Input
              {...props}
              type="text"
              value={medicineName}
              onChange={e => setMedicineName(e.target.value)}
              placeholder="Amoxicilina"
            />
          )}
        </Field>

        <div className="flex flex-wrap gap-4 [&>*]:min-w-[160px] [&>*]:flex-1">
          <Field label="Cantidad" required>
            {props => (
              <Input
                {...props}
                type="number"
                inputMode="decimal"
                min="0"
                step="0.5"
                value={quantity}
                onChange={e => setQuantity(e.target.value)}
              />
            )}
          </Field>
          <Field label="Unidad">
            {props => (
              <Select {...props} value={unit} onChange={e => setUnit(e.target.value)}>
                {DOSE_UNITS.map(u => (
                  <option key={u.value} value={u.value}>
                    {u.label}
                  </option>
                ))}
              </Select>
            )}
          </Field>
        </div>

        <div className="flex flex-wrap gap-4 [&>*]:min-w-[160px] [&>*]:flex-1">
          <Field label="Cada" required>
            {props => (
              <Input
                {...props}
                type="number"
                inputMode="numeric"
                min="1"
                value={every}
                onChange={e => setEvery(e.target.value)}
              />
            )}
          </Field>
          <Field label="Intervalo">
            {props => (
              <Select {...props} value={intervalUnit} onChange={e => setIntervalUnit(e.target.value)}>
                {INTERVAL_UNITS.map(u => (
                  <option key={u.value} value={u.value}>
                    {u.label}
                  </option>
                ))}
              </Select>
            )}
          </Field>
        </div>

        <Field label="Primera dosis" required>
          {props => (
            <Input {...props} type="datetime-local" value={startAt} onChange={e => setStartAt(e.target.value)} />
          )}
        </Field>

        <Field label="Hasta cuándo">
          {props => (
            <Select {...props} value={endMode} onChange={e => setEndMode(e.target.value)}>
              {END_MODES.map(mode => (
                <option key={mode.value} value={mode.value}>
                  {mode.label}
                </option>
              ))}
            </Select>
          )}
        </Field>

        {endMode === 'doses' && (
          <Field
            label="Días de tratamiento"
            required
            hint={derivedTotalDoses ? `${derivedTotalDoses} dosis en total.` : 'Ingresa un número de días mayor que cero.'}
          >
            {props => (
              <Input
                {...props}
                type="number"
                inputMode="numeric"
                min="1"
                value={durationDays}
                onChange={e => setDurationDays(e.target.value)}
              />
            )}
          </Field>
        )}

        {endMode === 'date' && (
          <Field label="Última dosis hasta" required>
            {props => (
              <Input {...props} type="datetime-local" value={endDate} onChange={e => setEndDate(e.target.value)} />
            )}
          </Field>
        )}

        <Field label="Notas" hint="Opcional. Por ejemplo: con comida.">
          {props => <Input {...props} type="text" value={notes} onChange={e => setNotes(e.target.value)} />}
        </Field>

        {preview.length > 0 && (
          <Section title="Primeras dosis" className="rounded-lg border border-line bg-surface p-4">
            <ul className="m-0 flex list-none flex-col gap-2 p-0">
              {preview.map(occurrence => (
                <li key={occurrence.getTime()} className="flex items-baseline gap-3">
                  <span className="text-data-sm">{formatTime(occurrence)}</span>
                  <span className="text-body-sm text-muted">{dayLabel(occurrence)}</span>
                </li>
              ))}
            </ul>
          </Section>
        )}

        <FormActions>
          <Button type="submit" disabled={saving || !medicineName.trim()}>
            {saving ? 'Guardando…' : 'Guardar tratamiento'}
          </Button>
          <Button variant="ghost" to={patientPath}>
            Cancelar
          </Button>
        </FormActions>
      </Form>
    </Page>
  )
}

export default TreatmentForm
