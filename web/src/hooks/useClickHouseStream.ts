import { useEffect, useState, useCallback, useRef } from 'react';
import type { MergeInfo, ServerStatus } from '../types';

export function useClickHouseStream() {
  const [merges, setMerges] = useState<MergeInfo[]>([]);
  const [status, setStatus] = useState<ServerStatus | null>(null);
  const [wsConnected, setWsConnected] = useState<boolean>(false);
  const wsRef = useRef<WebSocket | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const res = await fetch('/api/health');
      if (res.ok) {
        const data: ServerStatus = await res.json();
        setStatus(data);
      }
    } catch {
      // Ignored during connection transitions
    }
  }, []);

  const fetchMerges = useCallback(async () => {
    try {
      const res = await fetch('/api/merges');
      if (res.ok) {
        const data: MergeInfo[] = await res.json();
        setMerges(data);
      }
    } catch {
      // Ignored
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    fetchMerges();

    // Setup WebSocket for instant live stream
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws`;
    let ws: WebSocket | null = null;
    let reconnectTimeout: number | undefined;

    function connectWs() {
      try {
        ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          setWsConnected(true);
        };

        ws.onmessage = (event) => {
          try {
            const payload = JSON.parse(event.data);
            if (payload.type === 'merges_update' && Array.isArray(payload.data)) {
              setMerges(payload.data);
            }
          } catch {
            // Ignore parse errors
          }
        };

        ws.onclose = () => {
          setWsConnected(false);
          reconnectTimeout = window.setTimeout(connectWs, 3000);
        };

        ws.onerror = () => {
          setWsConnected(false);
        };
      } catch {
        setWsConnected(false);
      }
    }

    connectWs();

    // Fallback polling for status every 3 seconds
    const interval = setInterval(() => {
      fetchStatus();
      if (!wsConnected) {
        fetchMerges();
      }
    }, 3000);

    return () => {
      clearInterval(interval);
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (ws) {
        ws.close();
      }
    };
  }, [fetchStatus, fetchMerges, wsConnected]);

  return {
    merges,
    status,
    wsConnected,
    refetchStatus: fetchStatus,
    refetchMerges: fetchMerges,
  };
}
