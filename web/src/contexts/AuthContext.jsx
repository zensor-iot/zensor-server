import { createContext, useContext, useEffect, useState, useCallback } from 'react'

const AuthContext = createContext(null)

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null)
    const [loading, setLoading] = useState(true)
    const [mode, setMode] = useState('google')
    const [loginError, setLoginError] = useState(null)

    useEffect(() => {
        fetch('/auth/mode')
            .then((response) => (response.ok ? response.json() : null))
            .then((body) => {
                if (body?.mode) setMode(body.mode)
            })
            .catch((err) => console.error('Failed to fetch auth mode:', err))
    }, [])

    const refresh = useCallback(async () => {
        try {
            const response = await fetch('/v1/me')
            if (response.ok) {
                setUser(await response.json())
            } else {
                setUser(null)
            }
        } catch (err) {
            console.error('Failed to fetch current user:', err)
            setUser(null)
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        refresh()
    }, [refresh])

    // Global 401 handling: an expired or revoked session on any API call
    // drops the user back to the login page.
    useEffect(() => {
        const originalFetch = window.fetch
        window.fetch = async (...args) => {
            const response = await originalFetch(...args)
            const url = typeof args[0] === 'string' ? args[0] : (args[0]?.url || '')
            if (response.status === 401 && !url.includes('/v1/me') && !url.includes('/auth/')) {
                setUser(null)
            }
            return response
        }
        return () => {
            window.fetch = originalFetch
        }
    }, [])

    const login = () => {
        window.location.href = '/auth/login'
    }

    const loginWithCredentials = async (username, password) => {
        setLoginError(null)
        try {
            const response = await fetch('/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            })
            if (!response.ok) {
                setLoginError('Invalid username or password')
                return false
            }
            setUser(await response.json())
            return true
        } catch (err) {
            console.error('Login request failed:', err)
            setLoginError('Login failed')
            return false
        }
    }

    const logout = async () => {
        try {
            await fetch('/auth/logout', { method: 'POST' })
        } catch (err) {
            console.error('Logout request failed:', err)
        } finally {
            setUser(null)
        }
    }

    return (
        <AuthContext.Provider value={{ user, loading, mode, login, loginWithCredentials, loginError, logout, refresh }}>
            {children}
        </AuthContext.Provider>
    )
}

export const useAuth = () => useContext(AuthContext)
