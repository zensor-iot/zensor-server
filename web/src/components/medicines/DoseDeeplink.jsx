import { useEffect, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'
import { medicinesApi } from '../../config/api'
import { Button, ErrorState, LoadingState, Page } from '../../ui'

/** Landing route for dose reminder push notifications: resolves the dose's tenant and opens it in that agenda. */
const DoseDeeplink = () => {
  const { doseId } = useParams()
  const [target, setTarget] = useState(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function resolve() {
      try {
        const dose = await medicinesApi.getDose(doseId)
        const treatment = await medicinesApi.getTreatment(dose.treatment_id)
        if (!cancelled) setTarget(`/portal/${treatment.tenant_id}/medicines/agenda?dose=${encodeURIComponent(doseId)}`)
      } catch {
        if (!cancelled) setFailed(true)
      }
    }

    resolve()
    return () => {
      cancelled = true
    }
  }, [doseId])

  if (target) return <Navigate to={target} replace />

  return (
    <Page>
      {failed ? (
        <div className="flex flex-col items-start gap-4">
          <ErrorState title="No se pudo abrir la dosis">
            Puede que el tratamiento se haya eliminado o que no tengas acceso. Revisa la agenda de Medicinas desde el
            inicio.
          </ErrorState>
          <Button variant="secondary" to="/">
            Ir al inicio
          </Button>
        </div>
      ) : (
        <LoadingState>Abriendo la dosis…</LoadingState>
      )}
    </Page>
  )
}

export default DoseDeeplink
