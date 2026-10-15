import { useState, useEffect } from 'react';
import { metricsApi } from '../config/api';
import { getTodayWindow, computeHourlyGridEnergy, computeCumulativeGridEnergy } from '../utils/energyAccumulation';

const STEP = '5m';

// Tracks energy imported from and exported to the grid since local midnight,
// both grouped by hour and as running totals, from a signed grid power (W)
// metric. Refetches every 30s so the current hour keeps filling in.
const useGridEnergyToday = (metricName) => {
  const [hourly, setHourly] = useState([]);
  const [cumulative, setCumulative] = useState([]);
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
        if (!cancelled) {
          setHourly(computeHourlyGridEnergy(points, start, end));
          setCumulative(computeCumulativeGridEnergy(points));
        }
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

  const totals = cumulative.length > 0 ? cumulative[cumulative.length - 1] : null;
  const hasData = hourly.some(bucket => bucket.net !== null);

  return { hourly, cumulative, totals, hasData, loading, error };
};

export default useGridEnergyToday;
