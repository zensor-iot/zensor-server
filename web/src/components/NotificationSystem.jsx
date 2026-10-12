import { createContext, useCallback, useContext, useEffect, useMemo, useReducer, useState } from 'react'
import { AlertTriangle, CheckCircle2, CircleAlert, X } from 'lucide-react'
import { cx } from '../lib/zensor'

export const NOTIFICATION_TYPES = {
  SUCCESS: 'success',
  ERROR: 'error',
  WARNING: 'warning',
}

const NotificationContext = createContext(null)

const notificationReducer = (state, action) => {
  switch (action.type) {
    case 'ADD_NOTIFICATION':
      return [...state, { id: Date.now() + Math.random(), ...action.payload }]
    case 'REMOVE_NOTIFICATION':
      return state.filter(notification => notification.id !== action.payload)
    case 'CLEAR_ALL':
      return []
    default:
      return state
  }
}

export const NotificationProvider = ({ children }) => {
  const [notifications, dispatch] = useReducer(notificationReducer, [])

  const addNotification = useCallback(notification => dispatch({ type: 'ADD_NOTIFICATION', payload: notification }), [])
  const removeNotification = useCallback(id => dispatch({ type: 'REMOVE_NOTIFICATION', payload: id }), [])
  const clearAllNotifications = useCallback(() => dispatch({ type: 'CLEAR_ALL' }), [])

  const value = useMemo(
    () => ({ notifications, addNotification, removeNotification, clearAllNotifications }),
    [notifications, addNotification, removeNotification, clearAllNotifications],
  )

  return (
    <NotificationContext.Provider value={value}>
      {children}
      <NotificationContainer notifications={notifications} onRemove={removeNotification} />
    </NotificationContext.Provider>
  )
}

export const useNotifications = () => {
  const context = useContext(NotificationContext)
  if (!context) {
    throw new Error('useNotifications must be used within a NotificationProvider')
  }
  return context
}

const TONES = {
  [NOTIFICATION_TYPES.SUCCESS]: { icon: CheckCircle2, className: 'bg-ok-bg text-ok-fg' },
  [NOTIFICATION_TYPES.ERROR]: { icon: CircleAlert, className: 'bg-warn-bg text-warn-fg' },
  [NOTIFICATION_TYPES.WARNING]: { icon: AlertTriangle, className: 'bg-warn-bg text-warn-fg' },
}

const Notification = ({ notification, onRemove }) => {
  const [paused, setPaused] = useState(false)
  const tone = TONES[notification.type] || TONES[NOTIFICATION_TYPES.SUCCESS]
  const Icon = tone.icon

  useEffect(() => {
    if (paused || notification.autoHide === false) return undefined
    const timer = setTimeout(() => onRemove(notification.id), notification.duration || 5000)
    return () => clearTimeout(timer)
  }, [paused, notification, onRemove])

  return (
    <div
      role={notification.type === NOTIFICATION_TYPES.SUCCESS ? 'status' : 'alert'}
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      onFocus={() => setPaused(true)}
      onBlur={() => setPaused(false)}
      className={cx(
        'pointer-events-auto flex w-full items-start gap-3 rounded-lg py-3 pl-4 pr-1 shadow-overlay animate-scrim-in',
        tone.className,
      )}
    >
      <Icon size={20} strokeWidth={1.8} aria-hidden="true" className="mt-[2px] shrink-0" />
      <div className="flex min-w-0 flex-1 flex-col gap-1 py-[2px]">
        {notification.title && <p className="m-0 font-bold">{notification.title}</p>}
        <p className="m-0 text-body-sm">{notification.message}</p>
      </div>
      <button
        type="button"
        onClick={() => onRemove(notification.id)}
        aria-label="Cerrar aviso"
        className="-my-2 inline-flex min-h-touch min-w-touch shrink-0 items-center justify-center rounded-pill text-current"
      >
        <X size={18} strokeWidth={1.8} aria-hidden="true" />
      </button>
    </div>
  )
}

const NotificationContainer = ({ notifications, onRemove }) => (
  <div
    aria-live="polite"
    className="pointer-events-none fixed bottom-4 right-4 z-toast flex w-[min(24rem,calc(100vw-2rem))] flex-col gap-3"
  >
    {notifications.map(notification => (
      <Notification key={notification.id} notification={notification} onRemove={onRemove} />
    ))}
  </div>
)
