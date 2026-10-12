import { useCallback, useEffect, useRef, useState } from 'react'
import { maintenanceApi } from '../../config/api'
import { nextPendingExecution, sortByScheduledDesc } from '../../utils/maintenanceSchedule'

const cache = new Map()

async function loadOverview(tenantId) {
  const activities = await maintenanceApi.listAllActivities(tenantId)
  return Promise.all(
    activities.map(activity =>
      maintenanceApi
        .listExecutions(activity.id, 1, 100)
        .then(data => data.data || [])
        .catch(() => [])
        .then(executions => ({
          activity,
          executions: sortByScheduledDesc(executions),
          next: nextPendingExecution(executions),
        })),
    ),
  )
}

/**
 * Loads every maintenance activity of a tenant with its executions (newest first) and
 * next pending execution. Keeps the last result per tenant so remounts render instantly.
 */
export function useMaintenanceOverview(tenantId) {
  const [entry, setEntry] = useState({ tenantId, data: cache.get(tenantId) ?? null })
  const [loading, setLoading] = useState(!cache.has(tenantId))
  const [error, setError] = useState(null)
  const currentTenant = useRef(tenantId)
  currentTenant.current = tenantId

  const reload = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await loadOverview(tenantId)
      cache.set(tenantId, data)
      if (currentTenant.current === tenantId) setEntry({ tenantId, data })
    } catch (err) {
      if (currentTenant.current === tenantId) setError(err)
    } finally {
      if (currentTenant.current === tenantId) setLoading(false)
    }
  }, [tenantId])

  useEffect(() => {
    reload()
  }, [reload])

  const items = entry.tenantId === tenantId ? entry.data : cache.get(tenantId) ?? null
  return { items, loading, error, reload }
}
