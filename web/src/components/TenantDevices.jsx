import { useCallback, useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import DeviceCard, { DeviceDetailDrawer } from './DeviceCard'
import { getApiUrl } from '../config/api'
import { EmptyState, ErrorState, FilterChips, List, LoadingState, Notice, Page, PageHeader, StatusPill } from '../ui'

const FILTERS = {
  all: () => true,
  online: device => device.status === 'online',
  offline: device => device.status !== 'online',
}

const TenantDevices = () => {
  const { tenantId } = useParams()
  const [tenant, setTenant] = useState(null)
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [filter, setFilter] = useState('all')
  const [selectedId, setSelectedId] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [tenantResponse, devicesResponse] = await Promise.all([
        fetch(getApiUrl(`/tenants/${tenantId}`)),
        fetch(getApiUrl(`/tenants/${tenantId}/devices`)),
      ])
      if (!tenantResponse.ok) throw new Error(`no se pudo cargar la cuenta (HTTP ${tenantResponse.status})`)
      if (!devicesResponse.ok) throw new Error(`no se pudieron cargar los dispositivos (HTTP ${devicesResponse.status})`)
      setTenant(await tenantResponse.json())
      const body = await devicesResponse.json()
      setDevices(body.data || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [tenantId])

  useEffect(() => {
    load()
  }, [load])

  const counts = useMemo(
    () => ({
      all: devices.length,
      online: devices.filter(FILTERS.online).length,
      offline: devices.filter(FILTERS.offline).length,
    }),
    [devices],
  )

  const visible = devices.filter(FILTERS[filter])
  const selected = devices.find(device => device.id === selectedId) || null

  const header = (
    <PageHeader
      title={tenant ? `Dispositivos de ${tenant.name}` : 'Dispositivos'}
      description={[tenant?.email, tenant?.description].filter(Boolean).join(' · ') || undefined}
      back={{ to: '/', label: 'Cuentas' }}
      actions={tenant?.is_active === false && <StatusPill tone="neutral">Cuenta inactiva</StatusPill>}
    />
  )

  if (error) {
    return (
      <Page>
        {header}
        <ErrorState title="No se pudo cargar esta cuenta" onRetry={load}>
          {`Error: ${error}. Revisa tu conexión y vuelve a intentarlo.`}
        </ErrorState>
      </Page>
    )
  }

  return (
    <Page>
      {header}

      {loading && devices.length === 0 ? (
        <LoadingState>Cargando dispositivos…</LoadingState>
      ) : devices.length === 0 ? (
        <EmptyState title="Esta cuenta aún no tiene dispositivos">
          Pídele a un administrador que adopte tus dispositivos en esta cuenta para verlos aquí.
        </EmptyState>
      ) : (
        <>
          {counts.offline > 0 && filter !== 'offline' && (
            <Notice action="Ver" onAction={() => setFilter('offline')}>
              {counts.offline === 1 ? '1 dispositivo sin conexión.' : `${counts.offline} dispositivos sin conexión.`}
            </Notice>
          )}
          <FilterChips
            label="Filtrar dispositivos"
            value={filter}
            onChange={setFilter}
            options={[
              { value: 'all', label: 'Todos', count: counts.all },
              { value: 'online', label: 'En línea', count: counts.online },
              { value: 'offline', label: 'Sin conexión', count: counts.offline },
            ]}
          />
          {visible.length === 0 ? (
            <EmptyState>
              {filter === 'online'
                ? 'Ningún dispositivo está en línea. Revisa su alimentación y la cobertura del gateway.'
                : 'Todos los dispositivos están en línea.'}
            </EmptyState>
          ) : (
            <List>
              {visible.map(device => (
                <li key={device.id}>
                  <DeviceCard device={device} onOpen={next => setSelectedId(next.id)} />
                </li>
              ))}
            </List>
          )}
        </>
      )}

      <DeviceDetailDrawer device={selected} onClose={() => setSelectedId(null)} />
    </Page>
  )
}

export default TenantDevices
