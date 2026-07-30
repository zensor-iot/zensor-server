import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import {
  ArrowLeft,
  Cpu,
  Building,
  Loader2,
  AlertCircle,
  Plus,
  Activity,
  Wifi,
  Battery,
  Circle
} from 'lucide-react'
import DeviceCard from './DeviceCard'
import { getApiUrl } from '../config/api'

const TenantDevices = () => {
  const { tenantId } = useParams()
  const [tenant, setTenant] = useState(null)
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    fetchTenantInfo()
    fetchTenantDevices()
  }, [tenantId])

  const fetchTenantInfo = async () => {
    try {
      const response = await fetch(getApiUrl(`/tenants/${tenantId}`))
      if (!response.ok) {
        throw new Error(`Failed to fetch tenant: ${response.statusText}`)
      }
      const data = await response.json()
      setTenant(data)
    } catch (err) {
      setError(err.message)
    }
  }

  const fetchTenantDevices = async () => {
    try {
      setLoading(true)
      const response = await fetch(getApiUrl(`/tenants/${tenantId}/devices`))
      if (!response.ok) {
        throw new Error(`Failed to fetch devices: ${response.statusText}`)
      }
      const data = await response.json()
      // Simulate some device metrics for demo
      const devicesWithMetrics = (data.data || []).map(device => ({
        ...device,
        batteryLevel: Math.floor(Math.random() * 100),
        signalStrength: Math.floor(Math.random() * 100),
        status: Math.random() > 0.3 ? 'online' : 'offline',
        lastSeen: new Date(Date.now() - Math.random() * 86400000) // Random within last 24h
      }))
      setDevices(devicesWithMetrics)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }



  const handleViewDetails = (device) => {
    console.log('View device details:', device)
    // TODO: Navigate to device details page
  }

  const handleSendCommand = (device) => {
    console.log('Send command to device:', device)
    // TODO: Open command send modal
  }

  if (error) {
    return (
      <div className="error-container">
        <AlertCircle className="error-icon" />
        <p className="error-message">Error: {error}</p>
        <div className="error-actions">
          <button onClick={() => {
            setError(null)
            fetchTenantInfo()
            fetchTenantDevices()
          }} className="retry-button">
            Retry
          </button>
          <Link to="/" className="back-button">
            <ArrowLeft size={16} />
            Back to Tenants
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="tenant-devices">
      <div className="page-header">
        <div className="breadcrumb">
          <Link to="/" className="breadcrumb-link">
            <ArrowLeft size={16} />
            Back to Tenants
          </Link>
        </div>

        {tenant && (
          <div className="tenant-header">
            <div className="tenant-header-info">
              <h2>
                <Building className="page-icon" />
                {tenant.name} - Devices
              </h2>
              <p className="page-description">
                {tenant.email} • {tenant.description}
              </p>
            </div>
            <div className="tenant-header-actions">
              <span className={`status-badge ${tenant.is_active ? 'active' : 'inactive'}`}>
                {tenant.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Device Stats */}
      <div className="device-stats">
        <div className="stat-card">
          <div className="stat-icon">
            <Cpu className="icon" />
          </div>
          <div className="stat-content">
            <h3>Total Devices</h3>
            <p className="stat-number">{devices.length}</p>
          </div>
        </div>

        <div className="stat-card online">
          <div className="stat-icon">
            <Activity className="icon" />
          </div>
          <div className="stat-content">
            <h3>Online</h3>
            <p className="stat-number">{devices.filter(d => d.status === 'online').length}</p>
          </div>
        </div>

        <div className="stat-card offline">
          <div className="stat-icon">
            <Circle className="icon" />
          </div>
          <div className="stat-content">
            <h3>Offline</h3>
            <p className="stat-number">{devices.filter(d => d.status === 'offline').length}</p>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="loading-container">
          <Loader2 className="loading-spinner" />
          <p>Loading devices...</p>
        </div>
      ) : (
        <div className="devices-section">
          <div className="devices-header">
            <h3>
              <Cpu className="section-icon" />
              Device Fleet ({devices.length})
            </h3>
            <button className="add-device-button">
              <Plus size={16} />
              Adopt Device
            </button>
          </div>

          {devices.length === 0 ? (
            <div className="empty-state">
              <Cpu size={64} className="empty-icon" />
              <h3>No devices found</h3>
              <p>This tenant doesn't have any devices yet. Adopt some devices to get started.</p>
              <button className="primary-button">
                <Plus size={16} />
                Adopt First Device
              </button>
            </div>
          ) : (
            <div className="devices-grid">
              {devices.map((device) => (
                <DeviceCard
                  key={device.id}
                  device={device}
                  onViewDetails={handleViewDetails}
                  onSendCommand={handleSendCommand}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default TenantDevices 