import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Play, ArrowLeft, Clock, CheckCircle, XCircle, AlertCircle, Building, Cpu } from 'lucide-react'
import { useAdmin } from '../../hooks/useAdmin'
import { useNotification } from '../../hooks/useNotification'
import { scheduledTasksApi } from '../../config/api'
import './AdminTaskExecutions.css'

const AdminTaskExecutions = () => {
    const { tenantId, deviceId, taskId } = useParams()
    const { isAdmin, isLoading } = useAdmin()
    const { showError } = useNotification()
    const [executions, setExecutions] = useState([])
    const [scheduledTask, setScheduledTask] = useState(null)
    const [device, setDevice] = useState(null)
    const [tenant, setTenant] = useState(null)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        if (!isLoading && !isAdmin) {
            showError('Access denied. Admin privileges required.', 'Unauthorized')
            return
        }

        if (isAdmin && tenantId && deviceId && taskId) {
            fetchData()
        }
    }, [isAdmin, isLoading, tenantId, deviceId, taskId])

    const fetchData = async () => {
        try {
            setLoading(true)

            // Fetch tenant, device, scheduled task, and executions in parallel
            const [tenantResponse, deviceResponse, taskResponse, executionsResponse] = await Promise.all([
                fetch(`/api/tenants/${tenantId}`),
                fetch(`/api/devices/${deviceId}`),
                scheduledTasksApi.getScheduledTask(tenantId, deviceId, taskId),
                scheduledTasksApi.getTaskExecutions(tenantId, deviceId, taskId, 50)
            ])

            if (tenantResponse.ok) {
                const tenantData = await tenantResponse.json()
                setTenant(tenantData)
            }

            if (deviceResponse.ok) {
                const deviceData = await deviceResponse.json()
                setDevice(deviceData)
            }

            if (taskResponse) {
                setScheduledTask(taskResponse)
            }

            if (executionsResponse) {
                setExecutions(executionsResponse.tasks || [])
            }
        } catch (error) {
            console.error('Failed to fetch data:', error)
            showError('Failed to load task executions', 'Error')
        } finally {
            setLoading(false)
        }
    }

    const getStatusIcon = (status) => {
        switch (status) {
            case 'completed':
                return <CheckCircle size={16} className="status-icon completed" />
            case 'failed':
                return <XCircle size={16} className="status-icon failed" />
            case 'running':
                return <Play size={16} className="status-icon running" />
            default:
                return <AlertCircle size={16} className="status-icon unknown" />
        }
    }

    const getStatusColor = (status) => {
        switch (status) {
            case 'completed':
                return '#10b981'
            case 'failed':
                return '#ef4444'
            case 'running':
                return '#3b82f6'
            default:
                return '#6b7280'
        }
    }

    const formatDuration = (startTime, endTime) => {
        if (!startTime || !endTime) return 'N/A'
        const start = new Date(startTime)
        const end = new Date(endTime)
        const duration = end - start
        return `${Math.round(duration / 1000)}s`
    }

    if (isLoading) {
        return (
            <div className="admin-task-executions">
                <div className="loading">Loading admin panel...</div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="admin-task-executions">
                <div className="access-denied">
                    <Play size={48} />
                    <h2>Access Denied</h2>
                    <p>You need admin privileges to access this page.</p>
                </div>
            </div>
        )
    }

    return (
        <div className="admin-task-executions">
            <div className="admin-header">
                <div className="header-top">
                    <Link to={`/admin/tenants/${tenantId}/devices/${deviceId}/scheduled-tasks`} className="back-link">
                        <ArrowLeft size={20} />
                        Back to Scheduled Tasks
                    </Link>
                </div>
                <div className="admin-title">
                    <Play size={32} />
                    <div>
                        <h1>Task Executions</h1>
                        <div className="breadcrumb">
                            {tenant && (
                                <span className="breadcrumb-item">
                                    <Building size={16} />
                                    {tenant.name}
                                </span>
                            )}
                            {device && (
                                <span className="breadcrumb-item">
                                    <Cpu size={16} />
                                    {device.name}
                                </span>
                            )}
                            {scheduledTask && (
                                <span className="breadcrumb-item">
                                    <Clock size={16} />
                                    {scheduledTask.schedule}
                                </span>
                            )}
                        </div>
                    </div>
                </div>
                <p className="admin-subtitle">View execution history for this scheduled task</p>
            </div>

            {/* Task Info */}
            {scheduledTask && (
                <div className="task-info-card">
                    <h3>Scheduled Task Details</h3>
                    <div className="task-details">
                        <div className="detail-item">
                            <strong>Schedule:</strong> {scheduledTask.schedule}
                        </div>
                        <div className="detail-item">
                            <strong>Status:</strong>
                            <span className={`status-badge ${scheduledTask.is_active ? 'active' : 'inactive'}`}>
                                {scheduledTask.is_active ? 'Active' : 'Inactive'}
                            </span>
                        </div>
                        {scheduledTask.description && (
                            <div className="detail-item">
                                <strong>Description:</strong> {scheduledTask.description}
                            </div>
                        )}
                        <div className="detail-item">
                            <strong>Commands:</strong> {scheduledTask.commands?.length || 0} commands
                        </div>
                    </div>
                </div>
            )}

            {/* Executions List */}
            {loading ? (
                <div className="loading">Loading task executions...</div>
            ) : (
                <div className="executions-container">
                    <div className="executions-header">
                        <h2>Execution History</h2>
                        <div className="executions-stats">
                            <div className="stat">
                                <span className="stat-label">Total:</span>
                                <span className="stat-value">{executions.length}</span>
                            </div>
                            <div className="stat">
                                <span className="stat-label">Completed:</span>
                                <span className="stat-value completed">{executions.filter(e => e.status === 'completed').length}</span>
                            </div>
                            <div className="stat">
                                <span className="stat-label">Failed:</span>
                                <span className="stat-value failed">{executions.filter(e => e.status === 'failed').length}</span>
                            </div>
                        </div>
                    </div>

                    {executions.length > 0 ? (
                        <div className="executions-list">
                            {executions.map((execution) => (
                                <div key={execution.id} className="execution-card">
                                    <div className="execution-header">
                                        <div className="execution-info">
                                            <div className="execution-status">
                                                {getStatusIcon(execution.status)}
                                                <span className="status-text" style={{ color: getStatusColor(execution.status) }}>
                                                    {execution.status}
                                                </span>
                                            </div>
                                            <div className="execution-time">
                                                {new Date(execution.created_at).toLocaleString()}
                                            </div>
                                        </div>
                                        <div className="execution-duration">
                                            Duration: {formatDuration(execution.started_at, execution.completed_at)}
                                        </div>
                                    </div>

                                    <div className="execution-details">
                                        <div className="detail-row">
                                            <span className="detail-label">Started:</span>
                                            <span className="detail-value">
                                                {execution.started_at ? new Date(execution.started_at).toLocaleString() : 'N/A'}
                                            </span>
                                        </div>
                                        <div className="detail-row">
                                            <span className="detail-label">Completed:</span>
                                            <span className="detail-value">
                                                {execution.completed_at ? new Date(execution.completed_at).toLocaleString() : 'N/A'}
                                            </span>
                                        </div>
                                        {execution.error_message && (
                                            <div className="detail-row error">
                                                <span className="detail-label">Error:</span>
                                                <span className="detail-value error-message">{execution.error_message}</span>
                                            </div>
                                        )}
                                    </div>

                                    {execution.commands && execution.commands.length > 0 && (
                                        <div className="execution-commands">
                                            <h4>Commands Executed:</h4>
                                            <div className="commands-list">
                                                {execution.commands.map((command, index) => (
                                                    <div key={index} className="command-item">
                                                        <span className="command-index">#{command.index}</span>
                                                        <span className="command-value">Value: {command.value}</span>
                                                        <span className="command-priority">Priority: {command.priority}</span>
                                                        <span className="command-wait">Wait: {command.wait_for}</span>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="empty-state">
                            <Play size={48} />
                            <h3>No Executions Found</h3>
                            <p>This scheduled task hasn't been executed yet.</p>
                        </div>
                    )}
                </div>
            )}
        </div>
    )
}

export default AdminTaskExecutions
