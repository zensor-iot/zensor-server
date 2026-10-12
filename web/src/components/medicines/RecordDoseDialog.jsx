import { useState } from 'react'
import { Check, SkipForward } from 'lucide-react'
import { Button, Drawer, Field, KeyValue, StatusPill, Textarea } from '../../ui'
import { doseState, doseTextOf, formatDateTime } from '../../utils/medicineSchedule'

/** Right drawer to record a dose as administered or skipped, with an optional note. */
const RecordDoseDialog = ({ entry, onClose, onResolved }) => {
  const [shown, setShown] = useState(entry)
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(null)

  if (entry && entry !== shown) {
    setShown(entry)
    setNotes('')
  }

  const handle = async action => {
    try {
      setSaving(action)
      await onResolved(action, notes.trim() || null)
    } finally {
      setSaving(null)
    }
  }

  const state = shown ? doseState(shown.dose) : null
  const doseText = shown ? doseTextOf(shown.dose, shown.dose_text) : ''

  return (
    <Drawer
      open={Boolean(entry)}
      onClose={onClose}
      title={shown?.patient_name ?? ''}
      subtitle={
        shown && (
          <>
            <StatusPill tone={state.tone}>{state.label}</StatusPill>
            <span>{formatDateTime(shown.dose.scheduled_at)}</span>
          </>
        )
      }
      footer={
        <>
          <Button icon={Check} onClick={() => handle('administer')} disabled={Boolean(saving)}>
            {saving === 'administer' ? 'Registrando…' : 'Marcar como administrada'}
          </Button>
          <Button variant="secondary" icon={SkipForward} onClick={() => handle('skip')} disabled={Boolean(saving)}>
            {saving === 'skip' ? 'Saltando…' : 'Saltar dosis'}
          </Button>
        </>
      }
    >
      {shown && (
        <>
          <KeyValue
            items={[
              { label: 'Medicina', value: shown.medicine_name },
              { label: 'Dosis', value: doseText },
              { label: 'Programada', value: formatDateTime(shown.dose.scheduled_at), mono: true },
            ]}
          />
          <Field label="Notas" hint="Opcional. Por ejemplo: con comida o a medias.">
            {props => <Textarea {...props} value={notes} onChange={e => setNotes(e.target.value)} />}
          </Field>
        </>
      )}
    </Drawer>
  )
}

export default RecordDoseDialog
