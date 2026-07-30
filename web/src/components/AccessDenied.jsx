import { ShieldOff, LogIn } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import './LoginPage.css'

const AccessDenied = () => {
    const { login } = useAuth()

    return (
        <div className="login-page">
            <div className="login-card">
                <div className="login-logo">
                    <ShieldOff size={40} className="access-denied-icon" />
                    <h1>Access Denied</h1>
                </div>
                <p className="login-subtitle">
                    Your Google account is not enabled for this portal.
                    Ask an administrator to add your email address, then sign in again.
                </p>
                <button className="login-button" onClick={login}>
                    <LogIn size={18} />
                    Try another account
                </button>
            </div>
        </div>
    )
}

export default AccessDenied
