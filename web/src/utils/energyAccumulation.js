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
