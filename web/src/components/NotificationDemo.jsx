import React from 'react';
import { useNotification } from '../hooks/useNotification';
import { Bell, Zap, AlertTriangle, CheckCircle } from 'lucide-react';

const NotificationDemo = () => {
    const { showSuccess, showError, showWarning, showNotification, showApiNotification } = useNotification();

    const handleSuccessDemo = () => {
        showSuccess('This is a success notification!', 'Success');
    };

    const handleErrorDemo = () => {
        showError('This is an error notification!', 'Error');
    };

    const handleWarningDemo = () => {
        showWarning('This is a warning notification!', 'Warning');
    };

    const handleCustomDemo = () => {
        showNotification({
            type: 'success',
            title: 'Custom Notification',
            message: 'This is a custom notification with longer duration',
            duration: 10000,
            autoHide: false
        });
    };

    const handleApiDemo = async () => {
        // Simulate an API call
        const mockApiCall = new Promise((resolve, reject) => {
            setTimeout(() => {
                if (Math.random() > 0.5) {
                    resolve('Data saved successfully');
                } else {
                    reject(new Error('Failed to save data'));
                }
            }, 1000);
        });

        await showApiNotification(
            mockApiCall,
            'API operation completed successfully!',
            'API operation failed',
            'Success',
            'Error'
        );
    };

    const handlePersistentDemo = () => {
        showNotification({
            type: 'warning',
            title: 'Persistent Notification',
            message: 'This notification will stay until you close it manually',
            autoHide: false
        });
    };

    return (
        <div className="notification-demo">
            <div className="demo-header">
                <Bell className="demo-icon" />
                <h2>Notification System Demo</h2>
                <p>Click the buttons below to test different notification types</p>
            </div>

            <div className="demo-buttons">
                <button
                    className="demo-btn success"
                    onClick={handleSuccessDemo}
                >
                    <CheckCircle size={16} />
                    Success Notification
                </button>

                <button
                    className="demo-btn error"
                    onClick={handleErrorDemo}
                >
                    <AlertTriangle size={16} />
                    Error Notification
                </button>

                <button
                    className="demo-btn warning"
                    onClick={handleWarningDemo}
                >
                    <AlertTriangle size={16} />
                    Warning Notification
                </button>

                <button
                    className="demo-btn custom"
                    onClick={handleCustomDemo}
                >
                    <Zap size={16} />
                    Custom Notification
                </button>

                <button
                    className="demo-btn api"
                    onClick={handleApiDemo}
                >
                    <Zap size={16} />
                    API Demo (Random Success/Failure)
                </button>

                <button
                    className="demo-btn persistent"
                    onClick={handlePersistentDemo}
                >
                    <Bell size={16} />
                    Persistent Notification
                </button>
            </div>

            <div className="demo-features">
                <h3>Features:</h3>
                <ul>
                    <li>✅ Different colors for success (green), error (red), and warning (yellow)</li>
                    <li>✅ Auto-dismiss after a few seconds</li>
                    <li>✅ Pause auto-dismiss on hover</li>
                    <li>✅ Manual close button</li>
                    <li>✅ Smooth animations</li>
                    <li>✅ Mobile responsive</li>
                    <li>✅ Stack multiple notifications</li>
                    <li>✅ Custom duration and persistence options</li>
                </ul>
            </div>
        </div>
    );
};

export default NotificationDemo; 