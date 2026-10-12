import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { Check, ChevronDown } from 'lucide-react'
import { cx } from '../lib/zensor'

const TIMEZONES = [
  { value: 'America/Santiago', label: 'Santiago' },
  { value: 'America/Buenos_Aires', label: 'Buenos Aires' },
  { value: 'America/Lima', label: 'Lima' },
  { value: 'America/Sao_Paulo', label: 'São Paulo' },
  { value: 'America/Mexico_City', label: 'Ciudad de México' },
  { value: 'America/New_York', label: 'Hora del este (EE. UU. y Canadá)' },
  { value: 'America/Chicago', label: 'Hora central (EE. UU. y Canadá)' },
  { value: 'America/Denver', label: 'Hora de la montaña (EE. UU. y Canadá)' },
  { value: 'America/Los_Angeles', label: 'Hora del Pacífico (EE. UU. y Canadá)' },
  { value: 'America/Phoenix', label: 'Arizona' },
  { value: 'America/Anchorage', label: 'Alaska' },
  { value: 'America/Toronto', label: 'Toronto' },
  { value: 'Europe/London', label: 'Londres' },
  { value: 'Europe/Paris', label: 'París' },
  { value: 'Europe/Berlin', label: 'Berlín' },
  { value: 'Europe/Madrid', label: 'Madrid' },
  { value: 'Europe/Rome', label: 'Roma' },
  { value: 'Europe/Amsterdam', label: 'Ámsterdam' },
  { value: 'Europe/Brussels', label: 'Bruselas' },
  { value: 'Europe/Stockholm', label: 'Estocolmo' },
  { value: 'Europe/Vienna', label: 'Viena' },
  { value: 'Europe/Zurich', label: 'Zúrich' },
  { value: 'Europe/Athens', label: 'Atenas' },
  { value: 'Europe/Dublin', label: 'Dublín' },
  { value: 'Europe/Prague', label: 'Praga' },
  { value: 'Europe/Warsaw', label: 'Varsovia' },
  { value: 'Europe/Moscow', label: 'Moscú' },
  { value: 'Europe/Istanbul', label: 'Estambul' },
  { value: 'Asia/Dubai', label: 'Dubái' },
  { value: 'Asia/Tokyo', label: 'Tokio' },
  { value: 'Asia/Shanghai', label: 'Shanghái' },
  { value: 'Asia/Hong_Kong', label: 'Hong Kong' },
  { value: 'Asia/Singapore', label: 'Singapur' },
  { value: 'Asia/Bangkok', label: 'Bangkok' },
  { value: 'Asia/Seoul', label: 'Seúl' },
  { value: 'Asia/Delhi', label: 'India' },
  { value: 'Asia/Karachi', label: 'Karachi' },
  { value: 'Asia/Jakarta', label: 'Yakarta' },
  { value: 'Asia/Manila', label: 'Manila' },
  { value: 'Asia/Dhaka', label: 'Daca' },
  { value: 'Asia/Jerusalem', label: 'Jerusalén' },
  { value: 'Asia/Baghdad', label: 'Bagdad' },
  { value: 'Australia/Sydney', label: 'Sídney' },
  { value: 'Australia/Melbourne', label: 'Melbourne' },
  { value: 'Australia/Brisbane', label: 'Brisbane' },
  { value: 'Australia/Perth', label: 'Perth' },
  { value: 'Pacific/Auckland', label: 'Auckland' },
  { value: 'Pacific/Fiji', label: 'Fiyi' },
  { value: 'Pacific/Honolulu', label: 'Honolulu' },
  { value: 'UTC', label: 'UTC' },
]

const normalize = text =>
  text
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()

const timezoneLabel = value => TIMEZONES.find(tz => tz.value === value)?.label

/**
 * Searchable timezone combobox (WAI-ARIA combobox + listbox). Calls onChange with the IANA name.
 * Accepts the input props a Field passes (id, aria-describedby…).
 */
const TimezoneDropdown = ({ value, onChange, placeholder = 'Busca una zona horaria', ...inputProps }) => {
  const listId = useId()
  const optionId = index => `${listId}-option-${index}`
  const rootRef = useRef(null)
  const inputRef = useRef(null)
  const listRef = useRef(null)
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [highlighted, setHighlighted] = useState(-1)

  const selectedLabel = value ? (timezoneLabel(value) ?? value) : ''

  const options = useMemo(() => {
    const term = normalize(query.trim())
    if (!term) return TIMEZONES
    return TIMEZONES.filter(tz => normalize(tz.label).includes(term) || normalize(tz.value).includes(term))
  }, [query])

  useEffect(() => {
    if (!open) return undefined
    const onPointerDown = event => {
      if (rootRef.current?.contains(event.target)) return
      setOpen(false)
      setQuery('')
      setHighlighted(-1)
    }
    document.addEventListener('mousedown', onPointerDown)
    return () => document.removeEventListener('mousedown', onPointerDown)
  }, [open])

  useEffect(() => {
    if (!open || highlighted < 0) return
    listRef.current?.querySelector(`[data-index="${highlighted}"]`)?.scrollIntoView({ block: 'nearest' })
  }, [open, highlighted])

  const openList = () => {
    if (open) return
    setOpen(true)
    setQuery('')
    setHighlighted(Math.max(0, TIMEZONES.findIndex(tz => tz.value === value)))
  }

  const close = () => {
    setOpen(false)
    setQuery('')
    setHighlighted(-1)
  }

  const select = timezone => {
    onChange(timezone.value)
    close()
    inputRef.current?.focus()
  }

  const handleKeyDown = event => {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        if (!open) openList()
        else setHighlighted(index => Math.min(options.length - 1, index + 1))
        break
      case 'ArrowUp':
        event.preventDefault()
        if (!open) openList()
        else setHighlighted(index => Math.max(0, index - 1))
        break
      case 'Home':
        if (open) {
          event.preventDefault()
          setHighlighted(0)
        }
        break
      case 'End':
        if (open) {
          event.preventDefault()
          setHighlighted(options.length - 1)
        }
        break
      case 'Enter':
        if (open && options[highlighted]) {
          event.preventDefault()
          select(options[highlighted])
        }
        break
      case 'Escape':
        if (open) {
          event.preventDefault()
          event.stopPropagation()
          close()
        }
        break
      case 'Tab':
        close()
        break
      default:
    }
  }

  return (
    <div ref={rootRef} className="relative">
      <input
        ref={inputRef}
        type="text"
        role="combobox"
        aria-autocomplete="list"
        aria-expanded={open}
        aria-controls={listId}
        aria-activedescendant={open && highlighted >= 0 && options[highlighted] ? optionId(highlighted) : undefined}
        autoComplete="off"
        spellCheck={false}
        className="zs-control pr-12"
        placeholder={placeholder}
        value={open ? query : selectedLabel}
        onChange={event => {
          if (!open) setOpen(true)
          setQuery(event.target.value)
          setHighlighted(0)
        }}
        onClick={openList}
        onKeyDown={handleKeyDown}
        {...inputProps}
      />
      <ChevronDown
        size={18}
        strokeWidth={1.8}
        aria-hidden="true"
        className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-muted"
      />
      {open && (
        <ul
          ref={listRef}
          id={listId}
          role="listbox"
          aria-label="Zonas horarias"
          className="absolute inset-x-0 top-full z-sticky m-0 mt-2 max-h-[18rem] list-none overflow-y-auto rounded-md border border-line bg-surface p-1 shadow-overlay"
        >
          {options.length === 0 ? (
            <li className="px-4 py-3 text-body-sm text-muted">Sin resultados. Prueba con el nombre de una ciudad.</li>
          ) : (
            options.map((timezone, index) => {
              const selected = timezone.value === value
              return (
                <li
                  key={timezone.value}
                  id={optionId(index)}
                  data-index={index}
                  role="option"
                  aria-selected={selected}
                  onMouseDown={event => event.preventDefault()}
                  onClick={() => select(timezone)}
                  onMouseEnter={() => setHighlighted(index)}
                  className={cx(
                    'flex min-h-touch cursor-pointer items-center gap-3 rounded-sm px-3 py-2',
                    index === highlighted && 'bg-sunken',
                  )}
                >
                  <span className="flex min-w-0 flex-1 flex-col">
                    <span className="text-body-sm">{timezone.label}</span>
                    <span className="font-mono text-xs text-muted">{timezone.value}</span>
                  </span>
                  {selected && <Check size={20} strokeWidth={1.8} aria-hidden="true" className="text-ok-fg" />}
                </li>
              )
            })
          )}
        </ul>
      )}
    </div>
  )
}

export default TimezoneDropdown
