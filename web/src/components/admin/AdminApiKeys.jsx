import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { KeyRound, Plus, Trash2, ArrowLeft, AlertTriangle, Copy, Check } from 'lucide-react'
import { useNotification } from '../../hooks/useNotification'
import { createAPIKey, listAPIKeys, revokeAPIKey } from '../../utils/apiKeys'
import './AdminApiKeys.css'

const createErrorMessages = {
    duplicate_name: 'An API key with that name already exists',
    name_required: 'Name is required',
}

const AdminApiKeys = () => {
    const { showSuccess, showError } = useNotification()
    const [apiKeys, setApiKeys] = useState([])
    const [loading, setLoading] = useState(true)
    const [showCreateForm, setShowCreateForm] = useState(false)
    const [creating, setCreating] = useState(false)
    const [name, setName] = useState('')
    const [createdKey, setCreatedKey] = useState(null)
    const [copied, setCopied] = useState(false)

    useEffect(() => {
        fetchApiKeys()
    }, [])

    const fetchApiKeys = async () => {
        try {
            setLoading(true)
            setApiKeys(await listAPIKeys())
        } catch (error) {
            console.error('Failed to fetch api keys:', error)
            showError('Failed to load API keys', 'Error')
            setApiKeys([])
        } finally {
            setLoading(false)
        }
    }

    const handleCreate = async (event) => {
        event.preventDefault()
        try {
            setCreating(true)
            const created = await createAPIKey(name.trim())
            setShowCreateForm(false)
            setName('')
            setCreatedKey(created)
        } catch (error) {
            console.error('Failed to create api key:', error)
            showError(createErrorMessages[error.code] || 'Failed to create API key', 'Error')
        } finally {
            setCreating(false)
        }
    }

    const handleCopy = async () => {
        try {
            await copyToClipboard(createdKey.key)
            setCopied(true)
        } catch (error) {
            console.error('Failed to copy api key:', error)
            showError('Failed to copy — select the key and copy it manually', 'Error')
        }
    }

    const handleAcknowledge = () => {
        setCreatedKey(null)
        setCopied(false)
        showSuccess('API key created', 'Success')
        fetchApiKeys()
    }

    const handleRevoke = async (apiKey) => {
        if (!window.confirm(`Revoke "${apiKey.name}"? Any client using it will immediately start getting 401s.`)) {
            return
        }

        try {
            await revokeAPIKey(apiKey.id)
            showSuccess('API key revoked', 'Success')
            fetchApiKeys()
        } catch (error) {
            console.error('Failed to revoke api key:', error)
            showError('Failed to revoke API key', 'Error')
        }
    }

    const formatCreatedAt = (value) => {
        if (!value) return '—'
        return new Date(value).toLocaleString()
    }

    return (
        <div className="admin-api-keys">
            <div className="admin-header">
                <div className="header-top">
                    <Link to="/admin" className="back-link">
                        <ArrowLeft size={16} />
                        Back to Dashboard
                    </Link>
                </div>
                <div className="admin-title">
                    <KeyRound size={32} />
                    <h1>API Keys</h1>
                </div>
                <p className="admin-subtitle">Grant non-human clients access to the API</p>
            </div>

            <div className="admin-actions">
                <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
                    <Plus size={16} />
                    New API Key
                </button>
            </div>

            {showCreateForm && (
                <div className="form-overlay">
                    <div className="form-card">
                        <h2>Create API Key</h2>
                        <form onSubmit={handleCreate}>
                            <div className="form-group">
                                <label htmlFor="name">Name</label>
                                <input
                                    id="name"
                                    type="text"
                                    required
                                    autoFocus
                                    placeholder="grafana-sync"
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                />
                                <p className="form-hint">A label to identify this key later. It cannot be changed.</p>
                            </div>
                            <div className="form-actions">
                                <button
                                    type="button"
                                    className="btn btn-secondary"
                                    onClick={() => setShowCreateForm(false)}
                                    disabled={creating}
                                >
                                    Cancel
                                </button>
                                <button type="submit" className="btn btn-primary" disabled={creating}>
                                    {creating ? 'Creating...' : 'Create Key'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {createdKey && (
                <div className="form-overlay">
                    <div className="form-card reveal-card">
                        <div className="reveal-title">
                            <AlertTriangle size={20} />
                            <h2>Copy your API key now</h2>
                        </div>
                        <p className="reveal-warning">
                            This is the only time it will be shown. Once you close this dialog it cannot be
                            retrieved — you would have to create a new key.
                        </p>
                        <div className="reveal-key">
                            <input type="text" readOnly value={createdKey.key} onFocus={(e) => e.target.select()} />
                            <button type="button" className="btn btn-secondary" onClick={handleCopy}>
                                {copied ? <Check size={16} /> : <Copy size={16} />}
                                {copied ? 'Copied' : 'Copy'}
                            </button>
                        </div>
                        <dl className="reveal-meta">
                            <dt>Name</dt>
                            <dd>{createdKey.name}</dd>
                            <dt>Usage</dt>
                            <dd><code>Authorization: Bearer &lt;key&gt;</code></dd>
                        </dl>
                        <div className="form-actions">
                            <button type="button" className="btn btn-primary" onClick={handleAcknowledge}>
                                I&apos;ve saved it — close
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {loading ? (
                <div className="loading">Loading API keys...</div>
            ) : (
                <div className="api-keys-table-wrapper">
                    <table className="api-keys-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Key Prefix</th>
                                <th>Created</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {apiKeys.map((apiKey) => (
                                <tr key={apiKey.id}>
                                    <td>{apiKey.name}</td>
                                    <td><code className="key-prefix">{apiKey.key_prefix}…</code></td>
                                    <td>{formatCreatedAt(apiKey.created_at)}</td>
                                    <td>
                                        <div className="api-key-actions">
                                            <button
                                                className="btn-icon danger"
                                                onClick={() => handleRevoke(apiKey)}
                                                title="Revoke key"
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            {apiKeys.length === 0 && (
                                <tr>
                                    <td colSpan="4" className="empty-state">No API keys yet</td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}

async function copyToClipboard(value) {
    if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(value)
        return
    }

    const textarea = document.createElement('textarea')
    textarea.value = value
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'absolute'
    textarea.style.left = '-9999px'
    document.body.appendChild(textarea)
    textarea.select()

    try {
        if (!document.execCommand('copy')) {
            throw new Error('copy command was rejected')
        }
    } finally {
        document.body.removeChild(textarea)
    }
}

export default AdminApiKeys
