import { cx } from '../lib/zensor'

export function Table({ caption, className, children }) {
  return (
    <div className={cx('overflow-x-auto rounded-xl border border-line bg-surface', className)}>
      <table className="zs-table">
        {caption && <caption className="sr-only">{caption}</caption>}
        {children}
      </table>
    </div>
  )
}
