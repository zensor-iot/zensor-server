// Helpers for reading and building a treatment's dosing schedule.

const UNIT_LABELS = {
    drop: ['gota', 'gotas'],
    pill: ['pastilla', 'pastillas'],
    spoonful: ['cucharada', 'cucharadas']
}

export const DOSE_UNITS = [
    { value: 'drop', label: 'Gotas' },
    { value: 'pill', label: 'Pastillas' },
    { value: 'spoonful', label: 'Cucharadas' }
]

export const INTERVAL_UNITS = [
    { value: 'hour', label: 'horas' },
    { value: 'day', label: 'días' }
]

// describeDose renders a quantity and unit the way a person reads it.
export function describeDose(quantity, unit) {
    const labels = UNIT_LABELS[unit]
    if (!labels) {
        return `${quantity} ${unit}`
    }
    const singular = Math.abs(Number(quantity)) <= 1
    return `${quantity} ${singular ? labels[0] : labels[1]}`
}

// describeFrequency turns a schedule into "cada 8 horas" or "cada día".
export function describeFrequency(schedule) {
    if (!schedule || !schedule.every || !schedule.unit) {
        return ''
    }
    const unit = INTERVAL_UNITS.find(u => u.value === schedule.unit)
    if (!unit) {
        return ''
    }
    if (schedule.every === 1) {
        return schedule.unit === 'hour' ? 'cada hora' : 'cada día'
    }
    return `cada ${schedule.every} ${unit.label}`
}

// totalDosesForDuration converts a duration in days into the number of doses a
// schedule produces, which is what the API stores as the end condition.
export function totalDosesForDuration(days, every, unit) {
    const numericDays = Number(days)
    const numericEvery = Number(every)
    if (!Number.isFinite(numericDays) || !Number.isFinite(numericEvery) || numericDays <= 0 || numericEvery <= 0) {
        return null
    }
    const hours = unit === 'hour' ? numericEvery : numericEvery * 24
    return Math.max(1, Math.floor((numericDays * 24) / hours))
}

// previewOccurrences lists the first few dose times a schedule produces, so the
// form can show what it is about to create.
export function previewOccurrences(schedule, count = 3) {
    if (!schedule || !schedule.start_at || !schedule.every || !schedule.unit) {
        return []
    }
    const start = new Date(schedule.start_at)
    if (Number.isNaN(start.getTime())) {
        return []
    }

    const occurrences = []
    for (let n = 0; n < count; n++) {
        const next = new Date(start)
        if (schedule.unit === 'hour') {
            next.setHours(next.getHours() + n * schedule.every)
        } else {
            next.setDate(next.getDate() + n * schedule.every)
        }
        occurrences.push(next)
    }
    return occurrences
}

export function formatDateTime(value) {
    const date = value instanceof Date ? value : new Date(value)
    if (Number.isNaN(date.getTime())) {
        return ''
    }
    return date.toLocaleString([], {
        day: '2-digit',
        month: 'short',
        hour: '2-digit',
        minute: '2-digit'
    })
}

export function formatTime(value) {
    const date = value instanceof Date ? value : new Date(value)
    if (Number.isNaN(date.getTime())) {
        return ''
    }
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
