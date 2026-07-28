import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Building, Cpu, Play, Shield, Activity, Users, Database, RefreshCw } from 'lucide-react'
import { useAdmin } from '../../hooks/useAdmin'
import { useNotification } from '../../hooks/useNotification'
import './AdminDashboard.css'

const AdminDashboard = () => {
    const { isAdmin, isLoading } = useAdmin()
    const { showError } = useNotification()
    const [stats, setStats] = useState({
        totalTenants: 0,
        totalDevices: 0,
        systemHealth: 'healthy'
    })
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        if (!isLoading && !isAdmin) {
            showError('Access denied. Admin privileges required.', 'Unauthorized')
            return
        }

        if (isAdmin) {
            fetchSystemStats()
        }
    }, [isAdmin, isLoading])

    const fetchSystemStats = async () => {
        try {
            setLoading(true)

            // Fetch system-wide statistics
            const [tenantsResponse, devicesResponse] = await Promise.all([
                fetch('/api/tenants'),
                fetch('/api/devices')
            ])

            if (!tenantsResponse.ok || !devicesResponse.ok) {
                throw new Error('Failed to fetch basic system data')
            }

            const tenantsData = await tenantsResponse.json()
            const devicesData = await devicesResponse.json()

            // Handle paginated response format
            const tenants = Array.isArray(tenantsData) ? tenantsData : (tenantsData.data || [])
            const devices = Array.isArray(devicesData) ? devicesData : (devicesData.data || [])

            console.log('Admin Dashboard - Tenants:', tenants.length, 'Devices:', devices.length)

            setStats({
                totalTenants: tenants.length || 0,
                totalDevices: devices.length || 0,
                systemHealth: 'healthy'
            })
        } catch (error) {
            console.error('Failed to fetch system stats:', error)
            showError('Failed to load system statistics', 'Error')
        } finally {
            setLoading(false)
        }
    }

    if (isLoading) {
        return (
            <div className="admin-dashboard">
                <div className="loading">Loading admin dashboard...</div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="admin-dashboard">
                <div className="access-denied">
                    <Shield size={48} />
                    <h2>Access Denied</h2>
                    <p>You need admin privileges to access this page.</p>
                </div>
            </div>
        )
    }

    const quickActions = [
        {
            title: 'Manage Tenants',
            description: 'View and manage all tenant organizations',
            icon: Building,
            link: '/admin/tenants',
            color: 'blue'
        },
        {
            title: 'System Commands',
            description: 'Monitor and manage system-wide commands',
            icon: Play,
            link: '/admin/commands',
            color: 'green'
        },
        {
            title: 'System Health',
            description: 'Monitor system performance and health',
            icon: Activity,
            link: '/admin/health',
            color: 'orange'
        }
    ]

    const statCards = [
        {
            title: 'Total Tenants',
            value: stats.totalTenants,
            icon: Building,
            color: 'blue'
        },
        {
            title: 'Total Devices',
            value: stats.totalDevices,
            icon: Cpu,
            color: 'green'
        }
    ]

    return (
        <div className="admin-dashboard">
            <div className="admin-header">
                <div className="admin-title">
                    <Shield size={32} />
                    <h1>Admin Dashboard</h1>
                </div>
                <div className="admin-header-actions">
                    <button
                        className="refresh-button"
                        onClick={fetchSystemStats}
                        disabled={loading}
                        title="Refresh statistics"
                    >
                        <RefreshCw size={16} className={loading ? 'spinning' : ''} />
                        Refresh
                    </button>
                </div>
                <p className="admin-subtitle">System-wide management and monitoring</p>
            </div>

            {loading ? (
                <div className="loading">Loading system statistics...</div>
            ) : (
                <>
                    {/* Statistics Cards */}
                    <div className="stats-grid">
                        {statCards.map((stat, index) => (
                            <div key={index} className={`stat-card ${stat.color}`}>
                                <div className="stat-icon">
                                    <stat.icon size={24} />
                                </div>
                                <div className="stat-content">
                                    <div className="stat-value">{stat.value}</div>
                                    <div className="stat-title">{stat.title}</div>
                                </div>
                            </div>
                        ))}
                    </div>

                    {/* Quick Actions */}
                    <div className="quick-actions">
                        <h2>Quick Actions</h2>
                        <div className="actions-grid">
                            {quickActions.map((action, index) => (
                                <Link key={index} to={action.link} className={`action-card ${action.color}`}>
                                    <div className="action-icon">
                                        <action.icon size={24} />
                                    </div>
                                    <div className="action-content">
                                        <h3>{action.title}</h3>
                                        <p>{action.description}</p>
                                    </div>
                                </Link>
                            ))}
                        </div>
                    </div>

                    {/* System Health */}
                    <div className="system-health">
                        <h2>System Health</h2>
                        <div className="health-card">
                            <div className="health-status">
                                <div className={`health-indicator ${stats.systemHealth}`}></div>
                                <span className="health-text">System Status: {stats.systemHealth}</span>
                            </div>
                            <div className="health-details">
                                <div className="health-item">
                                    <Database size={16} />
                                    <span>Database: Connected</span>
                                </div>
                                <div className="health-item">
                                    <Activity size={16} />
                                    <span>API: Operational</span>
                                </div>
                                <div className="health-item">
                                    <Users size={16} />
                                    <span>WebSocket: Active</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </div>
    )
}

export default AdminDashboard
