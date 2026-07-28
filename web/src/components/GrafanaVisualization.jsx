import { X, Clock } from 'lucide-react'
import { useState } from 'react'
import config from '../config/api'

const GrafanaVisualization = ({ isOpen, onClose, deviceName }) => {
    const [timeRange, setTimeRange] = useState('6h') // Default to 6 hours

    // Time range options
    const timeRangeOptions = [
        { value: '1h', label: 'Last hour' },
        { value: '6h', label: 'Last 6 hours' },
        { value: '24h', label: 'Last 24 hours' }
    ]

    if (!isOpen) return null

    // Generate time range based on selection
    const now = Date.now()
    const getTimeRange = (range) => {
        switch (range) {
            case '1h':
                return { from: now - (1 * 60 * 60 * 1000), to: now }
            case '6h':
                return { from: now - (6 * 60 * 60 * 1000), to: now }
            case '24h':
                return { from: now - (24 * 60 * 60 * 1000), to: now }
            default:
                return { from: now - (6 * 60 * 60 * 1000), to: now }
        }
    }

    const { from, to } = getTimeRange(timeRange)

    // Build Grafana iframe URL with authentication and time range
    const baseUrl = `${config.grafanaBaseUrl}/grafana/d-solo/ces2boi7cdh4wc/sensor-data?panelId=1&var-device_name=${encodeURIComponent(deviceName)}&from=${from}&to=${to}`

    // Add API key if available
    const grafanaUrl = config.grafanaApiKey
        ? `${baseUrl}&authToken=${config.grafanaApiKey}`
        : baseUrl

    const handleIframeLoad = () => {
        console.log('Grafana iframe loaded successfully')
        setIframeLoading(false)
        setIframeError(false)
    }

    const handleTimeRangeChange = (newTimeRange) => {
        setTimeRange(newTimeRange)
    }

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content grafana-modal" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <h3>Sensor Data Visualization</h3>
                    <button onClick={onClose} className="modal-close-btn">
                        <X size={20} />
                    </button>
                </div>

                <div className="grafana-container">
                    <div className="grafana-info">
                        <div className="grafana-info-row">
                            <p>Device: <strong>{deviceName}</strong></p>
                            <div className="time-range-selector">
                                <Clock size={16} />
                                <select
                                    value={timeRange}
                                    onChange={(e) => handleTimeRangeChange(e.target.value)}
                                    className="time-range-select"
                                >
                                    {timeRangeOptions.map(option => (
                                        <option key={option.value} value={option.value}>
                                            {option.label}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </div>

                    <div className="grafana-iframe-container">
                        <iframe
                            src={grafanaUrl}
                            width="100%"
                            height="600"
                            frameBorder="0"
                            title="Grafana Sensor Data Visualization"
                            className="grafana-iframe"
                            loading="lazy"
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GrafanaVisualization 