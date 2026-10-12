import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Button, ErrorState, LoadingState, Page } from '../../ui'
import { maintenanceApi } from '../../config/api'

/**
 * Deeplink target for push notifications: resolves the execution and redirects to its
 * activity (history included) in the tenant portal.
 */
const ExecutionDeeplink = () => {
  const { executionId } = useParams()
  const navigate = useNavigate()
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let cancelled = false
    const resolve = async () => {
      try {
        const execution = await maintenanceApi.getExecution(executionId)
        const activity = await maintenanceApi.getActivity(execution.activity_id)
        if (!cancelled) {
          navigate(`/portal/${activity.tenant_id}/maintenance/activities/${activity.id}?tab=history`, { replace: true })
        }
      } catch {
        if (!cancelled) setFailed(true)
      }
    }
    resolve()
    return () => {
      cancelled = true
    }
  }, [executionId, navigate])

  return (
    <Page>
      {failed ? (
        <>
          <ErrorState title="No se pudo abrir esta mantención">
            Puede que la actividad se haya eliminado o que no haya conexión. Vuelve al inicio y búscala en Mantención.
          </ErrorState>
          <Button variant="secondary" to="/" className="self-start">
            Ir al inicio
          </Button>
        </>
      ) : (
        <LoadingState>Abriendo la mantención…</LoadingState>
      )}
    </Page>
  )
}

export default ExecutionDeeplink
