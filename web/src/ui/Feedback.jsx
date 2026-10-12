import { Loader2 } from 'lucide-react'
import { cx } from '../lib/zensor'
import Button from './Button'

const PILL_TONES = {
  ok: 'bg-ok-bg text-ok-fg',
  attention: 'bg-warn-bg text-warn-fg',
  live: 'bg-live-bg text-live-fg',
  neutral: 'bg-sunken text-muted',
  inverse: 'bg-inverse text-on-inverse',
}

export function StatusPill({ tone = 'neutral', className, children }) {
  return (
    <span
      className={cx(
        'inline-flex shrink-0 items-center whitespace-nowrap rounded-pill px-3 py-1 text-xs font-bold',
        PILL_TONES[tone],
        className,
      )}
    >
      {children}
    </span>
  )
}

const NOTICE_TONES = {
  attention: 'bg-warn-bg text-warn-fg',
  ok: 'bg-ok-bg text-ok-fg',
  live: 'bg-live-bg text-live-fg',
  neutral: 'bg-sunken text-ink',
}

export function Notice({ tone = 'attention', action, onAction, role, className, children }) {
  return (
    <div
      role={role ?? (tone === 'attention' ? 'alert' : 'status')}
      className={cx('flex flex-wrap items-center gap-4 rounded-lg px-5 py-3', NOTICE_TONES[tone], className)}
    >
      <p className="m-0 min-w-0 flex-1">{children}</p>
      {action && (
        <button
          type="button"
          onClick={onAction}
          className="min-h-touch rounded-pill border-control border-current bg-transparent px-4 text-md font-bold text-current transition duration-instant ease-standard active:scale-[.98]"
        >
          {action}
        </button>
      )}
    </div>
  )
}

export function EmptyState({ title, children, action, className }) {
  return (
    <div className={cx('rounded-xl border border-dashed border-line-strong px-6 py-8 text-muted', className)}>
      {title && <p className="m-0 mb-1 text-h3 text-ink">{title}</p>}
      <p className="zs-prose m-0">{children}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}

export function Spinner({ size = 20, className }) {
  return (
    <Loader2
      size={size}
      strokeWidth={1.8}
      aria-hidden="true"
      className={cx('animate-spin motion-reduce:animate-none', className)}
    />
  )
}

export function LoadingState({ children = 'Cargando…', className }) {
  return (
    <div role="status" className={cx('flex items-center gap-3 py-8 text-muted', className)}>
      <Spinner />
      <span>{children}</span>
    </div>
  )
}

export function ErrorState({ title = 'No se pudo cargar la información', children, onRetry, retryLabel = 'Reintentar' }) {
  return (
    <div role="alert" className="flex flex-col items-start gap-3 rounded-xl bg-warn-bg px-6 py-5 text-warn-fg">
      <p className="m-0 text-h3">{title}</p>
      {children && <p className="zs-prose m-0">{children}</p>}
      {onRetry && (
        <Button variant="secondary" onClick={onRetry} className="border-warn-fg text-warn-fg">
          {retryLabel}
        </Button>
      )}
    </div>
  )
}
