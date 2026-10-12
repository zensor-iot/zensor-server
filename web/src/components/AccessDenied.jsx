import { LogIn } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { Button } from '../ui'
import { AuthLayout } from './LoginPage'

const AccessDenied = () => {
  const { login } = useAuth()

  return (
    <AuthLayout title="Tu cuenta no tiene acceso">
      <p className="zs-prose m-0 text-body text-muted">
        Tu cuenta de Google todavía no está habilitada en Zensor. Pídele a un administrador que agregue tu correo y
        vuelve a iniciar sesión.
      </p>
      <Button variant="secondary" icon={LogIn} size="lg" block onClick={login}>
        Probar con otra cuenta
      </Button>
    </AuthLayout>
  )
}

export default AccessDenied
