import { useState, useEffect } from 'react'
import { User, Mail, Shield, UserCircle, Clock, Settings } from 'lucide-react'
import { useNotification } from '../hooks/useNotification'
import TimezoneDropdown from './TimezoneDropdown'

const Profile = () => {
    const { showSuccess, showError, showApiNotification } = useNotification()
    const [userInfo, setUserInfo] = useState(null)
    const [authorizedTenants, setAuthorizedTenants] = useState([])
    const [selectedTenantId, setSelectedTenantId] = useState(null)
    const [tenantConfig, setTenantConfig] = useState(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [isEditingConfig, setIsEditingConfig] = useState(false)
    const [configForm, setConfigForm] = useState({
        timezone: '',
        notification_email: ''
    })

    useEffect(() => {
        const fetchData = async () => {
            try {
                console.log('🔍 Fetching user info for profile...')
                const response = await fetch('/api/user')
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`)
                }
                const data = await response.json()
                console.log('👤 User info received for profile:', data)
                setUserInfo(data)

                // If we have an email, fetch user details with authorized tenants
                if (data.email) {
                    const userDetails = await fetchUserDetails(data.email)
                    if (userDetails && userDetails.tenants && userDetails.tenants.length > 0) {
                        // Fetch details for each authorized tenant
                        const tenantPromises = userDetails.tenants.map(tenantId => fetchTenantDetails(tenantId))
                        const tenantDetails = await Promise.all(tenantPromises)
                        const validTenants = tenantDetails.filter(tenant => tenant !== null)

                        setAuthorizedTenants(validTenants)

                        // Select the first tenant by default
                        if (validTenants.length > 0) {
                            setSelectedTenantId(validTenants[0].id)
                            await fetchTenantConfig(validTenants[0].id)
                        }
                    } else {
                        // Fallback to default tenant if no authorized tenants
                        const defaultTenantId = '550e8400-e29b-41d4-a716-446655440001'
                        setSelectedTenantId(defaultTenantId)
                        await fetchTenantConfig(defaultTenantId)
                    }
                } else {
                    // Fallback to default tenant if no email
                    const defaultTenantId = '550e8400-e29b-41d4-a716-446655440001'
                    setSelectedTenantId(defaultTenantId)
                    await fetchTenantConfig(defaultTenantId)
                }
            } catch (err) {
                console.error('❌ Failed to fetch user info:', err)
                setError(err.message)
            } finally {
                setLoading(false)
            }
        }

        fetchData()
    }, []) // eslint-disable-line react-hooks/exhaustive-deps

    const fetchUserDetails = async (email) => {
        try {
            console.log(`🔍 Fetching user details for email: ${email}`)
            const response = await fetch(`/api/users/${email}`)
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`)
            }
            const userDetails = await response.json()
            console.log('👤 User details received:', userDetails)
            return userDetails
        } catch (err) {
            console.error('❌ Failed to fetch user details:', err)
            showError(`Failed to fetch user details: ${err.message}`, 'Error')
            return null
        }
    }

    const fetchTenantDetails = async (tenantId) => {
        try {
            console.log(`🔍 Fetching tenant details for tenant: ${tenantId}`)
            const response = await fetch(`/api/tenants/${tenantId}`)
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`)
            }
            const tenantDetails = await response.json()
            console.log('🏢 Tenant details received:', tenantDetails)
            return tenantDetails
        } catch (err) {
            console.error('❌ Failed to fetch tenant details:', err)
            showError(`Failed to fetch tenant details: ${err.message}`, 'Error')
            return null
        }
    }

    const fetchTenantConfig = async (tenantId) => {
        try {
            const response = await fetch(`/api/tenants/${tenantId}/configuration`)
            if (!response.ok && response.status !== 404) {
                throw new Error(`HTTP ${response.status}`)
            }

            if (response.ok) {
                const data = await response.json()
                console.log('⚙️ Configuration received:', data)
                setTenantConfig(data)
                setConfigForm({
                    timezone: data.timezone || '',
                    notification_email: data.notification_email || ''
                })
            }
        } catch (err) {
            console.error('❌ Failed to fetch configuration:', err)
            // Configuration might not exist yet, that's ok
        }
    }

    const handleConfigEdit = () => {
        setIsEditingConfig(true)
    }

    const handleConfigCancel = () => {
        setIsEditingConfig(false)
        // Reset form to original values
        if (tenantConfig) {
            setConfigForm({
                timezone: tenantConfig.timezone || '',
                notification_email: tenantConfig.notification_email || ''
            })
        }
    }

    const handleTenantSelect = async (tenantId) => {
        setSelectedTenantId(tenantId)
        setIsEditingConfig(false)
        await fetchTenantConfig(tenantId)
    }

    const handleConfigSubmit = async (e) => {
        e.preventDefault()

        if (!selectedTenantId) {
            showError('No tenant selected', 'Error')
            return
        }

        const result = await showApiNotification(
            fetch(`/api/tenants/${selectedTenantId}/configuration`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(configForm)
            }).then(res => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`)
                return res.json()
            }),
            'Configuration updated successfully',
            'Failed to update configuration',
            'Success',
            'Error'
        )

        if (result.success) {
            setTenantConfig(result.data)
            setIsEditingConfig(false)
        }
    }

    const handleConfigInputChange = (e) => {
        const { name, value } = e.target
        setConfigForm(prev => ({
            ...prev,
            [name]: value
        }))
    }

    if (loading) {
        return (
            <div className="profile-container">
                <div className="page-header">
                    <h2>
                        <UserCircle className="page-icon" />
                        Profile
                    </h2>
                    <p className="page-description">Loading your profile information...</p>
                </div>
                <div className="loading-container">
                    <div className="loading-spinner"></div>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="profile-container">
                <div className="page-header">
                    <h2>
                        <UserCircle className="page-icon" />
                        Profile
                    </h2>
                    <p className="page-description">Error loading profile</p>
                </div>
                <div className="error-container">
                    <p className="error-message">Failed to load profile: {error}</p>
                </div>
            </div>
        )
    }

    const hasUserData = userInfo && (userInfo.name || userInfo.email || userInfo.user)

    if (!hasUserData) {
        return (
            <div className="profile-container">
                <div className="page-header">
                    <h2>
                        <UserCircle className="page-icon" />
                        Profile
                    </h2>
                    <p className="page-description">No user information available</p>
                </div>
                <div className="empty-container">
                    <UserCircle size={48} className="empty-icon" />
                    <p className="empty-message">You are not authenticated</p>
                </div>
            </div>
        )
    }

    const role = userInfo.isAdmin ? 'Admin' : 'User'

    return (
        <div className="profile-container">
            <div className="page-header">
                <h2>
                    <UserCircle className="page-icon" />
                    Profile
                </h2>
                <p className="page-description">Your account information</p>
            </div>

            <div className="profile-card">
                <div className="profile-header">
                    <div className="profile-avatar">
                        <User size={32} />
                    </div>
                    <div className="profile-title">
                        <h3>{userInfo.name || 'Unknown Name'}</h3>
                        <p className="profile-subtitle">{role}</p>
                    </div>
                </div>

                <div className="profile-details">
                    <div className="profile-field">
                        <div className="field-label">
                            <Mail size={16} />
                            Email
                        </div>
                        <div className="field-value">
                            {userInfo.email || 'No email provided'}
                        </div>
                    </div>

                    <div className="profile-field">
                        <div className="field-label">
                            <User size={16} />
                            Name
                        </div>
                        <div className="field-value">
                            {userInfo.name || 'No name provided'}
                        </div>
                    </div>

                    <div className="profile-field">
                        <div className="field-label">
                            <Shield size={16} />
                            Role
                        </div>
                        <div className="field-value">
                            <span className={`role-badge ${userInfo.isAdmin ? 'admin' : 'user'}`}>
                                {role}
                            </span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Configuration Section */}
            <div className="profile-card">
                <div className="profile-section-header">
                    <Settings size={20} />
                    <h3>Configuration</h3>
                </div>

                {/* Tenant Selector */}
                {authorizedTenants.length > 0 && (
                    <div className="profile-details">
                        <div className="profile-field">
                            <div className="field-label">
                                <Shield size={16} />
                                Tenant
                            </div>
                            <div className="field-value">
                                <select
                                    value={selectedTenantId || ''}
                                    onChange={(e) => handleTenantSelect(e.target.value)}
                                    className="form-input"
                                    style={{ marginTop: '8px' }}
                                >
                                    {authorizedTenants.map(tenant => (
                                        <option key={tenant.id} value={tenant.id}>
                                            {tenant.name}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </div>
                )}

                {!isEditingConfig ? (
                    <div className="profile-details">
                        <div className="profile-field">
                            <div className="field-label">
                                <Clock size={16} />
                                Timezone
                            </div>
                            <div className="field-value">
                                {tenantConfig?.timezone || 'Not set'}
                            </div>
                        </div>

                        <div className="profile-field">
                            <div className="field-label">
                                <Mail size={16} />
                                Notification Email
                            </div>
                            <div className="field-value">
                                {tenantConfig?.notification_email || 'Not set'}
                            </div>
                        </div>

                        <div className="profile-actions">
                            <button
                                onClick={handleConfigEdit}
                                className="btn-primary"
                            >
                                Edit Configuration
                            </button>
                        </div>
                    </div>
                ) : (
                    <form onSubmit={handleConfigSubmit} className="profile-form">
                        <div className="form-field">
                            <label htmlFor="timezone">
                                <Clock size={16} />
                                Timezone
                            </label>
                            <TimezoneDropdown
                                value={configForm.timezone}
                                onChange={(value) => setConfigForm(prev => ({ ...prev, timezone: value }))}
                                placeholder="Select a timezone..."
                            />
                        </div>

                        <div className="form-field">
                            <label htmlFor="notification_email">
                                <Mail size={16} />
                                Notification Email
                            </label>
                            <input
                                type="email"
                                id="notification_email"
                                name="notification_email"
                                value={configForm.notification_email}
                                onChange={handleConfigInputChange}
                                placeholder="e.g., notifications@example.com"
                                className="form-input"
                            />
                        </div>

                        <div className="profile-actions">
                            <button
                                type="submit"
                                className="btn-primary"
                            >
                                Save Changes
                            </button>
                            <button
                                type="button"
                                onClick={handleConfigCancel}
                                className="btn-secondary"
                            >
                                Cancel
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    )
}

export default Profile
