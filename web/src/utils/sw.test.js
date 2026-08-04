import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

// The service worker is served under the SPA base path (/ui/sw.js), so every
// URL it emits must resolve under that base, never the site root.
const swScriptUrl = 'https://portal.example/ui/sw.js'

const listeners = {}
const showNotification = vi.fn()

beforeAll(async () => {
    vi.stubGlobal('self', {
        addEventListener: (type, handler) => {
            listeners[type] = handler
        },
        registration: { showNotification },
        clients: { matchAll: vi.fn().mockResolvedValue([]), openWindow: vi.fn() },
        location: { href: swScriptUrl },
    })
    await import('../../public/sw.js')
})

beforeEach(() => {
    showNotification.mockClear()
})

function pushEvent(payload) {
    return {
        data: { json: () => payload },
        waitUntil: (promise) => promise,
    }
}

describe('service worker push notifications', () => {
    it('resolves the notification icon under the worker base path', async () => {
        await listeners.push(pushEvent({ title: 'T', body: 'B' }))

        const [, options] = showNotification.mock.calls[0]
        expect(options.icon).toBe('/ui/vite.svg')
    })

    it('defaults the deeplink to the worker base path', async () => {
        await listeners.push(pushEvent({ title: 'T', body: 'B' }))

        const [, options] = showNotification.mock.calls[0]
        expect(options.data.deeplink).toBe('/ui/')
    })
})
