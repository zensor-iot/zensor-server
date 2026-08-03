// Cron helpers for maintenance activity schedules (5-field cron: minute hour dom month dow).
// JS port of the mobile app's schedule utilities: a friendly builder for
// weekly/monthly/quarterly/yearly schedules plus reverse parsing with a
// fallback for cron expressions the builder cannot represent.

export const FREQUENCIES = ['weekly', 'monthly', 'quarterly', 'yearly']

export const WEEKDAYS = [
    { value: 0, label: 'Sunday' },
    { value: 1, label: 'Monday' },
    { value: 2, label: 'Tuesday' },
    { value: 3, label: 'Wednesday' },
    { value: 4, label: 'Thursday' },
    { value: 5, label: 'Friday' },
    { value: 6, label: 'Saturday' }
]

export const MONTHS = [
    { value: 1, label: 'January' },
    { value: 2, label: 'February' },
    { value: 3, label: 'March' },
    { value: 4, label: 'April' },
    { value: 5, label: 'May' },
    { value: 6, label: 'June' },
    { value: 7, label: 'July' },
    { value: 8, label: 'August' },
    { value: 9, label: 'September' },
    { value: 10, label: 'October' },
    { value: 11, label: 'November' },
    { value: 12, label: 'December' }
]

const QUARTER_MONTHS = '1,4,7,10'

export function buildCronExpression({ frequency, dayOfWeek = 1, dayOfMonth = 1, month = 1 }) {
    switch (frequency) {
        case 'weekly':
            return `0 9 * * ${dayOfWeek}`
        case 'monthly':
            return `0 9 ${dayOfMonth} * *`
        case 'quarterly':
            return `0 9 ${dayOfMonth} ${QUARTER_MONTHS} *`
        case 'yearly':
            return `0 9 ${dayOfMonth} ${month} *`
        default:
            throw new Error(`Unknown frequency: ${frequency}`)
    }
}

const isNumeric = (field) => /^\d+$/.test(field)

// Returns {frequency, dayOfWeek, dayOfMonth, month} for expressions the
// builder can represent, or null so callers can warn and fall back to raw
// cron editing.
export function parseCronExpression(cron) {
    if (typeof cron !== 'string') {
        return null
    }
    const fields = cron.trim().split(/\s+/)
    if (fields.length !== 5) {
        return null
    }
    const [, , dayOfMonth, month, dayOfWeek] = fields

    if (dayOfMonth === '*' && month === '*' && isNumeric(dayOfWeek)) {
        return { frequency: 'weekly', dayOfWeek: Number(dayOfWeek) % 7, dayOfMonth: 1, month: 1 }
    }
    if (isNumeric(dayOfMonth) && month === '*' && dayOfWeek === '*') {
        return { frequency: 'monthly', dayOfWeek: 1, dayOfMonth: Number(dayOfMonth), month: 1 }
    }
    if (isNumeric(dayOfMonth) && (month === QUARTER_MONTHS || month === '*/3') && dayOfWeek === '*') {
        return { frequency: 'quarterly', dayOfWeek: 1, dayOfMonth: Number(dayOfMonth), month: 1 }
    }
    if (isNumeric(dayOfMonth) && isNumeric(month) && dayOfWeek === '*') {
        return { frequency: 'yearly', dayOfWeek: 1, dayOfMonth: Number(dayOfMonth), month: Number(month) }
    }
    return null
}

export function describeSchedule(cron) {
    const parsed = parseCronExpression(cron)
    if (!parsed) {
        return cron
    }
    switch (parsed.frequency) {
        case 'weekly': {
            const weekday = WEEKDAYS.find(d => d.value === parsed.dayOfWeek)
            return `Weekly on ${weekday ? weekday.label : `day ${parsed.dayOfWeek}`}`
        }
        case 'monthly':
            return `Monthly on day ${parsed.dayOfMonth}`
        case 'quarterly':
            return `Quarterly on day ${parsed.dayOfMonth}`
        case 'yearly': {
            const month = MONTHS.find(m => m.value === parsed.month)
            return `Yearly on ${month ? month.label : `month ${parsed.month}`} ${parsed.dayOfMonth}`
        }
        default:
            return cron
    }
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
