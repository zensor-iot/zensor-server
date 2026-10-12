import { useState } from 'react'
import { Check } from 'lucide-react'
import { Button, Checkbox, Drawer, Field, Form, FormActions, Input, StatusPill } from '../../ui'
import { formatDateTime } from '../../lib/zensor'
import { maintenanceApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import {
  buildFieldValues,
  dueLabel,
  executionStatus,
  initialFieldValues,
  missingRequiredFields,
} from '../../utils/maintenanceSchedule'

const INPUT_TYPES = { number: 'number', date: 'date' }

/**
 * Form to mark an execution as done. completed_by starts with the signed-in user and
 * the activity's custom fields become typed inputs captured into field_values.
 */
export function CompleteExecutionForm({ activity, execution, userEmail, onCancel, onCompleted }) {
  const { showSuccess, showError } = useNotification()
  const fields = activity?.fields || []
  const [completedBy, setCompletedBy] = useState(userEmail || '')
  const [values, setValues] = useState(() => initialFieldValues(fields))
  const [errors, setErrors] = useState({})
  const [submitting, setSubmitting] = useState(false)

  const setValue = (name, value) => setValues(prev => ({ ...prev, [name]: value }))

  const handleSubmit = async e => {
    e.preventDefault()
    const nextErrors = {}
    if (!completedBy.trim()) nextErrors.completedBy = 'Escribe quién la hizo.'
    missingRequiredFields(fields, values).forEach(name => {
      const field = fields.find(f => f.name === name)
      nextErrors[name] = `Completa ${field?.display_name ?? name}.`
    })
    setErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) return

    setSubmitting(true)
    try {
      await maintenanceApi.completeExecution(execution.id, completedBy, buildFieldValues(fields, values))
      showSuccess(`“${activity.name}” quedó como hecha`)
      onCompleted()
    } catch {
      showError('Revisa tu conexión y vuelve a intentarlo.', 'No se pudo marcar como hecha')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form onSubmit={handleSubmit} noValidate>
      <Field label="Hecha por" required error={errors.completedBy}>
        {props => (
          <Input {...props} type="text" autoComplete="email" value={completedBy} onChange={e => setCompletedBy(e.target.value)} />
        )}
      </Field>

      {fields.map(field =>
        field.type === 'boolean' ? (
          <Checkbox
            key={field.name}
            label={field.display_name}
            checked={Boolean(values[field.name])}
            onChange={e => setValue(field.name, e.target.checked)}
          />
        ) : (
          <Field key={field.name} label={field.display_name} required={field.is_required} error={errors[field.name]}>
            {props => (
              <Input
                {...props}
                type={INPUT_TYPES[field.type] ?? 'text'}
                inputMode={field.type === 'number' ? 'decimal' : undefined}
                className={field.type === 'number' ? 'font-mono' : undefined}
                value={values[field.name]}
                onChange={e => setValue(field.name, e.target.value)}
              />
            )}
          </Field>
        ),
      )}

      <FormActions>
        <Button type="submit" icon={Check} disabled={submitting}>
          {submitting ? 'Guardando…' : 'Marcar como hecha'}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={submitting}>
          Cancelar
        </Button>
      </FormActions>
    </Form>
  )
}

/** Right drawer that wraps CompleteExecutionForm for a given { activity, execution }. */
const CompleteExecutionDialog = ({ target, userEmail, onClose, onCompleted }) => {
  const [shown, setShown] = useState(target)
  if (target && target !== shown) setShown(target)

  const status = shown ? executionStatus(shown.execution) : null

  return (
    <Drawer
      open={Boolean(target)}
      onClose={onClose}
      title="Marcar como hecha"
      subtitle={
        shown && (
          <>
            <StatusPill tone={status.tone}>{status.label}</StatusPill>
            <span>
              {shown.activity.name} · {dueLabel(shown.execution.scheduled_date).toLowerCase()}
            </span>
          </>
        )
      }
    >
      {shown && (
        <>
          <p className="m-0 text-body-sm text-muted">
            Programada para el <span className="font-mono">{formatDateTime(shown.execution.scheduled_date)}</span>.
          </p>
          <CompleteExecutionForm
            key={shown.execution.id}
            activity={shown.activity}
            execution={shown.execution}
            userEmail={userEmail}
            onCancel={onClose}
            onCompleted={onCompleted}
          />
        </>
      )}
    </Drawer>
  )
}

export default CompleteExecutionDialog
