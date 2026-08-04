// Native Web Push enrollment helpers. The PushSubscription JSON is stored
// verbatim as the token on the server (platform "web").

// The SPA (including sw.js) is served under Vite's base path, not the site root.
const serviceWorkerUrl = `${import.meta.env.BASE_URL}sw.js`

export const isWebPushSupported = () =>
    typeof navigator !== 'undefined' &&
    'serviceWorker' in navigator &&
    typeof window !== 'undefined' &&
    'PushManager' in window &&
    'Notification' in window

export class PermissionDeniedError extends Error {
    constructor() {
        super('Notification permission denied')
        this.name = 'PermissionDeniedError'
    }
}

function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
    const rawData = window.atob(base64)
    const outputArray = new Uint8Array(rawData.length)
    for (let i = 0; i < rawData.length; i++) {
        outputArray[i] = rawData.charCodeAt(i)
    }
    return outputArray
}

async function fetchVapidPublicKey() {
    const response = await fetch('/v1/push/vapid-public-key')
    if (!response.ok) {
        throw new Error('Web push is not configured on the server')
    }
    const data = await response.json()
    return data.public_key
}

export async function getExistingSubscription() {
    if (!isWebPushSupported()) {
        return null
    }
    const registration = await navigator.serviceWorker.getRegistration(serviceWorkerUrl)
    if (!registration) {
        return null
    }
    return registration.pushManager.getSubscription()
}

export async function enableBrowserNotifications(userId) {
    const publicKey = await fetchVapidPublicKey()
    const registration = await navigator.serviceWorker.register(serviceWorkerUrl)

    const permission = await Notification.requestPermission()
    if (permission !== 'granted') {
        throw new PermissionDeniedError()
    }

    const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(publicKey)
    })

    const response = await fetch(`/v1/users/${userId}/push-tokens`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            token: JSON.stringify(subscription),
            platform: 'web'
        })
    })
    if (!response.ok) {
        await subscription.unsubscribe()
        throw new Error(`Failed to register push token: HTTP ${response.status}`)
    }

    return subscription
}

export async function disableBrowserNotifications(userId) {
    const subscription = await getExistingSubscription()
    if (!subscription) {
        return
    }

    const token = JSON.stringify(subscription)
    await subscription.unsubscribe()

    const response = await fetch(`/v1/users/${userId}/push-tokens`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token })
    })
    if (!response.ok && response.status !== 404) {
        throw new Error(`Failed to unregister push token: HTTP ${response.status}`)
    }
}
