import { useNotifications, NOTIFICATION_TYPES } from '../components/NotificationSystem';

/**
 * Custom hook for easy notification usage
 * Provides convenient methods for showing different types of notifications
 */
export const useNotification = () => {
    const { addNotification } = useNotifications();

    /**
     * Show a success notification
     * @param {string} message - The notification message
     * @param {string} title - Optional title for the notification
     * @param {Object} options - Additional options (duration, autoHide)
     */
    const showSuccess = (message, title = null, options = {}) => {
        addNotification({
            type: NOTIFICATION_TYPES.SUCCESS,
            message,
            title,
            duration: 4000,
            ...options
        });
    };

    /**
     * Show an error notification
     * @param {string} message - The notification message
     * @param {string} title - Optional title for the notification
     * @param {Object} options - Additional options (duration, autoHide)
     */
    const showError = (message, title = null, options = {}) => {
        addNotification({
            type: NOTIFICATION_TYPES.ERROR,
            message,
            title,
            duration: 6000, // Longer duration for errors
            ...options
        });
    };

    /**
     * Show a warning notification
     * @param {string} message - The notification message
     * @param {string} title - Optional title for the notification
     * @param {Object} options - Additional options (duration, autoHide)
     */
    const showWarning = (message, title = null, options = {}) => {
        addNotification({
            type: NOTIFICATION_TYPES.WARNING,
            message,
            title,
            duration: 5000,
            ...options
        });
    };

    /**
     * Show a custom notification
     * @param {Object} notification - Complete notification object
     */
    const showNotification = (notification) => {
        addNotification(notification);
    };

    /**
     * Show a notification for API operations
     * @param {Promise} promise - The API promise
     * @param {string} successMessage - Message to show on success
     * @param {string} errorMessage - Message to show on error
     * @param {string} successTitle - Optional success title
     * @param {string} errorTitle - Optional error title
     */
    const showApiNotification = async (
        promise,
        successMessage = 'Operation completed successfully',
        errorMessage = 'An error occurred',
        successTitle = null,
        errorTitle = null
    ) => {
        try {
            const result = await promise;
            showSuccess(successMessage, successTitle);
            return { success: true, data: result };
        } catch (error) {
            const errorMsg = error.message || errorMessage;
            showError(errorMsg, errorTitle);
            return { success: false, error };
        }
    };

    return {
        showSuccess,
        showError,
        showWarning,
        showNotification,
        showApiNotification,
        NOTIFICATION_TYPES
    };
}; 