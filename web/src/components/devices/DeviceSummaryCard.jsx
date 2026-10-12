import { ChevronRight } from 'lucide-react'
import { cx } from '../../lib/zensor'
import { sensorReadings } from '../../utils/sensorReadings'
import { Card, DataValue, HumidityBar, StatusPill } from '../../ui'
import { deviceName, deviceStatus, isOnline, programSummary } from './deviceStatus'

function Readings({ readings }) {
  const humidity = readings.filter(reading => reading.kind === 'humidity')
  const others = readings.filter(reading => ['temperature', 'flow', 'pressure'].includes(reading.kind))

  return (
    <span className="flex flex-wrap items-center gap-6">
      {humidity.length > 0 && (
        <span className="flex min-w-[200px] flex-[1_1_200px] flex-col gap-3">
          {humidity.map(reading => (
            <HumidityBar key={reading.key} label={reading.label} value={Number(reading.value)} />
          ))}
        </span>
      )}
      {others.length > 0 && (
        <span className="flex flex-wrap gap-5">
          {others.map(reading => (
            <span key={reading.key} className="flex flex-col">
              <DataValue value={Number(reading.value)} unit={reading.unit} size="md" />
              <span className="text-caption text-muted">{reading.label}</span>
            </span>
          ))}
        </span>
      )}
    </span>
  )
}

export default function DeviceSummaryCard({ device, sensorData, programs, onOpen }) {
  const status = deviceStatus(device, sensorData)
  const online = isOnline(device)
  const readings = sensorReadings(sensorData)
  const summary = programSummary(programs)

  return (
    <Card
      as="button"
      type="button"
      interactive
      onClick={onOpen}
      className={cx('gap-5', !online && 'opacity-offline')}
    >
      <span className="flex items-start justify-between gap-3">
        <span className="flex min-w-0 flex-col gap-1">
          <span className="truncate text-title">{deviceName(device)}</span>
          <span className="text-body-sm text-muted">{status.detail}</span>
        </span>
        <StatusPill tone={status.tone}>{status.label}</StatusPill>
      </span>

      {online && readings.length > 0 ? (
        <Readings readings={readings} />
      ) : (
        <span className="flex min-h-control items-center text-body-sm text-muted">
          {online ? 'Esperando la primera lectura' : 'Sin datos recientes'}
        </span>
      )}

      <span className="flex items-center justify-between gap-3 border-t border-line pt-4 text-body-sm">
        <span className="text-ok-fg">{summary ?? 'Cargando programas…'}</span>
        <span className="inline-flex items-center gap-1 font-bold">
          Detalle
          <ChevronRight size={18} strokeWidth={1.8} aria-hidden="true" />
        </span>
      </span>
    </Card>
  )
}
