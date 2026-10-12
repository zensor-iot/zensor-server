import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { medicinesApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { Button, Field, Form, FormActions, Input, Page, PageHeader, Select } from '../../ui'
import { PATIENT_KINDS } from '../../utils/medicineSchedule'

const PatientForm = () => {
  const { tenantId } = useParams()
  const navigate = useNavigate()
  const { showSuccess, showError } = useNotification()
  const [name, setName] = useState('')
  const [kind, setKind] = useState('human')
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)
  const listPath = `/portal/${tenantId}/medicines`

  const handleSubmit = async event => {
    event.preventDefault()
    try {
      setSaving(true)
      const patient = await medicinesApi.createPatient({
        tenant_id: tenantId,
        name: name.trim(),
        kind,
        notes: notes.trim(),
      })
      showSuccess(`${patient.name} ya está en Medicinas.`, 'Paciente guardado')
      navigate(`${listPath}/patients/${patient.id}`)
    } catch {
      showError('Revisa los datos y tu conexión, y vuelve a intentarlo.', 'No se pudo guardar el paciente')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Page>
      <PageHeader
        title="Nuevo paciente"
        description="Una persona o un animal que toma medicamentos."
        back={{ to: listPath, label: 'Medicinas' }}
      />
      <Form onSubmit={handleSubmit} className="w-full max-w-[36rem]">
        <Field label="Nombre" required>
          {props => (
            <Input {...props} type="text" value={name} onChange={e => setName(e.target.value)} placeholder="Luna" />
          )}
        </Field>
        <Field label="Tipo">
          {props => (
            <Select {...props} value={kind} onChange={e => setKind(e.target.value)}>
              {PATIENT_KINDS.map(k => (
                <option key={k.value} value={k.value}>
                  {k.label}
                </option>
              ))}
            </Select>
          )}
        </Field>
        <Field label="Notas" hint="Opcional. Lo que ayude a dosificar, como especie o peso.">
          {props => (
            <Input
              {...props}
              type="text"
              value={notes}
              onChange={e => setNotes(e.target.value)}
              placeholder="Perra, 12 kg"
            />
          )}
        </Field>
        <FormActions>
          <Button type="submit" disabled={saving || !name.trim()}>
            {saving ? 'Guardando…' : 'Guardar paciente'}
          </Button>
          <Button variant="ghost" to={listPath}>
            Cancelar
          </Button>
        </FormActions>
      </Form>
    </Page>
  )
}

export default PatientForm
