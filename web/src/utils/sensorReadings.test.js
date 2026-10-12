import { describe, expect, it } from 'vitest'
import { classifySensor, isValveOpen, sensorReadings } from './sensorReadings'

const sample = {
  data: {
    relay: [{ Index: 1, Value: 0 }],
    temperature: [{ Index: 1, Value: 25 }],
    humidity: [
      { Index: 2, Value: 31 },
      { Index: 1, Value: 95 },
    ],
    water_flow: [{ index: 1, value: 8.6 }],
  },
}

describe('classifySensor', () => {
  it('maps known sensor types to labels and units', () => {
    expect(classifySensor('soil_humidity')).toMatchObject({ kind: 'humidity', label: 'Humedad', unit: '%' })
    expect(classifySensor('water_flow')).toMatchObject({ kind: 'flow', unit: 'L/min' })
    expect(classifySensor('relay')).toMatchObject({ kind: 'relay', label: 'Válvula' })
  })

  it('keeps unknown sensor types as they come', () => {
    expect(classifySensor('lux')).toMatchObject({ kind: 'other', label: 'lux', unit: null })
  })
})

describe('sensorReadings', () => {
  it('flattens and orders readings by kind and index', () => {
    expect(sensorReadings(sample).map(reading => reading.label)).toEqual([
      'Humedad #1',
      'Humedad #2',
      'Temperatura',
      'Caudal',
      'Válvula',
    ])
  })

  it('returns an empty list without data', () => {
    expect(sensorReadings(undefined)).toEqual([])
    expect(sensorReadings({ data: null })).toEqual([])
  })
})

describe('isValveOpen', () => {
  it('is true only when a relay reads 1', () => {
    expect(isValveOpen(sample)).toBe(false)
    expect(isValveOpen({ data: { relay: [{ Index: 1, Value: 1 }] } })).toBe(true)
  })
})
