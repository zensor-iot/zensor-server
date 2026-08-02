import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { Loader2 } from 'lucide-react'
import { maintenanceApi } from '../../config/api'
import './Maintenance.css'

// Deeplink target for push notifications: resolves the execution and
// redirects to its activity's history tab in the tenant portal.
const ExecutionDeeplink = () => {
    const { executionId } = useParams()
    const navigate = useNavigate()
    const [error, setError] = useState(null)

    useEffect(() => {
        let cancelled = false
        const resolve = async () => {
            try {
                const execution = await maintenanceApi.getExecution(executionId)
                const activity = await maintenanceApi.getActivity(execution.activity_id)
                if (!cancelled) {
                    navigate(
                        `/portal/${activity.tenant_id}/maintenance/activities/${activity.id}?tab=history`,
                        { replace: true }
                    )
                }
            } catch (err) {
                if (!cancelled) {
                    setError(err.message)
                }
            }
        }
        resolve()
        return () => {
            cancelled = true
        }
    }, [executionId, navigate])

    if (error) {
        return (
            <div className="maintenance-page">
                <div className="maintenance-empty">
                    <h3>Could not open this maintenance execution</h3>
                    <p>{error}</p>
                    <Link to="/" className="maintenance-btn">Go home</Link>
                </div>
            </div>
        )
    }

    return (
        <div className="maintenance-page">
            <div className="maintenance-empty">
                <Loader2 className="loading-spinner" />
                <p>Opening maintenance execution...</p>
            </div>
        </div>
    )
}

export default ExecutionDeeplink
