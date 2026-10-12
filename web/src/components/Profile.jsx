import { useCallback, useEffect, useState } from 'react'
import { useNotification } from '../hooks/useNotification'
import TimezoneDropdown from './TimezoneDropdown'
import {
  isWebPushSupported,
  getExistingSubscription,
  enableBrowserNotifications,
  disableBrowserNotifications,
  PermissionDeniedError,
} from '../utils/webPush'
import {
  Button,
  Card,
  EmptyState,
  ErrorState,
  Field,
  Form,
  FormActions,
  Input,
  KeyValue,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  Section,
  Select,
  Switch,
} from '../ui'

const DEFAULT_TENANT_ID = '550e8400-e29b-41d4-a716-446655440001'

const fetchJson = async url => {
  const response = await fetch(url)
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  return response.json()
}

const NotSet = () => <span className="text-muted">Sin definir</span>

const Profile = () => {
  const { showSuccess, showError, showApiNotification } = useNotification()
  const [userInfo, setUserInfo] = useState(null)
  const [authorizedTenants, setAuthorizedTenants] = useState([])
  const [selectedTenantId, setSelectedTenantId] = useState(null)
  const [tenantConfig, setTenantConfig] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [isEditingConfig, setIsEditingConfig] = useState(false)
  const [configForm, setConfigForm] = useState({ timezone: '', notification_email: '' })
  const [savingConfig, setSavingConfig] = useState(false)
  const [pushSupported] = useState(isWebPushSupported())
  const [pushEnabled, setPushEnabled] = useState(false)
  const [pushBusy, setPushBusy] = useState(false)
  const [pushPermissionDenied, setPushPermissionDenied] = useState(
    typeof Notification !== 'undefined' && Notification.permission === 'denied',
  )

  useEffect(() => {
    if (!pushSupported) return
    getExistingSubscription()
      .then(subscription => setPushEnabled(Boolean(subscription)))
      .catch(() => setPushEnabled(false))
  }, [pushSupported])

  const handlePushToggle = async () => {
    const userId = userInfo?.user_id
    if (!userId || pushBusy) return
    setPushBusy(true)
    try {
      if (pushEnabled) {
        await disableBrowserNotifications(userId)
        setPushEnabled(false)
        showSuccess('Notificaciones desactivadas en este navegador')
      } else {
        await enableBrowserNotifications(userId)
        setPushEnabled(true)
        setPushPermissionDenied(false)
        showSuccess('Notificaciones activadas en este navegador')
      }
    } catch (err) {
      if (err instanceof PermissionDeniedError) {
        setPushPermissionDenied(true)
      } else {
        showError(`No se pudieron cambiar las notificaciones: ${err.message}. Vuelve a intentarlo.`)
      }
    } finally {
      setPushBusy(false)
    }
  }

  const fetchTenantConfig = useCallback(async tenantId => {
    try {
      const response = await fetch(`/v1/tenants/${tenantId}/configuration`)
      if (!response.ok && response.status !== 404) throw new Error(`HTTP ${response.status}`)
      const data = response.ok ? await response.json() : null
      setTenantConfig(data)
      setConfigForm({ timezone: data?.timezone || '', notification_email: data?.notification_email || '' })
    } catch {
      setTenantConfig(null)
    }
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)

    const fetchUserDetails = async email => {
      try {
        return await fetchJson(`/v1/users/${email}`)
      } catch (err) {
        if (err.message !== 'HTTP 404') {
          showError('No se pudieron cargar tus cuentas autorizadas. Recarga la página para volver a intentarlo.')
        }
        return null
      }
    }

    const fetchTenantDetails = async tenantId => {
      try {
        return await fetchJson(`/v1/tenants/${tenantId}`)
      } catch {
        showError('No se pudo cargar una de tus cuentas. Recarga la página para volver a intentarlo.')
        return null
      }
    }

    try {
      const data = await fetchJson('/v1/me')
      setUserInfo(data)

      const userDetails = data.email ? await fetchUserDetails(data.email) : null
      if (userDetails?.tenants?.length > 0) {
        const tenantDetails = await Promise.all(userDetails.tenants.map(fetchTenantDetails))
        const validTenants = tenantDetails.filter(Boolean)
        setAuthorizedTenants(validTenants)
        if (validTenants.length > 0) {
          setSelectedTenantId(validTenants[0].id)
          await fetchTenantConfig(validTenants[0].id)
        }
      } else {
        setSelectedTenantId(DEFAULT_TENANT_ID)
        await fetchTenantConfig(DEFAULT_TENANT_ID)
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [fetchTenantConfig]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    load()
  }, [load])

  const handleConfigCancel = () => {
    setIsEditingConfig(false)
    setConfigForm({
      timezone: tenantConfig?.timezone || '',
      notification_email: tenantConfig?.notification_email || '',
    })
  }

  const handleTenantSelect = async tenantId => {
    setSelectedTenantId(tenantId)
    setIsEditingConfig(false)
    await fetchTenantConfig(tenantId)
  }

  const handleConfigSubmit = async event => {
    event.preventDefault()
    if (!selectedTenantId) {
      showError('Elige una cuenta antes de guardar la configuración.')
      return
    }

    setSavingConfig(true)
    const result = await showApiNotification(
      fetch(`/v1/tenants/${selectedTenantId}/configuration`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(configForm),
      }).then(res => {
        if (!res.ok) throw new Error(`No se pudo guardar la configuración (HTTP ${res.status}). Vuelve a intentarlo.`)
        return res.json()
      }),
      'Configuración guardada',
      'No se pudo guardar la configuración. Vuelve a intentarlo.',
    )
    setSavingConfig(false)

    if (result.success) {
      setTenantConfig(result.data)
      setIsEditingConfig(false)
    }
  }

  const header = <PageHeader title="Perfil" description="Tus datos, las notificaciones de este navegador y la configuración de cada cuenta." />

  if (loading) {
    return (
      <Page>
        {header}
        <LoadingState>Cargando tu perfil…</LoadingState>
      </Page>
    )
  }

  if (error) {
    return (
      <Page>
        {header}
        <ErrorState title="No se pudo cargar tu perfil" onRetry={load}>
          {`Error: ${error}. Revisa tu conexión y vuelve a intentarlo.`}
        </ErrorState>
      </Page>
    )
  }

  const hasUserData = userInfo && (userInfo.name || userInfo.email || userInfo.user_id)

  if (!hasUserData) {
    return (
      <Page>
        {header}
        <EmptyState title="No encontramos tu sesión">Cierra sesión y vuelve a entrar para ver tu perfil.</EmptyState>
      </Page>
    )
  }

  return (
    <Page>
      {header}

      <Section title="Tus datos">
        <Card>
          <KeyValue
            items={[
              { label: 'Nombre', value: userInfo.name || <NotSet /> },
              { label: 'Correo', value: userInfo.email || <NotSet /> },
            ]}
          />
        </Card>
      </Section>

      <Section title="Notificaciones" description="Recordatorios de mantención y alertas en este navegador.">
        <Card>
          {!pushSupported ? (
            <p className="m-0 text-body-sm text-muted">
              Este navegador no permite notificaciones push. Prueba con Chrome, Edge, Firefox o Safari actualizado.
            </p>
          ) : (
            <div className="flex items-center justify-between gap-4">
              <div className="flex min-w-0 flex-col gap-1">
                <span className="text-label">Notificaciones en este navegador</span>
                <span className="text-body-sm text-muted">
                  {pushBusy ? 'Actualizando…' : pushEnabled ? 'Activadas' : 'Desactivadas'}
                </span>
              </div>
              <Switch
                checked={pushEnabled}
                onChange={handlePushToggle}
                disabled={pushBusy || !userInfo?.user_id}
                label={pushEnabled ? 'Desactivar notificaciones' : 'Activar notificaciones'}
              />
            </div>
          )}
          {pushSupported && pushPermissionDenied && (
            <Notice>
              Las notificaciones están bloqueadas para este sitio. Permítelas en la configuración del sitio de tu
              navegador y vuelve a activarlas.
            </Notice>
          )}
        </Card>
      </Section>

      <Section
        title="Configuración de la cuenta"
        description="Zona horaria para los programas y correo que recibe los avisos."
        actions={
          !isEditingConfig && (
            <Button variant="secondary" onClick={() => setIsEditingConfig(true)}>
              Editar configuración
            </Button>
          )
        }
      >
        <Card>
          {authorizedTenants.length > 1 && (
            <Field label="Cuenta">
              {props => (
                <Select {...props} value={selectedTenantId || ''} onChange={event => handleTenantSelect(event.target.value)}>
                  {authorizedTenants.map(tenant => (
                    <option key={tenant.id} value={tenant.id}>
                      {tenant.name}
                    </option>
                  ))}
                </Select>
              )}
            </Field>
          )}

          {!isEditingConfig ? (
            <KeyValue
              items={[
                authorizedTenants.length === 1 && { label: 'Cuenta', value: authorizedTenants[0].name },
                {
                  label: 'Zona horaria',
                  value: tenantConfig?.timezone ? <span className="font-mono text-md">{tenantConfig.timezone}</span> : <NotSet />,
                },
                { label: 'Correo para avisos', value: tenantConfig?.notification_email || <NotSet /> },
              ]}
            />
          ) : (
            <Form onSubmit={handleConfigSubmit}>
              <Field label="Zona horaria" hint="Busca por ciudad o por nombre IANA, por ejemplo America/Santiago.">
                {props => (
                  <TimezoneDropdown
                    {...props}
                    value={configForm.timezone}
                    onChange={value => setConfigForm(prev => ({ ...prev, timezone: value }))}
                  />
                )}
              </Field>
              <Field label="Correo para avisos">
                {props => (
                  <Input
                    {...props}
                    type="email"
                    name="notification_email"
                    autoComplete="email"
                    value={configForm.notification_email}
                    onChange={event => setConfigForm(prev => ({ ...prev, notification_email: event.target.value }))}
                    placeholder="avisos@ejemplo.cl"
                  />
                )}
              </Field>
              <FormActions>
                <Button type="submit" disabled={savingConfig}>
                  {savingConfig ? 'Guardando…' : 'Guardar configuración'}
                </Button>
                <Button variant="ghost" onClick={handleConfigCancel} disabled={savingConfig}>
                  Cancelar
                </Button>
              </FormActions>
            </Form>
          )}
        </Card>
      </Section>
    </Page>
  )
}

export default Profile
