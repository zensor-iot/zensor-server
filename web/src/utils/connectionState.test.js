import { describe, expect, it } from 'vitest'
import { connectionState } from './connectionState'

describe('connectionState', () => {
  it('reports online when the socket is open', () => {
    expect(connectionState({ isConnected: true, connectionStatus: 'Connected' })).toEqual({ tone: 'ok', label: 'En línea' })
  })

  it('reports connecting while the first attempt is in progress', () => {
    expect(connectionState({ isConnected: false, connectionStatus: 'Connecting...' }).label).toBe('Conectando…')
  })

  it('reports connecting while retrying', () => {
    expect(connectionState({ isConnected: false, connectionStatus: 'Reconnecting... (2/5)' }).label).toBe('Conectando…')
  })

  it('reports offline after giving up', () => {
    expect(connectionState({ isConnected: false, connectionStatus: 'Failed to connect' })).toEqual({
      tone: 'attention',
      label: 'Sin conexión',
    })
  })

  it('reports offline after an error', () => {
    expect(connectionState({ isConnected: false, connectionStatus: 'Error' }).label).toBe('Sin conexión')
  })
})
