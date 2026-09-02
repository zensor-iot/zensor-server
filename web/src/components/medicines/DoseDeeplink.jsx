import { useEffect, useState } from 'react'
import { useParams, Navigate } from 'react-router-dom'
import { Loader2 } from 'lucide-react'
import { medicinesApi } from '../../config/api'
import './Medicines.css'

// DoseDeeplink is where a dose reminder push notification lands. It resolves the
// tenant the dose belongs to and forwards to that tenant's agenda.
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
                if (!cancelled) {
                    setTarget(`/portal/${treatment.tenant_id}/medicines/agenda`)
                }
            } catch {
                if (!cancelled) {
                    setFailed(true)
                }
            }
        }

        resolve()
        return () => { cancelled = true }
    }, [doseId])

    if (failed) {
        return <Navigate to="/" replace />
    }

    if (target) {
        return <Navigate to={target} replace />
    }

    return (
        <div className="medicines-page">
            <div className="medicines-loading"><Loader2 size={20} className="spin" /> Abriendo la dosis…</div>
        </div>
    )
}

export default DoseDeeplink
