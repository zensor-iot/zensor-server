import { useEffect, useState } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import { Cpu, LogOut, Menu, Moon, Pill, Radio, Shield, Sun, SunMedium, UserCircle, Wrench, X } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { useTenants } from '../contexts/TenantContext'
import { cx, theme } from '../lib/zensor'
import { IconButton, Lockup, Mark, Select } from '../ui'

const navLinkClass = ({ isActive }) =>
  cx(
    'flex min-h-touch items-center gap-3 rounded-pill px-4 text-label no-underline transition-colors duration-fast ease-standard',
    isActive ? 'bg-primary text-on-primary' : 'text-ink hover:bg-sunken',
  )

function NavItem({ to, icon: Icon, end, children }) {
  return (
    <li>
      <NavLink to={to} end={end} className={navLinkClass}>
        <Icon size={20} strokeWidth={1.8} aria-hidden="true" />
        {children}
      </NavLink>
    </li>
  )
}

function NavGroup({ label, children }) {
  return (
    <div className="flex flex-col gap-1">
      {label && <p className="m-0 px-4 pb-1 text-caption text-muted">{label}</p>}
      <ul className="m-0 flex list-none flex-col gap-1 p-0">{children}</ul>
    </div>
  )
}

function TenantSelector() {
  const { tenants, loading, currentTenantId } = useTenants()
  const navigate = useNavigate()

  if (!loading && tenants.length === 0) return null

  return (
    <div className="flex flex-col gap-2">
      <label htmlFor="tenant-selector" className="px-1 text-caption text-muted">
        Cuenta
      </label>
      <Select
        id="tenant-selector"
        value={currentTenantId || ''}
        disabled={loading}
        onChange={event => event.target.value && navigate(`/portal/${event.target.value}`)}
      >
        <option value="" disabled>
          {loading ? 'Cargando…' : 'Elige una cuenta'}
        </option>
        {tenants.map(tenant => (
          <option key={tenant.id} value={tenant.id}>
            {tenant.name}
          </option>
        ))}
      </Select>
    </div>
  )
}

function useThemeMode() {
  const [mode, setMode] = useState(theme.effective)
  useEffect(() => {
    const sync = () => setMode(theme.effective())
    const media = window.matchMedia?.('(prefers-color-scheme: dark)')
    document.addEventListener('zensor:theme', sync)
    media?.addEventListener('change', sync)
    return () => {
      document.removeEventListener('zensor:theme', sync)
      media?.removeEventListener('change', sync)
    }
  }, [])
  const toggle = () => theme.set(mode === 'dark' ? 'light' : 'dark')
  return { mode, toggle }
}

function SidebarFooter() {
  const { user, logout } = useAuth()
  const { mode, toggle } = useThemeMode()

  return (
    <div className="flex flex-col gap-3 border-t border-line pt-4">
      {user && (
        <div className="flex min-w-0 flex-col px-4">
          <span className="truncate text-label">{user.name || user.email}</span>
          {user.name && user.email && <span className="truncate text-caption text-muted">{user.email}</span>}
        </div>
      )}
      <ul className="m-0 flex list-none flex-col gap-1 p-0">
        <NavItem to="/profile" icon={UserCircle}>
          Perfil
        </NavItem>
        <li>
          <button type="button" onClick={toggle} className={cx(navLinkClass({ isActive: false }), 'w-full border-0 bg-transparent')}>
            {mode === 'dark' ? (
              <SunMedium size={20} strokeWidth={1.8} aria-hidden="true" />
            ) : (
              <Moon size={20} strokeWidth={1.8} aria-hidden="true" />
            )}
            {mode === 'dark' ? 'Tema claro' : 'Tema oscuro'}
          </button>
        </li>
        <li>
          <button type="button" onClick={logout} className={cx(navLinkClass({ isActive: false }), 'w-full border-0 bg-transparent')}>
            <LogOut size={20} strokeWidth={1.8} aria-hidden="true" />
            Cerrar sesión
          </button>
        </li>
      </ul>
    </div>
  )
}

function SidebarNav() {
  const { user } = useAuth()
  const { currentTenantId } = useTenants()

  return (
    <nav aria-label="Principal" className="flex flex-1 flex-col gap-6">
      <TenantSelector />
      {currentTenantId && (
        <NavGroup>
          <NavItem to={`/portal/${currentTenantId}`} icon={Cpu} end>
            Dispositivos
          </NavItem>
          <NavItem to={`/portal/${currentTenantId}/maintenance`} icon={Wrench}>
            Mantención
          </NavItem>
          <NavItem to={`/portal/${currentTenantId}/medicines`} icon={Pill}>
            Medicinas
          </NavItem>
        </NavGroup>
      )}
      <NavGroup label="General">
        <NavItem to="/energy" icon={Sun}>
          Energía
        </NavItem>
        <NavItem to="/live-messages" icon={Radio}>
          Mensajes en vivo
        </NavItem>
        {user?.is_admin && (
          <NavItem to="/admin" icon={Shield}>
            Administración
          </NavItem>
        )}
      </NavGroup>
    </nav>
  )
}

const AppShell = ({ children }) => {
  const location = useLocation()
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    setMenuOpen(false)
  }, [location.pathname])

  return (
    <div className="min-h-screen bg-canvas lg:flex">
      <a
        href="#contenido"
        className="sr-only z-toast rounded-pill bg-primary px-4 py-2 text-on-primary focus:not-sr-only focus:fixed focus:left-4 focus:top-4"
      >
        Saltar al contenido
      </a>

      <header className="sticky top-0 z-sticky flex items-center justify-between gap-3 border-b border-line bg-surface px-5 py-2 lg:hidden">
        <NavLink to="/" aria-label="Zensor, inicio" className="flex min-h-touch items-center">
          <Mark className="h-8 w-8" />
        </NavLink>
        <IconButton
          icon={menuOpen ? X : Menu}
          label={menuOpen ? 'Cerrar menú' : 'Abrir menú'}
          aria-expanded={menuOpen}
          aria-controls="sidebar"
          onClick={() => setMenuOpen(open => !open)}
        />
      </header>

      <aside
        id="sidebar"
        className={cx(
          'flex-col gap-6 border-line bg-surface px-4 py-6',
          'lg:sticky lg:top-0 lg:flex lg:h-screen lg:w-sidebar lg:shrink-0 lg:overflow-y-auto lg:border-r',
          menuOpen ? 'flex border-b' : 'hidden',
        )}
      >
        <NavLink to="/" aria-label="Zensor, inicio" className="hidden min-h-touch items-center px-3 lg:flex">
          <Lockup />
        </NavLink>
        <SidebarNav />
        <SidebarFooter />
      </aside>

      <main id="contenido" className="min-w-0 flex-1 px-5 py-6 md:px-10 md:py-8">
        {children}
      </main>
    </div>
  )
}

export default AppShell
