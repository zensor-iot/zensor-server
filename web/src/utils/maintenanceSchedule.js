// Calendar schedule helpers for maintenance activity schedules.
// Schedules are { start_date: ISO8601, every: number, unit: 'day'|'week'|'month'|'quarter'|'year' }.

export const UNITS = [
    { value: 'day', label: 'Day' },
    { value: 'week', label: 'Week' },
    { value: 'month', label: 'Month' },
    { value: 'quarter', label: 'Quarter' },
    { value: 'year', label: 'Year' }
]

const unitLabel = (unit) => {
    const found = UNITS.find(u => u.value === unit)
    return found ? found.label.toLowerCase() : unit
}

const pluralize = (count, label) => (count === 1 ? label : `${label}s`)

export function describeSchedule(schedule) {
    if (!schedule || typeof schedule !== 'object') {
        return 'Unknown schedule'
    }
    const { start_date: startDate, every, unit } = schedule
    if (!startDate || !every || !unit) {
        return 'Unknown schedule'
    }
    const date = new Date(startDate)
    const dateLabel = date.toLocaleDateString(undefined, { month: 'long', day: 'numeric', year: 'numeric' })
    if (every === 1) {
        return `Every ${unitLabel(unit)} starting ${dateLabel}`
    }
    return `Every ${every} ${pluralize(every, unitLabel(unit))} starting ${dateLabel}`
}

// Converts a schedule object to the payload shape the API expects, trimming
// the start date to whole seconds for stable comparisons.
export function toSchedulePayload(schedule) {
    if (!schedule || typeof schedule !== 'object') {
        return null
    }
    const { start_date: startDate, every, unit } = schedule
    if (!startDate || !every || !unit) {
        return null
    }
    return {
        start_date: new Date(startDate).toISOString(),
        every,
        unit
    }
}

// Builds a schedule payload from form state { startDate, every, unit }.
export function buildSchedule({ startDate, every, unit }) {
    if (!startDate || !every || !unit) {
        return null
    }
    return toSchedulePayload({ start_date: new Date(startDate).toISOString(), every, unit })
}

// Relative label for an execution's scheduled date ("Today", "In 3 days", "5 days overdue").
export function relativeDateLabel(dateString) {
    const date = new Date(dateString)
    const now = new Date()
    const startOfDay = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate())
    const diffDays = Math.round((startOfDay(date) - startOfDay(now)) / 86400000)

    if (diffDays === 0) return 'Today'
    if (diffDays === 1) return 'Tomorrow'
    if (diffDays === -1) return 'Yesterday'
    if (diffDays > 1) return `In ${diffDays} days`
    return `${-diffDays} days overdue`
}
