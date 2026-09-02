import { describe, it, expect } from 'vitest'
import {
    describeDose,
    describeFrequency,
    totalDosesForDuration,
    previewOccurrences
} from './medicineSchedule'

describe('describeDose', () => {
    it('pluralises above one', () => {
        expect(describeDose(15, 'drop')).toBe('15 gotas')
        expect(describeDose(2, 'spoonful')).toBe('2 cucharadas')
    })

    it('uses the singular at one or below', () => {
        expect(describeDose(1, 'pill')).toBe('1 pastilla')
        expect(describeDose(0.5, 'pill')).toBe('0.5 pastilla')
    })

    it('falls back to the raw unit when unknown', () => {
        expect(describeDose(3, 'sachet')).toBe('3 sachet')
    })
})

describe('describeFrequency', () => {
    it('reads a plural interval', () => {
        expect(describeFrequency({ every: 8, unit: 'hour' })).toBe('cada 8 horas')
        expect(describeFrequency({ every: 2, unit: 'day' })).toBe('cada 2 días')
    })

    it('reads a single interval without the number', () => {
        expect(describeFrequency({ every: 1, unit: 'hour' })).toBe('cada hora')
        expect(describeFrequency({ every: 1, unit: 'day' })).toBe('cada día')
    })

    it('returns nothing for an incomplete schedule', () => {
        expect(describeFrequency(null)).toBe('')
        expect(describeFrequency({ every: 8 })).toBe('')
    })
})

describe('totalDosesForDuration', () => {
    it('converts days into a dose count', () => {
        expect(totalDosesForDuration(7, 8, 'hour')).toBe(21)
        expect(totalDosesForDuration(10, 1, 'day')).toBe(10)
    })

    it('never returns less than one dose', () => {
        expect(totalDosesForDuration(1, 48, 'hour')).toBe(1)
    })

    it('rejects nonsense input', () => {
        expect(totalDosesForDuration(0, 8, 'hour')).toBeNull()
        expect(totalDosesForDuration(7, 0, 'hour')).toBeNull()
        expect(totalDosesForDuration('abc', 8, 'hour')).toBeNull()
    })
})

describe('previewOccurrences', () => {
    it('lists the first doses of an hourly schedule', () => {
        const start = new Date('2026-03-01T08:00:00Z')
        const occurrences = previewOccurrences(
            { start_at: start.toISOString(), every: 8, unit: 'hour' }, 3
        )
        expect(occurrences).toHaveLength(3)
        expect(occurrences[0].getTime()).toBe(start.getTime())
        expect(occurrences[1].getTime() - occurrences[0].getTime()).toBe(8 * 3600 * 1000)
    })

    it('keeps the wall clock time on a daily schedule', () => {
        const occurrences = previewOccurrences(
            { start_at: '2026-03-01T08:00:00Z', every: 1, unit: 'day' }, 2
        )
        expect(occurrences[0].getHours()).toBe(occurrences[1].getHours())
    })

    it('returns nothing for an incomplete schedule', () => {
        expect(previewOccurrences(null)).toEqual([])
        expect(previewOccurrences({ every: 8, unit: 'hour' })).toEqual([])
    })
})
