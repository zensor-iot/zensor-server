import { useState, useEffect } from 'react';
import { metricsApi } from '../config/api';

const useMetricHistory = (metricName, range) => {
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;
    const fetchHistory = async () => {
      setLoading(true);
      setError(null);
      try {
        const points = await metricsApi.queryRange(metricName, {
          start: Date.now() - range.ms,
          step: range.step,
        });
        if (!cancelled) setHistory(points);
      } catch (err) {
        if (!cancelled) setError(err.message);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    fetchHistory();
    const interval = setInterval(fetchHistory, 30000);
    return () => { cancelled = true; clearInterval(interval); };
  }, [metricName, range]);

  return { history, loading, error };
};

export default useMetricHistory;
