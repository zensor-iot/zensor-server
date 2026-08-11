// API configuration for the embedded SPA, served same-origin by the Go server (no proxy involved).
const config = {
    // API base URL - relative path, same-origin with the Go server that serves this SPA
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/v1',

    // Grafana base URL
    grafanaBaseUrl: import.meta.env.VITE_GRAFANA_BASE_URL || 'https://cardamomo.zensor-iot.net',

    // Grafana API Key for authentication (still client-side for Grafana)
    grafanaApiKey: import.meta.env.VITE_GRAFANA_API_KEY || '',

    // WebSocket base URL - same-origin with the Go server
    get wsBaseUrl() {
        if (typeof window !== 'undefined') {
            const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
            return `${wsProtocol}//${window.location.host}`
        }
        return 'ws://localhost:5173'
    },

    // API endpoints (relative to base URL)
    endpoints: {
        tenants: '/tenants',
        devices: '/devices',
        websocket: '/ws/device-messages',
        scheduledTasks: '/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks'
    }
}

// Helper functions for building URLs
export const getApiUrl = (path = '') => {
    return `${config.apiBaseUrl}${path}`
}

export const getWebSocketUrl = (path = '') => {
    return `${config.wsBaseUrl}${path}`
}

export const buildApiEndpoint = (endpoint, ...params) => {
    let path = config.endpoints[endpoint]
    if (!path) {
        throw new Error(`Unknown endpoint: ${endpoint}`)
    }

    // Replace parameters in the path
    params.forEach(param => {
        path = path.replace(/\{[^}]+\}/, param)
    })

    return getApiUrl(path)
}

// Scheduled Tasks API functions
export const scheduledTasksApi = {
    // Get all scheduled tasks for a device
    async getScheduledTasks(tenantId, deviceId, page = 1, limit = 10) {
        const url = buildApiEndpoint('scheduledTasks', tenantId, deviceId)
        const fullUrl = `${url}?page=${page}&limit=${limit}`
        console.log('🌐 API call to:', fullUrl)

        const response = await fetch(fullUrl)

        if (!response.ok) {
            throw new Error(`Failed to fetch scheduled tasks: ${response.status}`)
        }

        const data = await response.json()
        console.log('🌐 API response:', data)
        return data
    },

    // Create a new scheduled task
    async createScheduledTask(tenantId, deviceId, taskData) {
        const url = buildApiEndpoint('scheduledTasks', tenantId, deviceId)
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(taskData)
        })

        if (!response.ok) {
            throw new Error(`Failed to create scheduled task: ${response.status}`)
        }

        return response.json()
    },

    // Update an existing scheduled task
    async updateScheduledTask(tenantId, deviceId, taskId, taskData) {
        const url = buildApiEndpoint('scheduledTasks', tenantId, deviceId) + `/${taskId}`
        const response = await fetch(url, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(taskData)
        })

        if (!response.ok) {
            throw new Error(`Failed to update scheduled task: ${response.status}`)
        }

        return response.json()
    },

    // Delete a scheduled task
    async deleteScheduledTask(tenantId, deviceId, taskId) {
        const url = buildApiEndpoint('scheduledTasks', tenantId, deviceId) + `/${taskId}`
        const response = await fetch(url, {
            method: 'DELETE'
        })

        if (!response.ok) {
            throw new Error(`Failed to delete scheduled task: ${response.status}`)
        }
    },

    // Get a specific scheduled task
    async getScheduledTask(tenantId, deviceId, taskId) {
        const url = buildApiEndpoint('scheduledTasks', tenantId, deviceId) + `/${taskId}`
        const response = await fetch(url)

        if (!response.ok) {
            throw new Error(`Failed to fetch scheduled task: ${response.status}`)
        }

        return response.json()
    },

    // Get task executions for a scheduled task
    async getTaskExecutions(tenantId, deviceId, scheduledTaskId, limit = 3) {
        const url = buildApiEndpoint('scheduledTasks', tenantId, deviceId) + `/${scheduledTaskId}/tasks`
        const response = await fetch(`${url}?limit=${limit}`)

        if (!response.ok) {
            throw new Error(`Failed to fetch task executions: ${response.status}`)
        }

        return response.json()
    }
}

// Maintenance API functions
export const maintenanceApi = {
    async listActivities(tenantId, page = 1, limit = 10) {
        const response = await fetch(getApiUrl(`/maintenance/activities?tenant_id=${tenantId}&page=${page}&limit=${limit}`))
        if (!response.ok) {
            throw new Error(`Failed to fetch maintenance activities: ${response.status}`)
        }
        return response.json()
    },

    // Pages through every activity for the tenant (used by Up Next).
    async listAllActivities(tenantId) {
        const limit = 50
        let page = 1
        const activities = []
        for (;;) {
            const data = await this.listActivities(tenantId, page, limit)
            const items = data.data || []
            activities.push(...items)
            const total = data.pagination?.total ?? activities.length
            if (activities.length >= total || items.length === 0) {
                return activities
            }
            page++
        }
    },

    async getActivity(activityId) {
        const response = await fetch(getApiUrl(`/maintenance/activities/${activityId}`))
        if (!response.ok) {
            throw new Error(`Failed to fetch maintenance activity: ${response.status}`)
        }
        return response.json()
    },

    async createActivity(activityData) {
        const response = await fetch(getApiUrl('/maintenance/activities'), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(activityData)
        })
        if (!response.ok) {
            throw new Error(`Failed to create maintenance activity: ${response.status}`)
        }
        return response.json()
    },

    async updateActivity(activityId, updates) {
        const response = await fetch(getApiUrl(`/maintenance/activities/${activityId}`), {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updates)
        })
        if (!response.ok) {
            throw new Error(`Failed to update maintenance activity: ${response.status}`)
        }
        return response.json()
    },

    async deleteActivity(activityId) {
        const response = await fetch(getApiUrl(`/maintenance/activities/${activityId}`), {
            method: 'DELETE'
        })
        if (!response.ok) {
            throw new Error(`Failed to delete maintenance activity: ${response.status}`)
        }
    },

    async activateActivity(activityId) {
        const response = await fetch(getApiUrl(`/maintenance/activities/${activityId}/activate`), {
            method: 'POST'
        })
        if (!response.ok) {
            throw new Error(`Failed to activate maintenance activity: ${response.status}`)
        }
    },

    async deactivateActivity(activityId) {
        const response = await fetch(getApiUrl(`/maintenance/activities/${activityId}/deactivate`), {
            method: 'POST'
        })
        if (!response.ok) {
            throw new Error(`Failed to deactivate maintenance activity: ${response.status}`)
        }
    },

    async listExecutions(activityId, page = 1, limit = 50) {
        const response = await fetch(getApiUrl(`/maintenance/executions?activity_id=${activityId}&page=${page}&limit=${limit}`))
        if (!response.ok) {
            throw new Error(`Failed to fetch maintenance executions: ${response.status}`)
        }
        return response.json()
    },

    async getExecution(executionId) {
        const response = await fetch(getApiUrl(`/maintenance/executions/${executionId}`))
        if (!response.ok) {
            throw new Error(`Failed to fetch maintenance execution: ${response.status}`)
        }
        return response.json()
    },

    async completeExecution(executionId, completedBy, fieldValues = null) {
        const body = { completed_by: completedBy }
        if (fieldValues && Object.keys(fieldValues).length > 0) {
            body.field_values = fieldValues
        }
        const response = await fetch(getApiUrl(`/maintenance/executions/${executionId}/complete`), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        })
        if (!response.ok) {
            throw new Error(`Failed to complete maintenance execution: ${response.status}`)
        }
    }
}

// Device Commands API functions
export const deviceCommandsApi = {
    // Send a command to a device
    async sendCommand(deviceId, commandData) {
        const url = getApiUrl(`/devices/${deviceId}/commands`)
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(commandData)
        })

        if (!response.ok) {
            throw new Error(`Failed to send command: ${response.status}`)
        }

        return response.json()
    }
}

// VictoriaMetrics query API (reverse-proxied by the Go server at /v1/metrics/*)
export const metricsApi = {
    // Query a range of samples for a metric name using the PromQL query_range
    // endpoint. Resolves to [{ time, value }, ...] sorted chronologically,
    // merging every returned time series.
    async queryRange(metricName, { start, end, step = '60s' } = {}) {
        const now = Date.now()
        const from = start ?? now - 6 * 60 * 60 * 1000
        const to = end ?? now
        const params = new URLSearchParams({
            query: metricName,
            start: String(Math.floor(from / 1000)),
            end: String(Math.floor(to / 1000)),
            step,
        })
        const response = await fetch(`${getApiUrl('/metrics/query_range')}?${params}`)
        if (!response.ok) {
            throw new Error(`Failed to query metrics: ${response.status}`)
        }
        const payload = await response.json()
        const points = []
        for (const series of payload?.data?.result ?? []) {
            for (const [timestamp, value] of series.values ?? []) {
                const parsed = parseFloat(value)
                if (Number.isFinite(parsed)) {
                    points.push({ time: parseInt(timestamp, 10) * 1000, value: parsed })
                }
            }
        }
        points.sort((a, b) => a.time - b.time)
        return points
    },
}

export default config 