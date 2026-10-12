import { cx, formatNumber, gauge, gaugeDash, gaugeMarker, humidityState } from '../lib/zensor'
import { Meter } from './Controls'

export function DataValue({ value, unit, digits = 1, size = 'lg', className }) {
  const sizes = { lg: 'text-data-lg', md: 'text-data-md', sm: 'text-data-sm' }
  const display = typeof value === 'number' ? formatNumber(value, digits) : value
  return (
    <span className={cx(sizes[size], 'whitespace-nowrap', className)}>
      {display}
      {unit && <small className="ml-1 font-body text-xs font-regular text-muted">{unit}</small>}
    </span>
  )
}

export function SensorTile({ label, value, unit, digits = 1, size = 'lg', meter, className }) {
  return (
    <div className={cx('flex min-w-[190px] flex-1 flex-col gap-2 rounded-md border border-line bg-surface p-4', className)}>
      <span className="text-body-sm text-muted">{label}</span>
      <DataValue value={value} unit={unit} digits={digits} size={size} />
      {meter}
    </div>
  )
}

export function HumidityBar({ label, value, target }) {
  const state = humidityState(value)
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-baseline justify-between gap-3">
        <span className="text-body-sm text-muted">{label}</span>
        <span className="flex items-baseline gap-2">
          <span className="text-caption" style={{ color: state.fg }}>
            {state.label}
          </span>
          <DataValue value={value} unit="%" size="sm" />
        </span>
      </div>
      <Meter value={value} fill={state.fill} target={target} label={`${label}: ${formatNumber(value, 1)} %, ${state.label}`} />
    </div>
  )
}

export function HumidityGauge({ value, target, label = 'Humedad de suelo' }) {
  const state = humidityState(value)
  const dash = gaugeDash(value)
  const marker = target !== undefined ? gaugeMarker(target) : null
  const description = `${label} ${formatNumber(value, 0)} por ciento${target !== undefined ? `, objetivo ${target} por ciento` : ''}`

  return (
    <figure className="m-0 flex flex-col items-center gap-2">
      <svg viewBox="0 0 180 180" width="180" height="180" role="img" aria-label={description}>
        <g transform={`rotate(${gauge.rotate} 90 90)`}>
          <circle
            cx="90"
            cy="90"
            r={gauge.radius}
            fill="none"
            stroke="var(--zs-track)"
            strokeWidth={gauge.strokeWidth}
            strokeLinecap="round"
            strokeDasharray={dash.track}
          />
          <circle
            cx="90"
            cy="90"
            r={gauge.radius}
            fill="none"
            stroke={state.fill}
            strokeWidth={gauge.strokeWidth}
            strokeLinecap="round"
            strokeDasharray={dash.value}
            className="transition-all duration-slower ease-standard"
          />
        </g>
        {marker && <circle cx={marker.x} cy={marker.y} r="6" fill="var(--zs-brand-ink)" />}
        <text
          x="90"
          y="96"
          textAnchor="middle"
          fill="var(--zs-text-primary)"
          style={{ font: '500 var(--zs-text-2xl) var(--zs-font-mono)' }}
        >
          {formatNumber(value, 0)}
          <tspan fill="var(--zs-text-secondary)" style={{ fontSize: 'var(--zs-text-md)' }}>
            {' %'}
          </tspan>
        </text>
        <text x="90" y="120" textAnchor="middle" fill={state.fg} style={{ font: '500 var(--zs-text-sm) var(--zs-font-body)' }}>
          {state.label}
        </text>
      </svg>
      <figcaption className="text-caption text-muted">{label}</figcaption>
    </figure>
  )
}
