import { Routes, Route, Link, useLocation } from 'react-router-dom'
import { Activity, Building, Cpu, Radio, Shield, Sun } from 'lucide-react'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import LoginPage from './components/LoginPage'
import AccessDenied from './components/AccessDenied'
import AdminUsers from './components/admin/AdminUsers'
import VictronDashboard from './components/VictronDashboard'
import TenantList from './components/TenantList'
import TenantDevices from './components/TenantDevices'
import TenantPortal from './components/TenantPortal'
import DeviceMessagesLive from './components/DeviceMessagesLive'
import MaintenanceActivities from './components/maintenance/MaintenanceActivities'
import MaintenanceActivityCreate from './components/maintenance/MaintenanceActivityCreate'
import MaintenanceActivityDetail from './components/maintenance/MaintenanceActivityDetail'
import MaintenanceUpNext from './components/maintenance/MaintenanceUpNext'
import ExecutionDeeplink from './components/maintenance/ExecutionDeeplink'
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
import './App.css'

function AuthGate({ children }) {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (location.pathname === '/access-denied') {
    return <AccessDenied />
  }

  if (loading) {
    return <div className="auth-loading">Loading...</div>
  }

  if (!user) {
    return <LoginPage />
  }

  return children
}

function App() {
  return (
    <AuthProvider>
      <AuthGate>
        <AppContent />
      </AuthGate>
    </AuthProvider>
  )
}

function AppContent() {
  const location = useLocation()
  const { user } = useAuth()
  const isPortalPage = location.pathname.startsWith('/portal/')
  const isAdminPage = location.pathname.startsWith('/admin/')

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
        )}

        <main className={`main-content ${isPortalPage ? 'portal-main' : ''} ${isAdminPage ? 'admin-main' : ''}`}>
          <Routes>
            <Route path="/" element={<TenantList />} />
            <Route path="/tenants/:tenantId/devices" element={<TenantDevices />} />
            <Route path="/portal/:tenantId" element={<TenantPortal />} />
            <Route path="/portal/:tenantId/maintenance" element={<MaintenanceActivities />} />
            <Route path="/portal/:tenantId/maintenance/new" element={<MaintenanceActivityCreate />} />
            <Route path="/portal/:tenantId/maintenance/activities/:activityId" element={<MaintenanceActivityDetail />} />
            <Route path="/portal/:tenantId/maintenance/up-next" element={<MaintenanceUpNext />} />
            <Route path="/maintenance/executions/:executionId" element={<ExecutionDeeplink />} />
            <Route path="/live-messages" element={<DeviceMessagesLive />} />
            <Route path="/energy" element={<VictronDashboard />} />
            <Route path="/profile" element={<Profile />} />

            {/* Admin Routes */}
            <Route path="/admin" element={<AdminDashboard />} />
            <Route path="/admin/tenants" element={<AdminTenants />} />
            <Route path="/admin/tenants/:tenantId/devices" element={<AdminDevices />} />
            <Route path="/admin/tenants/:tenantId/devices/:deviceId/scheduled-tasks" element={<AdminScheduledTasks />} />
            <Route path="/admin/tenants/:tenantId/devices/:deviceId/scheduled-tasks/:taskId/executions" element={<AdminTaskExecutions />} />
            <Route path="/admin/commands" element={<AdminCommands />} />
            <Route path="/admin/health" element={<AdminHealth />} />
            <Route path="/admin/users" element={<AdminUsers />} />
          </Routes>
        </main>
      </div>
    </NotificationProvider>
  )
}

export default App
