import { Activity, LogIn } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import './LoginPage.css'

const LoginPage = () => {
    const { login } = useAuth()

    return (
        <div className="login-page">
            <div className="login-card">
                <div className="login-logo">
                    <Activity size={40} />
                    <h1>Zensor Portal</h1>
                </div>
                <p className="login-subtitle">Sign in to access your dashboards and devices</p>
                <button className="login-button" onClick={login}>
                    <LogIn size={18} />
                    Sign in with Google
                </button>
            </div>
        </div>
    )
}

export default LoginPage
