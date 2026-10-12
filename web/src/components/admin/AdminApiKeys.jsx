import { useCallback, useEffect, useState } from 'react'
import { Check, Copy, Plus, Trash2 } from 'lucide-react'
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
  KeyValue,
  LoadingState,
  Notice,
  Page,
  PageHeader,
  Table,
} from '../../ui'
import { formatDateTime } from '../../lib/zensor'
import { createAPIKey, listAPIKeys, revokeAPIKey } from '../../utils/apiKeys'

const FORM_ID = 'admin-api-key-form'

const CREATE_ERRORS = {
  duplicate_name: 'Ya existe una clave con ese nombre. Usa otro nombre.',
  name_required: 'Escribe un nombre para la clave.',
}

const AdminApiKeys = () => {
  const { showSuccess, showError } = useNotification()
  const [apiKeys, setApiKeys] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [step, setStep] = useState(null)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createdKey, setCreatedKey] = useState(null)
  const [copied, setCopied] = useState(false)
  const [confirmClose, setConfirmClose] = useState(false)
  const [toRevoke, setToRevoke] = useState(null)
  const [revoking, setRevoking] = useState(false)

  const fetchApiKeys = useCallback(async () => {
    try {
      setLoading(true)
      setLoadError(false)
      const keys = await listAPIKeys()
      setApiKeys(Array.isArray(keys) ? keys : [])
    } catch {
      setLoadError(true)
      setApiKeys([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchApiKeys()
  }, [fetchApiKeys])

  const openCreate = () => {
    setName('')
    setStep('create')
  }

  const finishReveal = () => {
    setConfirmClose(false)
    setStep(null)
    setCreatedKey(null)
    setCopied(false)
    fetchApiKeys()
  }

  const closeDrawer = () => {
    if (step === 'reveal') {
      if (copied) finishReveal()
      else setConfirmClose(true)
      return
    }
    if (!creating) setStep(null)
  }

  const handleCreate = async event => {
    event.preventDefault()
    try {
      setCreating(true)
      const created = await createAPIKey(name.trim())
      setCreatedKey(created)
      setCopied(false)
      setStep('reveal')
      showSuccess('Clave creada')
    } catch (error) {
      showError(CREATE_ERRORS[error.code] || 'No se pudo crear la clave. Vuelve a intentarlo.')
    } finally {
      setCreating(false)
    }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(createdKey.key)
      setCopied(true)
    } catch {
      showError('No se pudo copiar la clave. Selecciónala en el campo y cópiala a mano.')
    }
  }

  const handleRevoke = async () => {
    try {
      setRevoking(true)
      await revokeAPIKey(toRevoke.id)
      showSuccess('Clave revocada')
      setToRevoke(null)
      fetchApiKeys()
    } catch (error) {
      showError(
        error.code === 'not_found'
          ? 'Esa clave ya no existe. La lista se actualizó.'
          : 'No se pudo revocar la clave. Vuelve a intentarlo.',
      )
      if (error.code === 'not_found') {
        setToRevoke(null)
        fetchApiKeys()
      }
    } finally {
      setRevoking(false)
    }
  }

  const revealing = step === 'reveal' && createdKey

  return (
    <Page>
      <PageHeader
        title="Claves de API"
        description="Da acceso a la API a integraciones y scripts que no inician sesión."
        back={{ to: '/admin', label: 'Administración' }}
        actions={
          <Button icon={Plus} onClick={openCreate}>
            Crear clave
          </Button>
        }
      />

      {loading && <LoadingState>Cargando claves…</LoadingState>}

      {!loading && loadError && (
        <ErrorState title="No se pudieron cargar las claves" onRetry={fetchApiKeys}>
          Revisa tu conexión y vuelve a intentarlo.
        </ErrorState>
      )}

      {!loading && !loadError && apiKeys.length === 0 && (
        <EmptyState title="Aún no hay claves">
          Crea una clave cuando una integración necesite leer o escribir datos sin iniciar sesión.
        </EmptyState>
      )}

      {!loading && !loadError && apiKeys.length > 0 && (
        <Table caption="Claves de API">
          <thead>
            <tr>
              <th scope="col">Nombre</th>
              <th scope="col">Prefijo</th>
              <th scope="col">Creada</th>
              <th scope="col">
                <span className="sr-only">Acciones</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {apiKeys.map(apiKey => (
              <tr key={apiKey.id}>
                <td className="font-bold">{apiKey.name}</td>
                <td className="font-mono">{apiKey.key_prefix}…</td>
                <td>{apiKey.created_at ? formatDateTime(apiKey.created_at) : '—'}</td>
                <td>
                  <div className="flex justify-end">
                    <IconButton
                      icon={Trash2}
                      variant="danger"
                      label={`Revocar la clave ${apiKey.name}`}
                      onClick={() => setToRevoke(apiKey)}
                    />
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      <Drawer
        open={!!step && !confirmClose}
        onClose={closeDrawer}
        title={revealing ? 'Copia la clave ahora' : 'Crear clave'}
        subtitle={revealing ? createdKey.name : undefined}
        footer={
          revealing ? (
            <Button onClick={finishReveal}>Ya guardé la clave</Button>
          ) : (
            <>
              <Button type="submit" form={FORM_ID} disabled={creating}>
                {creating ? 'Creando…' : 'Crear clave'}
              </Button>
              <Button variant="ghost" onClick={closeDrawer} disabled={creating}>
                Cancelar
              </Button>
            </>
          )
        }
      >
        {revealing ? (
          <div className="flex flex-col gap-6">
            <Notice tone="attention">
              Esta es la única vez que verás la clave completa. Cópiala y guárdala en un lugar seguro: si cierras este
              panel sin hacerlo, tendrás que crear una nueva.
            </Notice>
            <Field label="Clave de API">
              {props => (
                <Input
                  {...props}
                  type="text"
                  readOnly
                  value={createdKey.key}
                  onFocus={event => event.target.select()}
                  className="font-mono"
                />
              )}
            </Field>
            <div>
              <Button variant="secondary" icon={copied ? Check : Copy} onClick={handleCopy}>
                {copied ? 'Clave copiada' : 'Copiar clave'}
              </Button>
              <span role="status" className="sr-only">
                {copied ? 'Clave copiada al portapapeles' : ''}
              </span>
            </div>
            <KeyValue
              items={[
                { label: 'Nombre', value: createdKey.name },
                { label: 'Uso', value: 'Authorization: Bearer <clave>', mono: true },
              ]}
            />
          </div>
        ) : (
          <Form id={FORM_ID} onSubmit={handleCreate}>
            <Field label="Nombre" hint="Sirve para reconocer la clave después. No se puede cambiar." required>
              {props => (
                <Input
                  {...props}
                  type="text"
                  autoComplete="off"
                  placeholder="grafana-sync"
                  value={name}
                  onChange={event => setName(event.target.value)}
                />
              )}
            </Field>
          </Form>
        )}
      </Drawer>

      <ConfirmDialog
        open={confirmClose}
        onClose={() => setConfirmClose(false)}
        onConfirm={finishReveal}
        title="¿Cerrar sin copiar la clave?"
        confirmLabel="Cerrar sin copiar"
        cancelLabel="Volver a la clave"
        destructive
      >
        No podrás ver esta clave otra vez. Si no la guardaste, tendrás que crear una nueva.
      </ConfirmDialog>

      <ConfirmDialog
        open={!!toRevoke}
        onClose={() => !revoking && setToRevoke(null)}
        onConfirm={handleRevoke}
        title={toRevoke ? `¿Revocar la clave ${toRevoke.name}?` : ''}
        confirmLabel={revoking ? 'Revocando…' : 'Revocar clave'}
        destructive
        busy={revoking}
      >
        Las integraciones que la usen dejan de tener acceso de inmediato.
      </ConfirmDialog>
    </Page>
  )
}

export default AdminApiKeys
