import { Link } from 'react-router-dom'
import { ChevronRight } from 'lucide-react'
import { useTenants } from '../contexts/TenantContext'
import { buttonClass, Card, EmptyState, ErrorState, List, LoadingState, Page, PageHeader, StatusPill } from '../ui'

const TenantList = () => {
  const { tenants, loading, error, reload } = useTenants()

  return (
    <Page>
      <PageHeader title="Cuentas" description="Elige una cuenta para ver sus dispositivos y programas." />

      {error ? (
        <ErrorState title="No se pudieron cargar las cuentas" onRetry={reload}>
          {`Error: ${error}. Revisa tu conexión y vuelve a intentarlo.`}
        </ErrorState>
      ) : loading && tenants.length === 0 ? (
        <LoadingState>Cargando cuentas…</LoadingState>
      ) : tenants.length === 0 ? (
        <EmptyState title="Aún no tienes cuentas">
          Pídele a un administrador que te agregue a una cuenta para empezar a ver tus dispositivos.
        </EmptyState>
      ) : (
        <List>
          {tenants.map(tenant => (
            <li key={tenant.id}>
              <Card
                as={Link}
                to={`/portal/${tenant.id}`}
                interactive
                row
                className="no-underline"
              >
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                  <div className="flex flex-wrap items-center gap-3">
                    <h2 className="m-0 break-words text-title">{tenant.name}</h2>
                    {tenant.is_active === false && <StatusPill tone="neutral">Inactiva</StatusPill>}
                  </div>
                  {tenant.email && <p className="m-0 break-words text-body-sm text-muted">{tenant.email}</p>}
                  {tenant.description && <p className="zs-prose m-0 text-body-sm text-muted">{tenant.description}</p>}
                </div>
                <span className={buttonClass({ variant: 'ghost', size: 'sm', className: 'pointer-events-none shrink-0' })}>
                  <span className="hidden sm:inline">Ver dispositivos</span>
                  <ChevronRight size={20} strokeWidth={1.8} aria-hidden="true" />
                </span>
              </Card>
            </li>
          ))}
        </List>
      )}
    </Page>
  )
}

export default TenantList
