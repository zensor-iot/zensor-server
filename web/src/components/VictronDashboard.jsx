import { useState, useEffect } from 'react';
import useWebSocket from '../hooks/useWebSocket';
import { getWebSocketUrl } from '../config/api';
import {
  Sun, Battery, Zap, Gauge, Thermometer, Activity,
  Wifi, WifiOff, RefreshCw, Power, Droplets
} from 'lucide-react';

const VictronDashboard = () => {
  const [systemStatus, setSystemStatus] = useState(null);
  const [lastUpdate, setLastUpdate] = useState(null);
  const [timeAgo, setTimeAgo] = useState('');

  const wsUrl = getWebSocketUrl('/ws/victron/status');

  const { isConnected, lastMessage, connectionError, connectionStatus } = useWebSocket(wsUrl);

  useEffect(() => {
    if (lastMessage && lastMessage.type === 'victron_status') {
      setSystemStatus(lastMessage);
      setLastUpdate(new Date());
    }
  }, [lastMessage]);

  useEffect(() => {
    const interval = setInterval(() => {
      if (lastUpdate) {
        const seconds = Math.floor((new Date() - lastUpdate) / 1000);
        if (seconds < 5) setTimeAgo('just now');
        else if (seconds < 60) setTimeAgo(`${seconds}s ago`);
        else setTimeAgo(`${Math.floor(seconds / 60)}m ago`);
      }
    }, 1000);
    return () => clearInterval(interval);
  }, [lastUpdate]);

  const statusColor = isConnected ? '#10b981' : '#ef4444';
  const summary = systemStatus?.system;
  const data = systemStatus?.data;

  const solarPower = summary?.solar_power ?? 0;
  const batteryPower = summary?.battery_power ?? 0;
  const loadPower = summary?.ac_load_power ?? 0;
  const gridPower = summary?.grid_power ?? 0;

  return (
    <div className="victron-dashboard">
      <div className="victron-header">
        <div className="victron-title-section">
          <Activity size={28} className="victron-logo-icon" />
          <h1>Energy System</h1>
        </div>
        <div className="victron-connection-status">
          <span className="connection-indicator" style={{ backgroundColor: statusColor }} />
          <span className="connection-text">{isConnected ? 'Connected' : 'Disconnected'}</span>
          {!isConnected && (
            <RefreshCw size={16} className="reconnect-icon" onClick={() => window.location.reload()} />
          )}
        </div>
      </div>

      {!systemStatus && (
        <div className="victron-loading">
          <Activity size={48} className="loading-spinner" />
          <p>Connecting to energy system...</p>
        </div>
      )}

      {connectionError && (
        <div className="victron-error-banner">
          <WifiOff size={20} />
          <span>Connection error: {connectionError}</span>
        </div>
      )}

      {systemStatus && (
        <>
          <div className="victron-summary-cards">
            <div className="victron-card solar">
              <div className="card-icon"><Sun size={32} /></div>
              <div className="card-content">
                <span className="card-label">Solar</span>
                <span className="card-value">{solarPower.toFixed(0)} <small>W</small></span>
              </div>
            </div>
            <div className="victron-card battery">
              <div className="card-icon"><Battery size={32} /></div>
              <div className="card-content">
                <span className="card-label">Battery</span>
                <span className="card-value">{summary?.battery_soc?.toFixed(0) ?? '--'} <small>%</small></span>
                <span className="card-sub">{summary?.battery_voltage?.toFixed(1) ?? '--'} V</span>
              </div>
            </div>
            <div className="victron-card load">
              <div className="card-icon"><Zap size={32} /></div>
              <div className="card-content">
                <span className="card-label">Load</span>
                <span className="card-value">{loadPower.toFixed(0)} <small>W</small></span>
              </div>
            </div>
            <div className="victron-card grid">
              <div className="card-icon"><Power size={32} /></div>
              <div className="card-content">
                <span className="card-label">Grid</span>
                <span className="card-value">{gridPower.toFixed(0)} <small>W</small></span>
              </div>
            </div>
          </div>

          <div className="victron-power-flow">
            <h2>Power Flow</h2>
            <div className="flow-diagram">
              <div className="flow-node solar-node">
                <Sun size={24} />
                <span>Solar</span>
                <strong>{solarPower.toFixed(0)} W</strong>
              </div>
              <div className="flow-arrow">
                <Zap size={20} />
              </div>
              <div className="flow-node battery-node">
                <Battery size={24} />
                <span>Battery</span>
                <strong>{summary?.battery_soc?.toFixed(0) ?? '--'}%</strong>
                <small>{batteryPower >= 0 ? `+${batteryPower.toFixed(0)} W` : `${batteryPower.toFixed(0)} W`}</small>
                {summary?.is_charging && <span className="badge charging">Charging</span>}
                {summary?.is_inverting && <span className="badge inverting">Inverting</span>}
              </div>
              <div className="flow-arrow">
                <Zap size={20} />
              </div>
              <div className="flow-node load-node">
                <Zap size={24} />
                <span>Load</span>
                <strong>{loadPower.toFixed(0)} W</strong>
              </div>
            </div>
          </div>

          {data?.batteries && data.batteries.length > 0 && (
            <div className="victron-section">
              <h2><Battery size={20} /> Batteries</h2>
              <div className="victron-detail-cards">
                {data.batteries.map((b, i) => (
                  <div key={i} className="detail-card">
                    <div className="detail-title">Battery #{b.instance}</div>
                    <div className="detail-grid">
                      <div className="detail-item">
                        <Gauge size={16} />
                        <div><small>SOC</small><strong>{b.soc?.toFixed(0) ?? '--'}%</strong></div>
                      </div>
                      <div className="detail-item">
                        <Zap size={16} />
                        <div><small>Voltage</small><strong>{b.voltage?.toFixed(2) ?? '--'}V</strong></div>
                      </div>
                      <div className="detail-item">
                        <Activity size={16} />
                        <div><small>Current</small><strong>{b.current?.toFixed(1) ?? '--'}A</strong></div>
                      </div>
                      <div className="detail-item">
                        <Zap size={16} />
                        <div><small>Power</small><strong>{b.power?.toFixed(0) ?? '--'}W</strong></div>
                      </div>
                      {b.temperature > 0 && (
                        <div className="detail-item">
                          <Thermometer size={16} />
                          <div><small>Temp</small><strong>{b.temperature?.toFixed(1)}°C</strong></div>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {data?.solar_chargers && data.solar_chargers.length > 0 && (
            <div className="victron-section">
              <h2><Sun size={20} /> Solar Chargers</h2>
              <div className="victron-detail-cards">
                {data.solar_chargers.map((s, i) => (
                  <div key={i} className="detail-card">
                    <div className="detail-title">Solar Charger #{s.instance}</div>
                    <div className="detail-grid">
                      <div className="detail-item">
                        <Zap size={16} />
                        <div><small>Power</small><strong>{s.power?.toFixed(0) ?? '--'}W</strong></div>
                      </div>
                      <div className="detail-item">
                        <Gauge size={16} />
                        <div><small>Voltage</small><strong>{s.voltage?.toFixed(1) ?? '--'}V</strong></div>
                      </div>
                      <div className="detail-item">
                        <Activity size={16} />
                        <div><small>Current</small><strong>{s.current?.toFixed(1) ?? '--'}A</strong></div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {data?.ac_loads && data.ac_loads.length > 0 && (
            <div className="victron-section">
              <h2><Zap size={20} /> AC Loads</h2>
              <div className="victron-detail-cards">
                {data.ac_loads.map((l, i) => (
                  <div key={i} className="detail-card">
                    <div className="detail-title">AC Load #{l.instance}</div>
                    <div className="detail-grid">
                      <div className="detail-item">
                        <Zap size={16} />
                        <div><small>Power</small><strong>{l.power?.toFixed(0) ?? '--'}W</strong></div>
                      </div>
                      <div className="detail-item">
                        <Gauge size={16} />
                        <div><small>Voltage</small><strong>{l.voltage?.toFixed(1) ?? '--'}V</strong></div>
                      </div>
                      <div className="detail-item">
                        <Activity size={16} />
                        <div><small>Current</small><strong>{l.current?.toFixed(1) ?? '--'}A</strong></div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {data?.vebus && data.vebus.length > 0 && (
            <div className="victron-section">
              <h2><Activity size={20} /> Inverter / Charger</h2>
              <div className="victron-detail-cards">
                {data.vebus.map((v, i) => (
                  <div key={i} className="detail-card">
                    <div className="detail-title">VE.Bus #{v.instance}</div>
                    <div className="detail-grid">
                      <div className="detail-item">
                        <Zap size={16} />
                        <div><small>Power</small><strong>{v.power?.toFixed(0) ?? '--'}W</strong></div>
                      </div>
                      <div className="detail-item">
                        <Gauge size={16} />
                        <div><small>State</small><strong>{getVebusState(v.state)}</strong></div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {data?.temperatures && data.temperatures.length > 0 && (
            <div className="victron-section">
              <h2><Thermometer size={20} /> Temperatures</h2>
              <div className="victron-detail-cards">
                {data.temperatures.map((t, i) => (
                  <div key={i} className="detail-card">
                    <div className="detail-title">Sensor #{t.instance}</div>
                    <div className="detail-grid">
                      <div className="detail-item">
                        <Thermometer size={16} />
                        <div><strong>{t.temperature?.toFixed(1)}°C</strong></div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {data?.tanks && data.tanks.length > 0 && (
            <div className="victron-section">
              <h2><Droplets size={20} /> Tanks</h2>
              <div className="victron-detail-cards">
                {data.tanks.map((t, i) => (
                  <div key={i} className="detail-card">
                    <div className="detail-title">Tank #{t.instance}</div>
                    <div className="detail-grid">
                      <div className="detail-item">
                        <Gauge size={16} />
                        <div><small>Level</small><strong>{t.level?.toFixed(0) ?? '--'}%</strong></div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      <div className="victron-footer">
        <span>Last update: {timeAgo || 'waiting...'}</span>
        <span className="connection-status-badge">
          <span className="dot" style={{ backgroundColor: statusColor }} />
          {connectionStatus}
        </span>
      </div>
    </div>
  );
};

function getVebusState(state) {
  const states = {
    0: 'Off',
    1: 'Low Power',
    2: 'Inverting',
    3: 'Charging',
    4: 'Equalizing',
    5: 'Float',
    6: 'Absorption',
    7: 'Bulk',
    9: 'Passthrough',
    10: 'Inverting & Charging',
    11: 'Power Assist',
  };
  return states[state] ?? `Unknown (${state})`;
}

export default VictronDashboard;
