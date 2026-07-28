import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Play, ArrowLeft, Send, Clock, CheckCircle, XCircle, AlertCircle, Building, Cpu } from 'lucide-react'
import { useAdmin } from '../../hooks/useAdmin'
import { useNotification } from '../../hooks/useNotification'
import { deviceCommandsApi } from '../../config/api'
import './AdminCommands.css'

const AdminCommands = () => {
    const { isAdmin, isLoading } = useAdmin()
    const { showSuccess, showError, showApiNotification } = useNotification()
    const [devices, setDevices] = useState([])
    const [recentCommands, setRecentCommands] = useState([])
    const [loading, setLoading] = useState(true)
    const [showCommandForm, setShowCommandForm] = useState(false)
    const [formData, setFormData] = useState({
        device_id: '',
        index: 1,
        value: 0,
        priority: 'NORMAL',
        wait_for: '0s',
        description: ''
    })

    useEffect(() => {
        if (!isLoading && !isAdmin) {
            showError('Access denied. Admin privileges required.', 'Unauthorized')
            return
        }

        if (isAdmin) {
            fetchData()
        }
    }, [isAdmin, isLoading])

    const fetchData = async () => {
        try {
            setLoading(true)

            // Fetch all devices
            const devicesResponse = await fetch('/api/devices')
            if (devicesResponse.ok) {
                const devicesData = await devicesResponse.json()
                setDevices(devicesData)
            }

            // TODO: Fetch recent commands from a system-wide endpoint
            // For now, we'll use an empty array
            setRecentCommands([])
        } catch (error) {
            console.error('Failed to fetch data:', error)
            showError('Failed to load system data', 'Error')
        } finally {
            setLoading(false)
        }
    }

    const handleSendCommand = async (e) => {
        e.preventDefault()

        if (!formData.device_id) {
            showError('Please select a device', 'Validation Error')
            return
        }

        const commandData = {
            index: parseInt(formData.index),
            value: parseInt(formData.value),
            priority: formData.priority,
            wait_for: formData.wait_for,
            description: formData.description
        }

        const result = await showApiNotification(
            deviceCommandsApi.sendCommand(formData.device_id, commandData),
            'Command sent successfully',
            'Failed to send command',
            'Success',
            'Error'
        )

        if (result.success) {
            setShowCommandForm(false)
            setFormData({
                device_id: '',
                index: 1,
                value: 0,
                priority: 'NORMAL',
                wait_for: '0s',
                description: ''
            })
            fetchData()
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

    if (isLoading) {
        return (
            <div className="admin-commands">
                <div className="loading">Loading admin panel...</div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="admin-commands">
                <div className="access-denied">
                    <Play size={48} />
                    <h2>Access Denied</h2>
                    <p>You need admin privileges to access this page.</p>
                </div>
            </div>
        )
    }

    return (
        <div className="admin-commands">
            <div className="admin-header">
                <div className="header-top">
                    <Link to="/admin" className="back-link">
                        <ArrowLeft size={20} />
                        Back to Dashboard
                    </Link>
                </div>
                <div className="admin-title">
                    <Play size={32} />
                    <h1>System Commands</h1>
                </div>
                <p className="admin-subtitle">Send commands to any device in the system</p>
            </div>

            <div className="admin-actions">
                <button
                    className="btn btn-primary"
                    onClick={() => setShowCommandForm(true)}
                >
                    <Send size={20} />
                    Send Command
                </button>
            </div>

            {/* Command Form */}
            {showCommandForm && (
                <div className="form-overlay">
                    <div className="form-card">
                        <h2>Send Device Command</h2>
                        <form onSubmit={handleSendCommand}>
                            <div className="form-group">
                                <label htmlFor="device_id">Device *</label>
                                <select
                                    id="device_id"
                                    value={formData.device_id}
                                    onChange={(e) => setFormData({ ...formData, device_id: e.target.value })}
                                    required
                                >
                                    <option value="">Select a device</option>
                                    {devices.map((device) => (
                                        <option key={device.id} value={device.id}>
                                            {device.name} ({device.device_type}) - {device.tenant?.name || 'Unknown Tenant'}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div className="form-row">
                                <div className="form-group">
                                    <label htmlFor="index">Index</label>
                                    <input
                                        type="number"
                                        id="index"
                                        value={formData.index}
                                        onChange={(e) => setFormData({ ...formData, index: parseInt(e.target.value) || 1 })}
                                        min="1"
                                        max="10"
                                    />
                                </div>

                                <div className="form-group">
                                    <label htmlFor="value">Value</label>
                                    <input
                                        type="number"
                                        id="value"
                                        value={formData.value}
                                        onChange={(e) => setFormData({ ...formData, value: parseInt(e.target.value) || 0 })}
                                        min="0"
                                        max="100"
                                    />
                                </div>
                            </div>

                            <div className="form-row">
                                <div className="form-group">
                                    <label htmlFor="priority">Priority</label>
                                    <select
                                        id="priority"
                                        value={formData.priority}
                                        onChange={(e) => setFormData({ ...formData, priority: e.target.value })}
                                    >
                                        <option value="LOW">Low</option>
                                        <option value="NORMAL">Normal</option>
                                        <option value="HIGH">High</option>
                                        <option value="URGENT">Urgent</option>
                                    </select>
                                </div>

                                <div className="form-group">
                                    <label htmlFor="wait_for">Wait For</label>
                                    <input
                                        type="text"
                                        id="wait_for"
                                        value={formData.wait_for}
                                        onChange={(e) => setFormData({ ...formData, wait_for: e.target.value })}
                                        placeholder="0s, 5m, 1h"
                                    />
                                </div>
                            </div>

                            <div className="form-group">
                                <label htmlFor="description">Description</label>
                                <textarea
                                    id="description"
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    rows={3}
                                    placeholder="Optional description for this command"
                                />
                            </div>

                            <div className="form-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowCommandForm(false)}>
                                    Cancel
                                </button>
                                <button type="submit" className="btn btn-primary">
                                    Send Command
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* System Overview */}
            <div className="system-overview">
                <h2>System Overview</h2>
                <div className="overview-cards">
                    <div className="overview-card">
                        <div className="card-icon">
                            <Cpu size={24} />
                        </div>
                        <div className="card-content">
                            <div className="card-value">{devices.length}</div>
                            <div className="card-label">Total Devices</div>
                        </div>
                    </div>

                    <div className="overview-card">
                        <div className="card-icon">
                            <Play size={24} />
                        </div>
                        <div className="card-content">
                            <div className="card-value">{recentCommands.length}</div>
                            <div className="card-label">Recent Commands</div>
                        </div>
                    </div>

                    <div className="overview-card">
                        <div className="card-icon">
                            <CheckCircle size={24} />
                        </div>
                        <div className="card-content">
                            <div className="card-value">
                                {recentCommands.filter(c => c.status === 'completed').length}
                            </div>
                            <div className="card-label">Completed</div>
                        </div>
                    </div>

                    <div className="overview-card">
                        <div className="card-icon">
                            <XCircle size={24} />
                        </div>
                        <div className="card-content">
                            <div className="card-value">
                                {recentCommands.filter(c => c.status === 'failed').length}
                            </div>
                            <div className="card-label">Failed</div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Recent Commands */}
            <div className="recent-commands">
                <h2>Recent Commands</h2>
                {recentCommands.length > 0 ? (
                    <div className="commands-list">
                        {recentCommands.map((command) => (
                            <div key={command.id} className="command-card">
                                <div className="command-header">
                                    <div className="command-info">
                                        <div className="command-status">
                                            {getStatusIcon(command.status)}
                                            <span className="status-text" style={{ color: getStatusColor(command.status) }}>
                                                {command.status}
                                            </span>
                                        </div>
                                        <div className="command-time">
                                            {new Date(command.created_at).toLocaleString()}
                                        </div>
                                    </div>
                                    <div className="command-device">
                                        <Cpu size={16} />
                                        {command.device?.name || 'Unknown Device'}
                                    </div>
                                </div>

                                <div className="command-details">
                                    <div className="detail-row">
                                        <span className="detail-label">Index:</span>
                                        <span className="detail-value">{command.index}</span>
                                    </div>
                                    <div className="detail-row">
                                        <span className="detail-label">Value:</span>
                                        <span className="detail-value">{command.value}</span>
                                    </div>
                                    <div className="detail-row">
                                        <span className="detail-label">Priority:</span>
                                        <span className="detail-value">{command.priority}</span>
                                    </div>
                                    {command.description && (
                                        <div className="detail-row">
                                            <span className="detail-label">Description:</span>
                                            <span className="detail-value">{command.description}</span>
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="empty-state">
                        <Play size={48} />
                        <h3>No Recent Commands</h3>
                        <p>Commands sent through the admin panel will appear here.</p>
                    </div>
                )}
            </div>
        </div>
    )
}

export default AdminCommands
