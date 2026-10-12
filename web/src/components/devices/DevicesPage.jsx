import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTenants } from '../../contexts/TenantContext'
import { CardGrid, EmptyState, ErrorState, FilterChips, LoadingState, Notice, Page, PageHeader, StatusPill } from '../../ui'
import DeviceDrawer from './DeviceDrawer'
import DeviceSummaryCard from './DeviceSummaryCard'
import { isOnline } from './deviceStatus'
import useTenantDevices from './useTenantDevices'

const FILTERS = {
  all: () => true,
  online: isOnline,
  offline: device => !isOnline(device),
}

const offlineNotice = count =>
  count === 1 ? '1 dispositivo sin conexión. No riega ni envía lecturas.' : `${count} dispositivos sin conexión. No riegan ni envían lecturas.`

export default function DevicesPage() {
  const { tenantId } = useParams()
  const { currentTenant } = useTenants()
  const {
    devices,
    loading,
    error,
    reload,
    sensorData,
    programs,
    loadPrograms,
    setDevicePrograms,
    renameDevice,
    liveConnected,
  } = useTenantDevices(tenantId)

  const [filter, setFilter] = useState('all')
  const [selectedId, setSelectedId] = useState(null)
  const [drawerOpen, setDrawerOpen] = useState(false)

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

  const openDevice = id => {
    setSelectedId(id)
    setDrawerOpen(true)
  }

  return (
    <Page>
      <PageHeader
        title="Dispositivos"
        description={currentTenant ? `Riego y sensores de ${currentTenant.name}.` : undefined}
        actions={
          <StatusPill tone={liveConnected ? 'ok' : 'neutral'}>
            {liveConnected ? 'Lecturas en vivo' : 'Conectando lecturas…'}
          </StatusPill>
        }
      />

      {error ? (
        <ErrorState title="No se pudieron cargar los dispositivos" onRetry={reload}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      ) : loading ? (
        <LoadingState>Cargando dispositivos…</LoadingState>
      ) : devices.length === 0 ? (
        <EmptyState title="Aún no hay dispositivos">
          Cuando conectes un dispositivo a esta cuenta aparecerá aquí. Pide a un administrador que lo agregue.
        </EmptyState>
      ) : (
        <>
          {counts.offline > 0 && filter !== 'offline' && (
            <Notice action="Ver" onAction={() => setFilter('offline')}>
              {offlineNotice(counts.offline)}
            </Notice>
          )}

          <FilterChips
            label="Filtro de dispositivos"
            value={filter}
            onChange={setFilter}
            options={[
              { value: 'all', label: 'Todos', count: counts.all },
              { value: 'online', label: 'En línea', count: counts.online },
              { value: 'offline', label: 'Sin conexión', count: counts.offline },
            ]}
          />

          {visible.length === 0 ? (
            <EmptyState>No hay dispositivos con este filtro. Elige «Todos» para ver la lista completa.</EmptyState>
          ) : (
            <CardGrid>
              {visible.map(device => (
                <DeviceSummaryCard
                  key={device.id}
                  device={device}
                  sensorData={sensorData[device.name]}
                  programs={programs[device.id]}
                  onOpen={() => openDevice(device.id)}
                />
              ))}
            </CardGrid>
          )}
        </>
      )}

      <DeviceDrawer
        key={selected?.id}
        open={drawerOpen}
        tenantId={tenantId}
        device={selected}
        sensorData={selected ? sensorData[selected.name] : undefined}
        programs={selected ? programs[selected.id] : undefined}
        onClose={() => setDrawerOpen(false)}
        onRename={renameDevice}
        onProgramsChange={setDevicePrograms}
        onReloadPrograms={loadPrograms}
      />
    </Page>
  )
}
