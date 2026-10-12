import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { getApiUrl } from '../config/api'

const LAST_TENANT_KEY = 'zensor-last-tenant'

const TenantContext = createContext(null)

const readLastTenant = () => {
  try {
    return localStorage.getItem(LAST_TENANT_KEY)
  } catch {
    return null
  }
}

const rememberTenant = tenantId => {
  try {
    localStorage.setItem(LAST_TENANT_KEY, tenantId)
  } catch {
    // sin almacenamiento: solo se pierde la preselección
  }
}

export const tenantIdFromPath = pathname => pathname.match(/^\/(?:portal|tenants)\/([^/]+)/)?.[1] || null

export function TenantProvider({ children }) {
  const location = useLocation()
  const [tenants, setTenants] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [lastTenantId, setLastTenantId] = useState(readLastTenant)

  const pathTenantId = tenantIdFromPath(location.pathname)

  const reload = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch(getApiUrl('/tenants'))
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const body = await response.json()
      setTenants(body.data || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    reload()
  }, [reload])

  useEffect(() => {
    if (pathTenantId && pathTenantId !== lastTenantId) {
      rememberTenant(pathTenantId)
      setLastTenantId(pathTenantId)
    }
  }, [pathTenantId, lastTenantId])

  const currentTenantId = pathTenantId || (tenants.some(t => t.id === lastTenantId) ? lastTenantId : null)
  const currentTenant = tenants.find(t => t.id === currentTenantId) || null

  const value = useMemo(
    () => ({ tenants, loading, error, reload, currentTenantId, currentTenant }),
    [tenants, loading, error, reload, currentTenantId, currentTenant],
  )

  return <TenantContext.Provider value={value}>{children}</TenantContext.Provider>
}

export const useTenants = () => {
  const context = useContext(TenantContext)
  if (!context) throw new Error('useTenants must be used within a TenantProvider')
  return context
}
