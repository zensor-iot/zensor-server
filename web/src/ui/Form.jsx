import { forwardRef, useId } from 'react'
import { ChevronDown } from 'lucide-react'
import { cx } from '../lib/zensor'

export function Field({ label, hint, error, required, className, children }) {
  const id = useId()
  const hintId = hint ? `${id}-hint` : undefined
  const errorId = error ? `${id}-error` : undefined
  const describedBy = [hintId, errorId].filter(Boolean).join(' ') || undefined

  return (
    <div className={cx('flex flex-col gap-2', className)}>
      <label htmlFor={id} className="text-label">
        {label}
        {required && <span className="text-muted"> (obligatorio)</span>}
      </label>
      {children({ id, 'aria-describedby': describedBy, 'aria-invalid': error ? true : undefined, required })}
      {hint && (
        <p id={hintId} className="m-0 text-caption text-muted">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} className="m-0 text-caption text-attention">
          {error}
        </p>
      )}
    </div>
  )
}

export const Input = forwardRef(function Input({ className, ...props }, ref) {
  return <input ref={ref} className={cx('zs-control', className)} {...props} />
})

export const Textarea = forwardRef(function Textarea({ className, ...props }, ref) {
  return <textarea ref={ref} className={cx('zs-control', className)} {...props} />
})

export const Select = forwardRef(function Select({ className, children, ...props }, ref) {
  return (
    <div className="relative">
      <select ref={ref} className={cx('zs-control', className)} {...props}>
        {children}
      </select>
      <ChevronDown
        size={18}
        strokeWidth={1.8}
        aria-hidden="true"
        className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-muted"
      />
    </div>
  )
})

export function Checkbox({ label, description, className, ...props }) {
  const id = useId()
  return (
    <label htmlFor={id} className={cx('flex min-h-touch cursor-pointer items-center gap-3', className)}>
      <input id={id} type="checkbox" className="zs-check shrink-0" {...props} />
      <span className="flex flex-col">
        <span>{label}</span>
        {description && <span className="text-caption text-muted">{description}</span>}
      </span>
    </label>
  )
}

export function FormActions({ className, children }) {
  return <div className={cx('flex flex-wrap items-center gap-3 pt-2', className)}>{children}</div>
}

export function Form({ className, children, ...props }) {
  return (
    <form className={cx('flex flex-col gap-5', className)} {...props}>
      {children}
    </form>
  )
}
