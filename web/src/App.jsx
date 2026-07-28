import { Routes, Route, Link, useLocation } from 'react-router-dom'
import { Activity, Building, Cpu, Radio, Shield } from 'lucide-react'
import TenantList from './components/TenantList'
import TenantDevices from './components/TenantDevices'
import TenantPortal from './components/TenantPortal'
import DeviceMessagesLive from './components/DeviceMessagesLive'
import UserInfo from './components/UserInfo'
import Profile from './components/Profile'
import AdminDashboard from './components/admin/AdminDashboard'
import AdminTenants from './components/admin/AdminTenants'
import AdminDevices from './components/admin/AdminDevices'
import AdminScheduledTasks from './components/admin/AdminScheduledTasks'
import AdminTaskExecutions from './components/admin/AdminTaskExecutions'
import AdminCommands from './components/admin/AdminCommands'
import AdminHealth from './components/admin/AdminHealth'
import { NotificationProvider } from './components/NotificationSystem'
import { useAdmin } from './hooks/useAdmin'
import './App.css'

function App() {
  const location = useLocation()
  const isPortalPage = location.pathname.startsWith('/portal/')
  const isAdminPage = location.pathname.startsWith('/admin/')
  const { isAdmin, isLoading } = useAdmin()

  return (
    <NotificationProvider>
      <div className="app">
        {!isPortalPage && !isAdminPage && (
          <header className="app-header">
            <div className="header-content">
              <div className="header-left">
                <div className="logo">
                  <Activity className="logo-icon" />
                  <h1>Zensor Portal</h1>
                </div>
                <nav className="nav">
                  <Link to="/" className="nav-link">
                    <Building size={20} />
                    Tenants
                  </Link>
                  <Link to="/live-messages" className="nav-link">
                    <Radio size={20} />
                    Live Messages
                  </Link>
                  {!isLoading && isAdmin && (
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
        )}

        <main className={`main-content ${isPortalPage ? 'portal-main' : ''} ${isAdminPage ? 'admin-main' : ''}`}>
          <Routes>
            <Route path="/" element={<TenantList />} />
            <Route path="/tenants/:tenantId/devices" element={<TenantDevices />} />
            <Route path="/portal/:tenantId" element={<TenantPortal />} />
            <Route path="/live-messages" element={<DeviceMessagesLive />} />
            <Route path="/profile" element={<Profile />} />

            {/* Admin Routes */}
            <Route path="/admin" element={<AdminDashboard />} />
            <Route path="/admin/tenants" element={<AdminTenants />} />
            <Route path="/admin/tenants/:tenantId/devices" element={<AdminDevices />} />
            <Route path="/admin/tenants/:tenantId/devices/:deviceId/scheduled-tasks" element={<AdminScheduledTasks />} />
            <Route path="/admin/tenants/:tenantId/devices/:deviceId/scheduled-tasks/:taskId/executions" element={<AdminTaskExecutions />} />
            <Route path="/admin/commands" element={<AdminCommands />} />
            <Route path="/admin/health" element={<AdminHealth />} />
          </Routes>
        </main>
      </div>
    </NotificationProvider>
  )
}

export default App
