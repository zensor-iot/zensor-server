import { useState } from 'react'
import { Activity, LogIn } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import './LoginPage.css'

const LoginPage = () => {
    const { mode, login, loginWithCredentials, loginError } = useAuth()
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [submitting, setSubmitting] = useState(false)

    const handleStaticSubmit = async (event) => {
        event.preventDefault()
        setSubmitting(true)
        await loginWithCredentials(username, password)
        setSubmitting(false)
    }

    return (
        <div className="login-page">
            <div className="login-card">
                <div className="login-logo">
                    <Activity size={40} />
                    <h1>Zensor Portal</h1>
                </div>
                <p className="login-subtitle">Sign in to access your dashboards and devices</p>
                {mode === 'static' ? (
                    <form className="login-form" onSubmit={handleStaticSubmit}>
                        <input
                            className="login-input"
                            type="text"
                            placeholder="Username"
                            value={username}
                            onChange={(event) => setUsername(event.target.value)}
                            autoFocus
                        />
                        <input
                            className="login-input"
                            type="password"
                            placeholder="Password"
                            value={password}
                            onChange={(event) => setPassword(event.target.value)}
                        />
                        {loginError && <p className="login-error">{loginError}</p>}
                        <button className="login-button" type="submit" disabled={submitting}>
                            <LogIn size={18} />
                            {submitting ? 'Signing in…' : 'Sign in'}
                        </button>
                    </form>
                ) : (
                    <button className="login-button" onClick={login}>
                        <LogIn size={18} />
                        Sign in with Google
                    </button>
                )}
            </div>
        </div>
    )
}

export default LoginPage
