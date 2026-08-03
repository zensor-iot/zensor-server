import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { CalendarClock, ArrowLeft, CheckCircle, Loader2 } from 'lucide-react'
import { maintenanceApi } from '../../config/api'
import { relativeDateLabel } from '../../utils/maintenanceSchedule'
import { useAuth } from '../../contexts/AuthContext'
import { useNotification } from '../../hooks/useNotification'
import CompleteExecutionDialog from './CompleteExecutionDialog'
import './Maintenance.css'

// Up Next: the earliest pending execution per active activity, soonest first.
const MaintenanceUpNext = () => {
    const { tenantId } = useParams()
    const { user } = useAuth()
    const { showError } = useNotification()
    const [items, setItems] = useState([])
    const [loading, setLoading] = useState(true)
    const [completing, setCompleting] = useState(null)

    const fetchUpNext = useCallback(async () => {
        try {
            setLoading(true)
            const activities = await maintenanceApi.listAllActivities(tenantId)

            const executionLists = await Promise.all(
                activities.map(activity =>
                    maintenanceApi.listExecutions(activity.id, 1, 100)
                        .then(data => ({ activity, executions: data.data || [] }))
                        .catch(() => ({ activity, executions: [] }))
                )
            )

            const upNext = executionLists
                .map(({ activity, executions }) => {
                    const pending = executions
                        .filter(e => !e.completed_at)
                        .sort((a, b) => new Date(a.scheduled_date) - new Date(b.scheduled_date))
                    return pending.length > 0 ? { activity, execution: pending[0] } : null
                })
                .filter(Boolean)
                .sort((a, b) => new Date(a.execution.scheduled_date) - new Date(b.execution.scheduled_date))

            setItems(upNext)
        } catch (err) {
            showError(err.message, 'Failed to load upcoming maintenance')
        } finally {
            setLoading(false)
        }
    }, [tenantId]) // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        fetchUpNext()
    }, [fetchUpNext])

    const isFuture = (execution) => new Date(execution.scheduled_date) > new Date()

    return (
        <div className="maintenance-page">
            <div className="maintenance-header">
                <h2>
                    <CalendarClock size={22} />
                    Up Next
                </h2>
                <Link to={`/portal/${tenantId}/maintenance`} className="maintenance-btn">
                    <ArrowLeft size={16} />
                    All activities
                </Link>
            </div>

            {loading ? (
                <div className="maintenance-empty">
                    <Loader2 className="loading-spinner" />
                    <p>Loading upcoming maintenance...</p>
                </div>
            ) : items.length === 0 ? (
                <div className="maintenance-empty">
                    <CheckCircle size={48} />
                    <h3>All caught up</h3>
                    <p>No pending maintenance executions.</p>
                </div>
            ) : (
                items.map(({ activity, execution }) => (
                    <div key={execution.id} className="maintenance-card">
                        <div className="maintenance-card-row">
                            <div>
                                <h3 className="maintenance-card-title">
                                    <Link to={`/portal/${tenantId}/maintenance/activities/${activity.id}?tab=history`}>
                                        {activity.name}
                                    </Link>
                                    {execution.is_overdue && (
                                        <span className="maintenance-badge overdue">
                                            {execution.overdue_days > 0
                                                ? `Overdue ${execution.overdue_days}d`
                                                : 'Overdue'}
                                        </span>
                                    )}
                                </h3>
                                <span className="maintenance-card-schedule">
                                    <CalendarClock size={14} />
                                    {relativeDateLabel(execution.scheduled_date)}
                                    {' · '}
                                    {new Date(execution.scheduled_date).toLocaleDateString()}
                                </span>
                            </div>
                            <div className="maintenance-card-actions">
                                {!isFuture(execution) && (
                                    <button
                                        className="maintenance-btn primary"
                                        onClick={() => setCompleting({ activity, execution })}
                                    >
                                        <CheckCircle size={16} />
                                        Complete
                                    </button>
                                )}
                            </div>
                        </div>
                    </div>
                ))
            )}

            {completing && (
                <CompleteExecutionDialog
                    activity={completing.activity}
                    execution={completing.execution}
                    userEmail={user?.email}
                    onClose={() => setCompleting(null)}
                    onCompleted={() => {
                        setCompleting(null)
                        fetchUpNext()
                    }}
                />
            )}
        </div>
    )
}

export default MaintenanceUpNext
