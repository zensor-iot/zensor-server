import { useEffect, useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { Building2, Cpu, Wrench, Pill } from 'lucide-react'
import { getApiUrl } from '../config/api'

const PortalShell = ({ children }) => {
  const location = useLocation()
  const tenantId = location.pathname.match(/^\/portal\/([^/]+)/)?.[1]
  const [tenant, setTenant] = useState(null)

  useEffect(() => {
    if (!tenantId) {
      setTenant(null)
      return
    }

    let cancelled = false
    fetch(getApiUrl(`/tenants/${tenantId}`))
      .then(response => (response.ok ? response.json() : null))
      .then(data => {
        if (!cancelled) setTenant(data)
      })
      .catch(() => {
        if (!cancelled) setTenant(null)
      })

    return () => {
      cancelled = true
    }
  }, [tenantId])

  if (!tenantId) {
    return children
  }

  return (
    <div className="portal-shell">
      <aside className="portal-sidebar">
        <div className="portal-sidebar-tenant" title={tenant?.name}>
          <Building2 size={18} />
          <span>{tenant?.name || 'Loading...'}</span>
        </div>

        <nav className="portal-sidebar-nav">
          <NavLink
            to={`/portal/${tenantId}`}
            end
            className={({ isActive }) => `portal-sidebar-link${isActive ? ' active' : ''}`}
          >
            <Cpu size={20} />
            Devices
          </NavLink>
          <NavLink
            to={`/portal/${tenantId}/maintenance`}
            className={({ isActive }) => `portal-sidebar-link${isActive ? ' active' : ''}`}
          >
            <Wrench size={20} />
            Maintenance
          </NavLink>
          <NavLink
            to={`/portal/${tenantId}/medicines`}
            className={({ isActive }) => `portal-sidebar-link${isActive ? ' active' : ''}`}
          >
            <Pill size={20} />
            Medicines
          </NavLink>
        </nav>
      </aside>

      <div className="portal-shell-content">
        {children}
      </div>
    </div>
  )
}

export default PortalShell
