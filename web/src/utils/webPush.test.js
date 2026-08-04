import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableBrowserNotifications, getExistingSubscription } from './webPush'

const expectedSwUrl = `${import.meta.env.BASE_URL}sw.js`

function stubBrowserGlobals({ register, getRegistration }) {
    vi.stubGlobal('navigator', {
        serviceWorker: { register, getRegistration },
    })
    vi.stubGlobal('window', {
        PushManager: function PushManager() {},
        Notification: function Notification() {},
        atob: (value) => globalThis.atob(value),
    })
    vi.stubGlobal('Notification', {
        requestPermission: vi.fn().mockResolvedValue('granted'),
    })
}

afterEach(() => {
    vi.unstubAllGlobals()
})

describe('webPush service worker path', () => {
    it('registers the service worker under the app base path', async () => {
        const subscription = { unsubscribe: vi.fn() }
        const register = vi.fn().mockResolvedValue({
            pushManager: { subscribe: vi.fn().mockResolvedValue(subscription) },
        })
        stubBrowserGlobals({ register, getRegistration: vi.fn() })
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ public_key: 'BPubKey' }),
        }))

        await enableBrowserNotifications('user-1')

        expect(register).toHaveBeenCalledWith(expectedSwUrl)
    })

    it('looks up the existing registration under the app base path', async () => {
        const getRegistration = vi.fn().mockResolvedValue(null)
        stubBrowserGlobals({ register: vi.fn(), getRegistration })

        await getExistingSubscription()

        expect(getRegistration).toHaveBeenCalledWith(expectedSwUrl)
    })
})
