import { useState, useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { User, LogOut, UserCircle } from 'lucide-react'

const UserInfo = () => {
    const [userInfo, setUserInfo] = useState(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [showMenu, setShowMenu] = useState(false)
    const menuRef = useRef(null)

    useEffect(() => {
        const fetchUserInfo = async () => {
            try {
                console.log('🔍 Fetching user info...')
                const response = await fetch('/api/user')
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`)
                }
                const data = await response.json()
                console.log('👤 User info received:', data)
                setUserInfo(data)
            } catch (err) {
                console.error('❌ Failed to fetch user info:', err)
                setError(err.message)
            } finally {
                setLoading(false)
            }
        }

        fetchUserInfo()
    }, [])

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (menuRef.current && !menuRef.current.contains(event.target)) {
                setShowMenu(false)
            }
        }

        if (showMenu) {
            document.addEventListener('mousedown', handleClickOutside)
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside)
        }
    }, [showMenu])

    const handleUserClick = () => {
        setShowMenu(!showMenu)
    }

    const handleLogout = () => {
        // TODO: Implement logout functionality
        console.log('Logout clicked')
        setShowMenu(false)
    }

    if (loading) {
        console.log('⏳ UserInfo: Loading...')
        return (
            <div className="user-info">
                <User size={20} />
                <span>Loading...</span>
            </div>
        )
    }

    if (error) {
        console.log('❌ UserInfo: Error state -', error)
        return (
            <div className="user-info">
                <User size={20} />
                <span>Error</span>
            </div>
        )
    }

    const hasUserData = userInfo && (userInfo.name || userInfo.email || userInfo.user)

    if (!hasUserData) {
        console.log('👻 UserInfo: No authentication - showing Guest')
        return (
            <div className="user-info">
                <User size={20} />
                <span className="user-name">Guest</span>
            </div>
        )
    }

    const displayEmail = userInfo.email || 'No email'
    console.log('✅ UserInfo: Authenticated user -', displayEmail)

    return (
        <div className="user-info-container" ref={menuRef}>
            <div
                className="user-info clickable"
                onClick={handleUserClick}
                title={`${userInfo.name || 'Unknown'} (${userInfo.email || 'No email'})`}
            >
                <User size={20} />
                <span className="user-email">{displayEmail}</span>
            </div>

            {showMenu && (
                <div className="user-menu">
                    <Link
                        to="/profile"
                        className="user-menu-item"
                        onClick={() => setShowMenu(false)}
                    >
                        <UserCircle size={16} />
                        Profile
                    </Link>
                    <button
                        className="user-menu-item"
                        onClick={handleLogout}
                    >
                        <LogOut size={16} />
                        Logout
                    </button>
                </div>
            )}
        </div>
    )
}

export default UserInfo 