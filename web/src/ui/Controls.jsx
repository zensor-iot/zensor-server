import { cx } from '../lib/zensor'

export function FilterChips({ label, options, value, onChange, className }) {
  return (
    <div role="group" aria-label={label} className={cx('flex flex-wrap gap-2', className)}>
      {options.map(option => {
        const pressed = option.value === value
        return (
          <button
            key={option.value}
            type="button"
            aria-pressed={pressed}
            onClick={() => onChange(option.value)}
            className={cx(
              'inline-flex min-h-touch items-center gap-2 rounded-pill border-control border-brand px-4 text-md font-bold transition duration-instant ease-standard active:scale-[.98]',
              pressed ? 'bg-primary text-on-primary' : 'bg-transparent text-ink hover:bg-sunken',
            )}
          >
            {option.label}
            {option.count !== undefined && <span className="font-mono font-medium">{option.count}</span>}
          </button>
        )
      })}
    </div>
  )
}

export function Switch({ checked, onChange, label, disabled, className }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      title={label}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={cx(
        'group inline-flex min-h-touch w-[56px] shrink-0 items-center justify-center rounded-pill disabled:cursor-not-allowed disabled:opacity-disabled',
        className,
      )}
    >
      <span
        aria-hidden="true"
        className={cx(
          'relative inline-block h-[26px] w-[44px] rounded-pill transition-colors duration-fast ease-standard',
          checked ? 'bg-ok-fill' : 'bg-line-strong',
        )}
      >
        <span
          className={cx(
            'absolute left-[3px] top-[3px] h-[20px] w-[20px] rounded-pill bg-surface transition-transform duration-fast ease-standard',
            checked && 'translate-x-[18px]',
          )}
        />
      </span>
    </button>
  )
}

export function Meter({ value, max = 100, fill = 'var(--zs-status-ok-fill)', target, label, className }) {
  const percent = Math.max(0, Math.min(100, (Number(value) / max) * 100 || 0))
  const targetPercent = target !== undefined ? Math.max(0, Math.min(100, (target / max) * 100)) : undefined
  return (
    <span
      role="meter"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      aria-label={label}
      className={cx('relative block h-2 overflow-hidden rounded-sm bg-track', className)}
    >
      <span
        className="block h-full transition-[width] duration-slower ease-standard"
        style={{ width: `${percent}%`, backgroundColor: fill }}
      />
      {targetPercent !== undefined && (
        <span className="absolute top-0 h-full w-[2px] bg-brand" style={{ left: `${targetPercent}%` }} />
      )}
    </span>
  )
}
