import { useState, useEffect } from 'react';
import { metricsApi } from '../config/api';
import { computeGridEnergyBuckets, computeCumulativeGridEnergy } from '../utils/energyAccumulation';

// Tracks energy imported from and exported to the grid over the selected
// range (ending now), both grouped into the range's bucket size and as running
// totals, from a signed grid power (W) metric. Refetches every 30s so the
// latest bucket keeps filling in.
const useGridEnergy = (metricName, range) => {
  const [buckets, setBuckets] = useState([]);
  const [cumulative, setCumulative] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;
    const fetchHistory = async () => {
      setLoading(true);
      setError(null);
      try {
        const end = new Date();
        const start = new Date(end.getTime() - range.ms);
        const points = await metricsApi.queryRange(metricName, {
          start: start.getTime(),
          end: end.getTime(),
          step: range.grid.step,
        });
        if (!cancelled) {
          setBuckets(computeGridEnergyBuckets(points, start, end, range.grid.bucket));
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
  }, [metricName, range]);

  const totals = cumulative.length > 0 ? cumulative[cumulative.length - 1] : null;
  const hasData = buckets.some(bucket => bucket.net !== null);

  return { buckets, cumulative, totals, hasData, loading, error };
};

export default useGridEnergy;
