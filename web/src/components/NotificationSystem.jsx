import React, { createContext, useContext, useReducer, useEffect, useState } from 'react';
import { X, CheckCircle, AlertCircle, AlertTriangle } from 'lucide-react';
import './NotificationSystem.css';

// Notification types
export const NOTIFICATION_TYPES = {
    SUCCESS: 'success',
    ERROR: 'error',
    WARNING: 'warning'
};

// Notification context
const NotificationContext = createContext();

// Reducer for managing notifications
const notificationReducer = (state, action) => {
    switch (action.type) {
        case 'ADD_NOTIFICATION':
            return {
                ...state,
                notifications: [...state.notifications, {
                    id: Date.now() + Math.random(),
                    ...action.payload,
                    createdAt: Date.now()
                }]
            };
        case 'REMOVE_NOTIFICATION':
            return {
                ...state,
                notifications: state.notifications.filter(notification => notification.id !== action.payload)
            };
        case 'CLEAR_ALL':
            return {
                ...state,
                notifications: []
            };
        default:
            return state;
    }
};

// Notification provider component
export const NotificationProvider = ({ children }) => {
    const [state, dispatch] = useReducer(notificationReducer, {
        notifications: []
    });

    const addNotification = (notification) => {
        dispatch({
            type: 'ADD_NOTIFICATION',
            payload: notification
        });
    };

    const removeNotification = (id) => {
        dispatch({
            type: 'REMOVE_NOTIFICATION',
            payload: id
        });
    };

    const clearAllNotifications = () => {
        dispatch({ type: 'CLEAR_ALL' });
    };

    return (
        <NotificationContext.Provider value={{
            notifications: state.notifications,
            addNotification,
            removeNotification,
            clearAllNotifications
        }}>
            {children}
            <NotificationContainer />
        </NotificationContext.Provider>
    );
};

// Hook to use notifications
export const useNotifications = () => {
    const context = useContext(NotificationContext);
    if (!context) {
        throw new Error('useNotifications must be used within a NotificationProvider');
    }
    return context;
};

// Individual notification component
const Notification = ({ notification, onRemove }) => {
    const [isHovered, setIsHovered] = useState(false);
    const [isVisible, setIsVisible] = useState(false);

    useEffect(() => {
        // Animate in
        const timer = setTimeout(() => setIsVisible(true), 10);
        return () => clearTimeout(timer);
    }, []);

    useEffect(() => {
        if (!isHovered && notification.autoHide !== false) {
            const timer = setTimeout(() => {
                onRemove(notification.id);
            }, notification.duration || 5000);

            return () => clearTimeout(timer);
        }
    }, [isHovered, notification, onRemove]);

    const getIcon = () => {
        switch (notification.type) {
            case NOTIFICATION_TYPES.SUCCESS:
                return <CheckCircle size={20} />;
            case NOTIFICATION_TYPES.ERROR:
                return <AlertCircle size={20} />;
            case NOTIFICATION_TYPES.WARNING:
                return <AlertTriangle size={20} />;
            default:
                return <CheckCircle size={20} />;
        }
    };

    const handleRemove = () => {
        setIsVisible(false);
        setTimeout(() => onRemove(notification.id), 200);
    };

    return (
        <div
            className={`notification ${notification.type} ${isVisible ? 'visible' : ''}`}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
        >
            <div className="notification-icon">
                {getIcon()}
            </div>
            <div className="notification-content">
                {notification.title && (
                    <div className="notification-title">{notification.title}</div>
                )}
                <div className="notification-message">{notification.message}</div>
            </div>
            <button
                className="notification-close"
                onClick={handleRemove}
                aria-label="Close notification"
            >
                <X size={16} />
            </button>
        </div>
    );
};

// Notification container component
const NotificationContainer = () => {
    const { notifications, removeNotification } = useNotifications();

    return (
        <div className="notification-container">
            {notifications.map(notification => (
                <Notification
                    key={notification.id}
                    notification={notification}
                    onRemove={removeNotification}
                />
            ))}
        </div>
    );
};

