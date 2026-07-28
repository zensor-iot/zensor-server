import React, { useState, useEffect } from 'react'
import {
  Cpu,
  Wifi,
  Battery,
  Circle,
  MoreVertical,
  Edit3,
  Check,
  X,
  Droplets,
  Clock,
  Thermometer,
  Gauge,
  BarChart3,
  Power,
  Square
} from 'lucide-react'
import { getApiUrl, deviceCommandsApi } from '../config/api'
import ScheduledIrrigation from './ScheduledIrrigationWithNotifications'
import { useNotification } from '../hooks/useNotification'
import GrafanaVisualization from './GrafanaVisualization'

const TenantDeviceCard = ({ device, sensorData, onUpdateDisplayName }) => {
  const [isEditing, setIsEditing] = useState(false)
  const [editedDisplayName, setEditedDisplayName] = useState(device.display_name || device.name)
  const [isUpdating, setIsUpdating] = useState(false)
  const [irrigationMinutes, setIrrigationMinutes] = useState(5)
  const [isIrrigating, setIsIrrigating] = useState(false)
  const [hasReceivedFirstMessage, setHasReceivedFirstMessage] = useState(false)
  const [showGrafanaModal, setShowGrafanaModal] = useState(false)

  // Notification hook
  const { showSuccess, showError, showWarning } = useNotification()

  const getStatusColor = (status) => {
    switch (status) {
      case 'online': return '#10b981'
      case 'offline': return '#ef4444'
      default: return '#6b7280'
    }
  }

  const getSignalBars = (strength) => {
    if (strength >= 80) return 4
    if (strength >= 60) return 3
    if (strength >= 40) return 2
    if (strength >= 20) return 1
    return 0
  }

  const formatLastSeen = (date) => {
    if (!date) return 'Never'

    const now = new Date()
    const diffInMinutes = Math.floor((now - date) / (1000 * 60))

    if (diffInMinutes < 1) return 'Just now'
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`

    const diffInHours = Math.floor(diffInMinutes / 60)
    if (diffInHours < 24) return `${diffInHours}h ago`

    const diffInDays = Math.floor(diffInHours / 24)
    return `${diffInDays}d ago`
  }

  const getSensorIcon = (sensorType) => {
    const type = sensorType.toLowerCase()
    if (type.includes('temp')) return <Thermometer className="sensor-icon" />
    if (type.includes('humid') || type.includes('water')) return <Droplets className="sensor-icon" />
    if (type.includes('flow') || type.includes('pressure')) return <Gauge className="sensor-icon" />
    if (type.includes('relay')) return <Power className="sensor-icon" />
    return <BarChart3 className="sensor-icon" />
  }

  const getSensorUnit = (sensorType) => {
    const type = sensorType.toLowerCase()
    if (type.includes('temp')) return '°C'
    if (type.includes('humid')) return '%'
    if (type.includes('water') && type.includes('flow')) return 'L/min'
    if (type.includes('pressure')) return 'bar'
    return ''
  }

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString()
  }

  const isRelayActive = () => {
    if (!sensorData || !sensorData.data) return false

    // Check if any relay sensor has value 1 (On)
    for (const [sensorType, readings] of Object.entries(sensorData.data)) {
      if (sensorType.toLowerCase().includes('relay') && Array.isArray(readings)) {
        for (const reading of readings) {
          const value = reading.Value || reading.value
          if (value === 1) return true
        }
      }
    }

    // If we have sensor data but no relay data, or relay is off (value = 0), return false
    return false
  }

  const renderSensorData = () => {
    if (!sensorData || !sensorData.data) {
      return (
        <div className="sensor-data-section">
          <div className="sensor-data-header">
            <h4>Sensor Data</h4>
            <span className="no-data">No recent data</span>
          </div>
        </div>
      )
    }

    const entries = Object.entries(sensorData.data)
    if (entries.length === 0) {
      return (
        <div className="sensor-data-section">
          <div className="sensor-data-header">
            <h4>Sensor Data</h4>
            <span className="no-data">No sensor data</span>
          </div>
        </div>
      )
    }

    return (
      <div className="sensor-data-section">
        <div className="sensor-data-header">
          <h4>Latest Sensor Data</h4>
          <span className="data-timestamp">{formatTimestamp(sensorData.timestamp)}</span>
        </div>
        <div className="sensor-data-container">
          {entries.map(([sensorType, readings]) => {
            if (!Array.isArray(readings) || readings.length === 0) return null

            const unit = getSensorUnit(sensorType)
            const isRelay = sensorType.toLowerCase().includes('relay')

            return (
              <div key={sensorType} className="sensor-type">
                <div className="sensor-header">
                  {getSensorIcon(sensorType)}
                  <span className="sensor-name">{sensorType}</span>
                  <span className="sensor-count">({readings.length})</span>
                </div>
                <div className="sensor-readings">
                  {readings.map((reading, idx) => {
                    const value = reading.Value || reading.value
                    const displayValue = isRelay
                      ? (value === 1 ? 'On' : 'Off')
                      : (typeof value === 'number' ? value.toFixed(2) : value)

                    return (
                      <div key={idx} className="sensor-reading">
                        <span className="reading-index">#{reading.Index || reading.index}</span>
                        <span className={`reading-value ${isRelay ? (value === 1 ? 'relay-on' : 'relay-off') : ''}`}>
                          {displayValue}
                          {unit && !isRelay && <span className="reading-unit">{unit}</span>}
                        </span>
                      </div>
                    )
                  })}
                </div>
              </div>
            )
          }).filter(Boolean)}
        </div>
      </div>
    )
  }

  const handleSaveDisplayName = async () => {
    if (editedDisplayName.trim() === '') return

    setIsUpdating(true)
    try {
      await onUpdateDisplayName(device.id, editedDisplayName.trim())
      setIsEditing(false)
      showSuccess('Device name updated successfully', 'Name Updated')
    } catch (error) {
      console.error('Failed to update display name:', error)
      // Reset to original value on error
      setEditedDisplayName(device.display_name || device.name)
      showError('Failed to update device name', 'Update Failed')
    } finally {
      setIsUpdating(false)
    }
  }

  const handleCancelEdit = () => {
    setEditedDisplayName(device.display_name || device.name)
    setIsEditing(false)
  }

  const handleIrrigation = async () => {
    if (isIrrigating) {
      showWarning('Irrigation is already in progress', 'Already Active')
      return
    }

    if (device.status === 'offline') {
      showError('Cannot start irrigation - device is offline', 'Device Offline')
      return
    }

    if (isRelayActive()) {
      showWarning('Cannot start irrigation - relay is already active', 'Relay Active')
      return
    }

    if (!hasReceivedFirstMessage) {
      showWarning('Waiting for device data - please try again in a moment', 'Device Data Pending')
      return
    }

    setIsIrrigating(true)
    try {
      // Ensure duration is within valid range (1-60 minutes)
      const duration = Math.max(1, Math.min(60, irrigationMinutes))

      const taskPayload = {
        commands: [
          {
            wait_for: "0s", // Immediate activation
            priority: "HIGH",
            index: 1,
            value: 1 // Activation value
          },
          {
            wait_for: `${duration}m`, // Wait for user-specified minutes (1-60 minutes)
            priority: "HIGH",
            index: 1,
            value: 0 // Deactivation value
          }
        ]
      }

      const response = await fetch(getApiUrl(`/devices/${device.id}/tasks`), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(taskPayload)
      })

      if (!response.ok) {
        throw new Error(`Failed to start irrigation`)
      }

      const result = await response.json()
      console.log('Irrigation task created:', result)

      // Show success notification
      showSuccess(
        `Irrigation queued for ${duration} minute${duration !== 1 ? 's' : ''}!`,
        'Irrigation Queued',
        { duration: 5000 }
      )

      // Note: We no longer use setTimeout here
      // The irrigation state will be reset when we receive a WebSocket message
      // indicating the relay has been turned off (value = 0)

    } catch (error) {
      console.error('Failed to start irrigation:', error)
      showError(
        `Failed to start irrigation: ${error.message}`,
        'Irrigation Failed',
        { duration: 6000 }
      )
      setIsIrrigating(false)
    }
  }

  const handleStopIrrigation = async () => {
    if (!isRelayActive()) {
      showWarning('Irrigation is not active', 'Irrigation Not Active')
      return
    }

    if (device.status === 'offline') {
      showError('Cannot stop irrigation - device is offline', 'Device Offline')
      return
    }

    if (!hasReceivedFirstMessage) {
      showWarning('Waiting for device data - please try again in a moment', 'Device Data Pending')
      return
    }

    try {
      const commandData = {
        sequence: [
          {
            index: 1,
            value: 0,
            priority: "HIGH"
          }
        ]
      }

      const response = await deviceCommandsApi.sendCommand(device.id, commandData)
      console.log('Stop irrigation command sent:', response)

      showSuccess('Stop irrigation command sent successfully', 'Command Sent', { duration: 4000 })
    } catch (error) {
      console.error('Failed to stop irrigation:', error)
      showError(
        `Failed to stop irrigation: ${error.message}`,
        'Command Failed',
        { duration: 6000 }
      )
    }
  }

  // Monitor sensor data changes to reset irrigation state when relay turns off
  useEffect(() => {
    // Set flag when we receive first message
    if (sensorData && sensorData.data && !hasReceivedFirstMessage) {
      setHasReceivedFirstMessage(true)
    }

    if (isIrrigating && sensorData && sensorData.data) {
      // Check if any relay sensor has been turned off (value = 0)
      let relayTurnedOff = false
      for (const [sensorType, readings] of Object.entries(sensorData.data)) {
        if (sensorType.toLowerCase().includes('relay') && Array.isArray(readings)) {
          for (const reading of readings) {
            const value = reading.Value || reading.value
            if (value === 0) {
              relayTurnedOff = true
              break
            }
          }
          if (relayTurnedOff) break
        }
      }

      // If relay has been turned off, reset irrigation state
      if (relayTurnedOff) {
        setIsIrrigating(false)
        console.log('Irrigation completed - relay turned off')

        // Show completion notification
        const validatedDuration = Math.max(1, Math.min(60, irrigationMinutes))
        showSuccess(
          `Irrigation completed successfully after ${validatedDuration} minute${validatedDuration !== 1 ? 's' : ''}`,
          'Irrigation Completed',
          { duration: 4000 }
        )
      }
    }
  }, [sensorData, isIrrigating, hasReceivedFirstMessage, irrigationMinutes])

  return (
    <div className="tenant-device-card">
      {/* Device Header */}
      <div className="device-header">
        <div className="device-identity">
          <div className="device-icon-wrapper">
            <Cpu className="device-icon" />
            <div
              className="status-indicator"
              style={{ backgroundColor: getStatusColor(device.status) }}
            />
          </div>
          <div className="device-info">
            {isEditing ? (
              <div className="device-name-edit">
                <input
                  type="text"
                  value={editedDisplayName}
                  onChange={(e) => setEditedDisplayName(e.target.value)}
                  className="device-name-input"
                  placeholder="Enter display name"
                  disabled={isUpdating}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleSaveDisplayName()
                    if (e.key === 'Escape') handleCancelEdit()
                  }}
                  autoFocus
                />
                <div className="edit-actions">
                  <button
                    onClick={handleSaveDisplayName}
                    disabled={isUpdating || editedDisplayName.trim() === ''}
                    className="edit-save-btn"
                  >
                    <Check size={14} />
                  </button>
                  <button
                    onClick={handleCancelEdit}
                    disabled={isUpdating}
                    className="edit-cancel-btn"
                  >
                    <X size={14} />
                  </button>
                </div>
              </div>
            ) : (
              <h3 className="device-name">
                {device.display_name || device.name}
                <button
                  onClick={() => setIsEditing(true)}
                  className="edit-name-btn"
                  title="Edit display name"
                >
                  <Edit3 size={14} />
                </button>
              </h3>
            )}
            <p className="device-status">
              <Circle
                size={8}
                fill={getStatusColor(device.status)}
                color={getStatusColor(device.status)}
              />
              {device.status === 'online' ? 'Online' : 'Offline'} • {formatLastSeen(device.lastSeen)}
            </p>
          </div>
        </div>
        <button className="device-menu">
          <MoreVertical size={20} />
        </button>
      </div>

      {/* Device Metrics */}
      <div className="device-metrics">
        <div className="metric">
          <div className="metric-icon">
            <Wifi size={16} />
          </div>
          <div className="metric-content">
            <span className="metric-label">Signal</span>
            <div className="signal-bars">
              {[1, 2, 3, 4].map(bar => (
                <div
                  key={bar}
                  className={`signal-bar ${bar <= getSignalBars(device.signalStrength) ? 'active' : ''}`}
                />
              ))}
            </div>
          </div>
        </div>

        <div className="metric">
          <div className="metric-icon">
            <Battery size={16} />
          </div>
          <div className="metric-content">
            <span className="metric-label">Battery</span>
            <div className="battery-indicator">
              <div
                className="battery-level"
                style={{ width: `${device.batteryLevel}%` }}
              />
              <span className="battery-text">{device.batteryLevel}%</span>
            </div>
          </div>
        </div>
      </div>

      {/* Sensor Data Section */}
      {renderSensorData()}

      {/* Irrigation Control */}
      <div className="irrigation-control">
        <div className="irrigation-header">
          <div className="irrigation-title">
            <Droplets size={18} />
            <span>Irrigation Control</span>
          </div>
          {isIrrigating && (
            <div className="irrigation-status">
              <Circle size={8} fill="#10b981" color="#10b981" />
              <span>Active</span>
            </div>
          )}
        </div>

        <div className="irrigation-settings">
          <div className="duration-control">
            <label htmlFor={`duration-${device.id}`} className="duration-label">
              <Clock size={14} />
              Duration (minutes)
            </label>
            <div className="duration-slider-container">
              <input
                id={`duration-${device.id}`}
                type="range"
                min="1"
                max="60"
                value={irrigationMinutes}
                onChange={(e) => setIrrigationMinutes(parseInt(e.target.value))}
                className="duration-slider"
                disabled={isIrrigating || device.status === 'offline' || isRelayActive() || !hasReceivedFirstMessage}
              />
              <div className="duration-value">
                <span>{irrigationMinutes}</span>
              </div>
            </div>
            <div className="duration-display">
              <Droplets size={14} />
              <span>Will irrigate for {irrigationMinutes} minute{irrigationMinutes !== 1 ? 's' : ''}</span>
            </div>
          </div>

          <div className="irrigation-buttons">
            <button
              className={`irrigation-btn ${isIrrigating ? 'active' : ''}`}
              onClick={handleIrrigation}
              disabled={isIrrigating || device.status === 'offline' || isRelayActive() || !hasReceivedFirstMessage}
            >
              <Droplets size={16} />
              {isIrrigating ? `Irrigating (${irrigationMinutes}m)` : 'Start Irrigation'}
            </button>

            {isRelayActive() && (
              <button
                className="irrigation-btn stop-btn"
                onClick={handleStopIrrigation}
                disabled={device.status === 'offline' || !hasReceivedFirstMessage}
              >
                <Square size={16} />
                Stop Irrigation
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Scheduled Irrigation */}
      <ScheduledIrrigation
        tenantId={device.tenant_id}
        deviceId={device.id}
        deviceName={device.display_name || device.name}
      />

      {/* Analytics Button */}
      <div className="device-actions">
        <button
          className="action-btn secondary"
          onClick={() => setShowGrafanaModal(true)}
          title="View Sensor Data Visualization"
        >
          <BarChart3 size={16} />
          Analytics
        </button>
      </div>

      {/* Grafana Visualization Modal */}
      <GrafanaVisualization
        isOpen={showGrafanaModal}
        onClose={() => setShowGrafanaModal(false)}
        deviceName={device.display_name || device.name}
      />
    </div>
  )
}

export default TenantDeviceCard 