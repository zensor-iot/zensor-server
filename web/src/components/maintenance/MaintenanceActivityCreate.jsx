import { useParams, useNavigate, Link } from 'react-router-dom'
import { Wrench, ArrowLeft } from 'lucide-react'
import ActivityForm from './ActivityForm'
import './Maintenance.css'

const MaintenanceActivityCreate = () => {
    const { tenantId } = useParams()
    const navigate = useNavigate()

    return (
        <div className="maintenance-page">
            <div className="maintenance-header">
                <h2>
                    <Wrench size={22} />
                    New Maintenance Activity
                </h2>
                <Link to={`/portal/${tenantId}/maintenance`} className="maintenance-btn">
                    <ArrowLeft size={16} />
                    Back to activities
                </Link>
            </div>

            <ActivityForm
                tenantId={tenantId}
                onSaved={() => navigate(`/portal/${tenantId}/maintenance`)}
                onCancel={() => navigate(`/portal/${tenantId}/maintenance`)}
            />
        </div>
    )
}

export default MaintenanceActivityCreate
