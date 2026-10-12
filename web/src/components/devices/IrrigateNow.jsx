import { useEffect, useId, useRef, useState } from 'react'
import { deviceCommandsApi, getApiUrl } from '../../config/api'
import { useNotification } from '../../hooks/useNotification'
import { formatRelative } from '../../lib/zensor'
import { DEFAULT_DURATION, MAX_DURATION, MIN_DURATION, clampDuration, irrigateNowPayload } from '../../utils/irrigationSchedule'
import { isValveOpen } from '../../utils/sensorReadings'
import { Button } from '../../ui'
import { deviceName, isOnline } from './deviceStatus'

const QUEUE_TIMEOUT_MS = 2 * 60 * 1000
const STOP_COMMAND = { sequence: [{ index: 1, value: 0, priority: 'HIGH' }] }

function blockedReason(device, sensorData) {
  if (!isOnline(device)) {
    return `No se puede regar: ${deviceName(device)} está sin conexión (última señal ${formatRelative(device.lastSeen)}). Revisa su alimentación y conexión a internet.`
  }
  if (!sensorData?.data) {
    return 'Esperando la primera lectura del dispositivo para confirmar el estado de la válvula. Suele tardar menos de un minuto.'
  }
  return null
}

export default function IrrigateNow({ device, sensorData }) {
  const durationId = useId()
  const { showSuccess, showError } = useNotification()
  const [minutes, setMinutes] = useState(DEFAULT_DURATION)
  const [queued, setQueued] = useState(false)
  const [busy, setBusy] = useState(false)

  const valveOpen = isValveOpen(sensorData)
  const reason = blockedReason(device, sensorData)
  const name = deviceName(device)

  const wasOpen = useRef(valveOpen)

  useEffect(() => {
    if (valveOpen) setQueued(false)
    if (wasOpen.current && !valveOpen) showSuccess(`${name} cerró la válvula.`, 'Riego terminado')
    wasOpen.current = valveOpen
  }, [valveOpen, name, showSuccess])

  useEffect(() => {
    if (!queued) return undefined
    const timer = setTimeout(() => setQueued(false), QUEUE_TIMEOUT_MS)
    return () => clearTimeout(timer)
  }, [queued])

  const start = async () => {
    setBusy(true)
    try {
      const response = await fetch(getApiUrl(`/devices/${device.id}/tasks`), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(irrigateNowPayload(minutes)),
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      setQueued(true)
      showSuccess(`${name} empezará a regar en unos segundos durante ${clampDuration(minutes)}\u00A0min.`, 'Riego en cola')
    } catch {
      showError(`No se pudo iniciar el riego: ${name} no respondió. Revisa su conexión y vuelve a intentarlo.`, 'Riego no iniciado')
    } finally {
      setBusy(false)
    }
  }

  const stop = async () => {
    setBusy(true)
    try {
      await deviceCommandsApi.sendCommand(device.id, STOP_COMMAND)
      showSuccess(`${name} cerrará la válvula en unos segundos.`, 'Deteniendo riego')
    } catch {
      showError(`No se pudo detener el riego en ${name}. Vuelve a intentarlo.`, 'Riego no detenido')
    } finally {
      setBusy(false)
    }
  }

  return (
    <section aria-labelledby={`${durationId}-title`} className="flex flex-col gap-4 rounded-xl bg-inverse p-6 text-on-inverse">
      <h3 id={`${durationId}-title`} className="m-0 text-h3">
        Regar ahora
      </h3>

      {reason ? (
        <p className="m-0 text-body-sm text-on-inverse-muted">{reason}</p>
      ) : valveOpen ? (
        <>
          <p className="m-0 text-body-sm text-on-inverse-muted">La válvula está abierta. El riego se detiene solo al terminar el tiempo.</p>
          <Button variant="accent" size="lg" block onClick={stop} disabled={busy}>
            Detener riego
          </Button>
        </>
      ) : (
        <>
          <div className="flex items-baseline justify-between gap-3">
            <label htmlFor={durationId} className="text-body-sm text-on-inverse-muted">
              Duración
            </label>
            <output htmlFor={durationId} className="text-data-lg">
              {minutes}
              <small className="ml-1 font-body text-xs font-regular text-on-inverse-muted">min</small>
            </output>
          </div>
          <input
            id={durationId}
            type="range"
            min={MIN_DURATION}
            max={MAX_DURATION}
            value={minutes}
            onChange={event => setMinutes(Number(event.target.value))}
            disabled={queued || busy}
            className="zs-range zs-range-inverse"
          />
          <Button variant="accent" size="lg" block onClick={start} disabled={queued || busy}>
            {queued ? 'Riego en cola…' : 'Iniciar riego'}
          </Button>
        </>
      )}
    </section>
  )
}
