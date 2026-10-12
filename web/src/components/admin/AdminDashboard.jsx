import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Activity, Building, ChevronRight, KeyRound, Play, Users } from 'lucide-react'
import { Card, CardGrid, Page, PageHeader } from '../../ui'
import { formatNumber } from '../../lib/zensor'
import { totalOf } from '../../utils/adminData'

const plural = (count, one, many) => `${formatNumber(count)} ${count === 1 ? one : many}`

const AREAS = [
  {
    key: 'tenants',
    title: 'Cuentas',
    description: 'Crea y edita cuentas, y revisa sus dispositivos.',
    icon: Building,
    to: '/admin/tenants',
  },
  {
    key: 'commands',
    title: 'Comandos',
    description: 'Revisa los comandos enviados a los dispositivos y su estado.',
    icon: Play,
    to: '/admin/commands',
  },
  {
    key: 'health',
    title: 'Salud del sistema',
    description: 'Comprueba que el servidor y sus servicios respondan.',
    icon: Activity,
    to: '/admin/health',
  },
  {
    key: 'users',
    title: 'Usuarios',
    description: 'Decide qué correos pueden iniciar sesión y quién administra.',
    icon: Users,
    to: '/admin/users',
  },
  {
    key: 'api-keys',
    title: 'Claves de API',
    description: 'Da acceso a integraciones y scripts, y revoca claves que ya no uses.',
    icon: KeyRound,
    to: '/admin/api-keys',
  },
]

const AdminDashboard = () => {
  const [counts, setCounts] = useState(null)

  const fetchCounts = useCallback(async () => {
    try {
      const [tenantsResponse, devicesResponse] = await Promise.all([
        fetch('/v1/tenants?limit=1'),
        fetch('/v1/devices?limit=1'),
      ])
      if (!tenantsResponse.ok || !devicesResponse.ok) return
      const [tenants, devices] = await Promise.all([tenantsResponse.json(), devicesResponse.json()])
      setCounts({ tenants: totalOf(tenants), devices: totalOf(devices) })
    } catch {
      setCounts(null)
    }
  }, [])

  useEffect(() => {
    fetchCounts()
  }, [fetchCounts])

  const metaFor = key => {
    if (key !== 'tenants' || !counts) return null
    return `${plural(counts.tenants, 'cuenta', 'cuentas')} · ${plural(counts.devices, 'dispositivo', 'dispositivos')}`
  }

  return (
    <Page>
      <PageHeader title="Administración" description="Gestiona cuentas, accesos y el estado del servidor." />

      <nav aria-label="Áreas de administración">
        <CardGrid>
          {AREAS.map(area => {
            const Icon = area.icon
            const meta = metaFor(area.key)
            return (
              <Card key={area.key} as={Link} to={area.to} interactive row className="items-start no-underline">
                <Icon size={20} strokeWidth={1.8} aria-hidden="true" className="mt-1 shrink-0 text-muted" />
                <span className="flex min-w-0 flex-1 flex-col gap-1">
                  <span className="text-h3">{area.title}</span>
                  <span className="text-body-sm text-muted">{area.description}</span>
                  {meta && <span className="text-caption text-muted">{meta}</span>}
                </span>
                <ChevronRight size={20} strokeWidth={1.8} aria-hidden="true" className="mt-1 shrink-0 text-muted" />
              </Card>
            )
          })}
        </CardGrid>
      </nav>
    </Page>
  )
}

export default AdminDashboard
