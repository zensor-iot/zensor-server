import { PawPrint, User } from 'lucide-react'
import { cx } from '../../lib/zensor'

export default function PatientKindIcon({ kind, className }) {
  const Icon = kind === 'animal' ? PawPrint : User
  return (
    <span
      aria-hidden="true"
      className={cx('inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-pill bg-sunken text-muted', className)}
    >
      <Icon size={20} strokeWidth={1.8} />
    </span>
  )
}
