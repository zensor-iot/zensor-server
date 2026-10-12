import { forwardRef } from 'react'
import { Link } from 'react-router-dom'
import { cx } from '../lib/zensor'

const VARIANTS = {
  primary: 'bg-primary text-on-primary hover:bg-primary-hover',
  secondary: 'border-control border-brand bg-transparent text-ink hover:bg-sunken',
  ghost: 'bg-transparent text-ink hover:bg-sunken',
  danger: 'bg-transparent text-attention hover:bg-warn-bg',
  accent: 'bg-accent-inverse text-on-accent-inverse',
  inverse: 'border-control border-on-inverse-muted bg-transparent text-on-inverse',
}

const SIZES = {
  md: 'min-h-touch px-5 text-md',
  lg: 'min-h-control px-6 text-base',
  sm: 'min-h-touch px-4 text-md',
}

const BASE =
  'inline-flex items-center justify-center gap-2 rounded-pill font-bold no-underline transition duration-instant ease-standard active:scale-[.98] disabled:cursor-not-allowed disabled:opacity-disabled disabled:active:scale-100'

export const buttonClass = ({ variant = 'primary', size = 'md', block = false, className } = {}) =>
  cx(BASE, VARIANTS[variant], SIZES[size], block && 'w-full', className)

const Button = forwardRef(function Button(
  { variant = 'primary', size = 'md', block = false, icon: Icon, to, href, type = 'button', className, children, ...props },
  ref,
) {
  const classes = buttonClass({ variant, size, block, className })
  const content = (
    <>
      {Icon && <Icon size={20} strokeWidth={1.8} aria-hidden="true" />}
      {children}
    </>
  )

  if (to) {
    return (
      <Link ref={ref} to={to} className={classes} {...props}>
        {content}
      </Link>
    )
  }
  if (href) {
    return (
      <a ref={ref} href={href} className={classes} {...props}>
        {content}
      </a>
    )
  }
  return (
    <button ref={ref} type={type} className={classes} {...props}>
      {content}
    </button>
  )
})

export const IconButton = forwardRef(function IconButton(
  { icon: Icon, label, variant = 'ghost', className, type = 'button', ...props },
  ref,
) {
  return (
    <button
      ref={ref}
      type={type}
      aria-label={label}
      title={label}
      className={cx(
        'inline-flex min-h-touch min-w-touch items-center justify-center rounded-pill transition duration-instant ease-standard active:scale-[.98] disabled:cursor-not-allowed disabled:opacity-disabled',
        variant === 'ghost' && 'text-ink hover:bg-sunken',
        variant === 'danger' && 'text-attention hover:bg-warn-bg',
        variant === 'inverse' && 'text-on-inverse hover:bg-inverse-raised',
        className,
      )}
      {...props}
    >
      <Icon size={20} strokeWidth={1.8} aria-hidden="true" />
    </button>
  )
})

export default Button
