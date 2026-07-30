import React, { useState, useEffect } from 'react'
import {
    Clock,
    Droplets,
    Plus,
    Edit3,
    Trash2,
    Play,
    Pause,
    Calendar,
    X,
    Check,
    Settings,
    List
} from 'lucide-react'
import { scheduledTasksApi } from '../config/api'
import { useNotification } from '../hooks/useNotification'
import './ScheduledIrrigation.css'

const ScheduledIrrigationWithNotifications = ({ tenantId, deviceId, deviceName }) => {
    const [scheduledTasks, setScheduledTasks] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [showModal, setShowModal] = useState(false)
    const [editingTask, setEditingTask] = useState(null)
    const [showExecutionsModal, setShowExecutionsModal] = useState(false)
    const [selectedTask, setSelectedTask] = useState(null)
    const [taskExecutions, setTaskExecutions] = useState([])
    const [executionsLoading, setExecutionsLoading] = useState(false)
    const [executionsError, setExecutionsError] = useState(null)

    // Notification hook with error handling
    let notificationFunctions = { showError: () => { }, showWarning: () => { }, showApiNotification: () => { } }
    try {
        notificationFunctions = useNotification()
    } catch (err) {
        console.error('Failed to initialize notification hook:', err)
    }
    const { showError, showWarning, showApiNotification } = notificationFunctions

    // Form state with interval scheduling options
    const [formData, setFormData] = useState({
        scheduleType: 'interval', // Always interval
        time: '06:00', // HH:MM format
        initialDay: new Date().toISOString().split('T')[0], // Today's date (YYYY-MM-DD format)
        dayInterval: 1, // Days between executions (1-15)
        duration: 5,
        isActive: true
    })

    // Load scheduled tasks
    const loadScheduledTasks = async () => {
        try {
            setLoading(true)
            setError(null)
            const response = await scheduledTasksApi.getScheduledTasks(tenantId, deviceId)

            // Filter out tasks with invalid scheduling data
            const validTasks = (response.data || []).filter(task => {
                if (!task) return false
                // Task is valid if it has either schedule or scheduling field
                return task.schedule || task.scheduling
            })
            setScheduledTasks(validTasks)
        } catch (err) {
            setError(err.message)
            console.error('Failed to load scheduled tasks:', err)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        if (tenantId && deviceId) {
            loadScheduledTasks()
        } else {
            console.warn('Missing required props:', { tenantId, deviceId })
            setLoading(false)
        }
    }, [tenantId, deviceId])

    // Convert human-friendly form data to scheduling configuration
    const formDataToScheduling = (data) => {
        try {
            if (!data || !data.time) {
                console.error('Invalid form data for scheduling conversion:', data)
                return {
                    type: 'interval',
                    initial_day: new Date().toISOString(),
                    day_interval: 1,
                    execution_time: '06:00'
                }
            }

            if (data.scheduleType === 'interval') {
                return {
                    type: 'interval',
                    initial_day: `${data.initialDay}T00:00:00Z`,
                    day_interval: data.dayInterval || 1,
                    execution_time: data.time
                }
            } else {
                // Legacy cron support
                const [hour, minute] = data.time.split(':')
                if (!hour || !minute) {
                    console.error('Invalid time format:', data.time)
                    return {
                        type: 'cron',
                        schedule: '0 6 * * *'
                    }
                }

                let cronExpression
                switch (data.frequency) {
                    case 'daily':
                        cronExpression = `0 ${minute} ${hour} * * *`
                        break
                    case 'weekly':
                        cronExpression = `0 ${minute} ${hour} * * ${data.dayOfWeek || '1'}`
                        break
                    case 'every2days':
                        cronExpression = `0 ${minute} ${hour} */2 * *`
                        break
                    case 'every3days':
                        cronExpression = `0 ${minute} ${hour} */3 * *`
                        break
                    default:
                        cronExpression = `0 ${minute} ${hour} * * *`
                }

                return {
                    type: 'cron',
                    schedule: cronExpression
                }
            }
        } catch (err) {
            console.error('Error converting form data to scheduling:', err, data)
            return {
                type: 'interval',
                initial_day: new Date().toISOString(),
                day_interval: 1,
                execution_time: '06:00'
            }
        }
    }

    // Convert scheduling configuration to human-friendly form data
    const schedulingToFormData = (scheduling) => {
        try {
            // Handle null, undefined, or empty values
            if (!scheduling || scheduling === null || scheduling === undefined) {
                console.warn('Empty or undefined scheduling data:', scheduling)
                return {
                    scheduleType: 'interval',
                    initialDay: new Date().toISOString().split('T')[0],
                    dayInterval: 1,
                    time: '06:00',
                    duration: 5,
                    isActive: true
                }
            }

            // Handle legacy format where schedule is a direct string
            if (typeof scheduling === 'string') {
                const cron = scheduling
                if (!cron || cron.trim() === '') {
                    console.warn('Empty cron expression:', cron)
                    return {
                        scheduleType: 'interval',
                        initialDay: new Date().toISOString().split('T')[0],
                        dayInterval: 1,
                        time: '06:00',
                        duration: 5,
                        isActive: true
                    }
                }

                const parts = cron.split(' ')
                if (parts.length !== 5 && parts.length !== 6) {
                    console.error('Invalid cron format (expected 5 or 6 parts):', cron)
                    return {
                        scheduleType: 'interval',
                        initialDay: new Date().toISOString().split('T')[0],
                        dayInterval: 1,
                        time: '06:00',
                        duration: 5,
                        isActive: true
                    }
                }

                // Handle both 5-part and 6-part formats
                let minute, hour, day, month, dayOfWeekCron
                if (parts.length === 6) {
                    // 6-part format: second minute hour day month dayOfWeek
                    [, minute, hour, day, month, dayOfWeekCron] = parts
                } else {
                    // 5-part format: minute hour day month dayOfWeek
                    [minute, hour, day, month, dayOfWeekCron] = parts
                }

                // Determine frequency
                let frequency = 'daily'
                let dayOfWeek = '1'

                if (dayOfWeekCron !== '*') {
                    frequency = 'weekly'
                    dayOfWeek = dayOfWeekCron
                } else if (day.startsWith('*/')) {
                    const interval = parseInt(day.substring(2))
                    if (interval === 2) {
                        frequency = 'every2days'
                    } else if (interval === 3) {
                        frequency = 'every3days'
                    }
                }

                return {
                    scheduleType: 'cron',
                    frequency,
                    time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`,
                    dayOfWeek,
                    duration: 5, // Will be set separately
                    isActive: true
                }
            }

            // Handle new scheduling object format
            if (scheduling.type === 'interval') {
                const initialDay = scheduling.initial_day ? scheduling.initial_day.split('T')[0] : new Date().toISOString().split('T')[0]
                return {
                    scheduleType: 'interval',
                    initialDay,
                    dayInterval: scheduling.day_interval || 1,
                    time: scheduling.execution_time || '06:00',
                    duration: 5, // Will be set separately
                    isActive: true
                }
            } else {
                // Legacy cron support
                const cron = scheduling.schedule
                const parts = cron.split(' ')
                if (parts.length !== 5 && parts.length !== 6) {
                    console.error('Invalid cron format (expected 5 or 6 parts):', cron)
                    return null
                }

                // Handle both 5-part and 6-part formats
                let minute, hour, day, month, dayOfWeekCron
                if (parts.length === 6) {
                    // 6-part format: second minute hour day month dayOfWeek
                    [, minute, hour, day, month, dayOfWeekCron] = parts
                } else {
                    // 5-part format: minute hour day month dayOfWeek
                    [minute, hour, day, month, dayOfWeekCron] = parts
                }

                // Determine frequency
                let frequency = 'daily'
                let dayOfWeek = '1'

                if (dayOfWeekCron !== '*') {
                    frequency = 'weekly'
                    dayOfWeek = dayOfWeekCron
                } else if (day.startsWith('*/')) {
                    const interval = parseInt(day.substring(2))
                    if (interval === 2) {
                        frequency = 'every2days'
                    } else if (interval === 3) {
                        frequency = 'every3days'
                    }
                }

                return {
                    scheduleType: 'cron',
                    frequency,
                    time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`,
                    dayOfWeek,
                    duration: 5, // Will be set separately
                    isActive: true
                }
            }
        } catch (err) {
            console.error('Error parsing scheduling configuration:', err, scheduling)
            return null
        }
    }

    // Handle form submission with notification
    const handleSubmit = async (e) => {
        e.preventDefault()

        const scheduling = formDataToScheduling(formData)
        // Ensure duration is within valid range (1-60 minutes)
        const duration = Math.max(1, Math.min(60, formData.duration))

        const taskData = {
            commands: [
                {
                    index: 1,
                    value: 1, // Activate relay
                    priority: "NORMAL",
                    wait_for: "0s" // Immediate activation
                },
                {
                    index: 1,
                    value: 0, // Deactivate relay
                    priority: "NORMAL",
                    wait_for: `${duration}m` // Wait for user's selected duration (1-60 minutes)
                }
            ],
            scheduling: scheduling,
            is_active: formData.isActive
        }

        console.log('Submitting task:', { editingTask, taskData, tenantId, deviceId })

        const operation = editingTask
            ? scheduledTasksApi.updateScheduledTask(tenantId, deviceId, editingTask.id, taskData)
            : scheduledTasksApi.createScheduledTask(tenantId, deviceId, taskData)

        const result = await showApiNotification(
            operation,
            editingTask
                ? 'Scheduled irrigation task updated successfully!'
                : 'Scheduled irrigation task created successfully!',
            editingTask
                ? 'Failed to update scheduled irrigation task'
                : 'Failed to create scheduled irrigation task',
            editingTask ? 'Task Updated' : 'Task Created',
            'Operation Failed'
        )

        if (result.success) {
            if (editingTask) {
                // Update existing task in UI
                setScheduledTasks(prev => prev.map(task =>
                    task.id === editingTask.id
                        ? { ...task, ...taskData, id: editingTask.id }
                        : task
                ))
            } else {
                // Add new task to UI using the returned UUID
                const returnedTask = result.data
                const newTask = {
                    id: returnedTask?.id || returnedTask?.uuid || Date.now(), // Use returned UUID or fallback
                    ...taskData,
                    is_active: formData.isActive
                }

                console.log('New task created with UUID:', newTask.id, 'API response:', returnedTask)
                setScheduledTasks(prev => [...prev, newTask])
            }

            // Reset form and close modal
            closeModal()
        }
    }

    // Handle task deletion with notification
    const handleDelete = async (taskId) => {
        if (!window.confirm('Are you sure you want to delete this scheduled irrigation task?')) {
            return
        }

        const result = await showApiNotification(
            scheduledTasksApi.deleteScheduledTask(tenantId, deviceId, taskId),
            'Scheduled irrigation task deleted successfully',
            'Failed to delete scheduled irrigation task',
            'Task Deleted',
            'Delete Failed'
        )

        if (result.success) {
            // Remove the task from the UI immediately
            setScheduledTasks(prev => prev.filter(task => task.id !== taskId))
        }
    }

    // Handle task toggle (active/inactive) with notification
    const handleToggleActive = async (task) => {
        // Ensure all wait_for values are properly formatted as strings
        const normalizedCommands = task.commands.map(cmd => ({
            ...cmd,
            wait_for: typeof cmd.wait_for === 'number' ? `${cmd.wait_for}s` : String(cmd.wait_for)
        }))

        const updatedTask = {
            commands: normalizedCommands,
            // Handle both legacy and new formats
            ...(task.scheduling ? { scheduling: task.scheduling } : { schedule: task.schedule }),
            is_active: !task.is_active
        }

        const result = await showApiNotification(
            scheduledTasksApi.updateScheduledTask(tenantId, deviceId, task.id, updatedTask),
            `Task ${!task.is_active ? 'activated' : 'deactivated'} successfully`,
            `Failed to ${!task.is_active ? 'activate' : 'deactivate'} task`,
            'Status Updated',
            'Update Failed'
        )

        if (result.success) {
            // Update the task's active state in the UI immediately
            setScheduledTasks(prev => prev.map(t =>
                t.id === task.id
                    ? { ...t, is_active: !task.is_active }
                    : t
            ))
        }
    }

    // Handle edit task
    const handleEdit = (task) => {
        setEditingTask(task)
        // Handle both legacy (task.schedule) and new (task.scheduling) formats
        const schedulingData = task.scheduling || task.schedule
        const formDataFromScheduling = schedulingToFormData(schedulingData)
        setFormData({
            ...formDataFromScheduling,
            duration: getTaskDuration(task),
            isActive: task.is_active
        })
        setShowModal(true)
    }

    // Handle add new task
    const handleAddNew = () => {
        setEditingTask(null)
        setFormData({
            scheduleType: 'interval',
            time: '06:00',
            initialDay: new Date().toISOString().split('T')[0], // Today's date
            dayInterval: 1,
            duration: 5,
            isActive: true
        })
        setShowModal(true)
    }

    const closeModal = () => {
        setShowModal(false)
        setEditingTask(null)
        setFormData({
            scheduleType: 'interval',
            time: '06:00',
            initialDay: new Date().toISOString().split('T')[0], // Today's date
            dayInterval: 1,
            duration: 5,
            isActive: true
        })
        setError(null)
    }

    // Parse scheduling configuration for display
    const parseScheduleExpression = (scheduling) => {
        try {
            console.log('🔍 Parsing scheduling data:', scheduling)

            // Handle null, undefined, or empty values
            if (!scheduling || scheduling === null || scheduling === undefined) {
                console.warn('Empty or undefined scheduling data:', scheduling)
                return 'Schedule not set'
            }

            // Handle legacy format where schedule is a direct string
            if (typeof scheduling === 'string') {
                const cron = scheduling
                if (!cron || cron.trim() === '') {
                    console.warn('Empty cron expression:', cron)
                    return 'Schedule not set'
                }

                const parts = cron.split(' ')
                if (parts.length !== 5 && parts.length !== 6) {
                    console.error('Invalid cron format (expected 5 or 6 parts):', cron)
                    return 'Invalid schedule'
                }

                // Handle both 5-part and 6-part formats
                let minute, hour, day, month, dayOfWeek
                if (parts.length === 6) {
                    // 6-part format: second minute hour day month dayOfWeek
                    [, minute, hour, day, month, dayOfWeek] = parts
                } else {
                    // 5-part format: minute hour day month dayOfWeek
                    [minute, hour, day, month, dayOfWeek] = parts
                }

                // Convert to readable format
                const time = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`

                if (dayOfWeek !== '*') {
                    const days = ['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
                    const dayIndex = parseInt(dayOfWeek)
                    const dayName = days[dayIndex] || 'Sun'
                    return `Weekly on ${dayName} at ${time}`
                } else if (day.startsWith('*/')) {
                    const interval = parseInt(day.substring(2))
                    return `Every ${interval} days at ${time}`
                } else {
                    return `Daily at ${time}`
                }
            }

            // Handle new scheduling object format
            console.log('🔍 Checking if scheduling is object:', typeof scheduling, scheduling)
            if (!scheduling || typeof scheduling !== 'object') {
                console.error('Invalid scheduling configuration:', scheduling)
                return 'Invalid schedule'
            }

            console.log('🔍 Scheduling type:', scheduling.type)
            if (scheduling.type === 'interval') {
                console.log('✅ Processing interval scheduling:', scheduling)
                const time = scheduling.execution_time || '06:00'
                const [hour, minute] = time.split(':')
                const timeDisplay = new Date(2000, 0, 1, parseInt(hour), parseInt(minute)).toLocaleTimeString([], {
                    hour: 'numeric',
                    minute: '2-digit',
                    hour12: true
                })

                const interval = scheduling.day_interval || 1
                const result = interval === 1 ? `Daily at ${timeDisplay}` : `Every ${interval} days at ${timeDisplay}`
                console.log('✅ Interval result:', result)
                return result
            } else if (scheduling.type === 'cron') {
                const cron = scheduling.schedule
                if (!cron || typeof cron !== 'string') {
                    console.error('Invalid cron expression:', cron)
                    return 'Invalid schedule'
                }

                const parts = cron.split(' ')
                if (parts.length !== 5 && parts.length !== 6) {
                    console.error('Invalid cron format (expected 5 or 6 parts):', cron)
                    return 'Invalid schedule'
                }

                // Handle both 5-part and 6-part formats
                let minute, hour, day, month, dayOfWeek
                if (parts.length === 6) {
                    // 6-part format: second minute hour day month dayOfWeek
                    [, minute, hour, day, month, dayOfWeek] = parts
                } else {
                    // 5-part format: minute hour day month dayOfWeek
                    [minute, hour, day, month, dayOfWeek] = parts
                }

                // Convert to readable format
                const time = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`

                if (dayOfWeek !== '*') {
                    const days = ['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
                    const dayIndex = parseInt(dayOfWeek)
                    const dayName = days[dayIndex] || 'Sun'
                    return `Weekly on ${dayName} at ${time}`
                } else if (day.startsWith('*/')) {
                    const interval = parseInt(day.substring(2))
                    return `Every ${interval} days at ${time}`
                } else {
                    return `Daily at ${time}`
                }
            }

            // If we get here, it's an unknown scheduling type
            console.error('❌ Unknown scheduling type:', scheduling.type, scheduling)
            return 'Unknown schedule type'
        } catch (err) {
            console.error('❌ Error parsing schedule expression:', err, scheduling)
            return 'Invalid schedule'
        }
    }

    // Calculate next execution datetime based on form data
    const getNextExecutionDateTime = (formData) => {
        try {
            const initialDate = new Date(formData.initialDay)
            const [hours, minutes] = formData.time.split(':').map(Number)

            // Set the time for the initial date
            initialDate.setHours(hours, minutes, 0, 0)

            const now = new Date()
            const intervalDays = formData.dayInterval || 1

            // The first execution is always the initial date
            let nextExecution = new Date(initialDate)

            // Only advance to the next occurrence if the initial date is in the past
            while (nextExecution <= now) {
                nextExecution.setDate(nextExecution.getDate() + intervalDays)
            }

            // Format the date nicely
            const isToday = nextExecution.toDateString() === now.toDateString()
            const isTomorrow = nextExecution.toDateString() === new Date(now.getTime() + 24 * 60 * 60 * 1000).toDateString()

            let dateStr
            if (isToday) {
                dateStr = 'Today'
            } else if (isTomorrow) {
                dateStr = 'Tomorrow'
            } else {
                dateStr = nextExecution.toLocaleDateString('en-US', {
                    weekday: 'long',
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                })
            }

            const timeStr = nextExecution.toLocaleTimeString('en-US', {
                hour: 'numeric',
                minute: '2-digit',
                hour12: true
            })

            return `${dateStr} at ${timeStr}`
        } catch (error) {
            return 'Invalid date'
        }
    }

    // Calculate next execution datetime from task scheduling data
    const getTaskNextExecution = (scheduling) => {
        try {
            if (!scheduling) return 'No schedule'

            // Handle legacy format where schedule is a direct string
            if (typeof scheduling === 'string') {
                return 'Legacy schedule'
            }

            // Handle new scheduling object format
            if (scheduling.type === 'interval') {
                const initialDate = new Date(scheduling.initial_day)
                const [hours, minutes] = scheduling.execution_time.split(':').map(Number)

                // Set the time for the initial date
                initialDate.setHours(hours, minutes, 0, 0)

                const now = new Date()
                const intervalDays = scheduling.day_interval || 1

                // The first execution is always the initial date
                let nextExecution = new Date(initialDate)

                // Only advance to the next occurrence if the initial date is in the past
                while (nextExecution <= now) {
                    nextExecution.setDate(nextExecution.getDate() + intervalDays)
                }

                // Format the date nicely
                const isToday = nextExecution.toDateString() === now.toDateString()
                const isTomorrow = nextExecution.toDateString() === new Date(now.getTime() + 24 * 60 * 60 * 1000).toDateString()

                let dateStr
                if (isToday) {
                    dateStr = 'Today'
                } else if (isTomorrow) {
                    dateStr = 'Tomorrow'
                } else {
                    dateStr = nextExecution.toLocaleDateString('en-US', {
                        weekday: 'long',
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric'
                    })
                }

                const timeStr = nextExecution.toLocaleTimeString('en-US', {
                    hour: 'numeric',
                    minute: '2-digit',
                    hour12: true
                })

                return `${dateStr} at ${timeStr}`
            }

            return 'Unknown schedule type'
        } catch (error) {
            return 'Invalid schedule'
        }
    }

    // Get task duration from commands (always returns 1-60 minutes)
    const getTaskDuration = (task) => {
        try {
            if (!task || !task.commands || task.commands.length < 2) return 5

            const deactivateCommand = task.commands.find(cmd => cmd.value === 0)
            if (deactivateCommand && deactivateCommand.wait_for) {
                // Ensure wait_for is a string before calling match
                const waitFor = String(deactivateCommand.wait_for)

                // Parse different time units and convert to minutes
                const minutesMatch = waitFor.match(/(\d+)m/)
                const secondsMatch = waitFor.match(/(\d+)s/)
                const hoursMatch = waitFor.match(/(\d+)h/)

                if (minutesMatch) {
                    const minutes = parseInt(minutesMatch[1])
                    return Math.max(1, Math.min(60, minutes))
                } else if (secondsMatch) {
                    const seconds = parseInt(secondsMatch[1])
                    const minutes = Math.ceil(seconds / 60)
                    return Math.max(1, Math.min(60, minutes))
                } else if (hoursMatch) {
                    const hours = parseInt(hoursMatch[1])
                    const minutes = hours * 60
                    return Math.max(1, Math.min(60, minutes))
                }
            }

            return 5 // Default duration
        } catch (err) {
            console.error('Error parsing task duration:', err, task)
            return 5
        }
    }

    // Get human-readable schedule description
    const getScheduleDescription = (data) => {
        try {
            if (!data || !data.time || typeof data.time !== 'string') {
                return 'Invalid schedule'
            }

            const timeStr = data.time
            const [hour, minute] = timeStr.split(':')
            const timeDisplay = new Date(2000, 0, 1, parseInt(hour), parseInt(minute)).toLocaleTimeString([], {
                hour: 'numeric',
                minute: '2-digit',
                hour12: true
            })

            const duration = data.duration || 5

            if (data.scheduleType === 'interval') {
                const interval = data.dayInterval || 1
                if (interval === 1) {
                    return `Daily at ${timeDisplay} for ${duration} minutes`
                } else {
                    return `Every ${interval} days at ${timeDisplay} for ${duration} minutes`
                }
            } else {
                // Legacy cron support
                switch (data.frequency) {
                    case 'daily':
                        return `Daily at ${timeDisplay} for ${duration} minutes`
                    case 'weekly':
                        const days = ['', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']
                        const dayName = days[parseInt(data.dayOfWeek)] || 'Sunday'
                        return `Weekly on ${dayName}s at ${timeDisplay} for ${duration} minutes`
                    case 'every2days':
                        return `Every 2 days at ${timeDisplay} for ${duration} minutes`
                    case 'every3days':
                        return `Every 3 days at ${timeDisplay} for ${duration} minutes`
                    default:
                        return `Daily at ${timeDisplay} for ${duration} minutes`
                }
            }
        } catch (error) {
            console.error('Error generating schedule description:', error, data)
            return 'Invalid schedule'
        }
    }

    // Load task executions
    const loadTaskExecutions = async (taskId) => {
        try {
            setExecutionsLoading(true)
            setExecutionsError(null)
            const response = await scheduledTasksApi.getTaskExecutions(tenantId, deviceId, taskId)
            setTaskExecutions(response.data || [])
        } catch (err) {
            setExecutionsError(err.message)
            console.error('Failed to load task executions:', err)
        } finally {
            setExecutionsLoading(false)
        }
    }

    const handleShowExecutions = async (task) => {
        setSelectedTask(task)
        setShowExecutionsModal(true)
        await loadTaskExecutions(task.id)
    }

    const closeExecutionsModal = () => {
        setShowExecutionsModal(false)
        setSelectedTask(null)
        setTaskExecutions([])
        setExecutionsError(null)
    }

    if (loading) {
        return (
            <div className="scheduled-irrigation">
                <div className="scheduled-irrigation-header">
                    <div className="scheduled-irrigation-title">
                        <Calendar size={18} />
                        <span>Scheduled Irrigation</span>
                    </div>
                </div>
                <div className="scheduled-irrigation-loading">
                    <div className="loading-spinner"></div>
                    <span>Loading scheduled tasks...</span>
                </div>
            </div>
        )
    }

    return (
        <>
            <div className="scheduled-irrigation">
                <div className="scheduled-irrigation-header">
                    <div className="scheduled-irrigation-title">
                        <Calendar size={18} />
                        <span>Scheduled Irrigation</span>
                    </div>
                    <button
                        className="add-schedule-btn"
                        onClick={handleAddNew}
                    >
                        <Plus size={16} />
                        Add Schedule
                    </button>
                </div>

                {/* Scheduled Tasks List */}
                <div className="scheduled-tasks-list">
                    {scheduledTasks.length === 0 ? (
                        <div className="empty-state">
                            <Calendar size={48} />
                            <h4>No scheduled irrigation tasks</h4>
                            <p>Create your first scheduled irrigation task to automate watering.</p>
                            <button
                                className="add-first-schedule-btn"
                                onClick={handleAddNew}
                            >
                                <Plus size={16} />
                                Add First Schedule
                            </button>
                        </div>
                    ) : (
                        scheduledTasks.map((task) => {
                            // Safety check for task validity
                            if (!task || (!task.schedule && !task.scheduling)) {
                                console.warn('Skipping invalid task:', task)
                                return null
                            }

                            return (
                                <div key={task.id} className="scheduled-task-item">
                                    <div className="task-info">
                                        <div className="task-header">
                                            <div className="task-schedule">
                                                <Clock size={16} />
                                                <span>{parseScheduleExpression(task.scheduling || task.schedule)}</span>
                                            </div>
                                            <div className="task-status">
                                                <span className="status-emoji">
                                                    {task.is_active ? '✅' : '📵'}
                                                </span>
                                                <span>{task.is_active ? 'Active' : 'Inactive'}</span>
                                            </div>
                                        </div>

                                        <div className="task-details">
                                            <div className="task-duration">
                                                <Droplets size={14} />
                                                <span>{getTaskDuration(task)} minutes</span>
                                            </div>
                                            <div className="task-next-execution">
                                                <Calendar size={14} />
                                                <span>Next: {getTaskNextExecution(task.scheduling || task.schedule)}</span>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="task-actions">
                                        <button
                                            onClick={() => handleShowExecutions(task)}
                                            className="executions-btn"
                                            title="View Executions"
                                        >
                                            <Calendar size={14} />
                                        </button>
                                        <button
                                            onClick={() => handleToggleActive(task)}
                                            className={`toggle-btn ${task.is_active ? 'pause' : 'play'}`}
                                            title={task.is_active ? 'Pause' : 'Activate'}
                                        >
                                            {task.is_active ? <Pause size={14} /> : <Play size={14} />}
                                        </button>
                                        <button
                                            onClick={() => handleEdit(task)}
                                            className="edit-btn"
                                            title="Edit"
                                        >
                                            <Edit3 size={14} />
                                        </button>
                                        <button
                                            onClick={() => handleDelete(task.id)}
                                            className="delete-btn"
                                            title="Delete"
                                        >
                                            <Trash2 size={14} />
                                        </button>
                                    </div>
                                </div>
                            )
                        })
                    )}
                </div>
            </div>

            {/* Modal */}
            {showModal && (
                <div className="modal-overlay" onClick={closeModal}>
                    <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3>{editingTask ? 'Edit Schedule' : 'New Schedule'}</h3>
                            <button onClick={closeModal} className="modal-close-btn">
                                <X size={20} />
                            </button>
                        </div>

                        {error && (
                            <div className="modal-error">
                                <span>{error}</span>
                                <button onClick={() => setError(null)}>
                                    <X size={14} />
                                </button>
                            </div>
                        )}

                        <form onSubmit={handleSubmit} className="modal-form">
                            {/* Schedule Preview */}
                            <div className="schedule-preview">
                                <div className="preview-label">Schedule Preview:</div>
                                <div className="preview-text">{getScheduleDescription(formData)}</div>
                            </div>

                            {/* Initial Date Selection - Always Visible */}
                            <div className="form-group">
                                <label htmlFor="initialDay">
                                    <Calendar size={16} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
                                    Start Date
                                </label>
                                <input
                                    id="initialDay"
                                    type="date"
                                    value={formData.initialDay}
                                    onChange={(e) => setFormData({ ...formData, initialDay: e.target.value })}
                                    className="form-input"
                                    min={new Date().toISOString().split('T')[0]}
                                    style={{
                                        padding: '12px 16px',
                                        border: '2px solid #e1e5e9',
                                        borderRadius: '8px',
                                        fontSize: '16px',
                                        backgroundColor: '#fff',
                                        color: '#333',
                                        width: '100%',
                                        boxSizing: 'border-box',
                                        marginBottom: '8px'
                                    }}
                                    required
                                />
                                <div style={{
                                    fontSize: '13px',
                                    color: '#666',
                                    marginTop: '4px',
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '4px'
                                }}>
                                    <Calendar size={12} />
                                    <span>Choose when your irrigation schedule should begin</span>
                                </div>
                            </div>

                            {/* Interval Scheduling Fields */}
                            <>

                                {/* Day Interval Selection */}
                                <div className="form-group">
                                    <label htmlFor="dayInterval">Day Interval</label>
                                    <div className="duration-slider-container">
                                        <input
                                            id="dayInterval"
                                            type="range"
                                            min="1"
                                            max="15"
                                            value={formData.dayInterval}
                                            onChange={(e) => setFormData({ ...formData, dayInterval: parseInt(e.target.value) })}
                                            className="duration-slider"
                                            required
                                        />
                                        <div className="duration-value">
                                            <span>{formData.dayInterval}</span>
                                        </div>
                                    </div>
                                    <div className="duration-display">
                                        <Calendar size={14} />
                                        <span>Will execute every {formData.dayInterval} day{formData.dayInterval !== 1 ? 's' : ''}</span>
                                    </div>
                                </div>
                            </>

                            {/* Time Selection */}
                            <div className="form-group">
                                <label htmlFor="time">Time</label>
                                <input
                                    id="time"
                                    type="time"
                                    value={formData.time}
                                    onChange={(e) => setFormData({ ...formData, time: e.target.value })}
                                    className="form-input"
                                    required
                                />
                            </div>

                            {/* Next Execution Display */}
                            <div className="form-group">
                                <div style={{
                                    background: '#f8f9fa',
                                    border: '1px solid #e9ecef',
                                    borderRadius: '8px',
                                    padding: '12px',
                                    fontSize: '14px'
                                }}>
                                    <div style={{ fontWeight: '600', marginBottom: '4px', color: '#495057' }}>
                                        Next Execution
                                    </div>
                                    <div style={{ color: '#6c757d' }}>
                                        {getNextExecutionDateTime(formData)}
                                    </div>
                                </div>
                            </div>


                            {/* Duration Selection */}
                            <div className="form-group">
                                <label htmlFor="duration">Duration (minutes)</label>
                                <div className="duration-slider-container">
                                    <input
                                        id="duration"
                                        type="range"
                                        min="1"
                                        max="60"
                                        value={formData.duration}
                                        onChange={(e) => setFormData({ ...formData, duration: parseInt(e.target.value) })}
                                        className="duration-slider"
                                        required
                                    />
                                    <div className="duration-value">
                                        <span>{formData.duration}</span>
                                    </div>
                                </div>
                                <div className="duration-display">
                                    <Droplets size={14} />
                                    <span>Will irrigate for {formData.duration} minute{formData.duration !== 1 ? 's' : ''}</span>
                                </div>
                            </div>

                            {/* Active Status */}
                            <div className="form-group">
                                <label className="switch-label">
                                    <span>Active</span>
                                    <div className="switch-container">
                                        <input
                                            type="checkbox"
                                            checked={formData.isActive}
                                            onChange={(e) => setFormData({ ...formData, isActive: e.target.checked })}
                                            className="switch-input"
                                        />
                                        <span className="switch-slider"></span>
                                    </div>
                                </label>
                            </div>

                            <div className="modal-actions">
                                <button type="submit" className="save-btn">
                                    <Check size={16} />
                                    {editingTask ? 'Update' : 'Create'} Schedule
                                </button>
                                <button type="button" onClick={closeModal} className="cancel-btn">
                                    Cancel
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* Executions Modal */}
            {showExecutionsModal && (
                <div className="modal-overlay" onClick={closeExecutionsModal}>
                    <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3>Last 3 Task Executions</h3>
                            <button onClick={closeExecutionsModal} className="modal-close-btn">
                                <X size={20} />
                            </button>
                        </div>

                        {selectedTask && (
                            <div className="task-info-header">
                                <div className="task-schedule">
                                    <Clock size={16} />
                                    <span>{parseScheduleExpression(selectedTask.scheduling || selectedTask.schedule)}</span>
                                </div>
                                <div className="task-duration">
                                    <Droplets size={14} />
                                    <span>{getTaskDuration(selectedTask)} minutes</span>
                                </div>
                            </div>
                        )}

                        {executionsLoading ? (
                            <div className="modal-loading">
                                <div className="loading-spinner"></div>
                                <span>Loading executions...</span>
                            </div>
                        ) : executionsError ? (
                            <div className="modal-error">
                                <span>{executionsError}</span>
                                <button onClick={() => setExecutionsError(null)}>
                                    <X size={14} />
                                </button>
                            </div>
                        ) : (
                            <>
                                <div className="executions-list">
                                    {taskExecutions.length === 0 ? (
                                        <div className="empty-state">
                                            <Calendar size={48} />
                                            <h4>No executions found</h4>
                                            <p>This scheduled task has not been executed yet.</p>
                                        </div>
                                    ) : (
                                        taskExecutions.map((execution) => {
                                            const startCommand = execution.commands[0];
                                            const endCommand = execution.commands[1];

                                            const formatDateTime = (dateString) => {
                                                if (!dateString) return 'Unknown';
                                                const date = new Date(dateString);
                                                const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
                                                const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

                                                const day = days[date.getDay()];
                                                const dayNum = date.getDate();
                                                const month = months[date.getMonth()];
                                                const hours = date.getHours().toString().padStart(2, '0');
                                                const minutes = date.getMinutes().toString().padStart(2, '0');

                                                return `${day} ${dayNum} ${month}, ${hours}:${minutes}`;
                                            };

                                            return (
                                                <div key={execution.id} className="execution-item">
                                                    <div className="execution-info">
                                                        <div className="execution-timestamp">
                                                            <Clock size={14} />
                                                            <span><strong>Started:</strong> {formatDateTime(startCommand?.sent_at)}</span>
                                                        </div>
                                                        <div className="execution-timestamp">
                                                            <Clock size={14} />
                                                            <span><strong>Ended:</strong> {endCommand?.sent_at ? formatDateTime(endCommand.sent_at) : <em>In progress</em>}</span>
                                                        </div>
                                                    </div>
                                                    <div className="execution-status">
                                                        <div className="status-indicator completed" />
                                                        <span>Completed</span>
                                                    </div>
                                                </div>
                                            );
                                        })
                                    )}
                                </div>
                                <div className="modal-actions">
                                    <button onClick={closeExecutionsModal} className="close-btn">
                                        Close
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            )}
        </>
    )
}

export default ScheduledIrrigationWithNotifications 