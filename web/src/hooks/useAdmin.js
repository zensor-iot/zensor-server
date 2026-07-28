import { useState, useEffect } from 'react'

export const useAdmin = () => {
    const [isAdmin, setIsAdmin] = useState(false)
    const [isLoading, setIsLoading] = useState(true)
    const [userInfo, setUserInfo] = useState(null)

    useEffect(() => {
        const checkAdminStatus = async () => {
            try {
                const response = await fetch('/api/user')
                if (response.ok) {
                    const user = await response.json()
                    setUserInfo(user)
                    setIsAdmin(user.isAdmin || false)
                } else {
                    setIsAdmin(false)
                }
            } catch (error) {
                console.error('Failed to check admin status:', error)
                setIsAdmin(false)
            } finally {
                setIsLoading(false)
            }
        }

        checkAdminStatus()
    }, [])

    return { isAdmin, isLoading, userInfo }
}
