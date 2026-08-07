import { Link, useLocation } from 'react-router-dom'
import { Activity, Building, Radio, Shield, Sun } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import UserInfo from './UserInfo'

const AppShell = ({ children }) => {
  const location = useLocation()
  const { user } = useAuth()
  const isAdminPage = location.pathname.startsWith('/admin')
  const isPortalPage = location.pathname.startsWith('/portal/')

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-content">
          <div className="header-left">
            <Link to="/" className="logo">
              <Activity className="logo-icon" />
              <h1>Zensor Portal</h1>
            </Link>
            <nav className="nav">
              <Link to="/" className="nav-link">
                <Building size={20} />
                Tenants
              </Link>
              <Link to="/live-messages" className="nav-link">
                <Radio size={20} />
                Live Messages
              </Link>
              <Link to="/energy" className="nav-link">
                <Sun size={20} />
                Energy
              </Link>
              {user?.is_admin && (
                <Link to="/admin" className="nav-link admin-link">
                  <Shield size={20} />
                  Admin
                </Link>
              )}
            </nav>
          </div>
          <UserInfo />
        </div>
      </header>

      <main className={`main-content ${isAdminPage ? 'admin-main' : ''} ${isPortalPage ? 'portal-main' : ''}`}>
        {children}
      </main>
    </div>
  )
}

export default AppShell
