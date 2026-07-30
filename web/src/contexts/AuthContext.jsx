import { createContext, useContext, useEffect, useState, useCallback } from 'react'

const AuthContext = createContext(null)

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null)
    const [loading, setLoading] = useState(true)

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
        <AuthContext.Provider value={{ user, loading, login, logout, refresh }}>
            {children}
        </AuthContext.Provider>
    )
}

export const useAuth = () => useContext(AuthContext)
