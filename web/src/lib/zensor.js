import tokens from '@design-system/tokens.json'

const THEME_KEY = 'zensor-theme'
const LOCALE = 'es-CL'
const NBSP = '\u00A0'

const thresholds = tokens.domain.humidityThresholds
const gaugeSpec = tokens.domain.gauge

export const theme = {
  get() {
    try {
      return localStorage.getItem(THEME_KEY) || 'system'
    } catch {
      return 'system'
    }
  },
  set(mode) {
    const root = document.documentElement
    if (mode === 'light' || mode === 'dark') root.setAttribute('data-theme', mode)
    else root.removeAttribute('data-theme')
    try {
      if (mode === 'system') localStorage.removeItem(THEME_KEY)
      else localStorage.setItem(THEME_KEY, mode)
    } catch {
      // sin almacenamiento: el tema vale solo para esta sesión
    }
    document.dispatchEvent(new CustomEvent('zensor:theme', { detail: { mode } }))
  },
  effective() {
    const mode = theme.get()
    if (mode !== 'system') return mode
    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  },
}

const HUMIDITY_LABELS = { attention: 'Bajo', ok: 'Óptimo', live: 'Saturado' }

export function humidityState(value) {
  const n = Number(value)
  const state = n < thresholds.low.max ? 'attention' : n >= thresholds.saturated.min ? 'live' : 'ok'
  return {
    state,
    label: HUMIDITY_LABELS[state],
    fill: `var(--zs-status-${state}-fill)`,
    bg: `var(--zs-status-${state}-bg)`,
    fg: `var(--zs-status-${state}-fg)`,
  }
}

export function formatNumber(n, digits = 0) {
  return new Intl.NumberFormat(LOCALE, { minimumFractionDigits: digits, maximumFractionDigits: digits }).format(n)
}

export function formatUnit(n, unit, digits = 0) {
  return `${formatNumber(n, digits)}${NBSP}${unit}`
}

export function formatTime(date) {
  return new Intl.DateTimeFormat(LOCALE, { hour: '2-digit', minute: '2-digit', hour12: false }).format(new Date(date))
}

export function formatDate(date) {
  return new Intl.DateTimeFormat(LOCALE, { day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(date))
}

export function formatDateTime(date) {
  return new Intl.DateTimeFormat(LOCALE, {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(date))
}

export function formatRelative(date, now = new Date()) {
  if (!date) return 'nunca'
  const minutes = Math.floor((now - new Date(date)) / 60000)
  if (minutes < 1) return 'ahora'
  if (minutes < 60) return `hace ${minutes}${NBSP}min`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `hace ${hours}${NBSP}h`
  return `hace ${Math.floor(hours / 24)}${NBSP}d`
}

export const gauge = {
  radius: gaugeSpec.radius,
  strokeWidth: gaugeSpec.strokeWidth,
  rotate: gaugeSpec.rotateDeg,
  arc: gaugeSpec.arcDegrees,
}

const round1 = x => Math.round(x * 10) / 10
const clampPercent = v => Math.max(0, Math.min(100, Number(v) || 0))

export function gaugeDash(value, radius = gauge.radius) {
  const circumference = 2 * Math.PI * radius
  const arc = circumference * (gauge.arc / 360)
  const filled = (arc * clampPercent(value)) / 100
  return {
    track: `${round1(arc)} ${round1(circumference - arc)}`,
    value: `${round1(filled)} ${round1(circumference - filled)}`,
  }
}

export function gaugeMarker(target, cx = 90, cy = 90, radius = gauge.radius) {
  const angle = ((gauge.rotate + (gauge.arc * clampPercent(target)) / 100) * Math.PI) / 180
  return { x: round1(cx + radius * Math.cos(angle)), y: round1(cy + radius * Math.sin(angle)) }
}

export function prefersReducedMotion() {
  return !!window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
}

export function cx(...classes) {
  return classes.filter(Boolean).join(' ')
}
