import { useState } from 'react'
import { ChevronLeft } from 'lucide-react'
import config from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { formatDateTime, formatRelative, humidityState } from '../../lib/zensor'
import { sensorReadings } from '../../utils/sensorReadings'
import {
  Button,
  Drawer,
  EmptyState,
  FilterChips,
  Form,
  FormActions,
  IconButton,
  Input,
  KeyValue,
  Meter,
  Section,
  SensorTile,
  StatusPill,
} from '../../ui'
import IrrigateNow from './IrrigateNow'
import { ProgramForm, ProgramList } from './Programs'
import { deviceName, deviceStatus, isOnline } from './deviceStatus'

const HISTORY_RANGES = [
  { value: '1h', label: '1 h', ms: 60 * 60 * 1000 },
  { value: '6h', label: '6 h', ms: 6 * 60 * 60 * 1000 },
  { value: '24h', label: '24 h', ms: 24 * 60 * 60 * 1000 },
]

function Sensors({ device, sensorData }) {
  const readings = sensorReadings(sensorData)

  if (!isOnline(device) || readings.length === 0) {
    return (
      <Section title="Sensores">
        <EmptyState>
          {isOnline(device)
            ? 'Aún no llegan lecturas. Aparecerán aquí apenas el dispositivo envíe datos.'
            : 'Sin datos recientes. Las lecturas vuelven cuando el dispositivo se reconecte.'}
        </EmptyState>
      </Section>
    )
  }

  return (
    <Section
      title="Sensores"
      description={sensorData?.timestamp ? `Última lectura ${formatDateTime(sensorData.timestamp)}` : undefined}
    >
      <div className="flex flex-wrap gap-3">
        {readings.map(reading => {
          if (reading.kind === 'relay') {
            return (
              <SensorTile
                key={reading.key}
                label={reading.label}
                value={Number(reading.value) === 1 ? 'Abierta' : 'Cerrada'}
                size="md"
              />
            )
          }
          const value = Number(reading.value)
          const meter =
            reading.kind === 'humidity' ? (
              <Meter value={value} fill={humidityState(value).fill} label={`${reading.label}: ${humidityState(value).label}`} />
            ) : null
          return (
            <SensorTile
              key={reading.key}
              label={reading.label}
              value={Number.isFinite(value) ? value : reading.value}
              unit={reading.unit}
              meter={meter}
            />
          )
        })}
      </div>
    </Section>
  )
}

function History({ device }) {
  const [range, setRange] = useState('6h')
  const span = HISTORY_RANGES.find(option => option.value === range).ms
  const to = Date.now()
  const params = new URLSearchParams({
    panelId: '1',
    'var-device_name': deviceName(device),
    from: String(to - span),
    to: String(to),
    theme: document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light',
  })
  if (config.grafanaApiKey) params.set('authToken', config.grafanaApiKey)
  const src = `${config.grafanaBaseUrl}/grafana/d-solo/ces2boi7cdh4wc/sensor-data?${params}`

  return (
    <Section title="Historial">
      <FilterChips label="Rango del historial" options={HISTORY_RANGES} value={range} onChange={setRange} />
      <iframe
        src={src}
        title={`Historial de sensores de ${deviceName(device)}, últimas ${range.replace('h', ' horas')}`}
        loading="lazy"
        className="h-[320px] w-full rounded-md border border-line bg-surface"
      />
    </Section>
  )
}

function DeviceInfo({ device, onRename }) {
  const { showSuccess, showError } = useNotification()
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(deviceName(device))
  const [saving, setSaving] = useState(false)

  const save = async event => {
    event.preventDefault()
    const trimmed = name.trim()
    if (!trimmed) return
    setSaving(true)
    try {
      await onRename(device.id, trimmed)
      showSuccess('Nombre guardado')
      setEditing(false)
    } catch {
      showError('No se pudo guardar el nombre. Vuelve a intentarlo.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Section
      title="Dispositivo"
      actions={
        !editing && (
          <Button variant="ghost" size="sm" onClick={() => setEditing(true)}>
            Cambiar nombre
          </Button>
        )
      }
    >
      {editing && (
        <Form onSubmit={save} className="gap-3">
          <label htmlFor={`rename-${device.id}`} className="text-label">
            Nombre visible
          </label>
          <Input id={`rename-${device.id}`} value={name} onChange={event => setName(event.target.value)} autoFocus />
          <FormActions className="pt-0">
            <Button type="submit" size="sm" disabled={saving || !name.trim()}>
              {saving ? 'Guardando…' : 'Guardar nombre'}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setName(deviceName(device))
                setEditing(false)
              }}
            >
              Cancelar
            </Button>
          </FormActions>
        </Form>
      )}
      <div className="rounded-md border border-line bg-surface px-4 py-3">
        <KeyValue
          items={[
            { label: 'Nombre', value: deviceName(device) },
            { label: 'Identificador', value: device.name, mono: true },
            {
              label: 'Última señal',
              value: device.lastSeen ? `${formatRelative(device.lastSeen)} · ${formatDateTime(device.lastSeen)}` : 'Nunca',
            },
          ]}
        />
      </div>
    </Section>
  )
}

export default function DeviceDrawer({
  open,
  tenantId,
  device,
  sensorData,
  programs,
  onClose,
  onRename,
  onProgramsChange,
  onReloadPrograms,
}) {
  const [editor, setEditor] = useState(null)

  if (!device) return null

  const status = deviceStatus(device, sensorData)
  const closeEditor = () => setEditor(null)

  return (
    <Drawer
      open={open}
      onClose={editor ? closeEditor : onClose}
      title={editor ? (editor.task ? 'Editar programa' : 'Nuevo programa') : deviceName(device)}
      subtitle={
        editor ? (
          deviceName(device)
        ) : (
          <>
            <StatusPill tone={status.tone}>{status.label}</StatusPill>
            <span>{status.detail}</span>
          </>
        )
      }
      actions={editor && <IconButton icon={ChevronLeft} label="Volver al detalle" onClick={closeEditor} />}
    >
      {editor ? (
        <ProgramForm
          tenantId={tenantId}
          device={device}
          task={editor.task}
          onCancel={closeEditor}
          onSaved={update => {
            onProgramsChange(device.id, update)
            closeEditor()
          }}
        />
      ) : (
        <>
          <IrrigateNow device={device} sensorData={sensorData} />
          <Sensors device={device} sensorData={sensorData} />
          <ProgramList
            tenantId={tenantId}
            device={device}
            entry={programs}
            onChange={update => onProgramsChange(device.id, update)}
            onAdd={() => setEditor({ task: null })}
            onEdit={task => setEditor({ task })}
            onRetry={() => onReloadPrograms(device.id)}
          />
          <History device={device} />
          <DeviceInfo key={device.id} device={device} onRename={onRename} />
        </>
      )}
    </Drawer>
  )
}
