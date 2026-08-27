import { useState, useEffect } from 'react';
import useWebSocket from '../hooks/useWebSocket';
import useMetricHistory from '../hooks/useMetricHistory';
import useCumulativeEnergyToday from '../hooks/useCumulativeEnergyToday';
import { getWebSocketUrl } from '../config/api';
import {
  Sun, Battery, Zap, Gauge, Thermometer, Activity,
  Wifi, WifiOff, RefreshCw, Power, Droplets, TrendingUp
} from 'lucide-react';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, ReferenceLine
} from 'recharts';

const SOC_METRIC_NAME = 'zensor_server_victron_battery_soc';
const LOAD_METRIC_NAME = 'zensor_server_victron_system_load_power';
const SOLAR_POWER_METRIC_NAME = 'zensor_server_victron_solar_power';

const TIME_RANGES = [
  { value: '1h', label: '1h', ms: 60 * 60 * 1000, step: '30s' },
  { value: '6h', label: '6h', ms: 6 * 60 * 60 * 1000, step: '60s' },
  { value: '24h', label: '24h', ms: 24 * 60 * 60 * 1000, step: '300s' },
  { value: '7d', label: '7d', ms: 7 * 24 * 60 * 60 * 1000, step: '3600s' },
];

const VictronDashboard = () => {
  const [systemStatus, setSystemStatus] = useState(null);
  const [lastUpdate, setLastUpdate] = useState(null);
  const [timeAgo, setTimeAgo] = useState('');
  const [socRange, setSocRange] = useState('6h');
  const [loadRange, setLoadRange] = useState('6h');

  const wsUrl = getWebSocketUrl('/ws/victron/status');

  const { isConnected, lastMessage, connectionError, connectionStatus } = useWebSocket(wsUrl);

  const socRangeConfig = TIME_RANGES.find((r) => r.value === socRange) ?? TIME_RANGES[1];
  const { history: socHistory, loading: socLoading, error: socError } = useMetricHistory(SOC_METRIC_NAME, socRangeConfig);

  const loadRangeConfig = TIME_RANGES.find((r) => r.value === loadRange) ?? TIME_RANGES[1];
  const { history: loadHistory, loading: loadHistoryLoading, error: loadHistoryError } = useMetricHistory(LOAD_METRIC_NAME, loadRangeConfig);

  const {
    history: solarEnergyHistory,
    totalKWh: solarEnergyTotalKWh,
    loading: solarEnergyLoading,
    error: solarEnergyError,
  } = useCumulativeEnergyToday(SOLAR_POWER_METRIC_NAME);

  const {
    history: loadEnergyHistory,
    totalKWh: loadEnergyTotalKWh,
    loading: loadEnergyLoading,
    error: loadEnergyError,
  } = useCumulativeEnergyToday(LOAD_METRIC_NAME);

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

      {connectionError && (
        <div className="victron-error-banner">
          <WifiOff size={20} />
          <span>Connection error: {connectionError}</span>
        </div>
      )}

      <div className="victron-layout">
        <div className="victron-main">
          {!systemStatus ? (
            <div className="victron-loading">
              <Activity size={48} className="loading-spinner" />
              <p>Connecting to energy system...</p>
            </div>
          ) : (
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
        </div>

        <aside className="victron-sidebar">
          <div className="victron-section">
            <div className="victron-section-header">
              <h2><TrendingUp size={20} /> Battery SOC History</h2>
              <div className="time-range-selector">
                {TIME_RANGES.map((range) => (
                  <button
                    key={range.value}
                    className={`range-btn ${socRange === range.value ? 'active' : ''}`}
                    onClick={() => setSocRange(range.value)}
                  >
                    {range.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="victron-chart-card">
              {socLoading && socHistory.length === 0 && (
                <div className="chart-placeholder">Loading SOC history...</div>
              )}
              {!socLoading && socError && (
                <div className="chart-error">Metrics unavailable: {socError}</div>
              )}
              {!socLoading && !socError && socHistory.length === 0 && (
                <div className="chart-placeholder">No SOC data in this time range.</div>
              )}
              {socHistory.length > 0 && (
                <ResponsiveContainer width="100%" height={320}>
                  <AreaChart data={socHistory}>
                    <defs>
                      <linearGradient id="socGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.35} />
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0.05} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="time"
                      type="number"
                      scale="time"
                      domain={['dataMin', 'dataMax']}
                      tickFormatter={formatChartTime}
                      tick={{ fontSize: 12 }}
                      stroke="#9ca3af"
                    />
                    <YAxis domain={[0, 100]} unit="%" tick={{ fontSize: 12 }} stroke="#9ca3af" />
                    <Tooltip
                      labelFormatter={(ts) => new Date(ts).toLocaleString()}
                      formatter={(value) => [`${Number(value).toFixed(1)}%`, 'SOC']}
                    />
                    <ReferenceLine y={20} stroke="#ef4444" strokeDasharray="4 4" />
                    <Area
                      type="monotone"
                      dataKey="value"
                      stroke="#10b981"
                      strokeWidth={2}
                      fill="url(#socGradient)"
                      dot={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>

          <div className="victron-section">
            <div className="victron-section-header">
              <h2><TrendingUp size={20} /> System Load History</h2>
              <div className="time-range-selector">
                {TIME_RANGES.map((range) => (
                  <button
                    key={range.value}
                    className={`range-btn ${loadRange === range.value ? 'active' : ''}`}
                    onClick={() => setLoadRange(range.value)}
                  >
                    {range.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="victron-chart-card">
              {loadHistoryLoading && loadHistory.length === 0 && (
                <div className="chart-placeholder">Loading load history...</div>
              )}
              {!loadHistoryLoading && loadHistoryError && (
                <div className="chart-error">Metrics unavailable: {loadHistoryError}</div>
              )}
              {!loadHistoryLoading && !loadHistoryError && loadHistory.length === 0 && (
                <div className="chart-placeholder">No load data in this time range.</div>
              )}
              {loadHistory.length > 0 && (
                <ResponsiveContainer width="100%" height={320}>
                  <AreaChart data={loadHistory}>
                    <defs>
                      <linearGradient id="loadGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.35} />
                        <stop offset="95%" stopColor="#f59e0b" stopOpacity={0.05} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="time"
                      type="number"
                      scale="time"
                      domain={['dataMin', 'dataMax']}
                      tickFormatter={formatChartTime}
                      tick={{ fontSize: 12 }}
                      stroke="#9ca3af"
                    />
                    <YAxis unit="W" tick={{ fontSize: 12 }} stroke="#9ca3af" />
                    <Tooltip
                      labelFormatter={(ts) => new Date(ts).toLocaleString()}
                      formatter={(value) => [`${Number(value).toFixed(0)} W`, 'Load']}
                    />
                    <Area
                      type="monotone"
                      dataKey="value"
                      stroke="#f59e0b"
                      strokeWidth={2}
                      fill="url(#loadGradient)"
                      dot={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>

          <div className="victron-section">
            <div className="victron-section-header">
              <h2><Sun size={20} /> Solar Energy Today{solarEnergyHistory.length > 0 ? ` — ${solarEnergyTotalKWh.toFixed(2)} kWh` : ''}</h2>
            </div>
            <div className="victron-chart-card">
              {solarEnergyLoading && solarEnergyHistory.length === 0 && (
                <div className="chart-placeholder">Loading today's solar energy...</div>
              )}
              {!solarEnergyLoading && solarEnergyError && (
                <div className="chart-error">Metrics unavailable: {solarEnergyError}</div>
              )}
              {!solarEnergyLoading && !solarEnergyError && solarEnergyHistory.length === 0 && (
                <div className="chart-placeholder">No solar data yet today.</div>
              )}
              {solarEnergyHistory.length > 0 && (
                <ResponsiveContainer width="100%" height={320}>
                  <AreaChart data={solarEnergyHistory}>
                    <defs>
                      <linearGradient id="solarEnergyGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#eab308" stopOpacity={0.35} />
                        <stop offset="95%" stopColor="#eab308" stopOpacity={0.05} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="time"
                      type="number"
                      scale="time"
                      domain={['dataMin', 'dataMax']}
                      tickFormatter={formatChartTime}
                      tick={{ fontSize: 12 }}
                      stroke="#9ca3af"
                    />
                    <YAxis unit="kWh" tick={{ fontSize: 12 }} stroke="#9ca3af" />
                    <Tooltip
                      labelFormatter={(ts) => new Date(ts).toLocaleString()}
                      formatter={(value) => [`${Number(value).toFixed(2)} kWh`, 'Solar energy']}
                    />
                    <Area
                      type="monotone"
                      dataKey="value"
                      stroke="#eab308"
                      strokeWidth={2}
                      fill="url(#solarEnergyGradient)"
                      dot={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>

          <div className="victron-section">
            <div className="victron-section-header">
              <h2><Zap size={20} /> Load Energy Today{loadEnergyHistory.length > 0 ? ` — ${loadEnergyTotalKWh.toFixed(2)} kWh` : ''}</h2>
            </div>
            <div className="victron-chart-card">
              {loadEnergyLoading && loadEnergyHistory.length === 0 && (
                <div className="chart-placeholder">Loading today's load energy...</div>
              )}
              {!loadEnergyLoading && loadEnergyError && (
                <div className="chart-error">Metrics unavailable: {loadEnergyError}</div>
              )}
              {!loadEnergyLoading && !loadEnergyError && loadEnergyHistory.length === 0 && (
                <div className="chart-placeholder">No load data yet today.</div>
              )}
              {loadEnergyHistory.length > 0 && (
                <ResponsiveContainer width="100%" height={320}>
                  <AreaChart data={loadEnergyHistory}>
                    <defs>
                      <linearGradient id="loadEnergyGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.35} />
                        <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0.05} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="time"
                      type="number"
                      scale="time"
                      domain={['dataMin', 'dataMax']}
                      tickFormatter={formatChartTime}
                      tick={{ fontSize: 12 }}
                      stroke="#9ca3af"
                    />
                    <YAxis unit="kWh" tick={{ fontSize: 12 }} stroke="#9ca3af" />
                    <Tooltip
                      labelFormatter={(ts) => new Date(ts).toLocaleString()}
                      formatter={(value) => [`${Number(value).toFixed(2)} kWh`, 'Load energy']}
                    />
                    <Area
                      type="monotone"
                      dataKey="value"
                      stroke="#8b5cf6"
                      strokeWidth={2}
                      fill="url(#loadEnergyGradient)"
                      dot={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>
        </aside>
      </div>

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

function formatChartTime(timestamp) {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

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
