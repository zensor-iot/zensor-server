import { Routes, Route, useLocation } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import LoginPage from './components/LoginPage'
import AccessDenied from './components/AccessDenied'
import AppShell from './components/AppShell'
import AdminUsers from './components/admin/AdminUsers'
import VictronDashboard from './components/VictronDashboard'
import TenantList from './components/TenantList'
import TenantDevices from './components/TenantDevices'
import DevicesPage from './components/devices/DevicesPage'
import DeviceMessagesLive from './components/DeviceMessagesLive'
import MaintenanceActivities from './components/maintenance/MaintenanceActivities'
import MaintenanceActivityCreate from './components/maintenance/MaintenanceActivityCreate'
import MaintenanceActivityDetail from './components/maintenance/MaintenanceActivityDetail'
import MaintenanceUpNext from './components/maintenance/MaintenanceUpNext'
import ExecutionDeeplink from './components/maintenance/ExecutionDeeplink'
import MedicinePatients from './components/medicines/MedicinePatients'
import PatientForm from './components/medicines/PatientForm'
import PatientDetail from './components/medicines/PatientDetail'
import TreatmentForm from './components/medicines/TreatmentForm'
import MedicineAgenda from './components/medicines/MedicineAgenda'
import DoseDeeplink from './components/medicines/DoseDeeplink'
import Profile from './components/Profile'
import AdminDashboard from './components/admin/AdminDashboard'
import AdminTenants from './components/admin/AdminTenants'
import AdminDevices from './components/admin/AdminDevices'
import AdminScheduledTasks from './components/admin/AdminScheduledTasks'
import AdminTaskExecutions from './components/admin/AdminTaskExecutions'
import AdminCommands from './components/admin/AdminCommands'
import AdminHealth from './components/admin/AdminHealth'
import AdminApiKeys from './components/admin/AdminApiKeys'
import { NotificationProvider } from './components/NotificationSystem'
import { TenantProvider } from './contexts/TenantContext'
import { LoadingState } from './ui'

function AuthGate({ children }) {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (location.pathname === '/access-denied') {
    return <AccessDenied />
  }

  if (loading) {
    return <LoadingState className="justify-center py-16" />
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
  const routes = (
    <Routes>
      <Route path="/" element={<TenantList />} />
      <Route path="/tenants/:tenantId/devices" element={<TenantDevices />} />
      <Route path="/portal/:tenantId" element={<DevicesPage />} />
      <Route path="/portal/:tenantId/maintenance" element={<MaintenanceActivities />} />
      <Route path="/portal/:tenantId/maintenance/new" element={<MaintenanceActivityCreate />} />
      <Route path="/portal/:tenantId/maintenance/activities/:activityId" element={<MaintenanceActivityDetail />} />
      <Route path="/portal/:tenantId/maintenance/up-next" element={<MaintenanceUpNext />} />
      <Route path="/maintenance/executions/:executionId" element={<ExecutionDeeplink />} />
      <Route path="/portal/:tenantId/medicines" element={<MedicinePatients />} />
      <Route path="/portal/:tenantId/medicines/agenda" element={<MedicineAgenda />} />
      <Route path="/portal/:tenantId/medicines/patients/new" element={<PatientForm />} />
      <Route path="/portal/:tenantId/medicines/patients/:patientId" element={<PatientDetail />} />
      <Route path="/portal/:tenantId/medicines/patients/:patientId/treatments/new" element={<TreatmentForm />} />
      <Route path="/medicines/doses/:doseId" element={<DoseDeeplink />} />
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
      <Route path="/admin/api-keys" element={<AdminApiKeys />} />
    </Routes>
  )

  return (
    <NotificationProvider>
      <TenantProvider>
        <AppShell>{routes}</AppShell>
      </TenantProvider>
    </NotificationProvider>
  )
}

export default App
