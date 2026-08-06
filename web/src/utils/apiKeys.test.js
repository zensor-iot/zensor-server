import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiKeyError, createAPIKey, listAPIKeys, revokeAPIKey } from './apiKeys'

function stubFetch(response) {
    const fetch = vi.fn().mockResolvedValue(response)
    vi.stubGlobal('fetch', fetch)
    return fetch
}

afterEach(() => {
    vi.unstubAllGlobals()
})

describe('listAPIKeys', () => {
    it('returns the keys from the admin endpoint', async () => {
        const keys = [{ id: 'key-1', name: 'grafana', key_prefix: 'zsk_ab12cd34', created_at: '2026-08-05T10:00:00Z' }]
        const fetch = stubFetch({ ok: true, status: 200, json: async () => keys })

        const result = await listAPIKeys()

        expect(fetch).toHaveBeenCalledWith('/v1/admin/api-keys')
        expect(result).toEqual(keys)
    })

    it('fails with request_failed when the server errors', async () => {
        stubFetch({ ok: false, status: 500 })

        await expect(listAPIKeys()).rejects.toMatchObject({ code: 'request_failed' })
    })
})

describe('createAPIKey', () => {
    it('posts the name and returns the created key with its plaintext', async () => {
        const created = {
            id: 'key-1',
            name: 'grafana',
            key: 'zsk_ab12cd34ef56',
            key_prefix: 'zsk_ab12cd34',
            created_at: '2026-08-05T10:00:00Z',
        }
        const fetch = stubFetch({ ok: true, status: 201, json: async () => created })

        const result = await createAPIKey('grafana')

        expect(fetch).toHaveBeenCalledWith('/v1/admin/api-keys', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: 'grafana' }),
        })
        expect(result).toEqual(created)
    })

    it('fails with duplicate_name when the name is already taken', async () => {
        stubFetch({ ok: false, status: 409 })

        await expect(createAPIKey('grafana')).rejects.toMatchObject({ code: 'duplicate_name' })
    })

    it('fails with name_required when the server rejects the name', async () => {
        stubFetch({ ok: false, status: 400 })

        await expect(createAPIKey('')).rejects.toMatchObject({ code: 'name_required' })
    })

    it('fails with request_failed when the network is unreachable', async () => {
        vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

        await expect(createAPIKey('grafana')).rejects.toMatchObject({ code: 'request_failed' })
    })
})

describe('revokeAPIKey', () => {
    it('deletes the key by id', async () => {
        const fetch = stubFetch({ ok: true, status: 204 })

        await revokeAPIKey('key-1')

        expect(fetch).toHaveBeenCalledWith('/v1/admin/api-keys/key-1', { method: 'DELETE' })
    })

    it('fails with not_found when the key no longer exists', async () => {
        stubFetch({ ok: false, status: 404 })

        await expect(revokeAPIKey('key-1')).rejects.toMatchObject({ code: 'not_found' })
    })
})

describe('ApiKeyError', () => {
    it('is thrown for failed requests so callers can branch on the code', async () => {
        stubFetch({ ok: false, status: 409 })

        await expect(createAPIKey('grafana')).rejects.toBeInstanceOf(ApiKeyError)
    })
})
