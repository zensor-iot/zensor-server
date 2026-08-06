const BASE_PATH = '/v1/admin/api-keys'

const STATUS_CODES = {
    400: 'name_required',
    404: 'not_found',
    409: 'duplicate_name',
}

export class ApiKeyError extends Error {
    constructor(code, message) {
        super(message)
        this.name = 'ApiKeyError'
        this.code = code
    }
}

function errorForStatus(status) {
    const code = STATUS_CODES[status] || 'request_failed'
    return new ApiKeyError(code, `api key request failed with status ${status}`)
}

async function request(url, ...options) {
    let response
    try {
        response = await fetch(url, ...options)
    } catch (error) {
        throw new ApiKeyError('request_failed', `api key request failed: ${error.message}`)
    }

    if (!response.ok) {
        throw errorForStatus(response.status)
    }

    return response
}

export async function listAPIKeys() {
    const response = await request(BASE_PATH)
    return response.json()
}

export async function createAPIKey(name) {
    const response = await request(BASE_PATH, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
    })
    return response.json()
}

export async function revokeAPIKey(id) {
    await request(`${BASE_PATH}/${id}`, { method: 'DELETE' })
}
