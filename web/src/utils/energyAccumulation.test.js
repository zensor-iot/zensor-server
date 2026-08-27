import { describe, expect, it } from 'vitest'
import { getTodayWindow, computeCumulativeEnergyKWh } from './energyAccumulation'

describe('getTodayWindow', () => {
    it('returns local midnight as start and the given instant as end', () => {
        const now = new Date(2026, 7, 26, 14, 30, 0)

        const { start, end } = getTodayWindow(now)

        expect(start).toEqual(new Date(2026, 7, 26, 0, 0, 0, 0))
        expect(end).toEqual(now)
    })
})

describe('computeCumulativeEnergyKWh', () => {
    it('returns an empty array for no samples', () => {
        expect(computeCumulativeEnergyKWh([])).toEqual([])
    })

    it('starts at zero energy on the first sample', () => {
        const result = computeCumulativeEnergyKWh([{ time: 1000, value: 500 }])

        expect(result).toEqual([{ time: 1000, value: 0 }])
    })

    it('accumulates energy via the trapezoidal rule between samples', () => {
        // 1000 W held constant for 1 hour = 1 kWh
        const oneHourMs = 60 * 60 * 1000
        const samples = [
            { time: 0, value: 1000 },
            { time: oneHourMs, value: 1000 },
        ]

        const result = computeCumulativeEnergyKWh(samples)

        expect(result).toEqual([
            { time: 0, value: 0 },
            { time: oneHourMs, value: 1 },
        ])
    })

    it('averages power across an interval when it changes between samples', () => {
        // power ramps 0 W -> 2000 W over 1 hour: average 1000 W -> 1 kWh
        const oneHourMs = 60 * 60 * 1000
        const samples = [
            { time: 0, value: 0 },
            { time: oneHourMs, value: 2000 },
        ]

        const result = computeCumulativeEnergyKWh(samples)

        expect(result[1].value).toBeCloseTo(1)
    })

    it('keeps accumulating across more than two samples', () => {
        const oneHourMs = 60 * 60 * 1000
        const samples = [
            { time: 0, value: 1000 },
            { time: oneHourMs, value: 1000 },
            { time: 2 * oneHourMs, value: 1000 },
        ]

        const result = computeCumulativeEnergyKWh(samples)

        expect(result.map((p) => p.value)).toEqual([0, 1, 2])
    })
})
