import { useState } from 'react'
import { LogIn } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { Button, Field, Form, Input, Lockup, Notice } from '../ui'

export function AuthLayout({ title, children }) {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center bg-canvas px-4 py-10 text-ink">
      <div className="flex w-full max-w-[26rem] flex-col items-center gap-8">
        <Lockup className="h-10 w-[150px]" />
        <section className="flex w-full flex-col gap-6 rounded-xl border border-line bg-surface p-6 sm:p-8">
          <h1 className="m-0 text-h2">{title}</h1>
          {children}
        </section>
      </div>
    </main>
  )
}

const LoginPage = () => {
  const { mode, login, loginWithCredentials, loginError } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleCredentialsSubmit = async event => {
    event.preventDefault()
    setSubmitting(true)
    await loginWithCredentials(username, password)
    setSubmitting(false)
  }

  return (
    <AuthLayout title="Iniciar sesión">
      <p className="m-0 text-body text-muted">Entra para ver tus dispositivos, programas y avisos.</p>
      {loginError && <Notice>{loginError}</Notice>}
      {mode === 'google' ? (
        <Button icon={LogIn} size="lg" block onClick={login}>
          Continuar con Google
        </Button>
      ) : (
        <Form onSubmit={handleCredentialsSubmit}>
          <Field label="Usuario">
            {props => (
              <Input
                {...props}
                type="text"
                autoComplete="username"
                value={username}
                onChange={event => setUsername(event.target.value)}
                autoFocus
              />
            )}
          </Field>
          <Field label="Contraseña">
            {props => (
              <Input
                {...props}
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={event => setPassword(event.target.value)}
              />
            )}
          </Field>
          <Button type="submit" icon={LogIn} size="lg" block disabled={submitting}>
            {submitting ? 'Iniciando sesión…' : 'Iniciar sesión'}
          </Button>
        </Form>
      )}
    </AuthLayout>
  )
}

export default LoginPage
