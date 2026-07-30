import { useState, useEffect, useRef, useCallback } from 'react';

const useWebSocket = (url) => {
    const [isConnected, setIsConnected] = useState(false);
    const [lastMessage, setLastMessage] = useState(null);
    const [connectionError, setConnectionError] = useState(null);
    const [connectionStatus, setConnectionStatus] = useState('Disconnected');
    const reconnectTimeoutRef = useRef(null);
    const reconnectAttempts = useRef(0);
    const socketRef = useRef(null);
    const urlRef = useRef(url);
    const maxReconnectAttempts = 5;
    const reconnectDelay = 3000; // 3 seconds

    // Update URL ref when URL changes
    useEffect(() => {
        urlRef.current = url;
    }, [url]);

    const connect = useCallback(() => {
        // Don't connect if we already have a connection
        if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
            console.log('WebSocket already connected, skipping connection attempt');
            return;
        }

        // Clean up existing connection
        if (socketRef.current) {
            socketRef.current.close();
            socketRef.current = null;
        }

        try {
            setConnectionError(null);
            setConnectionStatus('Connecting...');

            const ws = new WebSocket(urlRef.current);
            socketRef.current = ws;

            ws.onopen = () => {
                console.log('WebSocket connected');
                setIsConnected(true);
                setConnectionStatus('Connected');
                setConnectionError(null);
                reconnectAttempts.current = 0;
            };

            ws.onmessage = (event) => {
                try {
                    // Handle different data types (string, Blob, ArrayBuffer)
                    let messageData = event.data;

                    if (messageData instanceof Blob) {
                        messageData.text().then(text => {
                            try {
                                const data = JSON.parse(text);
                                setLastMessage(data);
                            } catch (err) {
                                console.error('Failed to parse Blob WebSocket message:', err);
                            }
                        });
                        return;
                    }

                    if (messageData instanceof ArrayBuffer) {
                        messageData = new TextDecoder().decode(messageData);
                    }

                    const data = JSON.parse(messageData);
                    setLastMessage(data);
                } catch (err) {
                    console.error('Failed to parse WebSocket message:', err, 'Data type:', typeof event.data, 'Data:', event.data);
                }
            };

            ws.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason);
                setIsConnected(false);
                socketRef.current = null;

                if (event.code !== 1000) { // Not a normal closure
                    setConnectionStatus('Disconnected');

                    // Attempt to reconnect
                    if (reconnectAttempts.current < maxReconnectAttempts) {
                        reconnectAttempts.current += 1;
                        setConnectionStatus(`Reconnecting... (${reconnectAttempts.current}/${maxReconnectAttempts})`);

                        reconnectTimeoutRef.current = setTimeout(() => {
                            connect();
                        }, reconnectDelay);
                    } else {
                        setConnectionStatus('Failed to connect');
                        setConnectionError('Maximum reconnection attempts reached');
                    }
                } else {
                    setConnectionStatus('Disconnected');
                }
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                setConnectionError('WebSocket connection error');
                setConnectionStatus('Error');
            };

        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            setConnectionError('Failed to create WebSocket connection');
            setConnectionStatus('Error');
        }
    }, []); // Remove url dependency to prevent reconnections

    const disconnect = useCallback(() => {
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        if (socketRef.current) {
            socketRef.current.close(1000, 'Intentional disconnect');
            socketRef.current = null;
        }

        setIsConnected(false);
        setConnectionStatus('Disconnected');
        reconnectAttempts.current = 0;
    }, []); // Remove socket dependency

    const retry = useCallback(() => {
        reconnectAttempts.current = 0;
        connect();
    }, [connect]);

    // Only run effect once on mount, and when URL changes
    useEffect(() => {
        connect();

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (socketRef.current) {
                socketRef.current.close(1000, 'Component unmounting');
            }
        };
    }, [url]); // Only depend on URL, not on connect/disconnect functions

    return {
        isConnected,
        lastMessage,
        connectionError,
        connectionStatus,
        retry,
        disconnect
    };
};

export default useWebSocket; 