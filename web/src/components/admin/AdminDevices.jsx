import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Pencil, Plus } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import {
  Button,
  Drawer,
  EmptyState,
  ErrorState,
  Field,
  Form,
  IconButton,
  Input,
  KeyValue,
  List,
  LoadingState,
  Page,
  PageHeader,
  Section,
  StatusPill,
  Table,
} from '../../ui'
import { formatDateTime, formatNumber, formatRelative } from '../../lib/zensor'
import { jsonOrThrow, listOf, taskStatus, totalOf } from '../../utils/adminData'
import { describeSchedule } from '../../utils/irrigationSchedule'

const EMPTY_CREATE = { name: '', display_name: '', dev_eui: '', app_eui: '', app_key: '' }
const FORM_ID = 'admin-device-form'
const RECENT_TASKS = 10

const STATUS = {
  online: { label: 'En línea', tone: 'ok' },
  offline: { label: 'Sin conexión', tone: 'attention' },
}

const statusOf = device => STATUS[device?.status] || { label: 'Sin datos', tone: 'neutral' }
const nameOf = device => device?.display_name || device?.name || 'Dispositivo sin nombre'
const lastSeen = device => (device?.last_message_received_at ? formatRelative(device.last_message_received_at) : 'Nunca')
const countLabel = value => (typeof value === 'number' ? formatNumber(value) : '—')

const fetchTotal = async url => {
  try {
    const response = await fetch(url)
    return response.ok ? totalOf(await response.json()) : null
  } catch {
    return null
  }
}

const AdminDevices = () => {
  const { tenantId } = useParams()
  const { showSuccess, showError } = useNotification()
  const [tenant, setTenant] = useState(null)
  const [devices, setDevices] = useState([])
  const [counts, setCounts] = useState({})
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [drawer, setDrawer] = useState(null)
  const [detail, setDetail] = useState(null)
  const [formData, setFormData] = useState(EMPTY_CREATE)
  const [saving, setSaving] = useState(false)

  const fetchCounts = useCallback(
    async deviceList => {
      const entries = await Promise.all(
        deviceList.map(async device => {
          const [tasks, programs] = await Promise.all([
            fetchTotal(`/v1/devices/${device.id}/tasks?limit=1`),
            fetchTotal(`/v1/tenants/${tenantId}/devices/${device.id}/scheduled-tasks?limit=1`),
          ])
          return [device.id, { tasks, programs }]
        }),
      )
      setCounts(Object.fromEntries(entries))
    },
    [tenantId],
  )

  const fetchTenantAndDevices = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const [tenantResponse, devicesResponse] = await Promise.all([
        fetch(`/v1/tenants/${tenantId}`),
        fetch(`/v1/tenants/${tenantId}/devices?limit=100`),
      ])
      if (tenantResponse.ok) setTenant(await tenantResponse.json())
      const list = listOf(await jsonOrThrow(devicesResponse, 'No se pudieron cargar los dispositivos'))
      setDevices(list)
      fetchCounts(list)
    } catch {
      setLoadError(true)
      setDevices([])
    } finally {
      setLoading(false)
    }
  }, [tenantId, fetchCounts])

  useEffect(() => {
    if (tenantId) fetchTenantAndDevices()
  }, [tenantId, fetchTenantAndDevices])

  const loadDetail = useCallback(
    async device => {
      setDetail({ loading: true })
      try {
        const [programsResponse, tasksResponse] = await Promise.all([
          fetch(`/v1/tenants/${tenantId}/devices/${device.id}/scheduled-tasks`),
          fetch(`/v1/devices/${device.id}/tasks?limit=${RECENT_TASKS}`),
        ])
        const [programs, tasks] = await Promise.all([
          jsonOrThrow(programsResponse, 'programs'),
          jsonOrThrow(tasksResponse, 'tasks'),
        ])
        setDetail({ programs: listOf(programs), tasks: listOf(tasks), tasksTotal: totalOf(tasks) })
      } catch {
        setDetail({ error: true })
      }
    },
    [tenantId],
  )

  const openDetail = device => {
    setDrawer({ mode: 'detail', device })
    loadDetail(device)
  }

  const openEdit = device => {
    setFormData({ ...EMPTY_CREATE, display_name: device.display_name || device.name || '' })
    setDrawer({ mode: 'edit', device })
  }

  const openCreate = () => {
    setFormData(EMPTY_CREATE)
    setDrawer({ mode: 'create' })
  }

  const closeDrawer = () => {
    if (saving) return
    setDrawer(null)
    setDetail(null)
  }

  const handleCreate = async () => {
    const payload = { name: formData.name.trim(), display_name: formData.display_name.trim() }
    for (const field of ['dev_eui', 'app_eui', 'app_key']) {
      const value = formData[field].trim()
      if (value) payload[field] = value
    }

    const createResponse = await fetch('/v1/devices', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    const created = await jsonOrThrow(
      createResponse,
      createResponse.status === 409
        ? 'Ya existe un dispositivo con ese nombre técnico. Usa otro nombre.'
        : 'No se pudo crear el dispositivo. Revisa los datos y vuelve a intentarlo.',
    )

    const adoptResponse = await fetch(`/v1/tenants/${tenantId}/adopt`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ device_id: created.id }),
    })
    if (!adoptResponse.ok) {
      throw new Error('El dispositivo se creó, pero no se pudo asignar a esta cuenta. Vuelve a intentarlo.')
    }
    showSuccess('Dispositivo agregado')
  }

  const handleUpdate = async () => {
    const response = await fetch(`/v1/devices/${drawer.device.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ display_name: formData.display_name.trim() }),
    })
    await jsonOrThrow(response, 'No se pudo guardar el dispositivo. Vuelve a intentarlo.')
    showSuccess('Dispositivo guardado')
  }

  const handleSubmit = async event => {
    event.preventDefault()
    try {
      setSaving(true)
      if (drawer.mode === 'create') await handleCreate()
      else await handleUpdate()
      setDrawer(null)
      setDetail(null)
      fetchTenantAndDevices()
    } catch (error) {
      showError(error.message)
    } finally {
      setSaving(false)
    }
  }

  const setField = field => event => setFormData(current => ({ ...current, [field]: event.target.value }))

  const device = drawer?.device
  const isForm = drawer?.mode === 'create' || drawer?.mode === 'edit'
  const drawerTitle = drawer?.mode === 'create' ? 'Agregar dispositivo' : drawer?.mode === 'edit' ? 'Editar dispositivo' : nameOf(device)

  const drawerSubtitle =
    drawer?.mode === 'detail' ? (
      <>
        <StatusPill tone={statusOf(device).tone}>{statusOf(device).label}</StatusPill>
        <span>Última señal: {lastSeen(device)}</span>
      </>
    ) : drawer?.mode === 'edit' ? (
      nameOf(device)
    ) : undefined

  const drawerFooter = isForm ? (
    <>
      <Button type="submit" form={FORM_ID} disabled={saving}>
        {saving ? 'Guardando…' : drawer.mode === 'create' ? 'Agregar dispositivo' : 'Guardar cambios'}
      </Button>
      <Button variant="ghost" onClick={closeDrawer} disabled={saving}>
        Cancelar
      </Button>
    </>
  ) : drawer?.mode === 'detail' ? (
    <Button variant="secondary" icon={Pencil} onClick={() => openEdit(device)}>
      Editar dispositivo
    </Button>
  ) : undefined

  return (
    <Page>
      <PageHeader
        title={tenant?.name ? `Dispositivos de ${tenant.name}` : 'Dispositivos'}
        description="Revisa el estado de cada dispositivo de la cuenta, sus programas y sus tareas recientes."
        back={{ to: '/admin/tenants', label: 'Cuentas' }}
        actions={
          <Button icon={Plus} onClick={openCreate}>
            Agregar dispositivo
          </Button>
        }
      />

      {loading && <LoadingState>Cargando dispositivos…</LoadingState>}

      {!loading && loadError && (
        <ErrorState title="No se pudieron cargar los dispositivos" onRetry={fetchTenantAndDevices}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )}

      {!loading && !loadError && devices.length === 0 && (
        <EmptyState title="Esta cuenta aún no tiene dispositivos">
          Agrega un dispositivo para empezar a recibir sus lecturas.
        </EmptyState>
      )}

      {!loading && !loadError && devices.length > 0 && (
        <Table caption="Dispositivos de la cuenta">
          <thead>
            <tr>
              <th scope="col">Dispositivo</th>
              <th scope="col">Estado</th>
              <th scope="col">Última señal</th>
              <th scope="col" className="zs-num">
                Programas
              </th>
              <th scope="col" className="zs-num">
                Tareas
              </th>
              <th scope="col">
                <span className="sr-only">Acciones</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {devices.map(item => {
              const status = statusOf(item)
              const itemCounts = counts[item.id] || {}
              return (
                <tr key={item.id}>
                  <td>
                    <button
                      type="button"
                      onClick={() => openDetail(item)}
                      className="min-h-touch bg-transparent p-0 text-left font-bold text-ink underline-offset-4 hover:underline"
                    >
                      {nameOf(item)}
                    </button>
                    {item.display_name && item.name && item.display_name !== item.name && (
                      <p className="m-0 font-mono text-caption text-muted">{item.name}</p>
                    )}
                  </td>
                  <td>
                    <StatusPill tone={status.tone}>{status.label}</StatusPill>
                  </td>
                  <td>{lastSeen(item)}</td>
                  <td className="zs-num">{countLabel(itemCounts.programs)}</td>
                  <td className="zs-num">{countLabel(itemCounts.tasks)}</td>
                  <td>
                    <div className="flex justify-end">
                      <IconButton icon={Pencil} label={`Editar ${nameOf(item)}`} onClick={() => openEdit(item)} />
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </Table>
      )}

      <Drawer open={!!drawer} onClose={closeDrawer} title={drawerTitle} subtitle={drawerSubtitle} footer={drawerFooter}>
        {drawer?.mode === 'detail' && (
          <DeviceDetail tenantId={tenantId} device={device} detail={detail} onRetry={() => loadDetail(device)} />
        )}

        {drawer?.mode === 'edit' && (
          <Form id={FORM_ID} onSubmit={handleSubmit}>
            <Field label="Nombre visible" hint="Es el nombre que ven las personas de la cuenta." required>
              {props => (
                <Input {...props} type="text" maxLength={100} value={formData.display_name} onChange={setField('display_name')} />
              )}
            </Field>
            <KeyValue items={[{ label: 'Nombre técnico', value: device?.name, mono: true }]} />
          </Form>
        )}

        {drawer?.mode === 'create' && (
          <Form id={FORM_ID} onSubmit={handleSubmit}>
            <Field label="Nombre técnico" hint="Identifica al dispositivo en la red. No se puede cambiar después." required>
              {props => (
                <Input {...props} type="text" autoComplete="off" value={formData.name} onChange={setField('name')} />
              )}
            </Field>
            <Field label="Nombre visible" hint="Si lo dejas vacío, se usa el nombre técnico.">
              {props => (
                <Input {...props} type="text" maxLength={100} value={formData.display_name} onChange={setField('display_name')} />
              )}
            </Field>
            <Field label="Dev EUI">
              {props => (
                <Input {...props} type="text" autoComplete="off" className="font-mono" value={formData.dev_eui} onChange={setField('dev_eui')} />
              )}
            </Field>
            <Field label="App EUI">
              {props => (
                <Input {...props} type="text" autoComplete="off" className="font-mono" value={formData.app_eui} onChange={setField('app_eui')} />
              )}
            </Field>
            <Field label="App key">
              {props => (
                <Input {...props} type="text" autoComplete="off" className="font-mono" value={formData.app_key} onChange={setField('app_key')} />
              )}
            </Field>
          </Form>
        )}
      </Drawer>
    </Page>
  )
}

function DeviceDetail({ tenantId, device, detail, onRetry }) {
  const programsUrl = `/admin/tenants/${tenantId}/devices/${device.id}/scheduled-tasks`

  return (
    <>
      <Section title="Dispositivo">
        <KeyValue
          items={[
            { label: 'Nombre técnico', value: device.name, mono: true },
            { label: 'Nombre visible', value: device.display_name },
            { label: 'Dev EUI', value: device.dev_eui, mono: true },
            { label: 'App EUI', value: device.app_eui, mono: true },
            {
              label: 'Última señal',
              value: device.last_message_received_at ? formatDateTime(device.last_message_received_at) : 'Nunca',
            },
            { label: 'ID', value: device.id, mono: true },
          ]}
        />
      </Section>

      {(!detail || detail.loading) && <LoadingState>Cargando programas y tareas…</LoadingState>}

      {detail?.error && (
        <ErrorState title="No se pudieron cargar los programas y tareas" onRetry={onRetry}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )}

      {detail?.programs && (
        <Section
          title="Programas"
          actions={
            <Button variant="secondary" size="sm" to={programsUrl}>
              Administrar programas
            </Button>
          }
        >
          {detail.programs.length === 0 ? (
            <EmptyState>Este dispositivo no tiene programas. Créalos desde Administrar programas.</EmptyState>
          ) : (
            <List>
              {detail.programs.map(program => (
                <li
                  key={program.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-line bg-surface px-4 py-3"
                >
                  <span className="flex min-w-0 flex-col gap-1">
                    <span>{describeSchedule(program)}</span>
                    <span className="text-caption text-muted">
                      {program.commands?.length === 1 ? '1 comando' : `${formatNumber(program.commands?.length || 0)} comandos`}
                    </span>
                  </span>
                  <StatusPill tone={program.is_active ? 'ok' : 'neutral'}>{program.is_active ? 'Activo' : 'Pausado'}</StatusPill>
                </li>
              ))}
            </List>
          )}
        </Section>
      )}

      {detail?.tasks && (
        <Section
          title="Tareas recientes"
          description={
            detail.tasksTotal > detail.tasks.length
              ? `Últimas ${formatNumber(detail.tasks.length)} de ${formatNumber(detail.tasksTotal)}`
              : undefined
          }
        >
          {detail.tasks.length === 0 ? (
            <EmptyState>Este dispositivo aún no ha recibido tareas.</EmptyState>
          ) : (
            <Table caption="Tareas recientes">
              <thead>
                <tr>
                  <th scope="col">Creada</th>
                  <th scope="col" className="zs-num">
                    Comandos
                  </th>
                  <th scope="col">Estado</th>
                </tr>
              </thead>
              <tbody>
                {detail.tasks.map(task => {
                  const status = taskStatus(task)
                  return (
                    <tr key={task.id}>
                      <td>{task.created_at ? formatDateTime(task.created_at) : '—'}</td>
                      <td className="zs-num">{formatNumber(task.commands?.length || 0)}</td>
                      <td>
                        <StatusPill tone={status.tone}>{status.label}</StatusPill>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </Table>
          )}
        </Section>
      )}
    </>
  )
}

export default AdminDevices
