import { useState, useEffect, useMemo } from 'react';
import useWebSocket from '../hooks/useWebSocket';
import { Activity, Wifi, WifiOff, RotateCcw, Eye, EyeOff, BarChart3, Thermometer, Droplets, Gauge, Clock, AlertCircle, CheckCircle, XCircle, RefreshCw, Power } from 'lucide-react';
import { getWebSocketUrl } from '../config/api';

const DeviceMessagesLive = () => {
  const [messages, setMessages] = useState([]);
  const [selectedDevice, setSelectedDevice] = useState('all');
  const [maxMessages, setMaxMessages] = useState(50);
  const [showRawData, setShowRawData] = useState(false);
  const [expandedMessages, setExpandedMessages] = useState(new Set());

  const wsUrl = getWebSocketUrl('/ws/device-messages');
  const { isConnected, lastMessage, error } = useWebSocket(wsUrl);

  // Add new messages to the list
  useEffect(() => {
    if (lastMessage && lastMessage.type === 'device_state') {
      setMessages(prev => {
        const newMessages = [{
          ...lastMessage,
          id: `${lastMessage.device_id}-${lastMessage.timestamp}-${Date.now()}`,
          receivedAt: new Date()
        }, ...prev];

        // Keep only the most recent messages
        return newMessages.slice(0, maxMessages);
      });
    }
  }, [lastMessage, maxMessages]);

  // Get unique device IDs for filtering
  const deviceIds = useMemo(() => [...new Set(messages.map(msg => msg.device_id))], [messages]);

  // Filter messages by selected device
  const filteredMessages = selectedDevice === 'all'
    ? messages
    : messages.filter(msg => msg.device_id === selectedDevice);

  // Calculate message statistics
  const messageStats = useMemo(() => {
    const stats = {
      totalMessages: messages.length,
      uniqueDevices: deviceIds.length,
      lastActivity: messages.length > 0 ? messages[0].receivedAt : null,
      sensorTypes: new Set()
    };

    messages.forEach(msg => {
      if (msg.data && typeof msg.data === 'object') {
        Object.keys(msg.data).forEach(key => stats.sensorTypes.add(key));
      }
    });

    return stats;
  }, [messages, deviceIds]);

  const clearMessages = () => {
    setMessages([]);
    setExpandedMessages(new Set());
  };

  const toggleMessageExpansion = (messageId) => {
    setExpandedMessages(prev => {
      const newSet = new Set(prev);
      if (newSet.has(messageId)) {
        newSet.delete(messageId);
      } else {
        newSet.add(messageId);
      }
      return newSet;
    });
  };

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString();
  };

  const getSensorIcon = (sensorType) => {
    const type = sensorType.toLowerCase();
    if (type.includes('temp')) return <Thermometer className="sensor-icon" />;
    if (type.includes('humid') || type.includes('water')) return <Droplets className="sensor-icon" />;
    if (type.includes('flow') || type.includes('pressure')) return <Gauge className="sensor-icon" />;
    if (type.includes('relay')) return <Power className="sensor-icon" />;
    return <BarChart3 className="sensor-icon" />;
  };

  const getSensorUnit = (sensorType) => {
    const type = sensorType.toLowerCase();
    if (type.includes('temp')) return '°C';
    if (type.includes('humid')) return '%';
    if (type.includes('water') && type.includes('flow')) return 'L/min';
    if (type.includes('pressure')) return 'bar';
    return '';
  };

  const formatSensorData = (data, messageId) => {
    if (!data) return 'No data';

    const entries = Object.entries(data);
    if (entries.length === 0) return 'No sensor data';

    const isExpanded = expandedMessages.has(messageId);

    return (
      <div className="sensor-data-container">
        {entries.map(([sensorType, readings]) => {
          if (!Array.isArray(readings) || readings.length === 0) return null;

          const unit = getSensorUnit(sensorType);
          const isRelay = sensorType.toLowerCase().includes('relay');

          return (
            <div key={sensorType} className="sensor-type">
              <div className="sensor-header">
                {getSensorIcon(sensorType)}
                <span className="sensor-name">{sensorType}</span>
                <span className="sensor-count">({readings.length} readings)</span>
              </div>
              <div className="sensor-readings">
                {readings.slice(0, isExpanded ? readings.length : 3).map((reading, idx) => {
                  const value = reading.Value || reading.value;
                  const displayValue = isRelay
                    ? (value === 1 ? 'On' : 'Off')
                    : (typeof value === 'number' ? value.toFixed(2) : value);

                  return (
                    <div key={idx} className="sensor-reading">
                      <span className="reading-index">#{reading.Index || reading.index}</span>
                      <span className={`reading-value ${isRelay ? (value === 1 ? 'relay-on' : 'relay-off') : ''}`}>
                        {displayValue}
                        {unit && !isRelay && <span className="reading-unit">{unit}</span>}
                      </span>
                    </div>
                  );
                })}
                {readings.length > 3 && (
                  <button
                    className="expand-button"
                    onClick={() => toggleMessageExpansion(messageId)}
                  >
                    {isExpanded ? 'Show less' : `Show ${readings.length - 3} more`}
                  </button>
                )}
              </div>
            </div>
          );
        }).filter(Boolean)}
      </div>
    );
  };

  const formatRawData = (message) => {
    return (
      <div className="raw-data">
        <pre>{JSON.stringify(message, null, 2)}</pre>
      </div>
    );
  };

  const getConnectionStatus = () => {
    if (error) {
      return {
        icon: <XCircle className="status-icon error" />,
        text: 'Connection Error',
        className: 'error'
      };
    }

    if (isConnected) {
      return {
        icon: <CheckCircle className="status-icon connected" />,
        text: 'Connected',
        className: 'connected'
      };
    }

    return {
      icon: <RefreshCw className="status-icon reconnecting" />,
      text: 'Connecting...',
      className: 'reconnecting'
    };
  };

  const status = getConnectionStatus();

  return (
    <div className="device-messages-live">
      <div className="header">
        <div className="title-section">
          <Activity className="icon" />
          <h2>Live Device Messages</h2>
        </div>

        <div className="status-section">
          <div className={`connection-status ${status.className}`}>
            {status.icon}
            <span>{status.text}</span>
          </div>

          {error && (
            <div className="error-display">
              <span className="error-text">{error}</span>
            </div>
          )}
        </div>
      </div>

      {/* Message Statistics */}
      <div className="stats-section">
        <div className="stat-card">
          <span className="stat-value">{messageStats.totalMessages}</span>
          <span className="stat-label">Total Messages</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{messageStats.uniqueDevices}</span>
          <span className="stat-label">Active Devices</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{messageStats.sensorTypes.size}</span>
          <span className="stat-label">Sensor Types</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">
            {messageStats.lastActivity ? formatTimestamp(messageStats.lastActivity) : 'None'}
          </span>
          <span className="stat-label">Last Activity</span>
        </div>
      </div>

      <div className="controls">
        <div className="filter-section">
          <label htmlFor="device-filter">Filter by device:</label>
          <select
            id="device-filter"
            value={selectedDevice}
            onChange={(e) => setSelectedDevice(e.target.value)}
          >
            <option value="all">All devices ({messages.length} messages)</option>
            {deviceIds.map(deviceId => {
              const count = messages.filter(msg => msg.device_id === deviceId).length;
              return (
                <option key={deviceId} value={deviceId}>
                  {deviceId} ({count} messages)
                </option>
              );
            })}
          </select>
        </div>

        <div className="actions">
          <button
            onClick={() => setShowRawData(!showRawData)}
            className={`toggle-button ${showRawData ? 'active' : ''}`}
          >
            {showRawData ? <EyeOff className="icon" /> : <Eye className="icon" />}
            {showRawData ? 'Hide' : 'Show'} Raw Data
          </button>

          <button onClick={clearMessages} className="clear-button">
            Clear Messages
          </button>

          <div className="max-messages">
            <label htmlFor="max-messages">Max messages:</label>
            <select
              id="max-messages"
              value={maxMessages}
              onChange={(e) => setMaxMessages(parseInt(e.target.value))}
            >
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
              <option value={200}>200</option>
            </select>
          </div>
        </div>
      </div>

      {error && (
        <div className="error-banner">
          <strong>Connection Error:</strong> {error}
        </div>
      )}

      <div className="messages-container">
        {filteredMessages.length === 0 ? (
          <div className="no-messages">
            {isConnected ? 'Waiting for device messages...' : 'Connect to start receiving messages'}
          </div>
        ) : (
          <div className="messages-list">
            {filteredMessages.map(message => (
              <div key={message.id} className="message-card">
                <div className="message-header">
                  <div className="device-info">
                    <strong className="device-id">{message.device_id}</strong>
                    <span className="timestamp">{formatTimestamp(message.timestamp)}</span>
                  </div>
                  <div className="received-time">
                    Received: {formatTimestamp(message.receivedAt)}
                  </div>
                </div>

                <div className="message-content">
                  {showRawData ? (
                    formatRawData(message)
                  ) : (
                    <div className="message-data">
                      {formatSensorData(message.data, message.id)}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <style>{`
        .device-messages-live {
          width: 100%;
          padding: 20px;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }

        .header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
          padding-bottom: 10px;
          border-bottom: 2px solid #e1e5e9;
        }

        .title-section {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .title-section h2 {
          margin: 0;
          color: #2c3e50;
        }

        .icon {
          width: 20px;
          height: 20px;
        }

        .status-section {
          display: flex;
          align-items: center;
          gap: 15px;
        }

        .connection-status {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          border-radius: 6px;
          font-weight: 500;
        }

        .connection-status.connected {
          background-color: #d4edda;
          color: #155724;
        }

        .connection-status.disconnected {
          background-color: #f8d7da;
          color: #721c24;
        }

        .status-icon {
          width: 16px;
          height: 16px;
        }

        .retry-button {
          display: flex;
          align-items: center;
          gap: 5px;
          padding: 6px 12px;
          background-color: #007bff;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
        }

        .retry-button:hover {
          background-color: #0056b3;
        }

        .stats-section {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 15px;
          margin-bottom: 20px;
        }

        .stat-card {
          background-color: white;
          padding: 15px;
          border-radius: 8px;
          border: 1px solid #e1e5e9;
          text-align: center;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .stat-value {
          display: block;
          font-size: 24px;
          font-weight: bold;
          color: #2c3e50;
          margin-bottom: 5px;
        }

        .stat-label {
          display: block;
          font-size: 12px;
          color: #6c757d;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .controls {
          display: flex;
          justify-content: space-between;
          align-items: center;
          flex-wrap: wrap;
          gap: 15px;
          margin-bottom: 20px;
          padding: 15px;
          background-color: #f8f9fa;
          border-radius: 8px;
        }

        .filter-section {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .actions {
          display: flex;
          align-items: center;
          gap: 15px;
        }

        .max-messages {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        select, .clear-button, .toggle-button {
          padding: 6px 10px;
          border: 1px solid #ced4da;
          border-radius: 4px;
          font-size: 14px;
          cursor: pointer;
        }

        .clear-button {
          background-color: #6c757d;
          color: white;
        }

        .clear-button:hover {
          background-color: #545b62;
        }

        .toggle-button {
          display: flex;
          align-items: center;
          gap: 5px;
          background-color: #fff;
          color: #6c757d;
        }

        .toggle-button.active {
          background-color: #007bff;
          color: white;
        }

        .toggle-button:hover {
          background-color: #e9ecef;
        }

        .toggle-button.active:hover {
          background-color: #0056b3;
        }

        .error-banner {
          background-color: #f8d7da;
          color: #721c24;
          padding: 12px;
          border-radius: 6px;
          margin-bottom: 20px;
          border: 1px solid #f5c6cb;
        }

        .messages-container {
          min-height: 400px;
        }

        .no-messages {
          text-align: center;
          color: #6c757d;
          padding: 60px 20px;
          font-style: italic;
        }

        .messages-list {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .message-card {
          background-color: white;
          border: 1px solid #e1e5e9;
          border-radius: 8px;
          padding: 15px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
          transition: box-shadow 0.2s;
        }

        .message-card:hover {
          box-shadow: 0 4px 8px rgba(0,0,0,0.15);
        }

        .message-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 15px;
          padding-bottom: 10px;
          border-bottom: 1px solid #e9ecef;
        }

        .device-info {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .device-id {
          color: #2c3e50;
          font-size: 16px;
        }

        .timestamp {
          color: #6c757d;
          font-size: 14px;
        }

        .received-time {
          color: #868e96;
          font-size: 12px;
        }

        .message-content {
          min-height: 20px;
        }

        .message-data {
          display: flex;
          flex-direction: column;
          gap: 15px;
        }

        .sensor-data-container {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
          gap: 15px;
        }

        /* Ensure maximum 4 columns for sensor types on large screens */
        @media (min-width: 1200px) {
          .sensor-data-container {
            grid-template-columns: repeat(4, 1fr);
          }
        }

        /* 3 columns on medium-large screens */
        @media (min-width: 900px) and (max-width: 1199px) {
          .sensor-data-container {
            grid-template-columns: repeat(3, 1fr);
          }
        }

        /* 2 columns on medium screens */
        @media (min-width: 600px) and (max-width: 899px) {
          .sensor-data-container {
            grid-template-columns: repeat(2, 1fr);
          }
        }

        .sensor-type {
          background-color: #f8f9fa;
          padding: 12px;
          border-radius: 6px;
          border: 1px solid #e9ecef;
        }

        .sensor-header {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 10px;
        }

        .sensor-icon {
          width: 16px;
          height: 16px;
          color: #6c757d;
        }

        .sensor-name {
          font-weight: 600;
          color: #495057;
          text-transform: capitalize;
        }

        .sensor-count {
          color: #6c757d;
          font-size: 12px;
        }

        .sensor-readings {
          display: flex;
          flex-direction: column;
          gap: 6px;
          width: 100%;
        }

        .sensor-reading {
          display: flex;
          align-items: center;
          justify-content: space-between;
          gap: 12px;
          background-color: white;
          padding: 8px 12px;
          border-radius: 4px;
          font-size: 13px;
          border: 1px solid #dee2e6;
          transition: all 0.2s ease;
          width: 100%;
        }

        .sensor-reading:hover {
          border-color: #007bff;
          box-shadow: 0 2px 4px rgba(0,123,255,0.1);
        }

        .reading-index {
          color: #6c757d;
          font-size: 11px;
          font-weight: 500;
          flex-shrink: 0;
        }

        .reading-value {
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-weight: 600;
          color: #2c3e50;
          text-align: right;
        }

        .reading-value.relay-on {
          color: #28a745;
          font-weight: 700;
        }

        .reading-value.relay-off {
          color: #dc3545;
          font-weight: 700;
        }

        .reading-unit {
          color: #6c757d;
          font-size: 11px;
          margin-left: 2px;
        }

        .expand-button {
          background: none;
          border: 1px solid #007bff;
          color: #007bff;
          padding: 4px 8px;
          border-radius: 3px;
          font-size: 11px;
          cursor: pointer;
        }

        .expand-button:hover {
          background-color: #007bff;
          color: white;
        }

        .raw-data {
          background-color: #f8f9fa;
          border: 1px solid #e9ecef;
          border-radius: 4px;
          padding: 15px;
          overflow-x: auto;
        }

        .raw-data pre {
          margin: 0;
          font-size: 12px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          color: #495057;
        }

        @media (max-width: 768px) {
          .header {
            flex-direction: column;
            align-items: flex-start;
            gap: 10px;
          }

          .stats-section {
            grid-template-columns: repeat(2, 1fr);
          }

          .controls {
            flex-direction: column;
            align-items: flex-start;
          }

          .message-header {
            flex-direction: column;
            align-items: flex-start;
            gap: 5px;
          }

          .sensor-data-container {
            grid-template-columns: 1fr;
            gap: 12px;
          }

          .sensor-readings {
            gap: 4px;
          }

          .actions {
            flex-wrap: wrap;
            width: 100%;
          }
        }
      `}</style>
    </div>
  );
};

export default DeviceMessagesLive; 