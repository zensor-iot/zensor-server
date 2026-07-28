import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Clock, ArrowLeft, Eye, Edit, Trash2, Plus, Cpu, Building, Play, Pause } from 'lucide-react'
import { useAdmin } from '../../hooks/useAdmin'
import { useNotification } from '../../hooks/useNotification'
import { scheduledTasksApi } from '../../config/api'
import './AdminScheduledTasks.css'

const AdminScheduledTasks = () => {
    const { tenantId, deviceId } = useParams()
    const { isAdmin, isLoading } = useAdmin()
    const { showSuccess, showError, showApiNotification } = useNotification()
    const [tasks, setTasks] = useState([])
    const [device, setDevice] = useState(null)
    const [tenant, setTenant] = useState(null)
    const [loading, setLoading] = useState(true)
    const [showCreateForm, setShowCreateForm] = useState(false)
    const [editingTask, setEditingTask] = useState(null)
    const [formData, setFormData] = useState({
        schedule: '',
        duration: 5,
        is_active: true,
        description: ''
    })

    useEffect(() => {
        if (!isLoading && !isAdmin) {
            showError('Access denied. Admin privileges required.', 'Unauthorized')
            return
        }

        if (isAdmin && tenantId && deviceId) {
            fetchData()
        }
    }, [isAdmin, isLoading, tenantId, deviceId])

    const fetchData = async () => {
        try {
            setLoading(true)

            // Fetch tenant, device, and scheduled tasks in parallel
            const [tenantResponse, deviceResponse, tasksResponse] = await Promise.all([
                fetch(`/api/tenants/${tenantId}`),
                fetch(`/api/devices/${deviceId}`),
                scheduledTasksApi.getScheduledTasks(tenantId, deviceId)
            ])

            if (tenantResponse.ok) {
                const tenantData = await tenantResponse.json()
                setTenant(tenantData)
            }

            if (deviceResponse.ok) {
                const deviceData = await deviceResponse.json()
                setDevice(deviceData)
            }

            setTasks(tasksResponse.tasks || [])
        } catch (error) {
            console.error('Failed to fetch data:', error)
            showError('Failed to load scheduled tasks', 'Error')
        } finally {
            setLoading(false)
        }
    }

    const handleCreateTask = async (e) => {
        e.preventDefault()

        const taskData = {
            schedule: formData.schedule,
            is_active: formData.is_active,
            description: formData.description,
            commands: [
                {
                    index: 1,
                    value: 1,
                    priority: "NORMAL",
                    wait_for: "0s"
                },
                {
                    index: 1,
                    value: 0,
                    priority: "NORMAL",
                    wait_for: `${formData.duration}m`
                }
            ]
        }

        const result = await showApiNotification(
            scheduledTasksApi.createScheduledTask(tenantId, deviceId, taskData),
            'Scheduled task created successfully',
            'Failed to create scheduled task',
            'Success',
            'Error'
        )

        if (result.success) {
            setShowCreateForm(false)
            setFormData({ schedule: '', duration: 5, is_active: true, description: '' })
            fetchData()
        }
    }

    const handleUpdateTask = async (e) => {
        e.preventDefault()

        const taskData = {
            schedule: formData.schedule,
            is_active: formData.is_active,
            description: formData.description,
            commands: [
                {
                    index: 1,
                    value: 1,
                    priority: "NORMAL",
                    wait_for: "0s"
                },
                {
                    index: 1,
                    value: 0,
                    priority: "NORMAL",
                    wait_for: `${formData.duration}m`
                }
            ]
        }

        const result = await showApiNotification(
            scheduledTasksApi.updateScheduledTask(tenantId, deviceId, editingTask.id, taskData),
            'Scheduled task updated successfully',
            'Failed to update scheduled task',
            'Success',
            'Error'
        )

        if (result.success) {
            setEditingTask(null)
            setFormData({ schedule: '', duration: 5, is_active: true, description: '' })
            fetchData()
        }
    }

    const handleDeleteTask = async (taskId) => {
        if (!window.confirm('Are you sure you want to delete this scheduled task? This action cannot be undone.')) {
            return
        }

        const result = await showApiNotification(
            scheduledTasksApi.deleteScheduledTask(tenantId, deviceId, taskId),
            'Scheduled task deleted successfully',
            'Failed to delete scheduled task',
            'Success',
            'Error'
        )

        if (result.success) {
            fetchData()
        }
    }

    const handleToggleActive = async (task) => {
        const newStatus = !task.is_active

        const result = await showApiNotification(
            scheduledTasksApi.updateScheduledTask(tenantId, deviceId, task.id, {
                ...task,
                is_active: newStatus
            }),
            `Task ${newStatus ? 'activated' : 'deactivated'} successfully`,
            `Failed to ${newStatus ? 'activate' : 'deactivate'} task`,
            'Success',
            'Error'
        )

        if (result.success) {
            fetchData()
        }
    }

    const startEdit = (task) => {
        setEditingTask(task)
        setFormData({
            schedule: task.schedule || '',
            duration: 5, // Default duration
            is_active: task.is_active || true,
            description: task.description || ''
        })
    }

    const cancelEdit = () => {
        setEditingTask(null)
        setFormData({ schedule: '', duration: 5, is_active: true, description: '' })
    }

    const parseCronExpression = (cron) => {
        // Simple cron parser for display
        const parts = cron.split(' ')
        if (parts.length >= 6) {
            const [second, minute, hour, day, month, dayOfWeek] = parts
            if (minute === '0' && hour !== '*') {
                return `Daily at ${hour}:00`
            }
            if (minute !== '*' && hour !== '*') {
                return `Daily at ${hour}:${minute.padStart(2, '0')}`
            }
        }
        return cron
    }

    if (isLoading) {
        return (
            <div className="admin-scheduled-tasks">
                <div className="loading">Loading admin panel...</div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="admin-scheduled-tasks">
                <div className="access-denied">
                    <Clock size={48} />
                    <h2>Access Denied</h2>
                    <p>You need admin privileges to access this page.</p>
                </div>
            </div>
        )
    }

    return (
        <div className="admin-scheduled-tasks">
            <div className="admin-header">
                <div className="header-top">
                    <Link to={`/admin/tenants/${tenantId}/devices`} className="back-link">
                        <ArrowLeft size={20} />
                        Back to Devices
                    </Link>
                </div>
                <div className="admin-title">
                    <Clock size={32} />
                    <div>
                        <h1>Scheduled Tasks</h1>
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
                        </div>
                    </div>
                </div>
                <p className="admin-subtitle">Manage scheduled irrigation tasks for this device</p>
            </div>

            <div className="admin-actions">
                <button
                    className="btn btn-primary"
                    onClick={() => setShowCreateForm(true)}
                >
                    <Plus size={20} />
                    Create Task
                </button>
            </div>

            {/* Create/Edit Form */}
            {(showCreateForm || editingTask) && (
                <div className="form-overlay">
                    <div className="form-card">
                        <h2>{editingTask ? 'Edit Scheduled Task' : 'Create New Scheduled Task'}</h2>
                        <form onSubmit={editingTask ? handleUpdateTask : handleCreateTask}>
                            <div className="form-group">
                                <label htmlFor="schedule">Cron Schedule *</label>
                                <input
                                    type="text"
                                    id="schedule"
                                    value={formData.schedule}
                                    onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                                    placeholder="0 0 6 * * * (Daily at 6 AM)"
                                    required
                                />
                                <small>Format: second minute hour day month dayOfWeek</small>
                            </div>

                            <div className="form-group">
                                <label htmlFor="duration">Duration (minutes)</label>
                                <input
                                    type="number"
                                    id="duration"
                                    value={formData.duration}
                                    onChange={(e) => setFormData({ ...formData, duration: parseInt(e.target.value) || 5 })}
                                    min="1"
                                    max="60"
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="description">Description</label>
                                <textarea
                                    id="description"
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    rows={3}
                                    placeholder="Optional description for this scheduled task"
                                />
                            </div>

                            <div className="form-group">
                                <label className="checkbox-label">
                                    <input
                                        type="checkbox"
                                        checked={formData.is_active}
                                        onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                                    />
                                    <span>Active</span>
                                </label>
                            </div>

                            <div className="form-actions">
                                <button type="button" className="btn btn-secondary" onClick={editingTask ? cancelEdit : () => setShowCreateForm(false)}>
                                    Cancel
                                </button>
                                <button type="submit" className="btn btn-primary">
                                    {editingTask ? 'Update Task' : 'Create Task'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* Tasks List */}
            {loading ? (
                <div className="loading">Loading scheduled tasks...</div>
            ) : (
                <div className="tasks-grid">
                    {tasks.map((task) => (
                        <div key={task.id} className="task-card">
                            <div className="task-header">
                                <div className="task-info">
                                    <h3>{parseCronExpression(task.schedule)}</h3>
                                    <span className={`status-badge ${task.is_active ? 'active' : 'inactive'}`}>
                                        {task.is_active ? 'Active' : 'Inactive'}
                                    </span>
                                </div>
                                <div className="task-actions">
                                    <Link
                                        to={`/admin/tenants/${tenantId}/devices/${deviceId}/scheduled-tasks/${task.id}/executions`}
                                        className="btn-icon"
                                        title="View Executions"
                                    >
                                        <Eye size={16} />
                                    </Link>
                                    <button
                                        className="btn-icon"
                                        onClick={() => handleToggleActive(task)}
                                        title={task.is_active ? 'Deactivate' : 'Activate'}
                                    >
                                        {task.is_active ? <Pause size={16} /> : <Play size={16} />}
                                    </button>
                                    <button
                                        className="btn-icon"
                                        onClick={() => startEdit(task)}
                                        title="Edit Task"
                                    >
                                        <Edit size={16} />
                                    </button>
                                    <button
                                        className="btn-icon danger"
                                        onClick={() => handleDeleteTask(task.id)}
                                        title="Delete Task"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            </div>

                            <div className="task-details">
                                <p className="task-schedule">
                                    <strong>Schedule:</strong> {task.schedule}
                                </p>
                                {task.description && (
                                    <p className="task-description">{task.description}</p>
                                )}
                                <p className="task-commands">
                                    <strong>Commands:</strong> {task.commands?.length || 0} commands
                                </p>
                            </div>

                            <div className="task-stats">
                                <div className="stat">
                                    <Clock size={16} />
                                    <span>Created: {new Date(task.created_at).toLocaleDateString()}</span>
                                </div>
                                <div className="stat">
                                    <Play size={16} />
                                    <span>Executions: {task.execution_count || 0}</span>
                                </div>
                            </div>
                        </div>
                    ))}

                    {tasks.length === 0 && (
                        <div className="empty-state">
                            <Clock size={48} />
                            <h3>No Scheduled Tasks Found</h3>
                            <p>Create your first scheduled task to get started.</p>
                        </div>
                    )}
                </div>
            )}
        </div>
    )
}

export default AdminScheduledTasks
