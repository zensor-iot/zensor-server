import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Activity, Database, Server, Wifi, CheckCircle, XCircle, AlertCircle, RefreshCw } from 'lucide-react'
import { useAdmin } from '../../hooks/useAdmin'
import { useNotification } from '../../hooks/useNotification'
import './AdminHealth.css'

const AdminHealth = () => {
    const { isAdmin, isLoading } = useAdmin()
    const { showError } = useNotification()
    const [healthData, setHealthData] = useState({
        api: { status: 'checking', responseTime: 0, lastCheck: null },
        database: { status: 'checking', responseTime: 0, lastCheck: null },
        websocket: { status: 'checking', responseTime: 0, lastCheck: null }
    })
    const [loading, setLoading] = useState(true)
    const [lastRefresh, setLastRefresh] = useState(null)

    useEffect(() => {
        if (!isLoading && !isAdmin) {
            showError('Access denied. Admin privileges required.', 'Unauthorized')
            return
        }

        if (isAdmin) {
            checkSystemHealth()
        }
    }, [isAdmin, isLoading])

    const checkSystemHealth = async () => {
        try {
            setLoading(true)
            const startTime = Date.now()

            // Check API health
            const apiStartTime = Date.now()
            try {
                const apiResponse = await fetch('/health', { timeout: 5000 })
                const apiResponseTime = Date.now() - apiStartTime

                if (apiResponse.ok) {
                    const apiData = await apiResponse.json()
                    setHealthData(prev => ({
                        ...prev,
                        api: {
                            status: 'healthy',
                            responseTime: apiResponseTime,
                            lastCheck: new Date(),
                            details: apiData
                        }
                    }))
                } else {
                    setHealthData(prev => ({
                        ...prev,
                        api: {
                            status: 'unhealthy',
                            responseTime: apiResponseTime,
                            lastCheck: new Date(),
                            error: `HTTP ${apiResponse.status}`
                        }
                    }))
                }
            } catch (error) {
                const apiResponseTime = Date.now() - apiStartTime
                setHealthData(prev => ({
                    ...prev,
                    api: {
                        status: 'unhealthy',
                        responseTime: apiResponseTime,
                        lastCheck: new Date(),
                        error: error.message
                    }
                }))
            }

            // Check database health (via API)
            const dbStartTime = Date.now()
            try {
                const dbResponse = await fetch('/api/tenants', { timeout: 5000 })
                const dbResponseTime = Date.now() - dbStartTime

                if (dbResponse.ok) {
                    setHealthData(prev => ({
                        ...prev,
                        database: {
                            status: 'healthy',
                            responseTime: dbResponseTime,
                            lastCheck: new Date()
                        }
                    }))
                } else {
                    setHealthData(prev => ({
                        ...prev,
                        database: {
                            status: 'unhealthy',
                            responseTime: dbResponseTime,
                            lastCheck: new Date(),
                            error: `HTTP ${dbResponse.status}`
                        }
                    }))
                }
            } catch (error) {
                const dbResponseTime = Date.now() - dbStartTime
                setHealthData(prev => ({
                    ...prev,
                    database: {
                        status: 'unhealthy',
                        responseTime: dbResponseTime,
                        lastCheck: new Date(),
                        error: error.message
                    }
                }))
            }

            // Check WebSocket health (simplified check)
            const wsStartTime = Date.now()
            try {
                // Try to create a WebSocket connection
                const wsUrl = window.location.origin.replace(/^http/, 'ws') + '/ws/device-messages'
                const ws = new WebSocket(wsUrl)

                const wsPromise = new Promise((resolve, reject) => {
                    const timeout = setTimeout(() => {
                        ws.close()
                        reject(new Error('WebSocket connection timeout'))
                    }, 5000)

                    ws.onopen = () => {
                        clearTimeout(timeout)
                        ws.close()
                        resolve(true)
                    }

                    ws.onerror = (error) => {
                        clearTimeout(timeout)
                        reject(error)
                    }
                })

                await wsPromise
                const wsResponseTime = Date.now() - wsStartTime

                setHealthData(prev => ({
                    ...prev,
                    websocket: {
                        status: 'healthy',
                        responseTime: wsResponseTime,
                        lastCheck: new Date()
                    }
                }))
            } catch (error) {
                const wsResponseTime = Date.now() - wsStartTime
                setHealthData(prev => ({
                    ...prev,
                    websocket: {
                        status: 'unhealthy',
                        responseTime: wsResponseTime,
                        lastCheck: new Date(),
                        error: error.message
                    }
                }))
            }

            setLastRefresh(new Date())
        } catch (error) {
            console.error('Failed to check system health:', error)
            showError('Failed to check system health', 'Error')
        } finally {
            setLoading(false)
        }
    }

    const getStatusIcon = (status) => {
        switch (status) {
            case 'healthy':
                return <CheckCircle size={20} className="status-icon healthy" />
            case 'unhealthy':
                return <XCircle size={20} className="status-icon unhealthy" />
            case 'checking':
                return <AlertCircle size={20} className="status-icon checking" />
            default:
                return <AlertCircle size={20} className="status-icon unknown" />
        }
    }

    const getStatusText = (status) => {
        switch (status) {
            case 'healthy':
                return 'Healthy'
            case 'unhealthy':
                return 'Unhealthy'
            case 'checking':
                return 'Checking...'
            default:
                return 'Unknown'
        }
    }

    if (isLoading) {
        return (
            <div className="admin-health">
                <div className="loading">Loading admin panel...</div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="admin-health">
                <div className="access-denied">
                    <Activity size={48} />
                    <h2>Access Denied</h2>
                    <p>You need admin privileges to access this page.</p>
                </div>
            </div>
        )
    }

    const overallStatus = Object.values(healthData).every(service => service.status === 'healthy')
        ? 'healthy'
        : Object.values(healthData).some(service => service.status === 'unhealthy')
            ? 'unhealthy'
            : 'checking'

    return (
        <div className="admin-health">
            <div className="admin-header">
                <div className="header-top">
                    <Link to="/admin" className="back-link">
                        <ArrowLeft size={20} />
                        Back to Dashboard
                    </Link>
                </div>
                <div className="admin-title">
                    <Activity size={32} />
                    <h1>System Health</h1>
                </div>
                <p className="admin-subtitle">Monitor system performance and health status</p>
            </div>

            <div className="health-actions">
                <button
                    className="btn btn-primary"
                    onClick={checkSystemHealth}
                    disabled={loading}
                >
                    <RefreshCw size={20} className={loading ? 'spinning' : ''} />
                    {loading ? 'Checking...' : 'Refresh Health'}
                </button>
                {lastRefresh && (
                    <span className="last-refresh">
                        Last checked: {lastRefresh.toLocaleTimeString()}
                    </span>
                )}
            </div>

            {/* Overall System Status */}
            <div className="overall-status">
                <div className={`status-card ${overallStatus}`}>
                    <div className="status-header">
                        {getStatusIcon(overallStatus)}
                        <h2>Overall System Status</h2>
                    </div>
                    <div className="status-details">
                        <span className="status-text">{getStatusText(overallStatus)}</span>
                        {overallStatus === 'healthy' && (
                            <p>All systems are operational</p>
                        )}
                        {overallStatus === 'unhealthy' && (
                            <p>Some systems are experiencing issues</p>
                        )}
                        {overallStatus === 'checking' && (
                            <p>Checking system status...</p>
                        )}
                    </div>
                </div>
            </div>

            {/* Individual Service Status */}
            <div className="services-grid">
                <div className="service-card">
                    <div className="service-header">
                        <Server size={24} />
                        <h3>API Server</h3>
                    </div>
                    <div className="service-status">
                        {getStatusIcon(healthData.api.status)}
                        <span>{getStatusText(healthData.api.status)}</span>
                    </div>
                    {healthData.api.responseTime > 0 && (
                        <div className="service-metrics">
                            <div className="metric">
                                <span className="metric-label">Response Time:</span>
                                <span className="metric-value">{healthData.api.responseTime}ms</span>
                            </div>
                            {healthData.api.lastCheck && (
                                <div className="metric">
                                    <span className="metric-label">Last Check:</span>
                                    <span className="metric-value">{healthData.api.lastCheck.toLocaleTimeString()}</span>
                                </div>
                            )}
                        </div>
                    )}
                    {healthData.api.error && (
                        <div className="service-error">
                            <span className="error-text">{healthData.api.error}</span>
                        </div>
                    )}
                </div>

                <div className="service-card">
                    <div className="service-header">
                        <Database size={24} />
                        <h3>Database</h3>
                    </div>
                    <div className="service-status">
                        {getStatusIcon(healthData.database.status)}
                        <span>{getStatusText(healthData.database.status)}</span>
                    </div>
                    {healthData.database.responseTime > 0 && (
                        <div className="service-metrics">
                            <div className="metric">
                                <span className="metric-label">Response Time:</span>
                                <span className="metric-value">{healthData.database.responseTime}ms</span>
                            </div>
                            {healthData.database.lastCheck && (
                                <div className="metric">
                                    <span className="metric-label">Last Check:</span>
                                    <span className="metric-value">{healthData.database.lastCheck.toLocaleTimeString()}</span>
                                </div>
                            )}
                        </div>
                    )}
                    {healthData.database.error && (
                        <div className="service-error">
                            <span className="error-text">{healthData.database.error}</span>
                        </div>
                    )}
                </div>

                <div className="service-card">
                    <div className="service-header">
                        <Wifi size={24} />
                        <h3>WebSocket</h3>
                    </div>
                    <div className="service-status">
                        {getStatusIcon(healthData.websocket.status)}
                        <span>{getStatusText(healthData.websocket.status)}</span>
                    </div>
                    {healthData.websocket.responseTime > 0 && (
                        <div className="service-metrics">
                            <div className="metric">
                                <span className="metric-label">Response Time:</span>
                                <span className="metric-value">{healthData.websocket.responseTime}ms</span>
                            </div>
                            {healthData.websocket.lastCheck && (
                                <div className="metric">
                                    <span className="metric-label">Last Check:</span>
                                    <span className="metric-value">{healthData.websocket.lastCheck.toLocaleTimeString()}</span>
                                </div>
                            )}
                        </div>
                    )}
                    {healthData.websocket.error && (
                        <div className="service-error">
                            <span className="error-text">{healthData.websocket.error}</span>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}

export default AdminHealth
