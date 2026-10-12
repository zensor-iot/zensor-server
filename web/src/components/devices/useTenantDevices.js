import { useCallback, useEffect, useState } from 'react'
import useWebSocket from '../../hooks/useWebSocket'
import { getApiUrl, getWebSocketUrl, scheduledTasksApi } from '../../config/api'

const REFRESH_MS = 30000

const toDevice = device => ({
  ...device,
  lastSeen: device.last_message_received_at ? new Date(device.last_message_received_at) : null,
})

const sameDevices = (a, b) =>
  a.length === b.length &&
  a.every(
    (device, i) =>
      device.id === b[i].id &&
      device.status === b[i].status &&
      device.display_name === b[i].display_name &&
      device.lastSeen?.getTime() === b[i].lastSeen?.getTime(),
  )

export default function useTenantDevices(tenantId) {
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [sensorData, setSensorData] = useState({})
  const [programs, setPrograms] = useState({})

  const { isConnected, lastMessage } = useWebSocket(getWebSocketUrl('/ws/device-messages'))

  useEffect(() => {
    if (lastMessage?.type !== 'device_state') return
    setSensorData(prev => ({
      ...prev,
      [lastMessage.device_id]: { data: lastMessage.data, timestamp: lastMessage.timestamp, receivedAt: new Date() },
    }))
  }, [lastMessage])

  const fetchDevices = useCallback(async () => {
    const response = await fetch(getApiUrl(`/tenants/${tenantId}/devices`))
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const body = await response.json()
    return (body.data || []).map(toDevice)
  }, [tenantId])

  const loadPrograms = useCallback(
    async deviceId => {
      try {
        const response = await scheduledTasksApi.getScheduledTasks(tenantId, deviceId)
        const tasks = (response.data || []).filter(task => task && (task.schedule || task.scheduling))
        setPrograms(prev => ({ ...prev, [deviceId]: { tasks, error: null } }))
      } catch (err) {
        setPrograms(prev => ({ ...prev, [deviceId]: { tasks: prev[deviceId]?.tasks || [], error: err.message } }))
      }
    },
    [tenantId],
  )

  const reload = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const list = await fetchDevices()
      setDevices(list)
      list.forEach(device => loadPrograms(device.id))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [fetchDevices, loadPrograms])

  useEffect(() => {
    setDevices([])
    setPrograms({})
    reload()
  }, [reload])

  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const list = await fetchDevices()
        setDevices(prev => (sameDevices(prev, list) ? prev : list))
      } catch {
        // la próxima actualización vuelve a intentarlo; la lista visible sigue siendo válida
      }
    }, REFRESH_MS)
    return () => clearInterval(interval)
  }, [fetchDevices])

  const renameDevice = useCallback(async (deviceId, displayName) => {
    const response = await fetch(getApiUrl(`/devices/${deviceId}`), {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ display_name: displayName }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    setDevices(prev => prev.map(device => (device.id === deviceId ? { ...device, display_name: displayName } : device)))
  }, [])

  const setDevicePrograms = useCallback((deviceId, update) => {
    setPrograms(prev => ({ ...prev, [deviceId]: { tasks: update(prev[deviceId]?.tasks || []), error: null } }))
  }, [])

  return {
    devices,
    loading,
    error,
    reload,
    sensorData,
    programs,
    loadPrograms,
    setDevicePrograms,
    renameDevice,
    liveConnected: isConnected,
  }
}
