import { useEffect, useId, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { cx, prefersReducedMotion } from '../lib/zensor'
import { IconButton } from './Button'
import Button from './Button'

const EXIT_MS = 140
const overlayStack = []
const FOCUSABLE = 'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

function useOverlay(open, onClose, panelRef, initialFocusRef) {
  const [mounted, setMounted] = useState(open)
  const [closing, setClosing] = useState(false)
  const onCloseRef = useRef(onClose)
  onCloseRef.current = onClose

  useEffect(() => {
    if (open) {
      setMounted(true)
      setClosing(false)
      return undefined
    }
    if (!mounted) return undefined
    if (prefersReducedMotion()) {
      setMounted(false)
      return undefined
    }
    setClosing(true)
    const timer = setTimeout(() => {
      setMounted(false)
      setClosing(false)
    }, EXIT_MS)
    return () => clearTimeout(timer)
  }, [open, mounted])

  useEffect(() => {
    if (!open) return undefined
    const previous = document.activeElement
    const panel = panelRef.current
    const target = initialFocusRef?.current || panel?.querySelector('[data-autofocus]') || panel?.querySelector(FOCUSABLE)
    target?.focus()

    const token = {}
    overlayStack.push(token)

    const onKeyDown = event => {
      if (overlayStack[overlayStack.length - 1] !== token) return
      if (event.key === 'Escape') {
        event.stopPropagation()
        onCloseRef.current()
        return
      }
      if (event.key !== 'Tab' || !panel) return
      const focusable = [...panel.querySelectorAll(FOCUSABLE)]
      if (focusable.length === 0) return
      const first = focusable[0]
      const last = focusable[focusable.length - 1]
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault()
        last.focus()
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault()
        first.focus()
      }
    }

    const { overflow } = document.body.style
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', onKeyDown)
    return () => {
      overlayStack.splice(overlayStack.indexOf(token), 1)
      document.removeEventListener('keydown', onKeyDown)
      document.body.style.overflow = overflow
      if (previous && typeof previous.focus === 'function') previous.focus()
    }
  }, [open, panelRef, initialFocusRef])

  return { mounted, closing }
}

/**
 * Right-side panel for detail views and secondary flows. Traps focus, closes on Escape
 * and scrim click, and restores focus to the element that opened it.
 */
export default function Drawer({ open, onClose, title, subtitle, actions, footer, initialFocusRef, children }) {
  const panelRef = useRef(null)
  const titleId = useId()
  const { mounted, closing } = useOverlay(open, onClose, panelRef, initialFocusRef)

  if (!mounted) return null

  return createPortal(
    <>
      <div
        aria-hidden="true"
        onClick={onClose}
        className={cx(
          'fixed inset-0 z-scrim bg-scrim',
          closing ? 'opacity-0 transition-opacity duration-fast ease-accelerate' : 'animate-scrim-in',
        )}
      />
      <aside
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className={cx(
          'fixed inset-y-0 right-0 z-drawer flex w-drawer max-w-full flex-col bg-canvas text-ink shadow-drawer',
          closing ? 'zs-anim-drawer-out' : 'animate-drawer-in',
        )}
      >
        <header className="flex shrink-0 items-start gap-3 border-b border-line px-6 py-5">
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <h2 id={titleId} className="m-0 break-words text-h2">
              {title}
            </h2>
            {subtitle && <div className="flex flex-wrap items-center gap-2 text-body-sm text-muted">{subtitle}</div>}
          </div>
          {actions}
          <IconButton icon={X} label="Cerrar" onClick={onClose} data-autofocus />
        </header>
        <div className="flex flex-1 flex-col gap-8 overflow-y-auto px-6 py-6">{children}</div>
        {footer && <footer className="flex shrink-0 flex-wrap gap-3 border-t border-line px-6 py-4">{footer}</footer>}
      </aside>
    </>,
    document.body,
  )
}

/**
 * Centered dialog reserved for short confirmations (e.g. destructive actions).
 */
export function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  children,
  confirmLabel,
  cancelLabel = 'Cancelar',
  destructive = false,
  busy = false,
}) {
  const panelRef = useRef(null)
  const cancelRef = useRef(null)
  const titleId = useId()
  const { mounted, closing } = useOverlay(open, onClose, panelRef, cancelRef)

  if (!mounted) return null

  return createPortal(
    <div className="fixed inset-0 z-drawer flex items-center justify-center p-4">
      <div
        aria-hidden="true"
        onClick={onClose}
        className={cx(
          'absolute inset-0 bg-scrim',
          closing ? 'opacity-0 transition-opacity duration-fast ease-accelerate' : 'animate-scrim-in',
        )}
      />
      <div
        ref={panelRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className={cx(
          'relative flex w-full max-w-[28rem] flex-col gap-4 rounded-xl bg-surface p-6 text-ink shadow-modal',
          closing ? 'opacity-0 transition-opacity duration-fast ease-accelerate' : 'animate-scrim-in',
        )}
      >
        <h2 id={titleId} className="m-0 text-h3">
          {title}
        </h2>
        {children && <div className="text-body-sm text-muted">{children}</div>}
        <div className="flex flex-wrap justify-end gap-3 pt-2">
          <Button ref={cancelRef} variant="ghost" onClick={onClose} disabled={busy}>
            {cancelLabel}
          </Button>
          <Button
            variant={destructive ? 'secondary' : 'primary'}
            className={destructive ? 'border-warn-fill text-attention hover:bg-warn-bg' : undefined}
            onClick={onConfirm}
            disabled={busy}
          >
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>,
    document.body,
  )
}
