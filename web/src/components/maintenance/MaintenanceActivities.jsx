import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Wrench, Plus, CalendarClock, Pause, Play, Trash2, Loader2, ChevronLeft, ChevronRight } from 'lucide-react'
import { maintenanceApi } from '../../config/api'
import { describeSchedule } from '../../utils/maintenanceSchedule'
import { useNotification } from '../../hooks/useNotification'
import './Maintenance.css'

const PAGE_SIZE = 10

const MaintenanceActivities = () => {
    const { tenantId } = useParams()
    const { showSuccess, showError } = useNotification()
    const [activities, setActivities] = useState([])
    const [page, setPage] = useState(1)
    const [total, setTotal] = useState(0)
    const [loading, setLoading] = useState(true)

    const fetchActivities = useCallback(async (targetPage) => {
        try {
            setLoading(true)
            const data = await maintenanceApi.listActivities(tenantId, targetPage, PAGE_SIZE)
            setActivities(data.data || [])
            setTotal(data.pagination?.total ?? (data.data || []).length)
        } catch (err) {
            showError(err.message, 'Failed to load maintenance activities')
        } finally {
            setLoading(false)
        }
    }, [tenantId]) // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        fetchActivities(page)
    }, [fetchActivities, page])

    const handleToggleActive = async (activity) => {
        try {
            if (activity.is_active) {
                await maintenanceApi.deactivateActivity(activity.id)
                showSuccess(`"${activity.name}" paused`)
            } else {
                await maintenanceApi.activateActivity(activity.id)
                showSuccess(`"${activity.name}" resumed`)
            }
            fetchActivities(page)
        } catch (err) {
            showError(err.message, 'Failed to update activity')
        }
    }

    const handleDelete = async (activity) => {
        if (!window.confirm(`Delete "${activity.name}"? This cannot be undone.`)) {
            return
        }
        try {
            await maintenanceApi.deleteActivity(activity.id)
            showSuccess(`"${activity.name}" deleted`)
            fetchActivities(page)
        } catch (err) {
            showError(err.message, 'Failed to delete activity')
        }
    }

    const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

    return (
        <div className="maintenance-page">
            <div className="maintenance-header">
                <h2>
                    <Wrench size={22} />
                    Maintenance
                </h2>
                <div className="maintenance-header-actions">
                    <Link to={`/portal/${tenantId}/maintenance/up-next`} className="maintenance-btn">
                        <CalendarClock size={16} />
                        Up Next
                    </Link>
                    <Link to={`/portal/${tenantId}/maintenance/new`} className="maintenance-btn primary">
                        <Plus size={16} />
                        New Activity
                    </Link>
                </div>
            </div>

            {loading ? (
                <div className="maintenance-empty">
                    <Loader2 className="loading-spinner" />
                    <p>Loading maintenance activities...</p>
                </div>
            ) : activities.length === 0 ? (
                <div className="maintenance-empty">
                    <Wrench size={48} />
                    <h3>No maintenance activities yet</h3>
                    <p>Create your first activity to start tracking recurring maintenance.</p>
                </div>
            ) : (
                <>
                    {activities.map(activity => (
                        <div key={activity.id} className="maintenance-card">
                            <div className="maintenance-card-row">
                                <div>
                                    <h3 className="maintenance-card-title">
                                        <Link to={`/portal/${tenantId}/maintenance/activities/${activity.id}`}>
                                            {activity.name}
                                        </Link>
                                        <span className={`maintenance-badge ${activity.is_active ? 'active' : 'paused'}`}>
                                            {activity.is_active ? 'Active' : 'Paused'}
                                        </span>
                                    </h3>
                                    {activity.description && (
                                        <p className="maintenance-card-description">{activity.description}</p>
                                    )}
                                    <span className="maintenance-card-schedule">
                                        <CalendarClock size={14} />
                                        {describeSchedule(activity.schedule)}
                                    </span>
                                </div>
                                <div className="maintenance-card-actions">
                                    <button className="maintenance-btn" onClick={() => handleToggleActive(activity)}>
                                        {activity.is_active ? <Pause size={16} /> : <Play size={16} />}
                                        {activity.is_active ? 'Pause' : 'Resume'}
                                    </button>
                                    <button className="maintenance-btn danger" onClick={() => handleDelete(activity)}>
                                        <Trash2 size={16} />
                                        Delete
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}

                    <div className="maintenance-pagination">
                        <button
                            className="maintenance-btn"
                            disabled={page <= 1}
                            onClick={() => setPage(page - 1)}
                        >
                            <ChevronLeft size={16} />
                            Previous
                        </button>
                        <span>Page {page} of {totalPages}</span>
                        <button
                            className="maintenance-btn"
                            disabled={page >= totalPages}
                            onClick={() => setPage(page + 1)}
                        >
                            Next
                            <ChevronRight size={16} />
                        </button>
                    </div>
                </>
            )}
        </div>
    )
}

export default MaintenanceActivities
