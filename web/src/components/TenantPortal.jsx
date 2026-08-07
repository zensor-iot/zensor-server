import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import {
  Cpu,
  Activity,
  Wifi,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  AlertCircle,
  RefreshCw
} from 'lucide-react'
import TenantDeviceCard from './TenantDeviceCard'
import useWebSocket from '../hooks/useWebSocket'
import { getApiUrl, getWebSocketUrl } from '../config/api'

const TenantPortal = () => {
  const { tenantId } = useParams()
  const [tenant, setTenant] = useState(null)
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [deviceSensorData, setDeviceSensorData] = useState({}) // Store latest sensor data for each device
  const [lastUpdated, setLastUpdated] = useState(new Date())

  // WebSocket connection for real-time sensor data
  const wsUrl = getWebSocketUrl('/ws/device-messages')
  const { isConnected, lastMessage } = useWebSocket(wsUrl)

  // Process incoming WebSocket messages to store latest sensor data
  useEffect(() => {
    if (lastMessage && lastMessage.type === 'device_state') {
      setDeviceSensorData(prev => ({
        ...prev,
        [lastMessage.device_id]: {
          data: lastMessage.data,
          timestamp: lastMessage.timestamp,
          receivedAt: new Date()
        }
      }))
    }
  }, [lastMessage])

  useEffect(() => {
    fetchTenantInfo()
    fetchTenantDevices()
  }, [tenantId])

  // Partial refresh for device status only (no loading state)
  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(getApiUrl(`/tenants/${tenantId}/devices`))
        if (response.ok) {
          const data = await response.json()
          const updatedDevices = (data.data || []).map(device => ({
            ...device,
            status: device.status,
            lastSeen: device.last_message_received_at ? new Date(device.last_message_received_at) : null,
            batteryLevel: 100, // Always show 100% battery
            signalStrength: Math.floor(Math.random() * 100)
          }))

          // Only update if devices have actually changed
          setDevices(prevDevices => {
            const hasChanges = JSON.stringify(prevDevices.map(d => ({ id: d.id, status: d.status, lastSeen: d.lastSeen }))) !==
              JSON.stringify(updatedDevices.map(d => ({ id: d.id, status: d.status, lastSeen: d.lastSeen })))

            if (hasChanges) {
              setLastUpdated(new Date())
              return updatedDevices
            }
            return prevDevices
          })
        }
      } catch (err) {
        // Silently fail for partial refresh - don't show error to user
        console.warn('Partial refresh failed:', err.message)
      }
    }, 30000) // Refresh every 30 seconds

    return () => clearInterval(interval)
  }, [tenantId])

  const handleManualRefresh = async () => {
    try {
      const response = await fetch(getApiUrl(`/v1/tenants/${tenantId}/devices`))
      if (response.ok) {
        const data = await response.json()
        const updatedDevices = (data.data || []).map(device => ({
          ...device,
          status: device.status,
          lastSeen: device.last_message_received_at ? new Date(device.last_message_received_at) : null,
          batteryLevel: 100, // Always show 100% battery
          signalStrength: Math.floor(Math.random() * 100)
        }))

        setDevices(updatedDevices)
        setLastUpdated(new Date())
      }
    } catch (err) {
      console.warn('Manual refresh failed:', err.message)
    }
  }

  const fetchTenantInfo = async () => {
    try {
      const response = await fetch(getApiUrl(`/tenants/${tenantId}`))
      if (!response.ok) {
        throw new Error(`Failed to fetch tenant information`)
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
        throw new Error(`Failed to fetch devices`)
      }
      const data = await response.json()
      // Use real device status and timestamp from API
      const devicesWithStatus = (data.data || []).map(device => ({
        ...device,
        // Use real status from API
        status: device.status,
        // Use real last message timestamp from API
        lastSeen: device.last_message_received_at ? new Date(device.last_message_received_at) : null,
        // Keep simulated values for metrics not yet implemented in API
        batteryLevel: 100, // Always show 100% battery
        signalStrength: Math.floor(Math.random() * 100)
      }))
      setDevices(devicesWithStatus)
      setLastUpdated(new Date())
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateDisplayName = async (deviceId, displayName) => {
    try {
      const response = await fetch(getApiUrl(`/devices/${deviceId}`), {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          display_name: displayName
        })
      })

      if (!response.ok) {
        throw new Error(`Failed to update device display name`)
      }

      // Update the local devices state
      setDevices(prevDevices =>
        prevDevices.map(device =>
          device.id === deviceId
            ? { ...device, display_name: displayName }
            : device
        )
      )
    } catch (err) {
      throw new Error(`Failed to update display name: ${err.message}`)
    }
  }

  const deviceStats = {
    total: devices.length,
    online: devices.filter(d => d.status === 'online').length,
    offline: devices.filter(d => d.status === 'offline').length
  }

  if (error) {
    return (
      <div className="tenant-portal error">
        <div className="error-container">
          <AlertCircle className="error-icon" />
          <h2>Unable to load your devices</h2>
          <p className="error-message">{error}</p>
          <button onClick={() => {
            setError(null)
            fetchTenantInfo()
            fetchTenantDevices()
          }} className="retry-button">
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="tenant-portal">
      {/* Portal Header */}
      <header className="portal-header">
        <div className="portal-header-content">
          <div className="portal-branding">
            <Activity className="portal-logo" />
            <div className="portal-title">
              <h1>Device Management Portal</h1>
              {tenant && (
                <p className="tenant-name">{tenant.name}</p>
              )}
            </div>
          </div>
          <div className="portal-status">
            <div className="last-updated">
              <Clock size={16} />
              <span>Last updated: {lastUpdated.toLocaleTimeString()}</span>
              <button
                onClick={handleManualRefresh}
                className="refresh-button"
                title="Refresh device data"
              >
                <RefreshCw size={16} />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Dashboard Stats */}
      <section className="dashboard-stats">
        <div className="stat-card">
          <div className="stat-icon">
            <Cpu className="icon" />
          </div>
          <div className="stat-content">
            <h3>Total Devices</h3>
            <p className="stat-number">{deviceStats.total}</p>
          </div>
        </div>

        <div className="stat-card online">
          <div className="stat-icon">
            <CheckCircle className="icon" />
          </div>
          <div className="stat-content">
            <h3>Online</h3>
            <p className="stat-number">{deviceStats.online}</p>
          </div>
        </div>

        <div className="stat-card offline">
          <div className="stat-icon">
            <XCircle className="icon" />
          </div>
          <div className="stat-content">
            <h3>Offline</h3>
            <p className="stat-number">{deviceStats.offline}</p>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">
            <Wifi className="icon" />
          </div>
          <div className="stat-content">
            <h3>Health Score</h3>
            <p className="stat-number">{Math.round((deviceStats.online / deviceStats.total) * 100)}%</p>
          </div>
        </div>
      </section>

      {/* Device Management Section */}
      <section className="device-management">
        <div className="section-header">
          <h2>Your Devices</h2>
        </div>

        {/* Device Grid */}
        {loading ? (
          <div className="loading-container">
            <Loader2 className="loading-spinner" />
            <p>Loading your devices...</p>
          </div>
        ) : devices.length === 0 ? (
          <div className="empty-state">
            <Cpu size={64} className="empty-icon" />
            <h3>No devices found</h3>
            <p>Contact your administrator to add devices to your fleet</p>
          </div>
        ) : (
          <div className="device-grid">
            {devices.map((device) => (
              <TenantDeviceCard
                key={device.id}
                device={{ ...device, tenant_id: tenantId }}
                sensorData={deviceSensorData[device.name]} // Pass latest sensor data
                onUpdateDisplayName={handleUpdateDisplayName}
              />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

export default TenantPortal 