import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Cpu, ArrowLeft, Eye, Edit, Trash2, Plus, Building, Clock, Activity, Calendar } from 'lucide-react'
import { useAdmin } from '../../hooks/useAdmin'
import { useNotification } from '../../hooks/useNotification'
import './AdminDevices.css'

const AdminDevices = () => {
    const { tenantId } = useParams()
    const { isAdmin, isLoading } = useAdmin()
    const { showSuccess, showError, showApiNotification } = useNotification()
    const [devices, setDevices] = useState([])
    const [tenant, setTenant] = useState(null)
    const [taskCounts, setTaskCounts] = useState({})
    const [scheduledTaskCounts, setScheduledTaskCounts] = useState({})
    const [loading, setLoading] = useState(true)
    const [selectedDevice, setSelectedDevice] = useState(null)
    const [selectedAction, setSelectedAction] = useState(null)
    const [rightPanelData, setRightPanelData] = useState(null)
    const [showCreateForm, setShowCreateForm] = useState(false)
    const [editingDevice, setEditingDevice] = useState(null)
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        device_type: 'sensor',
        location: '',
        status: 'active'
    })

    useEffect(() => {
        if (!isLoading && !isAdmin) {
            showError('Access denied. Admin privileges required.', 'Unauthorized')
            return
        }

        if (isAdmin && tenantId) {
            fetchTenantAndDevices()
        }
    }, [isAdmin, isLoading, tenantId])

    const fetchTenantAndDevices = async () => {
        try {
            setLoading(true)

            // Fetch tenant info and devices in parallel
            const [tenantResponse, devicesResponse] = await Promise.all([
                fetch(`/api/tenants/${tenantId}`),
                fetch(`/api/tenants/${tenantId}/devices`)
            ])

            if (tenantResponse.ok) {
                const tenantData = await tenantResponse.json()
                setTenant(tenantData)
            }

            if (devicesResponse.ok) {
                const devicesData = await devicesResponse.json()
                console.log('Devices API response:', devicesData) // Debug log

                // Handle paginated response format (data.data) or direct array
                const devicesArray = Array.isArray(devicesData) ? devicesData : (devicesData.data || [])
                console.log('Devices count:', devicesArray.length)
                setDevices(devicesArray)

                // Fetch task counts and scheduled task counts for each device
                fetchTaskCounts(devicesArray)
                fetchScheduledTaskCounts(devicesArray)
            } else {
                throw new Error(`HTTP ${devicesResponse.status}`)
            }
        } catch (error) {
            console.error('Failed to fetch tenant and devices:', error)
            showError('Failed to load tenant and devices', 'Error')
            setDevices([]) // Ensure devices is always an array
        } finally {
            setLoading(false)
        }
    }

    const fetchTaskCounts = async (devicesArray) => {
        try {
            const taskCountPromises = devicesArray.map(async (device) => {
                try {
                    const response = await fetch(`/api/devices/${device.id}/tasks`)
                    if (response.ok) {
                        const data = await response.json()
                        // Use pagination.total for accurate count across all pages
                        const totalCount = data.pagination?.total || 0
                        return { deviceId: device.id, count: totalCount }
                    }
                    return { deviceId: device.id, count: 0 }
                } catch (error) {
                    console.warn(`Failed to fetch tasks for device ${device.id}:`, error)
                    return { deviceId: device.id, count: 0 }
                }
            })

            const counts = await Promise.all(taskCountPromises)
            const taskCountMap = {}
            counts.forEach(({ deviceId, count }) => {
                taskCountMap[deviceId] = count
            })

            setTaskCounts(taskCountMap)
            console.log('Task counts:', taskCountMap)
        } catch (error) {
            console.error('Failed to fetch task counts:', error)
        }
    }

    const fetchScheduledTaskCounts = async (devicesArray) => {
        try {
            const scheduledTaskCountPromises = devicesArray.map(async (device) => {
                try {
                    const response = await fetch(`/api/tenants/${tenantId}/devices/${device.id}/scheduled-tasks`)
                    if (response.ok) {
                        const data = await response.json()
                        // Use pagination.total for accurate count across all pages
                        const totalCount = data.pagination?.total || 0
                        return { deviceId: device.id, count: totalCount }
                    }
                    return { deviceId: device.id, count: 0 }
                } catch (error) {
                    console.warn(`Failed to fetch scheduled tasks for device ${device.id}:`, error)
                    return { deviceId: device.id, count: 0 }
                }
            })

            const counts = await Promise.all(scheduledTaskCountPromises)
            const scheduledTaskCountMap = {}
            counts.forEach(({ deviceId, count }) => {
                scheduledTaskCountMap[deviceId] = count
            })

            setScheduledTaskCounts(scheduledTaskCountMap)
            console.log('Scheduled task counts:', scheduledTaskCountMap)
        } catch (error) {
            console.error('Failed to fetch scheduled task counts:', error)
        }
    }

    const formatCronSchedule = (scheduling) => {
        try {
            if (!scheduling) return 'No schedule defined';

            // Debug logging
            console.log('formatCronSchedule input:', scheduling, 'type:', typeof scheduling);

            // Handle object format
            if (typeof scheduling === 'object' && scheduling !== null) {
                const { type, initial_day, day_interval, execution_time, next_execution } = scheduling;

                if (type === 'interval') {
                    // Format initial day
                    const initialDate = new Date(initial_day);
                    const formattedInitialDate = initialDate.toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric'
                    });

                    // Format execution time
                    const timeStr = execution_time || '00:00';

                    // Handle different intervals
                    if (day_interval === 1) {
                        return `Every day at ${timeStr} since ${formattedInitialDate}`;
                    } else if (day_interval > 1) {
                        return `Every ${day_interval} days at ${timeStr} since ${formattedInitialDate}`;
                    } else {
                        return `Interval schedule at ${timeStr} since ${formattedInitialDate}`;
                    }
                }

                // Handle other types if needed
                return `Schedule type: ${type}`;
            }

            // Handle string format (cron expressions)
            if (typeof scheduling === 'string') {
                const cronStr = String(scheduling);
                const parts = cronStr.split(' ');
                if (parts.length !== 6) {
                    return cronStr;
                }

                const [second, minute, hour, day, month, dayOfWeek] = parts;

                // Handle common patterns
                if (second === '0' && minute !== '*' && hour !== '*' && day === '*' && month === '*' && dayOfWeek === '*') {
                    // Daily at specific time
                    const timeStr = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
                    return `Every day at ${timeStr}`;
                }

                if (second === '0' && minute === '0' && hour !== '*' && day === '*' && month === '*' && dayOfWeek === '*') {
                    // Daily at specific hour
                    return `Every day at ${hour.padStart(2, '0')}:00`;
                }

                if (second === '0' && minute !== '*' && hour !== '*' && day !== '*' && day !== '*' && month === '*' && dayOfWeek === '*') {
                    // Specific day of month
                    const timeStr = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
                    return `Every ${day} of the month at ${timeStr}`;
                }

                if (second === '0' && minute !== '*' && hour !== '*' && day === '*' && month === '*' && dayOfWeek !== '*') {
                    // Specific day of week
                    const timeStr = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
                    const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
                    const dayName = dayNames[parseInt(dayOfWeek)] || dayOfWeek;
                    return `Every ${dayName} at ${timeStr}`;
                }

                // Handle intervals (e.g., every 3 days)
                if (day.startsWith('*/') && month === '*' && dayOfWeek === '*') {
                    const interval = day.substring(2);
                    const timeStr = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
                    return `Every ${interval} days at ${timeStr}`;
                }

                // Default fallback for string
                return cronStr;
            }

            // Default fallback
            return String(scheduling || 'Invalid schedule');
        } catch (error) {
            console.error('Error formatting schedule:', error, 'input:', scheduling);
            return String(scheduling || 'Invalid schedule');
        }
    };

    const handleDeviceAction = async (device, action) => {
        setSelectedDevice(device)
        setSelectedAction(action)

        try {
            switch (action) {
                case 'view-scheduled-tasks':
                    const scheduledTasksResponse = await fetch(`/api/tenants/${tenantId}/devices/${device.id}/scheduled-tasks`)
                    if (scheduledTasksResponse.ok) {
                        const data = await scheduledTasksResponse.json()
                        setRightPanelData({
                            type: 'scheduled-tasks',
                            data: data.data || [],
                            pagination: data.pagination
                        })
                    }
                    break

                case 'view-tasks':
                    const tasksResponse = await fetch(`/api/devices/${device.id}/tasks`)
                    if (tasksResponse.ok) {
                        const data = await tasksResponse.json()
                        setRightPanelData({
                            type: 'tasks',
                            data: data.data || [],
                            pagination: data.pagination
                        })
                    }
                    break

                case 'edit':
                    setRightPanelData({
                        type: 'edit-form',
                        device: device
                    })
                    break

                case 'view-details':
                    setRightPanelData({
                        type: 'device-details',
                        device: device
                    })
                    break

                default:
                    setRightPanelData(null)
            }
        } catch (error) {
            console.error(`Failed to fetch data for ${action}:`, error)
            showError(`Failed to load ${action} data`, 'Error')
        }
    }

    const handleCreateDevice = async (e) => {
        e.preventDefault()

        const result = await showApiNotification(
            fetch(`/api/tenants/${tenantId}/devices`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            }).then(res => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`)
                return res.json()
            }),
            'Device created successfully',
            'Failed to create device',
            'Success',
            'Error'
        )

        if (result.success) {
            setShowCreateForm(false)
            setFormData({ name: '', description: '', device_type: 'sensor', location: '', status: 'active' })
            fetchTenantAndDevices()
            setRightPanelData(null) // Clear the right panel
        }
    }

    const handleUpdateDevice = async (e) => {
        e.preventDefault()

        const result = await showApiNotification(
            fetch(`/api/devices/${editingDevice.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            }).then(res => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`)
                return res.json()
            }),
            'Device updated successfully',
            'Failed to update device',
            'Success',
            'Error'
        )

        if (result.success) {
            setEditingDevice(null)
            setFormData({ name: '', description: '', device_type: 'sensor', location: '', status: 'active' })
            fetchTenantAndDevices()
            setRightPanelData(null) // Clear the right panel
        }
    }

    const handleDeleteDevice = async (deviceId) => {
        if (!window.confirm('Are you sure you want to delete this device? This action cannot be undone.')) {
            return
        }

        const result = await showApiNotification(
            fetch(`/api/devices/${deviceId}`, {
                method: 'DELETE'
            }).then(res => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`)
                return res.json()
            }),
            'Device deleted successfully',
            'Failed to delete device',
            'Success',
            'Error'
        )

        if (result.success) {
            fetchTenantAndDevices()
        }
    }

    const startEdit = (device) => {
        setEditingDevice(device)
        setFormData({
            name: device.name || '',
            description: device.description || '',
            device_type: device.device_type || 'sensor',
            location: device.location || '',
            status: device.status || 'active'
        })
        // Also set the right panel data for the edit form
        setRightPanelData({
            type: 'edit-form',
            device: device
        })
    }

    const cancelEdit = () => {
        setEditingDevice(null)
        setFormData({ name: '', description: '', device_type: 'sensor', location: '', status: 'active' })
    }

    if (isLoading) {
        return (
            <div className="admin-devices">
                <div className="loading">Loading admin panel...</div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="admin-devices">
                <div className="access-denied">
                    <Cpu size={48} />
                    <h2>Access Denied</h2>
                    <p>You need admin privileges to access this page.</p>
                </div>
            </div>
        )
    }

    return (
        <div className="admin-devices">
            <div className="admin-header">
                <div className="header-top">
                    <Link to="/admin/tenants" className="back-link">
                        <ArrowLeft size={20} />
                        Back to Tenants
                    </Link>
                </div>
                <div className="admin-title">
                    <Cpu size={32} />
                    <div>
                        <h1>Device Management</h1>
                        {tenant && (
                            <p className="tenant-info">
                                <Building size={16} />
                                {tenant.name}
                            </p>
                        )}
                    </div>
                </div>
                <p className="admin-subtitle">Manage devices for this tenant</p>
            </div>

            <div className="admin-actions">
                <button
                    className="btn btn-primary"
                    onClick={() => {
                        setShowCreateForm(true)
                        setRightPanelData({
                            type: 'create-form'
                        })
                    }}
                >
                    <Plus size={20} />
                    Add Device
                </button>
            </div>


            {/* Main Content Area - 40/60 Split */}
            <div className="main-content">
                {/* Left Section - Device Cards (40%) */}
                <div className="devices-section">
                    {loading ? (
                        <div className="loading">Loading devices...</div>
                    ) : (
                        <div className="devices-list">
                            {Array.isArray(devices) && devices.map((device) => (
                                <div
                                    key={device.id}
                                    className={`device-card ${selectedDevice?.id === device.id ? 'selected' : ''}`}
                                >
                                    <div className="device-header">
                                        <div className="device-info">
                                            <h3>{device.name}</h3>
                                            <span className={`status-badge ${device.status}`}>
                                                {device.status}
                                            </span>
                                        </div>
                                        <div className="device-actions">
                                            <button
                                                className="btn-icon"
                                                onClick={() => handleDeviceAction(device, 'view-scheduled-tasks')}
                                                title="View Scheduled Tasks"
                                            >
                                                <Clock size={16} />
                                            </button>
                                            <button
                                                className="btn-icon"
                                                onClick={() => handleDeviceAction(device, 'view-tasks')}
                                                title="View Tasks"
                                            >
                                                <Activity size={16} />
                                            </button>
                                            <button
                                                className="btn-icon"
                                                onClick={() => handleDeviceAction(device, 'view-details')}
                                                title="View Details"
                                            >
                                                <Eye size={16} />
                                            </button>
                                            <button
                                                className="btn-icon"
                                                onClick={() => handleDeviceAction(device, 'edit')}
                                                title="Edit Device"
                                            >
                                                <Edit size={16} />
                                            </button>
                                            <button
                                                className="btn-icon danger"
                                                onClick={() => handleDeleteDevice(device.id)}
                                                title="Delete Device"
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </div>
                                    </div>

                                    <div className="device-details">
                                        {device.description && (
                                            <p className="device-description">{device.description}</p>
                                        )}
                                        {device.location && (
                                            <p className="device-location">
                                                <strong>Location:</strong> {device.location}
                                            </p>
                                        )}
                                        <p className="device-type">
                                            <strong>Type:</strong> {device.device_type}
                                        </p>
                                    </div>

                                    <div className="device-stats">
                                        <div className="stat">
                                            <Clock size={16} />
                                            <span>Tasks: {taskCounts[device.id] || 0}</span>
                                        </div>
                                        <div className="stat">
                                            <Calendar size={16} />
                                            <span>Scheduled: {scheduledTaskCounts[device.id] || 0}</span>
                                        </div>
                                        <div className="stat">
                                            <Activity size={16} />
                                            <span>Last Seen: {device.last_message_received_at ? new Date(device.last_message_received_at).toLocaleDateString('en-US', {
                                                month: 'short',
                                                day: 'numeric',
                                                hour: '2-digit',
                                                minute: '2-digit',
                                                hour12: false
                                            }) : 'Never'}</span>
                                        </div>
                                    </div>
                                </div>
                            ))}

                            {(!Array.isArray(devices) || devices.length === 0) && (
                                <div className="empty-state">
                                    <Cpu size={48} />
                                    <h3>No Devices Found</h3>
                                    <p>Add your first device to get started.</p>
                                </div>
                            )}
                        </div>
                    )}
                </div>

                {/* Right Section - Details Panel (60%) */}
                <div className="details-section">
                    {rightPanelData ? (
                        <div className="details-content">
                            {rightPanelData.type === 'scheduled-tasks' && (
                                <div className="scheduled-tasks-panel">
                                    <h3>Scheduled Tasks</h3>
                                    <div className="tasks-list">
                                        {rightPanelData.data.map((task) => (
                                            <div key={task.id} className="task-item">
                                                <div className="task-header">
                                                    <span className={`status-badge ${task.is_active ? 'active' : 'inactive'}`}>
                                                        {task.is_active ? 'Active' : 'Inactive'}
                                                    </span>
                                                </div>

                                                <div className="task-details">
                                                    <div className="task-detail-item">
                                                        <strong>Scheduling:</strong> {formatCronSchedule(task.scheduling)}
                                                    </div>

                                                    <div className="task-detail-item">
                                                        <strong>Commands:</strong>
                                                        <ul className="commands-list">
                                                            {task.commands && task.commands.map((command, index) => (
                                                                <li key={index} className="command-item">
                                                                    <span className="command-index">Index: {command.index}</span>
                                                                    <span className="command-value">Value: {command.value}</span>
                                                                    <span className="command-wait">Wait: {command.wait_for || '0s'}</span>
                                                                </li>
                                                            ))}
                                                        </ul>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                        {rightPanelData.data.length === 0 && (
                                            <p className="no-data">No scheduled tasks found</p>
                                        )}
                                    </div>
                                </div>
                            )}

                            {rightPanelData.type === 'tasks' && (
                                <div className="tasks-panel">
                                    <h3>Tasks</h3>
                                    <div className="tasks-table">
                                        <table>
                                            <thead>
                                                <tr>
                                                    <th>Name</th>
                                                    <th>Status</th>
                                                    <th>Created</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {rightPanelData.data.map((task) => (
                                                    <tr key={task.id}>
                                                        <td>{task.name || 'Unnamed Task'}</td>
                                                        <td>
                                                            <span className={`status-badge ${task.status || 'unknown'}`}>
                                                                {task.status || 'Unknown'}
                                                            </span>
                                                        </td>
                                                        <td>{task.created_at ? new Date(task.created_at).toLocaleDateString() : 'N/A'}</td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                        {rightPanelData.data.length === 0 && (
                                            <p className="no-data">No tasks found</p>
                                        )}
                                    </div>
                                </div>
                            )}

                            {rightPanelData.type === 'device-details' && (
                                <div className="device-details-panel">
                                    <h3>Device Details</h3>
                                    <div className="details-grid">
                                        <div className="detail-item">
                                            <strong>Name:</strong> {rightPanelData.device.name}
                                        </div>
                                        <div className="detail-item">
                                            <strong>Type:</strong> {rightPanelData.device.device_type}
                                        </div>
                                        <div className="detail-item">
                                            <strong>Status:</strong>
                                            <span className={`status-badge ${rightPanelData.device.status}`}>
                                                {rightPanelData.device.status}
                                            </span>
                                        </div>
                                        <div className="detail-item">
                                            <strong>Device EUI:</strong> {rightPanelData.device.dev_eui || 'N/A'}
                                        </div>
                                        <div className="detail-item">
                                            <strong>App EUI:</strong> {rightPanelData.device.app_eui || 'N/A'}
                                        </div>
                                        <div className="detail-item">
                                            <strong>Last Message:</strong> {rightPanelData.device.last_message_received_at ? new Date(rightPanelData.device.last_message_received_at).toLocaleString() : 'Never'}
                                        </div>
                                        {rightPanelData.device.description && (
                                            <div className="detail-item full-width">
                                                <strong>Description:</strong> {rightPanelData.device.description}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            )}

                            {rightPanelData.type === 'create-form' && (
                                <div className="create-form-panel">
                                    <h3>Add New Device</h3>
                                    <form onSubmit={handleCreateDevice}>
                                        <div className="form-group">
                                            <label htmlFor="create-name">Device Name *</label>
                                            <input
                                                type="text"
                                                id="create-name"
                                                value={formData.name}
                                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                                required
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="create-description">Description</label>
                                            <textarea
                                                id="create-description"
                                                value={formData.description}
                                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                                rows={3}
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="create-device_type">Device Type</label>
                                            <select
                                                id="create-device_type"
                                                value={formData.device_type}
                                                onChange={(e) => setFormData({ ...formData, device_type: e.target.value })}
                                            >
                                                <option value="sensor">Sensor</option>
                                                <option value="actuator">Actuator</option>
                                                <option value="gateway">Gateway</option>
                                                <option value="controller">Controller</option>
                                            </select>
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="create-location">Location</label>
                                            <input
                                                type="text"
                                                id="create-location"
                                                value={formData.location}
                                                onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="create-status">Status</label>
                                            <select
                                                id="create-status"
                                                value={formData.status}
                                                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                                            >
                                                <option value="active">Active</option>
                                                <option value="inactive">Inactive</option>
                                                <option value="maintenance">Maintenance</option>
                                            </select>
                                        </div>
                                        <div className="form-actions">
                                            <button type="button" className="btn btn-secondary" onClick={() => {
                                                setRightPanelData(null)
                                                setShowCreateForm(false)
                                            }}>
                                                Cancel
                                            </button>
                                            <button type="submit" className="btn btn-primary">
                                                Add Device
                                            </button>
                                        </div>
                                    </form>
                                </div>
                            )}

                            {rightPanelData.type === 'edit-form' && (
                                <div className="edit-form-panel">
                                    <h3>Edit Device</h3>
                                    <form onSubmit={handleUpdateDevice}>
                                        <div className="form-group">
                                            <label htmlFor="edit-name">Device Name *</label>
                                            <input
                                                type="text"
                                                id="edit-name"
                                                value={formData.name}
                                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                                required
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="edit-description">Description</label>
                                            <textarea
                                                id="edit-description"
                                                value={formData.description}
                                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                                rows={3}
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="edit-device_type">Device Type</label>
                                            <select
                                                id="edit-device_type"
                                                value={formData.device_type}
                                                onChange={(e) => setFormData({ ...formData, device_type: e.target.value })}
                                            >
                                                <option value="sensor">Sensor</option>
                                                <option value="actuator">Actuator</option>
                                                <option value="gateway">Gateway</option>
                                                <option value="controller">Controller</option>
                                            </select>
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="edit-location">Location</label>
                                            <input
                                                type="text"
                                                id="edit-location"
                                                value={formData.location}
                                                onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor="edit-status">Status</label>
                                            <select
                                                id="edit-status"
                                                value={formData.status}
                                                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                                            >
                                                <option value="active">Active</option>
                                                <option value="inactive">Inactive</option>
                                                <option value="maintenance">Maintenance</option>
                                            </select>
                                        </div>
                                        <div className="form-actions">
                                            <button type="button" className="btn btn-secondary" onClick={() => setRightPanelData(null)}>
                                                Cancel
                                            </button>
                                            <button type="submit" className="btn btn-primary">
                                                Update Device
                                            </button>
                                        </div>
                                    </form>
                                </div>
                            )}
                        </div>
                    ) : (
                        <div className="placeholder">
                            <div className="placeholder-content">
                                <Eye size={48} />
                                <h3>Select an action to display details</h3>
                                <p>Choose a device action to view information in this panel</p>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}

export default AdminDevices
