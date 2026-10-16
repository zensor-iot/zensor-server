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

const QUARTER_HOUR_MS = 15 * 60 * 1000

// Bucket sizes for grouping grid energy. Sub-day buckets step in absolute
// time so daylight-saving changes cannot repeat or skip a bucket; day buckets
// follow local midnights so each bar matches a calendar day.
export const GRID_BUCKETS = {
    quarterHour: {
        floor: date => date.setMinutes(date.getMinutes() - (date.getMinutes() % 15), 0, 0),
        next: time => time + QUARTER_HOUR_MS,
    },
    hour: {
        floor: date => date.setMinutes(0, 0, 0),
        next: time => time + HOUR_MS,
    },
    day: {
        floor: date => date.setHours(0, 0, 0, 0),
        next: time => {
            const date = new Date(time)
            date.setDate(date.getDate() + 1)
            return date.getTime()
        },
    },
}

// Groups signed grid power samples (watts) into energy buckets of the given
// size, from the bucket holding `start` up to the bucket holding `end`. Each
// bucket carries its own `end` so callers can label partial or uneven days.
// Imported energy is positive and exported energy negative, so the chart can
// draw them on either side of zero; buckets without any sample keep null
// values instead of a fake zero.
export function computeGridEnergyBuckets(points, start, end, bucketSize = 'hour') {
    const { floor, next } = GRID_BUCKETS[bucketSize]
    const first = new Date(start)
    floor(first)
    const buckets = []
    for (let t = first.getTime(); t <= end.getTime(); t = next(t)) {
        buckets.push({ time: t, end: next(t), imported: null, exported: null, net: null })
    }
    if (buckets.length === 0) return buckets

    const bucketFor = time => {
        let low = 0
        let high = buckets.length - 1
        while (low < high) {
            const middle = Math.ceil((low + high) / 2)
            if (buckets[middle].time <= time) low = middle
            else high = middle - 1
        }
        const bucket = buckets[low]
        return time >= bucket.time && time < bucket.end ? bucket : undefined
    }
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
