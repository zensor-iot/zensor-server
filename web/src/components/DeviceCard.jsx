import { useState } from 'react'
import { ChevronRight, Copy } from 'lucide-react'
import config from '../config/api'
import { useNotification } from '../hooks/useNotification'
import { cx, formatRelative } from '../lib/zensor'
import { Card, Drawer, FilterChips, IconButton, Section, StatusPill } from '../ui'

const deviceName = device => device.display_name || device.name

const isOnline = device => device.status === 'online'

const statusLine = device => `${isOnline(device) ? 'En línea' : 'Sin conexión'} · ${formatRelative(device.last_message_received_at)}`

function DeviceStatusPill({ device }) {
  return isOnline(device) ? (
    <StatusPill tone="ok">En línea</StatusPill>
  ) : (
    <StatusPill tone="attention">Sin conexión</StatusPill>
  )
}

const TIME_RANGES = [
  { value: '1h', label: 'Última hora', ms: 60 * 60 * 1000 },
  { value: '6h', label: 'Últimas 6 h', ms: 6 * 60 * 60 * 1000 },
  { value: '24h', label: 'Últimas 24 h', ms: 24 * 60 * 60 * 1000 },
]

const grafanaPanelUrl = (name, rangeMs) => {
  const to = Date.now()
  const from = to - rangeMs
  const url = `${config.grafanaBaseUrl}/grafana/d-solo/ces2boi7cdh4wc/sensor-data?panelId=1&var-device_name=${encodeURIComponent(name)}&from=${from}&to=${to}`
  return config.grafanaApiKey ? `${url}&authToken=${config.grafanaApiKey}` : url
}

function SensorHistory({ device }) {
  const [range, setRange] = useState('6h')
  const selected = TIME_RANGES.find(option => option.value === range)
  const name = deviceName(device)

  return (
    <Section title="Lecturas" description="Historial de sensores desde Grafana.">
      <FilterChips label="Período" options={TIME_RANGES} value={range} onChange={setRange} />
      <iframe
        key={range}
        src={grafanaPanelUrl(name, selected.ms)}
        title={`Lecturas de ${name}, ${selected.label.toLowerCase()}`}
        loading="lazy"
        className="h-[420px] w-full rounded-md border border-line bg-surface"
      />
    </Section>
  )
}

function CopyableValue({ label, value, display }) {
  const { showSuccess, showError } = useNotification()

  if (!value) return <span className="text-muted">Sin dato</span>

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      showSuccess(`${label} copiado`)
    } catch {
      showError(`No se pudo copiar ${label}. Selecciónalo y cópialo a mano.`)
    }
  }

  return (
    <span className="-my-3 flex items-center gap-1">
      <code className="min-w-0 break-all font-mono text-md">{display ?? value}</code>
      <IconButton icon={Copy} label={`Copiar ${label}`} onClick={copy} className="shrink-0" />
    </span>
  )
}

/**
 * Right-side detail for a device of the read-only tenant device list: identifiers and sensor history.
 */
export function DeviceDetailDrawer({ device: current, onClose }) {
  const [device, setDevice] = useState(current)
  if (current && current !== device) setDevice(current)

  return (
    <Drawer
      open={Boolean(current)}
      onClose={onClose}
      title={device ? deviceName(device) : ''}
      subtitle={
        device && (
          <>
            <DeviceStatusPill device={device} />
            <span>{statusLine(device)}</span>
          </>
        )
      }
    >
      {device && (
        <>
          <SensorHistory device={device} />
          <Section title="Dispositivo">
            <dl className="m-0 grid grid-cols-[auto_1fr] items-center gap-x-6 gap-y-4">
              <dt className="text-body-sm text-muted">Nombre técnico</dt>
              <dd className="m-0 min-w-0 break-words">{device.name}</dd>
              <dt className="text-body-sm text-muted">ID</dt>
              <dd className="m-0 min-w-0">
                <CopyableValue label="ID" value={device.id} />
              </dd>
              <dt className="text-body-sm text-muted">App EUI</dt>
              <dd className="m-0 min-w-0">
                <CopyableValue label="App EUI" value={device.app_eui} />
              </dd>
              <dt className="text-body-sm text-muted">Dev EUI</dt>
              <dd className="m-0 min-w-0">
                <CopyableValue label="Dev EUI" value={device.dev_eui} />
              </dd>
              <dt className="text-body-sm text-muted">App Key</dt>
              <dd className="m-0 min-w-0">
                <CopyableValue label="App Key" value={device.app_key} display={device.app_key && `${device.app_key.slice(0, 8)}…`} />
              </dd>
            </dl>
          </Section>
        </>
      )}
    </Drawer>
  )
}

const DeviceCard = ({ device, onOpen }) => {
  const online = isOnline(device)
  return (
    <Card
      as="button"
      type="button"
      interactive
      onClick={() => onOpen(device)}
      aria-label={`${deviceName(device)}, ${statusLine(device)}. Ver detalle`}
      row
      className={cx(!online && 'opacity-offline')}
    >
      <span className="flex min-w-0 flex-1 flex-col gap-1">
        <span className="break-words text-title">{deviceName(device)}</span>
        <span className="text-body-sm text-muted">{statusLine(device)}</span>
      </span>
      <DeviceStatusPill device={device} />
      <ChevronRight size={20} strokeWidth={1.8} aria-hidden="true" className="shrink-0 text-muted" />
    </Card>
  )
}

export default DeviceCard
