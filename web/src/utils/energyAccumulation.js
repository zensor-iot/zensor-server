// Returns the local-midnight-to-now window for "today", in the browser's
// own timezone (there is no per-user timezone wired into this dashboard).
export function getTodayWindow(now = new Date()) {
    const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0)
    return { start, end: now }
}

// Converts a chronologically sorted series of instantaneous power samples
// (watts) into cumulative energy (kWh) since the first sample, integrating
// with the trapezoidal rule so a changing power draw between samples is
// averaged rather than assumed constant.
export function computeCumulativeEnergyKWh(points) {
    if (points.length === 0) return []

    const result = [{ time: points[0].time, value: 0 }]
    let cumulativeKWh = 0

    for (let i = 1; i < points.length; i++) {
        const previous = points[i - 1]
        const current = points[i]
        const hours = (current.time - previous.time) / (60 * 60 * 1000)
        const averageWatts = (previous.value + current.value) / 2
        cumulativeKWh += (averageWatts * hours) / 1000
        result.push({ time: current.time, value: cumulativeKWh })
    }

    return result
}

const HOUR_MS = 60 * 60 * 1000

// Samples further apart than this are treated as a gap in the data rather
// than integrated, so an outage does not get filled with invented energy.
export const MAX_SAMPLE_GAP_MS = 15 * 60 * 1000

// Splits the energy of one segment between two signed grid power samples
// (positive = imported from the grid, negative = exported to it) into its
// imported and exported parts. When the segment crosses zero the crossing
// point is interpolated linearly so each side is integrated separately.
export function splitGridSegmentKWh(previous, current) {
    const hours = (current.time - previous.time) / HOUR_MS
    const a = previous.value
    const b = current.value
    const area = (x, y, h) => ((x + y) / 2) * h / 1000

    if (a >= 0 && b >= 0) return { importedKWh: area(a, b, hours), exportedKWh: 0 }
    if (a <= 0 && b <= 0) return { importedKWh: 0, exportedKWh: -area(a, b, hours) }

    const zeroAt = (a / (a - b)) * hours
    const first = area(a, 0, zeroAt)
    const second = area(0, b, hours - zeroAt)
    return a > 0
        ? { importedKWh: first, exportedKWh: -second }
        : { importedKWh: second, exportedKWh: -first }
}

// Groups signed grid power samples (watts) into hourly energy buckets from
// the hour of `start` up to the hour of `end`. Imported energy is positive
// and exported energy negative, so the chart can draw them on either side of
// zero; hours without any sample keep null values instead of a fake zero.
export function computeHourlyGridEnergy(points, start, end) {
    const first = new Date(start)
    first.setMinutes(0, 0, 0)
    const buckets = []
    for (let t = first.getTime(); t <= end.getTime(); t += HOUR_MS) {
        buckets.push({ time: t, imported: null, exported: null, net: null })
    }
    if (buckets.length === 0) return buckets

    const bucketFor = time => buckets[Math.floor((time - buckets[0].time) / HOUR_MS)]
    const add = (bucket, importedKWh, exportedKWh) => {
        bucket.imported = (bucket.imported ?? 0) + importedKWh
        bucket.exported = (bucket.exported ?? 0) - exportedKWh
        bucket.net = bucket.imported + bucket.exported
    }

    for (const point of points) {
        const bucket = bucketFor(point.time)
        if (bucket && bucket.net === null) add(bucket, 0, 0)
    }

    for (let i = 1; i < points.length; i++) {
        const previous = points[i - 1]
        const current = points[i]
        if (current.time - previous.time > MAX_SAMPLE_GAP_MS) continue
        const bucket = bucketFor(previous.time)
        if (!bucket) continue
        const { importedKWh, exportedKWh } = splitGridSegmentKWh(previous, current)
        add(bucket, importedKWh, exportedKWh)
    }

    return buckets
}

// Converts signed grid power samples (watts) into running totals (kWh) since
// the first sample: imported (positive), exported (negative) and their net
// balance, which is above zero while more energy came from the grid than went
// back to it.
export function computeCumulativeGridEnergy(points) {
    if (points.length === 0) return []

    const result = [{ time: points[0].time, imported: 0, exported: 0, net: 0 }]
    let imported = 0
    let exported = 0

    for (let i = 1; i < points.length; i++) {
        const previous = points[i - 1]
        const current = points[i]
        if (current.time - previous.time <= MAX_SAMPLE_GAP_MS) {
            const segment = splitGridSegmentKWh(previous, current)
            imported += segment.importedKWh
            exported -= segment.exportedKWh
        }
        result.push({ time: current.time, imported, exported, net: imported + exported })
    }

    return result
}
