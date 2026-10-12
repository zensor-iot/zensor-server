import lockupSvg from '@design-system/logos/zensor-lockup.svg?raw'
import markSvg from '@design-system/logos/zensor-mark.svg?raw'
import { cx } from '../lib/zensor'

const themed = svg =>
  svg
    .replaceAll('fill="#2B3A2A"', 'fill="var(--zs-brand-ink)"')
    .replaceAll('fill="#B9683E"', 'fill="var(--zs-accent)"')
    .replace('<svg ', '<svg width="100%" height="100%" ')

const LOCKUP = themed(lockupSvg)
const MARK = themed(markSvg)

export function Lockup({ className }) {
  return <span className={cx('block h-8 w-[120px]', className)} dangerouslySetInnerHTML={{ __html: LOCKUP }} />
}

export function Mark({ className }) {
  return <span className={cx('block h-8 w-8', className)} dangerouslySetInnerHTML={{ __html: MARK }} />
}
