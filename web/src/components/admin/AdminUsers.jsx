import { useState, useEffect } from 'react'
import { Users, Plus, Trash2, Shield, ShieldOff } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import { useAuth } from '../../contexts/AuthContext'
import './AdminUsers.css'

const AdminUsers = () => {
    const { showSuccess, showError } = useNotification()
    const { user: currentUser, mode } = useAuth()
    const [users, setUsers] = useState([])
    const [loading, setLoading] = useState(true)
    const [showCreateForm, setShowCreateForm] = useState(false)
    const [formData, setFormData] = useState({ email: '', is_admin: false })

    useEffect(() => {
        if (mode === 'static') {
            setLoading(false)
            return
        }
        fetchUsers()
    }, [mode])

    const fetchUsers = async () => {
        try {
            setLoading(true)
            const response = await fetch('/v1/admin/allowed-users')
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`)
            }
            setUsers(await response.json())
        } catch (error) {
            console.error('Failed to fetch allowed users:', error)
            showError('Failed to load allowed users', 'Error')
            setUsers([])
        } finally {
            setLoading(false)
        }
    }

    const handleCreate = async (event) => {
        event.preventDefault()
        try {
            const response = await fetch('/v1/admin/allowed-users', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData)
            })

            if (response.status === 409) {
                showError('This email is already on the allowlist', 'Duplicate')
                return
            }
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`)
            }

            showSuccess('User added to the allowlist', 'Success')
            setFormData({ email: '', is_admin: false })
            setShowCreateForm(false)
            fetchUsers()
        } catch (error) {
            console.error('Failed to add allowed user:', error)
            showError('Failed to add user — check the email address', 'Error')
        }
    }

    const handleToggleAdmin = async (user) => {
        try {
            const response = await fetch(`/v1/admin/allowed-users/${user.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ is_admin: !user.is_admin })
            })
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`)
            }
            fetchUsers()
        } catch (error) {
            console.error('Failed to update allowed user:', error)
            showError('Failed to update user', 'Error')
        }
    }

    const handleDelete = async (user) => {
        if (!window.confirm(`Remove ${user.email} from the allowlist? Their sessions will be revoked immediately.`)) {
            return
        }

        try {
            const response = await fetch(`/v1/admin/allowed-users/${user.id}`, {
                method: 'DELETE'
            })
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`)
            }
            showSuccess('User removed from the allowlist', 'Success')
            fetchUsers()
        } catch (error) {
            console.error('Failed to remove allowed user:', error)
            showError('Failed to remove user', 'Error')
        }
    }

    const formatLastLogin = (value) => {
        if (!value) return 'Never'
        return new Date(value).toLocaleString()
    }

    return (
        <div className="admin-users">
            <div className="admin-header">
                <div className="admin-title">
                    <Users size={32} />
                    <h1>User Access</h1>
                </div>
                <p className="admin-subtitle">Control which Google accounts can sign in to the portal</p>
            </div>

            {mode === 'static' && (
                <div className="static-mode-notice">
                    User access management is unavailable in static auth mode. Switch <code>auth.mode</code> to
                    "google" in the server configuration to manage an allowlist.
                </div>
            )}

            {mode !== 'static' && (
                <div className="admin-actions">
                    <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
                        <Plus size={16} />
                        Add User
                    </button>
                </div>
            )}

            {showCreateForm && (
                <div className="form-overlay">
                    <div className="form-card">
                        <h2>Add Allowed User</h2>
                        <form onSubmit={handleCreate}>
                            <div className="form-group">
                                <label htmlFor="email">Email</label>
                                <input
                                    id="email"
                                    type="email"
                                    required
                                    placeholder="person@example.com"
                                    value={formData.email}
                                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                />
                            </div>
                            <div className="form-group checkbox-group">
                                <label>
                                    <input
                                        type="checkbox"
                                        checked={formData.is_admin}
                                        onChange={(e) => setFormData({ ...formData, is_admin: e.target.checked })}
                                    />
                                    Administrator
                                </label>
                            </div>
                            <div className="form-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowCreateForm(false)}>
                                    Cancel
                                </button>
                                <button type="submit" className="btn btn-primary">
                                    Add User
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {mode !== 'static' && (loading ? (
                <div className="loading">Loading allowed users...</div>
            ) : (
                <div className="users-table-wrapper">
                    <table className="users-table">
                        <thead>
                            <tr>
                                <th>Email</th>
                                <th>Name</th>
                                <th>Role</th>
                                <th>Last Login</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {users.map((user) => (
                                <tr key={user.id}>
                                    <td>{user.email}</td>
                                    <td>{user.display_name || '—'}</td>
                                    <td>
                                        <span className={`role-badge ${user.is_admin ? 'admin' : 'member'}`}>
                                            {user.is_admin ? 'Admin' : 'Member'}
                                        </span>
                                    </td>
                                    <td>{formatLastLogin(user.last_login_at)}</td>
                                    <td>
                                        <div className="user-actions">
                                            <button
                                                className="btn-icon"
                                                onClick={() => handleToggleAdmin(user)}
                                                disabled={currentUser?.user_id === user.id}
                                                title={user.is_admin ? 'Revoke admin' : 'Make admin'}
                                            >
                                                {user.is_admin ? <ShieldOff size={16} /> : <Shield size={16} />}
                                            </button>
                                            <button
                                                className="btn-icon danger"
                                                onClick={() => handleDelete(user)}
                                                disabled={currentUser?.user_id === user.id}
                                                title="Remove from allowlist"
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            {users.length === 0 && (
                                <tr>
                                    <td colSpan="5" className="empty-state">No allowed users yet</td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            ))}
        </div>
    )
}

export default AdminUsers
