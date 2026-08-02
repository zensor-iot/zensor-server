import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom'
import { Wrench, ArrowLeft, Loader2 } from 'lucide-react'
import ActivityForm from './ActivityForm'
import { maintenanceApi } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import './Maintenance.css'

const formatDate = (value) => (value ? new Date(value).toLocaleDateString() : '')

const MaintenanceActivityDetail = () => {
    const { tenantId, activityId } = useParams()
    const navigate = useNavigate()
    const [searchParams, setSearchParams] = useSearchParams()
    const { showError } = useNotification()

    const tab = searchParams.get('tab') === 'history' ? 'history' : 'details'
    const [activity, setActivity] = useState(null)
    const [executions, setExecutions] = useState([])
    const [loading, setLoading] = useState(true)
    const [historyLoading, setHistoryLoading] = useState(false)

    const fetchActivity = useCallback(async () => {
        try {
            setLoading(true)
            const data = await maintenanceApi.getActivity(activityId)
            setActivity(data)
        } catch (err) {
            showError(err.message, 'Failed to load activity')
        } finally {
            setLoading(false)
        }
    }, [activityId]) // eslint-disable-line react-hooks/exhaustive-deps

    const fetchHistory = useCallback(async () => {
        try {
            setHistoryLoading(true)
            const data = await maintenanceApi.listExecutions(activityId, 1, 100)
            const sorted = [...(data.data || [])].sort(
                (a, b) => new Date(b.scheduled_date) - new Date(a.scheduled_date)
            )
            setExecutions(sorted)
        } catch (err) {
            showError(err.message, 'Failed to load execution history')
        } finally {
            setHistoryLoading(false)
        }
    }, [activityId]) // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        fetchActivity()
    }, [fetchActivity])

    useEffect(() => {
        if (tab === 'history') {
            fetchHistory()
        }
    }, [tab, fetchHistory])

    const selectTab = (nextTab) => {
        setSearchParams(nextTab === 'history' ? { tab: 'history' } : {})
    }

    if (loading) {
        return (
            <div className="maintenance-page">
                <div className="maintenance-empty">
                    <Loader2 className="loading-spinner" />
                    <p>Loading activity...</p>
                </div>
            </div>
        )
    }

    if (!activity) {
        return (
            <div className="maintenance-page">
                <div className="maintenance-empty">
                    <h3>Activity not found</h3>
                    <Link to={`/portal/${tenantId}/maintenance`} className="maintenance-btn">
                        <ArrowLeft size={16} />
                        Back to activities
                    </Link>
                </div>
            </div>
        )
    }

    return (
        <div className="maintenance-page">
            <div className="maintenance-header">
                <h2>
                    <Wrench size={22} />
                    {activity.name}
                    <span className={`maintenance-badge ${activity.is_active ? 'active' : 'paused'}`}>
                        {activity.is_active ? 'Active' : 'Paused'}
                    </span>
                </h2>
                <Link to={`/portal/${tenantId}/maintenance`} className="maintenance-btn">
                    <ArrowLeft size={16} />
                    Back to activities
                </Link>
            </div>

            <div className="maintenance-tabs">
                <button
                    className={`maintenance-tab ${tab === 'details' ? 'active' : ''}`}
                    onClick={() => selectTab('details')}
                >
                    Details
                </button>
                <button
                    className={`maintenance-tab ${tab === 'history' ? 'active' : ''}`}
                    onClick={() => selectTab('history')}
                >
                    History
                </button>
            </div>

            {tab === 'details' ? (
                <ActivityForm
                    tenantId={tenantId}
                    activity={activity}
                    onSaved={fetchActivity}
                    onCancel={() => navigate(`/portal/${tenantId}/maintenance`)}
                />
            ) : historyLoading ? (
                <div className="maintenance-empty">
                    <Loader2 className="loading-spinner" />
                    <p>Loading history...</p>
                </div>
            ) : executions.length === 0 ? (
                <div className="maintenance-empty">
                    <p>No executions yet for this activity.</p>
                </div>
            ) : (
                <table className="maintenance-history-table">
                    <thead>
                        <tr>
                            <th>Scheduled</th>
                            <th>Status</th>
                            <th>Completed at</th>
                            <th>Completed by</th>
                            <th>Field values</th>
                        </tr>
                    </thead>
                    <tbody>
                        {executions.map(execution => (
                            <tr key={execution.id}>
                                <td>{formatDate(execution.scheduled_date)}</td>
                                <td>
                                    {execution.completed_at ? (
                                        <span className="maintenance-badge completed">Completed</span>
                                    ) : execution.is_overdue ? (
                                        <span className="maintenance-badge overdue">
                                            {execution.overdue_days > 0
                                                ? `Overdue ${execution.overdue_days}d`
                                                : 'Overdue'}
                                        </span>
                                    ) : (
                                        <span className="maintenance-badge paused">Pending</span>
                                    )}
                                </td>
                                <td>{formatDate(execution.completed_at)}</td>
                                <td>{execution.completed_by || ''}</td>
                                <td>
                                    {execution.field_values && Object.keys(execution.field_values).length > 0 && (
                                        <ul className="maintenance-field-values">
                                            {Object.entries(execution.field_values).map(([key, value]) => (
                                                <li key={key}>
                                                    <strong>{key}</strong>: {String(value)}
                                                </li>
                                            ))}
                                        </ul>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </div>
    )
}

export default MaintenanceActivityDetail
