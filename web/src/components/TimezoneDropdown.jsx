import { useState, useRef, useEffect } from 'react'
import { Search, Clock, ChevronDown } from 'lucide-react'

// Comprehensive list of major IANA timezones
const TIMEZONES = [
    { value: 'America/New_York', label: 'Eastern Time (US & Canada)' },
    { value: 'America/Chicago', label: 'Central Time (US & Canada)' },
    { value: 'America/Denver', label: 'Mountain Time (US & Canada)' },
    { value: 'America/Los_Angeles', label: 'Pacific Time (US & Canada)' },
    { value: 'America/Phoenix', label: 'Arizona' },
    { value: 'America/Anchorage', label: 'Alaska' },
    { value: 'America/Toronto', label: 'Toronto' },
    { value: 'America/Mexico_City', label: 'Mexico City' },
    { value: 'America/Sao_Paulo', label: 'São Paulo' },
    { value: 'America/Buenos_Aires', label: 'Buenos Aires' },
    { value: 'America/Santiago', label: 'Santiago' },
    { value: 'America/Lima', label: 'Lima' },
    { value: 'Europe/London', label: 'London' },
    { value: 'Europe/Paris', label: 'Paris' },
    { value: 'Europe/Berlin', label: 'Berlin' },
    { value: 'Europe/Madrid', label: 'Madrid' },
    { value: 'Europe/Rome', label: 'Rome' },
    { value: 'Europe/Amsterdam', label: 'Amsterdam' },
    { value: 'Europe/Brussels', label: 'Brussels' },
    { value: 'Europe/Stockholm', label: 'Stockholm' },
    { value: 'Europe/Vienna', label: 'Vienna' },
    { value: 'Europe/Zurich', label: 'Zurich' },
    { value: 'Europe/Athens', label: 'Athens' },
    { value: 'Europe/Dublin', label: 'Dublin' },
    { value: 'Europe/Prague', label: 'Prague' },
    { value: 'Europe/Warsaw', label: 'Warsaw' },
    { value: 'Europe/Moscow', label: 'Moscow' },
    { value: 'Europe/Istanbul', label: 'Istanbul' },
    { value: 'Asia/Dubai', label: 'Dubai' },
    { value: 'Asia/Tokyo', label: 'Tokyo' },
    { value: 'Asia/Shanghai', label: 'Shanghai' },
    { value: 'Asia/Hong_Kong', label: 'Hong Kong' },
    { value: 'Asia/Singapore', label: 'Singapore' },
    { value: 'Asia/Bangkok', label: 'Bangkok' },
    { value: 'Asia/Seoul', label: 'Seoul' },
    { value: 'Asia/Delhi', label: 'India' },
    { value: 'Asia/Karachi', label: 'Karachi' },
    { value: 'Asia/Jakarta', label: 'Jakarta' },
    { value: 'Asia/Manila', label: 'Manila' },
    { value: 'Asia/Dhaka', label: 'Dhaka' },
    { value: 'Asia/Jerusalem', label: 'Jerusalem' },
    { value: 'Asia/Baghdad', label: 'Baghdad' },
    { value: 'Australia/Sydney', label: 'Sydney' },
    { value: 'Australia/Melbourne', label: 'Melbourne' },
    { value: 'Australia/Brisbane', label: 'Brisbane' },
    { value: 'Australia/Perth', label: 'Perth' },
    { value: 'Pacific/Auckland', label: 'Auckland' },
    { value: 'Pacific/Fiji', label: 'Fiji' },
    { value: 'Pacific/Honolulu', label: 'Honolulu' },
    { value: 'UTC', label: 'UTC' }
]

const TimezoneDropdown = ({ value, onChange, placeholder = 'Select timezone...' }) => {
    const [isOpen, setIsOpen] = useState(false)
    const [searchTerm, setSearchTerm] = useState('')
    const [highlightedIndex, setHighlightedIndex] = useState(0)
    const dropdownRef = useRef(null)
    const inputRef = useRef(null)

    // Filter timezones based on search term
    const filteredTimezones = TIMEZONES.filter(tz =>
        tz.value.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tz.label.toLowerCase().includes(searchTerm.toLowerCase())
    )

    // Get current timezone label
    const currentTimezone = TIMEZONES.find(tz => tz.value === value)
    const displayValue = currentTimezone ? currentTimezone.label : value || placeholder

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsOpen(false)
                setSearchTerm('')
                setHighlightedIndex(0)
            }
        }

        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [])

    useEffect(() => {
        if (isOpen && inputRef.current) {
            inputRef.current.focus()
        }
    }, [isOpen])

    useEffect(() => {
        setHighlightedIndex(0)
    }, [searchTerm])

    const handleSelect = (timezoneValue) => {
        onChange(timezoneValue)
        setIsOpen(false)
        setSearchTerm('')
        setHighlightedIndex(0)
    }

    const handleKeyDown = (e) => {
        if (!isOpen) {
            if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
                e.preventDefault()
                setIsOpen(true)
            }
            return
        }

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault()
                setHighlightedIndex(prev =>
                    prev < filteredTimezones.length - 1 ? prev + 1 : prev
                )
                break
            case 'ArrowUp':
                e.preventDefault()
                setHighlightedIndex(prev => prev > 0 ? prev - 1 : 0)
                break
            case 'Enter':
                e.preventDefault()
                if (filteredTimezones[highlightedIndex]) {
                    handleSelect(filteredTimezones[highlightedIndex].value)
                }
                break
            case 'Escape':
                e.preventDefault()
                setIsOpen(false)
                setSearchTerm('')
                break
            default:
                break
        }
    }

    // Scroll highlighted item into view
    useEffect(() => {
        if (isOpen && highlightedIndex >= 0 && highlightedIndex < filteredTimezones.length) {
            const element = document.getElementById(`timezone-option-${highlightedIndex}`)
            if (element) {
                element.scrollIntoView({ block: 'nearest' })
            }
        }
    }, [highlightedIndex, isOpen, filteredTimezones.length])

    return (
        <div className="timezone-dropdown" ref={dropdownRef}>
            <div
                className={`timezone-dropdown-trigger ${isOpen ? 'open' : ''}`}
                onClick={() => setIsOpen(!isOpen)}
                onKeyDown={handleKeyDown}
                tabIndex={0}
                role="combobox"
                aria-expanded={isOpen}
                aria-haspopup="listbox"
                aria-label="Timezone selector"
            >
                <Clock size={16} />
                <span className="timezone-dropdown-value">
                    {displayValue}
                </span>
                <ChevronDown size={16} className={`timezone-dropdown-chevron ${isOpen ? 'open' : ''}`} />
            </div>

            {isOpen && (
                <div className="timezone-dropdown-menu">
                    <div className="timezone-dropdown-search">
                        <Search size={16} />
                        <input
                            ref={inputRef}
                            type="text"
                            placeholder="Search timezones..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="timezone-dropdown-input"
                            autoComplete="off"
                        />
                    </div>

                    <div className="timezone-dropdown-list">
                        {filteredTimezones.length === 0 ? (
                            <div className="timezone-dropdown-empty">
                                No timezones found
                            </div>
                        ) : (
                            filteredTimezones.map((timezone, index) => (
                                <div
                                    key={timezone.value}
                                    id={`timezone-option-${index}`}
                                    className={`timezone-dropdown-item ${index === highlightedIndex ? 'highlighted' : ''
                                        } ${value === timezone.value ? 'selected' : ''}`}
                                    onClick={() => handleSelect(timezone.value)}
                                    onMouseEnter={() => setHighlightedIndex(index)}
                                >
                                    <div className="timezone-dropdown-item-value">{timezone.value}</div>
                                    <div className="timezone-dropdown-item-label">{timezone.label}</div>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            )}
        </div>
    )
}

export default TimezoneDropdown

