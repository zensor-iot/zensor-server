import { useState, useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { User, LogOut, UserCircle } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

const UserInfo = () => {
    const { user, logout } = useAuth()
    const [showMenu, setShowMenu] = useState(false)
    const menuRef = useRef(null)

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

    const handleLogout = async () => {
        setShowMenu(false)
        await logout()
    }

    if (!user) {
        return (
            <div className="user-info">
                <User size={20} />
                <span className="user-name">Guest</span>
            </div>
        )
    }

    const displayEmail = user.email || 'No email'

    return (
        <div className="user-info-container" ref={menuRef}>
            <div
                className="user-info clickable"
                onClick={handleUserClick}
                title={`${user.name || 'Unknown'} (${user.email || 'No email'})`}
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
