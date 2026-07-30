import { Cpu, Signal, Wifi, Key, Circle, Clock, BarChart3 } from 'lucide-react'
import { useState } from 'react'
import GrafanaVisualization from './GrafanaVisualization'

const DeviceCard = ({ device, onViewDetails, onSendCommand }) => {
  const [showGrafanaModal, setShowGrafanaModal] = useState(false)
  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text)
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'online': return '#10b981'
      case 'offline': return '#ef4444'
      default: return '#6b7280'
    }
  }

  const formatLastSeen = (date) => {
    if (!date) return 'Never'

    const lastSeenDate = typeof date === 'string' ? new Date(date) : date
    const now = new Date()
    const diffInMinutes = Math.floor((now - lastSeenDate) / (1000 * 60))

    if (diffInMinutes < 1) return 'Just now'
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`

    const diffInHours = Math.floor(diffInMinutes / 60)
    if (diffInHours < 24) return `${diffInHours}h ago`

    const diffInDays = Math.floor(diffInHours / 24)
    return `${diffInDays}d ago`
  }

  return (
    <div className="device-card">
      <div className="device-card-header">
        <div className="device-icon">
          <Cpu size={24} />
          <div
            className="status-indicator"
            style={{ backgroundColor: getStatusColor(device.status) }}
          />
        </div>
        <div className="device-info">
          <h4 className="device-name">{device.display_name || device.name}</h4>
          <p className="device-id">ID: {device.id}</p>
          <p className="device-status">
            <Circle
              size={8}
              fill={getStatusColor(device.status)}
              color={getStatusColor(device.status)}
            />
            {device.status === 'online' ? 'Online' : 'Offline'} • {formatLastSeen(device.last_message_received_at)}
          </p>
        </div>
      </div>

      <div className="device-details">
        <div className="detail-row">
          <Clock className="detail-icon" />
          <span className="detail-label">Last Seen:</span>
          <span className="detail-value">
            {formatLastSeen(device.last_message_received_at)}
          </span>
        </div>

        <div className="detail-row">
          <Signal className="detail-icon" />
          <span className="detail-label">App EUI:</span>
          <code
            className="detail-value clickable"
            onClick={() => copyToClipboard(device.app_eui)}
            title="Click to copy"
          >
            {device.app_eui}
          </code>
        </div>

        <div className="detail-row">
          <Wifi className="detail-icon" />
          <span className="detail-label">Dev EUI:</span>
          <code
            className="detail-value clickable"
            onClick={() => copyToClipboard(device.dev_eui)}
            title="Click to copy"
          >
            {device.dev_eui}
          </code>
        </div>

        <div className="detail-row">
          <Key className="detail-icon" />
          <span className="detail-label">App Key:</span>
          <code
            className="detail-value clickable"
            onClick={() => copyToClipboard(device.app_key)}
            title="Click to copy"
          >
            {device.app_key.substring(0, 16)}...
          </code>
        </div>

        {device.tenant_id && (
          <div className="detail-row">
            <Cpu className="detail-icon" />
            <span className="detail-label">Tenant:</span>
            <span className="detail-value">
              {device.tenant_id}
            </span>
          </div>
        )}
      </div>

      <div className="device-actions">
        <button
          className="action-button primary"
          onClick={() => onViewDetails && onViewDetails(device)}
        >
          View Details
        </button>
        <button
          className="action-button secondary"
          onClick={() => onSendCommand && onSendCommand(device)}
        >
          Send Command
        </button>
        <button
          className="action-button secondary"
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

export default DeviceCard 