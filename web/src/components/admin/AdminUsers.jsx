import { useCallback, useEffect, useState } from 'react'
import { Plus, Shield, ShieldOff, Trash2 } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import { useAuth } from '../../contexts/AuthContext'
import {
  Button,
  Checkbox,
  ConfirmDialog,
  Drawer,
  EmptyState,
  ErrorState,
  Field,
  Form,
  IconButton,
  Input,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  StatusPill,
  Table,
} from '../../ui'
import { formatDateTime } from '../../lib/zensor'

const EMPTY_FORM = { email: '', is_admin: false }
const FORM_ID = 'admin-user-form'

const AdminUsers = () => {
  const { showSuccess, showError } = useNotification()
  const { user: currentUser, mode } = useAuth()
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [formData, setFormData] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [updatingId, setUpdatingId] = useState(null)
  const [toRemove, setToRemove] = useState(null)
  const [removing, setRemoving] = useState(false)

  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const response = await fetch('/v1/admin/allowed-users')
      if (!response.ok) throw new Error()
      const body = await response.json()
      setUsers(Array.isArray(body) ? body : [])
    } catch {
      setLoadError(true)
      setUsers([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (mode === 'static') {
      setLoading(false)
      return
    }
    fetchUsers()
  }, [mode, fetchUsers])

  const openCreate = () => {
    setFormData(EMPTY_FORM)
    setDrawerOpen(true)
  }

  const closeDrawer = () => {
    if (!saving) setDrawerOpen(false)
  }

  const handleCreate = async event => {
    event.preventDefault()
    try {
      setSaving(true)
      const response = await fetch('/v1/admin/allowed-users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...formData, email: formData.email.trim() }),
      })
      if (response.status === 409) {
        showError('Ese correo ya tiene acceso. Búscalo en la lista.')
        return
      }
      if (!response.ok) throw new Error()
      showSuccess('Usuario agregado')
      setDrawerOpen(false)
      fetchUsers()
    } catch {
      showError('No se pudo agregar el usuario. Revisa el correo y vuelve a intentarlo.')
    } finally {
      setSaving(false)
    }
  }

  const handleToggleAdmin = async user => {
    try {
      setUpdatingId(user.id)
      const response = await fetch(`/v1/admin/allowed-users/${user.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_admin: !user.is_admin }),
      })
      if (!response.ok) throw new Error()
      showSuccess(user.is_admin ? `${user.email} ya no administra` : `${user.email} ahora administra`)
      fetchUsers()
    } catch {
      showError('No se pudo cambiar el rol. Vuelve a intentarlo.')
    } finally {
      setUpdatingId(null)
    }
  }

  const handleRemove = async () => {
    try {
      setRemoving(true)
      const response = await fetch(`/v1/admin/allowed-users/${toRemove.id}`, { method: 'DELETE' })
      if (!response.ok) throw new Error()
      showSuccess('Acceso quitado')
      setToRemove(null)
      fetchUsers()
    } catch {
      showError('No se pudo quitar el acceso. Vuelve a intentarlo.')
    } finally {
      setRemoving(false)
    }
  }

  const isStatic = mode === 'static'

  return (
    <Page>
      <PageHeader
        title="Usuarios"
        description="Decide qué correos de Google pueden iniciar sesión y quién administra."
        back={{ to: '/admin', label: 'Administración' }}
        actions={
          !isStatic && (
            <Button icon={Plus} onClick={openCreate}>
              Agregar usuario
            </Button>
          )
        }
      />

      {isStatic && (
        <Notice tone="neutral">
          El servidor usa acceso con credenciales fijas, así que no hay lista de usuarios que gestionar. Para usarla,
          cambia <code className="font-mono">auth.mode</code> a <code className="font-mono">google</code> en la
          configuración del servidor.
        </Notice>
      )}

      {!isStatic && loading && <LoadingState>Cargando usuarios…</LoadingState>}

      {!isStatic && !loading && loadError && (
        <ErrorState title="No se pudieron cargar los usuarios" onRetry={fetchUsers}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )}

      {!isStatic && !loading && !loadError && users.length === 0 && (
        <EmptyState title="Aún no hay usuarios con acceso">
          Agrega un correo de Google para que esa persona pueda iniciar sesión.
        </EmptyState>
      )}

      {!isStatic && !loading && !loadError && users.length > 0 && (
        <Table caption="Usuarios con acceso">
          <thead>
            <tr>
              <th scope="col">Correo</th>
              <th scope="col">Nombre</th>
              <th scope="col">Rol</th>
              <th scope="col">Último ingreso</th>
              <th scope="col">
                <span className="sr-only">Acciones</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {users.map(user => {
              const isSelf = currentUser?.user_id === user.id
              return (
                <tr key={user.id}>
                  <td className="break-all">
                    {user.email}
                    {isSelf && <span className="text-caption text-muted"> (tú)</span>}
                  </td>
                  <td>{user.display_name || '—'}</td>
                  <td>
                    <StatusPill tone={user.is_admin ? 'ok' : 'neutral'}>
                      {user.is_admin ? 'Administrador' : 'Miembro'}
                    </StatusPill>
                  </td>
                  <td>{user.last_login_at ? formatDateTime(user.last_login_at) : 'Nunca'}</td>
                  <td>
                    <div className="flex justify-end gap-1">
                      <IconButton
                        icon={user.is_admin ? ShieldOff : Shield}
                        label={
                          user.is_admin
                            ? `Quitar permisos de administrador a ${user.email}`
                            : `Dar permisos de administrador a ${user.email}`
                        }
                        onClick={() => handleToggleAdmin(user)}
                        disabled={isSelf || updatingId === user.id}
                      />
                      <IconButton
                        icon={Trash2}
                        variant="danger"
                        label={`Quitar acceso a ${user.email}`}
                        onClick={() => setToRemove(user)}
                        disabled={isSelf}
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
        open={drawerOpen}
        onClose={closeDrawer}
        title="Agregar usuario"
        footer={
          <>
            <Button type="submit" form={FORM_ID} disabled={saving}>
              {saving ? 'Agregando…' : 'Agregar usuario'}
            </Button>
            <Button variant="ghost" onClick={closeDrawer} disabled={saving}>
              Cancelar
            </Button>
          </>
        }
      >
        <Form id={FORM_ID} onSubmit={handleCreate}>
          <Field label="Correo de Google" hint="La persona inicia sesión con esta cuenta de Google." required>
            {props => (
              <Input
                {...props}
                type="email"
                autoComplete="off"
                placeholder="nombre@ejemplo.cl"
                value={formData.email}
                onChange={event => setFormData(current => ({ ...current, email: event.target.value }))}
              />
            )}
          </Field>
          <Checkbox
            label="Administrador"
            description="Puede gestionar cuentas, usuarios y claves de API."
            checked={formData.is_admin}
            onChange={event => setFormData(current => ({ ...current, is_admin: event.target.checked }))}
          />
        </Form>
      </Drawer>

      <ConfirmDialog
        open={!!toRemove}
        onClose={() => !removing && setToRemove(null)}
        onConfirm={handleRemove}
        title={toRemove ? `¿Quitar el acceso a ${toRemove.email}?` : ''}
        confirmLabel={removing ? 'Quitando…' : 'Quitar acceso'}
        destructive
        busy={removing}
      >
        Sus sesiones se cierran de inmediato y ya no podrá iniciar sesión.
      </ConfirmDialog>
    </Page>
  )
}

export default AdminUsers
