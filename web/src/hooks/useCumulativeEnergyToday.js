import { useState, useEffect } from 'react';
import { metricsApi } from '../config/api';
import { getTodayWindow, computeCumulativeEnergyKWh } from '../utils/energyAccumulation';

const STEP = '5m';

// Tracks cumulative energy (kWh) for a power (W) metric from local midnight
// up to now, re-deriving the window and refetching every 30s so the chart
// keeps growing through the day and resets itself at the next midnight.
const useCumulativeEnergyToday = (metricName) => {
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;
    const fetchHistory = async () => {
      setLoading(true);
      setError(null);
      try {
        const { start, end } = getTodayWindow();
        const points = await metricsApi.queryRange(metricName, {
          start: start.getTime(),
          end: end.getTime(),
          step: STEP,
        });
        if (!cancelled) setHistory(computeCumulativeEnergyKWh(points));
      } catch (err) {
        if (!cancelled) setError(err.message);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    fetchHistory();
    const interval = setInterval(fetchHistory, 30000);
    return () => { cancelled = true; clearInterval(interval); };
  }, [metricName]);

  const totalKWh = history.length > 0 ? history[history.length - 1].value : 0;

  return { history, totalKWh, loading, error };
};

export default useCumulativeEnergyToday;
