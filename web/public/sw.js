/* Service worker for native Web Push notifications. */

self.addEventListener('push', (event) => {
    let payload = {}
    try {
        payload = event.data ? event.data.json() : {}
    } catch {
        payload = { body: event.data ? event.data.text() : '' }
    }

    const title = payload.title || 'Zensor'
    const options = {
        body: payload.body || '',
        icon: '/vite.svg',
        data: { deeplink: payload.deeplink || '/' }
    }

    event.waitUntil(self.registration.showNotification(title, options))
})

self.addEventListener('notificationclick', (event) => {
    event.notification.close()
    const deeplink = (event.notification.data && event.notification.data.deeplink) || '/'

    event.waitUntil(
        self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
            for (const client of clientList) {
                const url = new URL(client.url)
                if (url.pathname === deeplink && 'focus' in client) {
                    return client.focus()
                }
            }
            for (const client of clientList) {
                if ('focus' in client && 'navigate' in client) {
                    return client.focus().then((focused) => focused.navigate(deeplink))
                }
            }
            return self.clients.openWindow(deeplink)
        })
    )
})
