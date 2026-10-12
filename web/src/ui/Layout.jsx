import { forwardRef } from 'react'
import { ChevronLeft } from 'lucide-react'
import { Link } from 'react-router-dom'
import { cx } from '../lib/zensor'

export function Page({ className, children }) {
  return <div className={cx('mx-auto flex w-full max-w-content flex-col gap-8', className)}>{children}</div>
}

export function PageHeader({ title, description, back, actions, children }) {
  return (
    <header className="flex flex-col gap-4">
      {back && (
        <Link
          to={back.to}
          className="-ml-2 inline-flex min-h-touch w-fit items-center gap-1 rounded-pill px-2 text-label text-muted no-underline hover:text-ink"
        >
          <ChevronLeft size={20} strokeWidth={1.8} aria-hidden="true" />
          {back.label}
        </Link>
      )}
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex min-w-0 flex-col gap-2">
          <h1 className="m-0 text-h1-mobile md:text-h1">{title}</h1>
          {description && <p className="zs-prose m-0 text-muted">{description}</p>}
        </div>
        {actions && <div className="flex flex-wrap items-center gap-3">{actions}</div>}
      </div>
      {children}
    </header>
  )
}

export function Section({ title, description, actions, className, children, as: Tag = 'section' }) {
  return (
    <Tag className={cx('flex flex-col gap-4', className)}>
      {(title || actions) && (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex min-w-0 flex-col gap-1">
            {title && <h2 className="m-0 text-h3">{title}</h2>}
            {description && <p className="m-0 text-body-sm text-muted">{description}</p>}
          </div>
          {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
        </div>
      )}
      {children}
    </Tag>
  )
}

export const Card = forwardRef(function Card(
  { as: Tag = 'div', interactive = false, row = false, className, children, ...props },
  ref,
) {
  return (
    <Tag
      ref={ref}
      className={cx(
        'flex gap-4 rounded-xl border border-line bg-surface p-5 text-left text-ink',
        row ? 'flex-row items-center' : 'flex-col',
        interactive &&
          'w-full cursor-pointer font-[inherit] transition duration-instant ease-standard hover:border-line-strong active:scale-[.98]',
        className,
      )}
      {...props}
    >
      {children}
    </Tag>
  )
})

export function CardGrid({ className, children }) {
  return <div className={cx('flex flex-wrap gap-5 [&>*]:min-w-0 [&>*]:flex-[1_1_440px]', className)}>{children}</div>
}

export function List({ className, children }) {
  return <ul className={cx('m-0 flex list-none flex-col gap-3 p-0', className)}>{children}</ul>
}

export function KeyValue({ items, className }) {
  return (
    <dl className={cx('m-0 grid grid-cols-[auto_1fr] gap-x-6 gap-y-3', className)}>
      {items
        .filter(item => item && item.value !== undefined && item.value !== null && item.value !== '')
        .map(item => (
          <div key={item.label} className="contents">
            <dt className="text-body-sm text-muted">{item.label}</dt>
            <dd className={cx('m-0 min-w-0 break-words', item.mono && 'font-mono text-md')}>{item.value}</dd>
          </div>
        ))}
    </dl>
  )
}
