import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import {
  Button,
  ConfirmDialog,
  Drawer,
  EmptyState,
  ErrorState,
  Field,
  Form,
  IconButton,
  Input,
  LoadingState,
  Page,
  PageHeader,
  StatusPill,
  Table,
  Textarea,
} from '../../ui'
import { formatNumber } from '../../lib/zensor'
import { jsonOrThrow, listOf, totalOf } from '../../utils/adminData'

const EMPTY_FORM = { name: '', email: '', description: '' }
const FORM_ID = 'admin-tenant-form'

const saveErrorFor = status => {
  if (status === 409) return 'Ya existe una cuenta con ese nombre o la cuenta cambió mientras la editabas. Recarga y vuelve a intentarlo.'
  if (status === 400) return 'Revisa el nombre y el correo de contacto e inténtalo de nuevo.'
  return 'No se pudo guardar la cuenta. Vuelve a intentarlo.'
}

const AdminTenants = () => {
  const { showSuccess, showError } = useNotification()
  const [tenants, setTenants] = useState([])
  const [deviceCounts, setDeviceCounts] = useState({})
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [drawer, setDrawer] = useState(null)
  const [formData, setFormData] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [toDelete, setToDelete] = useState(null)
  const [deleting, setDeleting] = useState(false)

  const fetchDeviceCounts = useCallback(async tenantList => {
    const entries = await Promise.all(
      tenantList.map(async tenant => {
        try {
          const response = await fetch(`/v1/tenants/${tenant.id}/devices?limit=1`)
          if (!response.ok) return [tenant.id, null]
          return [tenant.id, totalOf(await response.json())]
        } catch {
          return [tenant.id, null]
        }
      }),
    )
    setDeviceCounts(Object.fromEntries(entries))
  }, [])

  const fetchTenants = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const response = await fetch('/v1/tenants?limit=100')
      const body = await jsonOrThrow(response, 'No se pudieron cargar las cuentas')
      const list = listOf(body)
      setTenants(list)
      fetchDeviceCounts(list)
    } catch {
      setLoadError(true)
      setTenants([])
    } finally {
      setLoading(false)
    }
  }, [fetchDeviceCounts])

  useEffect(() => {
    fetchTenants()
  }, [fetchTenants])

  const openCreate = () => {
    setFormData(EMPTY_FORM)
    setDrawer({ mode: 'create' })
  }

  const openEdit = tenant => {
    setFormData({ name: tenant.name || '', email: tenant.email || '', description: tenant.description || '' })
    setDrawer({ mode: 'edit', tenant })
  }

  const closeDrawer = () => {
    if (!saving) setDrawer(null)
  }

  const handleSubmit = async event => {
    event.preventDefault()
    const editing = drawer?.mode === 'edit'
    const payload = {
      name: formData.name.trim(),
      email: formData.email.trim(),
      description: formData.description.trim(),
    }
    try {
      setSaving(true)
      const response = await fetch(editing ? `/v1/tenants/${drawer.tenant.id}` : '/v1/tenants', {
        method: editing ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(editing ? { ...payload, version: drawer.tenant.version } : payload),
      })
      if (!response.ok) throw new Error(saveErrorFor(response.status))
      showSuccess(editing ? 'Cuenta guardada' : 'Cuenta creada')
      setDrawer(null)
      fetchTenants()
    } catch (error) {
      showError(error.message)
    } finally {
      setSaving(false)
    }
  }

  const handleToggleActive = async () => {
    const tenant = drawer.tenant
    const action = tenant.is_active ? 'deactivate' : 'activate'
    try {
      setSaving(true)
      const response = await fetch(`/v1/tenants/${tenant.id}/${action}`, { method: 'POST' })
      if (!response.ok) throw new Error()
      showSuccess(tenant.is_active ? 'Cuenta desactivada' : 'Cuenta activada')
      setDrawer(null)
      fetchTenants()
    } catch {
      showError(
        tenant.is_active
          ? 'No se pudo desactivar la cuenta. Vuelve a intentarlo.'
          : 'No se pudo activar la cuenta. Vuelve a intentarlo.',
      )
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    try {
      setDeleting(true)
      const response = await fetch(`/v1/tenants/${toDelete.id}`, { method: 'DELETE' })
      if (!response.ok) throw new Error()
      showSuccess('Cuenta eliminada')
      setToDelete(null)
      fetchTenants()
    } catch {
      showError('No se pudo eliminar la cuenta. Vuelve a intentarlo.')
    } finally {
      setDeleting(false)
    }
  }

  const setField = field => event => setFormData(current => ({ ...current, [field]: event.target.value }))
  const editing = drawer?.mode === 'edit'

  return (
    <Page>
      <PageHeader
        title="Cuentas"
        description="Cada cuenta agrupa los dispositivos y las personas de un cliente."
        back={{ to: '/admin', label: 'Administración' }}
        actions={
          <Button icon={Plus} onClick={openCreate}>
            Crear cuenta
          </Button>
        }
      />

      {loading && <LoadingState>Cargando cuentas…</LoadingState>}

      {!loading && loadError && (
        <ErrorState title="No se pudieron cargar las cuentas" onRetry={fetchTenants}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )}

      {!loading && !loadError && tenants.length === 0 && (
        <EmptyState title="Aún no hay cuentas">Crea la primera cuenta para empezar a asignarle dispositivos.</EmptyState>
      )}

      {!loading && !loadError && tenants.length > 0 && (
        <Table caption="Cuentas">
          <thead>
            <tr>
              <th scope="col">Cuenta</th>
              <th scope="col">Correo de contacto</th>
              <th scope="col">Estado</th>
              <th scope="col" className="zs-num">
                Dispositivos
              </th>
              <th scope="col">
                <span className="sr-only">Acciones</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {tenants.map(tenant => {
              const count = deviceCounts[tenant.id]
              return (
                <tr key={tenant.id}>
                  <td>
                    <Link
                      to={`/admin/tenants/${tenant.id}/devices`}
                      className="font-bold text-ink underline-offset-4 hover:underline"
                    >
                      {tenant.name}
                    </Link>
                    {tenant.description && <p className="m-0 text-body-sm text-muted">{tenant.description}</p>}
                  </td>
                  <td className="break-all">{tenant.email || '—'}</td>
                  <td>
                    <StatusPill tone={tenant.is_active ? 'ok' : 'neutral'}>
                      {tenant.is_active ? 'Activa' : 'Inactiva'}
                    </StatusPill>
                  </td>
                  <td className="zs-num">{typeof count === 'number' ? formatNumber(count) : '—'}</td>
                  <td>
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="sm" to={`/admin/tenants/${tenant.id}/devices`}>
                        Ver dispositivos
                      </Button>
                      <IconButton icon={Pencil} label={`Editar ${tenant.name}`} onClick={() => openEdit(tenant)} />
                      <IconButton
                        icon={Trash2}
                        variant="danger"
                        label={`Eliminar ${tenant.name}`}
                        onClick={() => setToDelete(tenant)}
                      />
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </Table>
      )}

      <Drawer
        open={!!drawer}
        onClose={closeDrawer}
        title={editing ? 'Editar cuenta' : 'Crear cuenta'}
        subtitle={editing ? drawer.tenant.name : undefined}
        footer={
          <>
            <Button type="submit" form={FORM_ID} disabled={saving}>
              {saving ? 'Guardando…' : editing ? 'Guardar cambios' : 'Crear cuenta'}
            </Button>
            <Button variant="ghost" onClick={closeDrawer} disabled={saving}>
              Cancelar
            </Button>
          </>
        }
      >
        <Form id={FORM_ID} onSubmit={handleSubmit}>
          <Field label="Nombre" required>
            {props => <Input {...props} type="text" maxLength={100} value={formData.name} onChange={setField('name')} />}
          </Field>
          <Field label="Correo de contacto" hint="Recibe los avisos de la cuenta." required>
            {props => (
              <Input
                {...props}
                type="email"
                autoComplete="email"
                placeholder="nombre@ejemplo.cl"
                value={formData.email}
                onChange={setField('email')}
              />
            )}
          </Field>
          <Field label="Descripción">
            {props => (
              <Textarea {...props} rows={3} maxLength={500} value={formData.description} onChange={setField('description')} />
            )}
          </Field>
        </Form>

        {editing && (
          <section className="flex flex-col gap-3 border-t border-line pt-6">
            <h3 className="m-0 text-h3">Estado</h3>
            <p className="m-0 text-body-sm text-muted">
              {drawer.tenant.is_active
                ? 'La cuenta está activa.'
                : 'La cuenta está inactiva.'}
            </p>
            <div>
              <Button variant="secondary" onClick={handleToggleActive} disabled={saving}>
                {drawer.tenant.is_active ? 'Desactivar cuenta' : 'Activar cuenta'}
              </Button>
            </div>
          </section>
        )}
      </Drawer>

      <ConfirmDialog
        open={!!toDelete}
        onClose={() => !deleting && setToDelete(null)}
        onConfirm={handleDelete}
        title={toDelete ? `¿Eliminar la cuenta ${toDelete.name}?` : ''}
        confirmLabel={deleting ? 'Eliminando…' : 'Eliminar cuenta'}
        destructive
        busy={deleting}
      >
        La cuenta deja de aparecer en Zensor. Esta acción no se puede deshacer desde aquí.
      </ConfirmDialog>
    </Page>
  )
}

export default AdminTenants
