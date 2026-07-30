import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Building, Loader2, AlertCircle, ChevronRight } from 'lucide-react'
import { getApiUrl } from '../config/api'

const TenantList = () => {
  const [tenants, setTenants] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    fetchTenants()
  }, [])

  const fetchTenants = async () => {
    try {
      setLoading(true)
      const response = await fetch(getApiUrl('/tenants'))
      if (!response.ok) {
        throw new Error(`Failed to fetch tenants: ${response.status}`)
      }
      const data = await response.json()
      setTenants(data.data || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (error) {
    return (
      <div className="page-container">
        <div className="error-container">
          <AlertCircle className="error-icon" />
          <h2>Unable to load tenants</h2>
          <p className="error-message">{error}</p>
          <button onClick={fetchTenants} className="retry-button">
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <Building className="page-icon" />
        <div>
          <h2>Tenant Organizations</h2>
          <p className="page-description">
            Select a tenant to manage their devices and infrastructure
          </p>
        </div>
      </div>

      {loading ? (
        <div className="loading-container">
          <Loader2 className="loading-spinner" />
          <p>Loading tenants...</p>
        </div>
      ) : tenants.length === 0 ? (
        <div className="empty-state">
          <Building size={64} className="empty-icon" />
          <h3>No tenants found</h3>
          <p>Contact your administrator to set up tenant organizations</p>
        </div>
      ) : (
        <div className="tenant-grid">
          {tenants.map((tenant) => (
            <Link
              key={tenant.id}
              to={`/portal/${tenant.id}`}
              className="tenant-card"
            >
              <div className="tenant-card-header">
                <h3 className="tenant-name">{tenant.name}</h3>
                <p className="tenant-email">{tenant.email}</p>
              </div>
              {tenant.description && (
                <p className="tenant-description">{tenant.description}</p>
              )}
              <div className="tenant-card-footer">
                <div className="tenant-meta">
                  <span className="status-badge active">Active</span>
                </div>
                <ChevronRight className="nav-arrow" />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

export default TenantList 