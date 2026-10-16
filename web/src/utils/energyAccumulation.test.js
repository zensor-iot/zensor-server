import { describe, expect, it } from 'vitest'
import {
    getTodayWindow,
    computeCumulativeEnergyKWh,
    splitGridSegmentKWh,
    computeGridEnergyBuckets,
    computeCumulativeGridEnergy,
} from './energyAccumulation'

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

describe('splitGridSegmentKWh', () => {
    const oneHourMs = 60 * 60 * 1000

    it('counts a positive segment as imported energy', () => {
        const result = splitGridSegmentKWh({ time: 0, value: 1000 }, { time: oneHourMs, value: 1000 })

        expect(result).toEqual({ importedKWh: 1, exportedKWh: 0 })
    })

    it('counts a negative segment as exported energy', () => {
        const result = splitGridSegmentKWh({ time: 0, value: -500 }, { time: oneHourMs, value: -500 })

        expect(result).toEqual({ importedKWh: 0, exportedKWh: 0.5 })
    })

    it('splits a segment that crosses zero at the interpolated crossing', () => {
        // +1000 W to -1000 W over 1 h crosses zero at 30 min: 0.25 kWh each way
        const result = splitGridSegmentKWh({ time: 0, value: 1000 }, { time: oneHourMs, value: -1000 })

        expect(result.importedKWh).toBeCloseTo(0.25, 10)
        expect(result.exportedKWh).toBeCloseTo(0.25, 10)
    })
})

describe('computeGridEnergyBuckets', () => {
    const start = new Date(2026, 9, 15, 0, 0, 0)
    const at = (hour, minute = 0) => new Date(2026, 9, 15, hour, minute, 0).getTime()

    it('creates one bucket per hour from the start hour to the end hour', () => {
        const result = computeGridEnergyBuckets([], start, new Date(2026, 9, 15, 2, 30))

        expect(result.map(b => b.time)).toEqual([at(0), at(1), at(2)])
    })

    it('leaves hours without samples empty instead of zero', () => {
        const result = computeGridEnergyBuckets([], start, new Date(2026, 9, 15, 0, 30))

        expect(result[0]).toEqual({ time: at(0), end: at(1), imported: null, exported: null, net: null })
    })

    it('puts imported energy above zero and exported energy below zero', () => {
        const points = [
            { time: at(0), value: 1200 },
            { time: at(0, 15), value: 1200 },
            { time: at(1), value: -600 },
            { time: at(1, 15), value: -600 },
        ]

        const result = computeGridEnergyBuckets(points, start, new Date(2026, 9, 15, 1, 45), 'hour')

        expect(result[0].imported).toBeCloseTo(0.3, 10)
        expect(result[0].exported).toBe(0)
        expect(result[1].imported).toBe(0)
        expect(result[1].exported).toBeCloseTo(-0.15, 10)
        expect(result[1].net).toBeCloseTo(-0.15, 10)
    })

    it('skips segments across a gap in the data', () => {
        const points = [
            { time: at(0), value: 1000 },
            { time: at(0, 50), value: 1000 },
        ]

        const result = computeGridEnergyBuckets(points, start, new Date(2026, 9, 15, 0, 55))

        expect(result[0]).toEqual({ time: at(0), end: at(1), imported: 0, exported: 0, net: 0 })
    })

    it('creates quarter-hour buckets aligned to the quarter', () => {
        const result = computeGridEnergyBuckets([], new Date(2026, 9, 15, 0, 7), new Date(2026, 9, 15, 0, 40), 'quarterHour')

        expect(result.map(b => b.time)).toEqual([at(0), at(0, 15), at(0, 30)])
    })

    it('creates day buckets from local midnight and adds each segment to its day', () => {
        const day = (date, hour = 0) => new Date(2026, 9, date, hour, 0, 0).getTime()
        const points = [
            { time: day(14, 12), value: 1200 },
            { time: day(14, 12) + 10 * 60 * 1000, value: 1200 },
            { time: day(15, 9), value: -600 },
            { time: day(15, 9) + 10 * 60 * 1000, value: -600 },
        ]

        const result = computeGridEnergyBuckets(points, new Date(2026, 9, 14, 8, 30), new Date(2026, 9, 15, 18, 0), 'day')

        expect(result.map(b => b.time)).toEqual([day(14), day(15)])
        expect(result[0].end).toBe(day(15))
        expect(result[0].imported).toBeCloseTo(0.2, 10)
        expect(result[1].exported).toBeCloseTo(-0.1, 10)
    })
})

describe('computeCumulativeGridEnergy', () => {
    const tenMinutesMs = 10 * 60 * 1000

    it('returns an empty array for no samples', () => {
        expect(computeCumulativeGridEnergy([])).toEqual([])
    })

    it('accumulates imported, exported and net energy', () => {
        // 10 min at +600 W = 0.1 kWh in, then 10 min at -1200 W = 0.2 kWh out
        const points = [
            { time: 0, value: 600 },
            { time: tenMinutesMs, value: 600 },
            { time: tenMinutesMs + 1, value: -1200 },
            { time: 2 * tenMinutesMs + 1, value: -1200 },
        ]

        const last = computeCumulativeGridEnergy(points).at(-1)

        expect(last.imported).toBeCloseTo(0.1, 3)
        expect(last.exported).toBeCloseTo(-0.2, 3)
        expect(last.net).toBeCloseTo(-0.1, 3)
    })
})
