import { useEffect, useState } from 'react'
import {
  Area,
  AreaChart,
  Bar,
  CartesianGrid,
  ComposedChart,
  Line,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import useWebSocket from '../hooks/useWebSocket'
import useMetricHistory from '../hooks/useMetricHistory'
import useCumulativeEnergyToday from '../hooks/useCumulativeEnergyToday'
import useGridEnergyToday from '../hooks/useGridEnergyToday'
import { getWebSocketUrl } from '../config/api'
import { connectionState } from '../utils/connectionState'
import {
  Card,
  CardGrid,
  DataValue,
  EmptyState,
  FilterChips,
  KeyValue,
  LoadingState,
  Meter,
  Notice,
  Page,
  PageHeader,
  Section,
  StatusPill,
} from '../ui'
import { cx, formatDateTime, formatNumber, formatRelative, formatTime, formatUnit, prefersReducedMotion } from '../lib/zensor'

const SOC_METRIC_NAME = 'zensor_server_victron_battery_soc'
const LOAD_METRIC_NAME = 'zensor_server_victron_system_load_power'
const SOLAR_POWER_METRIC_NAME = 'zensor_server_victron_solar_power'
const GRID_POWER_METRIC_NAME = 'zensor_server_victron_grid_power'

const GRID_SERIES = {
  imported: { label: 'Sacado de la red', color: 'var(--zs-status-attention-fill)' },
  exported: { label: 'Inyectado a la red', color: 'var(--zs-status-ok-fill)' },
  net: { label: 'Balance', color: 'var(--zs-brand-ink)' },
}

const LOW_SOC = 20
const DAY_MS = 24 * 60 * 60 * 1000

const TIME_RANGES = [
  { value: '1h', label: '1 h', ms: 60 * 60 * 1000, step: '30s' },
  { value: '6h', label: '6 h', ms: 6 * 60 * 60 * 1000, step: '60s' },
  { value: '24h', label: '24 h', ms: DAY_MS, step: '300s' },
  { value: '7d', label: '7 d', ms: 7 * DAY_MS, step: '3600s' },
]

const VEBUS_STATES = {
  0: 'Apagado',
  1: 'Bajo consumo',
  2: 'Falla',
  3: 'Carga inicial',
  4: 'Absorción',
  5: 'Flotación',
  6: 'Almacenamiento',
  7: 'Ecualización',
  8: 'Paso directo',
  9: 'Invirtiendo',
  10: 'Asistencia de potencia',
  11: 'Fuente de poder',
  252: 'Control externo',
}

const EQUIPMENT_KEYS = ['batteries', 'solar_chargers', 'ac_loads', 'vebus', 'temperatures', 'tanks']

const isNumber = value => typeof value === 'number' && Number.isFinite(value)
const unitOrEmpty = (value, unit, digits = 0) => (isNumber(value) ? formatUnit(value, unit, digits) : undefined)

const shortDay = new Intl.DateTimeFormat('es-CL', { day: 'numeric', month: 'short' })

const CHART_CLASS =
  'h-[240px] w-full [&_.recharts-cartesian-axis-tick-value]:fill-muted [&_.recharts-cartesian-axis-tick-value]:font-mono [&_.recharts-cartesian-axis-tick-value]:text-xs'

const HOUR_MS = 60 * 60 * 1000

function vebusState(state) {
  if (!isNumber(state)) return undefined
  return VEBUS_STATES[state] ?? `Desconocido (${state})`
}

function useNow(intervalMs) {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])
  return now
}

function summarize(history, unit, digits) {
  const values = history.map(point => point.value).filter(isNumber)
  if (values.length === 0) return ''
  const last = values[values.length - 1]
  return `último ${formatUnit(last, unit, digits)}, mínimo ${formatUnit(Math.min(...values), unit, digits)}, máximo ${formatUnit(Math.max(...values), unit, digits)}`
}

function ChartTooltip({ active, payload, label, unit, digits, seriesLabel }) {
  if (!active || !payload?.length) return null
  return (
    <div className="flex flex-col gap-1 rounded-md border border-line bg-surface px-3 py-2 text-ink shadow-overlay">
      <span className="text-caption text-muted">{formatDateTime(label)}</span>
      <span className="text-body-sm">
        {seriesLabel}{' '}
        <span className="font-mono font-medium">{formatUnit(payload[0].value, unit, digits)}</span>
      </span>
    </div>
  )
}

function ChartCard({ title, headline, description, hasData, loading, error, emptyText, footer, children }) {
  let body
  if (hasData) {
    body = (
      <div role="img" aria-label={description} className={CHART_CLASS}>
        <ResponsiveContainer width="100%" height="100%">
          {children}
        </ResponsiveContainer>
      </div>
    )
  } else if (loading) {
    body = <LoadingState>Cargando historial…</LoadingState>
  } else if (error) {
    body = (
      <p role="alert" className="m-0 py-8 text-body-sm text-attention">
        No se pudo cargar el historial ({error}). Se volverá a intentar en 30 segundos.
      </p>
    )
  } else {
    body = <p className="m-0 py-8 text-body-sm text-muted">{emptyText}</p>
  }

  return (
    <Card as="figure" className="m-0">
      <figcaption className="flex flex-wrap items-baseline justify-between gap-2">
        <span className="text-h3">{title}</span>
        {headline}
      </figcaption>
      {body}
      {hasData && footer}
    </Card>
  )
}

function MetricChart({ title, history, loading, error, unit, digits, spanMs, domain, reference, emptyText, headline }) {
  const tickTime = ts => (spanMs > DAY_MS ? shortDay.format(new Date(ts)) : formatTime(ts))

  return (
    <ChartCard
      title={title}
      headline={headline}
      description={`${title}: ${summarize(history, unit, digits)}`}
      hasData={history.length > 0}
      loading={loading}
      error={error}
      emptyText={emptyText}
      footer={reference && <p className="m-0 text-caption text-muted">La línea punteada marca {reference.label}.</p>}
    >
      <AreaChart data={history} margin={{ top: 8, right: 8, bottom: 0, left: 0 }} accessibilityLayer={false}>
        <CartesianGrid vertical={false} stroke="var(--zs-border-default)" />
        <XAxis
          dataKey="time"
          type="number"
          scale="time"
          domain={['dataMin', 'dataMax']}
          tickFormatter={tickTime}
          axisLine={false}
          tickLine={false}
          minTickGap={32}
        />
        <YAxis
          domain={domain}
          tickFormatter={value => formatUnit(value, unit, digits > 1 ? 1 : 0)}
          axisLine={false}
          tickLine={false}
          width={72}
        />
        <Tooltip
          cursor={{ stroke: 'var(--zs-border-strong)', strokeWidth: 1 }}
          content={<ChartTooltip unit={unit} digits={digits} seriesLabel={title} />}
        />
        {reference && <ReferenceLine y={reference.y} stroke="var(--zs-status-attention-fill)" strokeDasharray="4 4" />}
        <Area
          type="monotone"
          dataKey="value"
          name={title}
          stroke="var(--zs-brand-ink)"
          strokeWidth={2}
          fill="var(--zs-brand-ink)"
          fillOpacity={0.12}
          dot={false}
          activeDot={{ r: 4, stroke: 'var(--zs-bg-surface)', strokeWidth: 2, fill: 'var(--zs-brand-ink)' }}
          isAnimationActive={!prefersReducedMotion()}
        />
      </AreaChart>
    </ChartCard>
  )
}

function symmetricScale(rows) {
  const largest = Math.max(
    0,
    ...rows.flatMap(row => [row.imported, row.exported, row.net]).filter(isNumber).map(Math.abs),
  )
  const magnitude = largest > 0 ? 10 ** Math.floor(Math.log10(largest)) : 1
  const bound = magnitude * ([1, 2, 4, 6, 8, 10].find(step => step * magnitude >= largest) ?? 10)
  const half = bound / 2
  return {
    domain: [-bound, bound],
    ticks: [-bound, -half, 0, half, bound],
    digits: Math.max(0, -Math.floor(Math.log10(half))),
  }
}

const signedKWh = value => `${value > 0 ? '+' : ''}${formatUnit(value, 'kWh', 2)}`

function GridTooltip({ active, payload, title }) {
  const row = payload?.[0]?.payload
  if (!active || !row || row.net === null) return null
  return (
    <div className="flex flex-col gap-1 rounded-md border border-line bg-surface px-3 py-2 text-ink shadow-overlay">
      <span className="text-caption text-muted">{title(row)}</span>
      <span className="text-body-sm">
        {GRID_SERIES.imported.label}{' '}
        <span className="font-mono font-medium">{formatUnit(row.imported, 'kWh', 2)}</span>
      </span>
      <span className="text-body-sm">
        {GRID_SERIES.exported.label}{' '}
        <span className="font-mono font-medium">{formatUnit(Math.abs(row.exported), 'kWh', 2)}</span>
      </span>
      <span className="text-body-sm">
        {GRID_SERIES.net.label} <span className="font-mono font-medium">{signedKWh(row.net)}</span>{' '}
        <span className="text-muted">{row.net > 0 ? '(sacas más de lo que inyectas)' : row.net < 0 ? '(inyectas más de lo que sacas)' : ''}</span>
      </span>
    </div>
  )
}

function GridLegend() {
  return (
    <ul className="m-0 flex list-none flex-wrap gap-4 p-0 text-caption text-muted">
      {Object.entries(GRID_SERIES).map(([key, series]) => (
        <li key={key} className="flex items-center gap-2">
          <span
            aria-hidden="true"
            className={cx('inline-block', key === 'net' ? 'h-1 w-4' : 'h-3 w-3 rounded-sm')}
            style={{ background: series.color }}
          />
          {series.label}
        </li>
      ))}
      <li>Sobre 0 sacas de la red; bajo 0 inyectas.</li>
    </ul>
  )
}

function GridAxes({ rows, xAxis }) {
  const scale = symmetricScale(rows)
  return [
    <CartesianGrid key="grid" vertical={false} stroke="var(--zs-border-default)" />,
    xAxis,
    <YAxis
      key="y"
      domain={scale.domain}
      ticks={scale.ticks}
      tickFormatter={value => formatUnit(value, 'kWh', scale.digits)}
      axisLine={false}
      tickLine={false}
      width={72}
    />,
    <ReferenceLine key="zero" y={0} stroke="var(--zs-border-strong)" strokeWidth={1.5} />,
  ]
}

function hourTicks(rows, everyHours = 3) {
  if (rows.length === 0) return []
  const first = new Date(rows[0].time)
  first.setMinutes(0, 0, 0)
  const ticks = []
  for (let t = first.getTime(); t <= rows[rows.length - 1].time; t += everyHours * HOUR_MS) {
    if (t >= rows[0].time) ticks.push(t)
  }
  return ticks
}

const hourRange = row => `${formatTime(row.time)}–${formatTime(row.time + HOUR_MS)}`

function GridHourlyChart({ grid }) {
  const { hourly, totals, hasData, loading, error } = grid
  const description = totals
    ? `Energía con la red por hora hoy: sacado ${formatUnit(totals.imported, 'kWh', 2)}, inyectado ${formatUnit(Math.abs(totals.exported), 'kWh', 2)}, balance ${signedKWh(totals.net)}`
    : 'Energía con la red por hora hoy'
  const animate = !prefersReducedMotion()

  return (
    <ChartCard
      title="Red por hora"
      description={description}
      hasData={hasData}
      loading={loading}
      error={error}
      emptyText="Sin datos de la red hoy."
      footer={<GridLegend />}
    >
      <ComposedChart data={hourly} stackOffset="sign" margin={{ top: 8, right: 8, bottom: 0, left: 0 }} accessibilityLayer={false}>
        {GridAxes({
          rows: hourly,
          xAxis: (
            <XAxis
              key="x"
              dataKey="time"
              tickFormatter={formatTime}
              axisLine={false}
              tickLine={false}
              minTickGap={16}
            />
          ),
        })}
        <Tooltip cursor={{ fill: 'var(--zs-border-default)', fillOpacity: 0.4 }} content={<GridTooltip title={hourRange} />} />
        <Bar dataKey="imported" stackId="grid" fill={GRID_SERIES.imported.color} isAnimationActive={animate} />
        <Bar dataKey="exported" stackId="grid" fill={GRID_SERIES.exported.color} isAnimationActive={animate} />
        <Line
          type="monotone"
          dataKey="net"
          stroke={GRID_SERIES.net.color}
          strokeWidth={2}
          dot={{ r: 3, stroke: 'var(--zs-bg-surface)', strokeWidth: 1, fill: GRID_SERIES.net.color }}
          isAnimationActive={animate}
        />
      </ComposedChart>
    </ChartCard>
  )
}

function GridCumulativeChart({ grid }) {
  const { cumulative, totals, hasData, loading, error } = grid
  const description = totals
    ? `Energía acumulada con la red hoy: sacado ${formatUnit(totals.imported, 'kWh', 2)}, inyectado ${formatUnit(Math.abs(totals.exported), 'kWh', 2)}, balance ${signedKWh(totals.net)}`
    : 'Energía acumulada con la red hoy'
  const animate = !prefersReducedMotion()
  const areaProps = { type: 'monotone', strokeWidth: 2, fillOpacity: 0.12, dot: false, isAnimationActive: animate }

  return (
    <ChartCard
      title="Red acumulada"
      headline={totals && <span className="font-mono text-data-sm whitespace-nowrap">{signedKWh(totals.net)}</span>}
      description={description}
      hasData={hasData}
      loading={loading}
      error={error}
      emptyText="Sin datos de la red hoy."
      footer={<GridLegend />}
    >
      <ComposedChart data={cumulative} margin={{ top: 8, right: 8, bottom: 0, left: 0 }} accessibilityLayer={false}>
        {GridAxes({
          rows: cumulative,
          xAxis: (
            <XAxis
              key="x"
              dataKey="time"
              type="number"
              scale="time"
              domain={['dataMin', 'dataMax']}
              ticks={hourTicks(cumulative)}
              tickFormatter={formatTime}
              axisLine={false}
              tickLine={false}
              minTickGap={32}
            />
          ),
        })}
        <Tooltip
          cursor={{ stroke: 'var(--zs-border-strong)', strokeWidth: 1 }}
          content={<GridTooltip title={row => formatDateTime(row.time)} />}
        />
        <Area {...areaProps} dataKey="imported" stroke={GRID_SERIES.imported.color} fill={GRID_SERIES.imported.color} />
        <Area {...areaProps} dataKey="exported" stroke={GRID_SERIES.exported.color} fill={GRID_SERIES.exported.color} />
        <Line type="monotone" dataKey="net" stroke={GRID_SERIES.net.color} strokeWidth={2} dot={false} isAnimationActive={animate} />
      </ComposedChart>
    </ChartCard>
  )
}

function Tile({ label, value, unit, digits = 0, children }) {
  return (
    <div className="flex min-w-[190px] flex-1 flex-col gap-2 rounded-md border border-line bg-surface p-4">
      <span className="text-body-sm text-muted">{label}</span>
      {isNumber(value) ? (
        <DataValue value={value} unit={unit} digits={digits} />
      ) : (
        <span className="text-body-sm text-muted">Sin datos recientes</span>
      )}
      {children}
    </div>
  )
}

function BatteryTile({ summary }) {
  const soc = summary.battery_soc
  const low = isNumber(soc) && soc < LOW_SOC
  const power = summary.battery_power
  return (
    <Tile label="Batería" value={soc} unit="%">
      {isNumber(soc) && (
        <Meter
          value={soc}
          fill={low ? 'var(--zs-status-attention-fill)' : 'var(--zs-status-ok-fill)'}
          label={`Estado de carga ${formatNumber(soc, 0)} %${low ? ', bajo' : ''}`}
        />
      )}
      <div className="flex flex-wrap items-center gap-2 text-caption text-muted">
        {low && <span className="text-attention">Carga baja</span>}
        {isNumber(summary.battery_voltage) && (
          <span className="font-mono">{formatUnit(summary.battery_voltage, 'V', 1)}</span>
        )}
        {isNumber(power) && (
          <span className="font-mono">
            {power > 0 ? '+' : ''}
            {formatUnit(power, 'W')}
          </span>
        )}
      </div>
      {(summary.is_charging || summary.is_inverting) && (
        <div className="flex flex-wrap gap-2">
          {summary.is_charging && <StatusPill tone="ok">Cargando</StatusPill>}
          {summary.is_inverting && <StatusPill tone="neutral">Invirtiendo</StatusPill>}
        </div>
      )}
    </Tile>
  )
}

function EquipmentCard({ title, items }) {
  const visible = items.filter(item => item.value !== undefined)
  return (
    <Card className="min-w-[220px] flex-1 basis-[240px]">
      <h3 className="m-0 text-label">{title}</h3>
      {visible.length > 0 ? (
        <KeyValue items={visible.map(item => ({ ...item, mono: item.mono ?? true }))} />
      ) : (
        <p className="m-0 text-body-sm text-muted">Sin datos recientes</p>
      )}
    </Card>
  )
}

function EquipmentGroup({ title, entries, render }) {
  if (!entries?.length) return null
  return (
    <Section title={title} as="div">
      <div className="flex flex-wrap gap-4">
        {entries.map((entry, index) => (
          <EquipmentCard key={`${entry.instance}-${index}`} {...render(entry)} />
        ))}
      </div>
    </Section>
  )
}

const VictronDashboard = () => {
  const [systemStatus, setSystemStatus] = useState(null)
  const [lastUpdate, setLastUpdate] = useState(null)
  const [range, setRange] = useState('6h')
  const now = useNow(15000)

  const wsUrl = getWebSocketUrl('/ws/victron/status')
  const { isConnected, lastMessage, connectionError, connectionStatus, retry } = useWebSocket(wsUrl)
  const connection = connectionState({ isConnected, connectionStatus })

  const rangeConfig = TIME_RANGES.find(r => r.value === range) ?? TIME_RANGES[1]
  const soc = useMetricHistory(SOC_METRIC_NAME, rangeConfig)
  const load = useMetricHistory(LOAD_METRIC_NAME, rangeConfig)
  const solarToday = useCumulativeEnergyToday(SOLAR_POWER_METRIC_NAME)
  const loadToday = useCumulativeEnergyToday(LOAD_METRIC_NAME)
  const gridToday = useGridEnergyToday(GRID_POWER_METRIC_NAME)

  useEffect(() => {
    if (lastMessage && lastMessage.type === 'victron_status') {
      setSystemStatus(lastMessage)
      setLastUpdate(new Date())
    }
  }, [lastMessage])

  const summary = systemStatus?.system
  const data = systemStatus?.data
  const hasEquipment = EQUIPMENT_KEYS.some(key => data?.[key]?.length > 0)

  const todayHeadline = (history, total) =>
    history.length > 0 ? <DataValue value={total} unit="kWh" digits={2} size="sm" /> : null

  return (
    <Page>
      <PageHeader
        title="Energía"
        description="Producción solar, batería y consumo del sistema Victron, en vivo."
        actions={
          <div className="flex flex-wrap items-center gap-3">
            {lastUpdate && (
              <span className="text-caption text-muted">
                Actualizado {formatRelative(lastUpdate, now)} · {formatTime(lastUpdate)}
              </span>
            )}
            <StatusPill tone={connection.tone}>{connection.label}</StatusPill>
          </div>
        }
      />

      {connectionError && !isConnected && (
        <Notice action="Reintentar conexión" onAction={retry}>
          Se perdió la conexión con el sistema de energía. Revisa tu red y vuelve a intentarlo.
        </Notice>
      )}

      <Section title="Ahora">
        {!summary ? (
          connection.tone === 'attention' ? (
            <EmptyState title="Sin datos recientes">
              No llegan lecturas del sistema de energía. Cuando se restablezca la conexión, verás aquí la producción
              solar, la batería y el consumo.
            </EmptyState>
          ) : (
            <LoadingState>Esperando lecturas del sistema de energía…</LoadingState>
          )
        ) : (
          <div className={cx('flex flex-wrap gap-4', !isConnected && 'opacity-offline')}>
            <Tile label="Solar" value={summary.solar_power} unit="W" />
            <BatteryTile summary={summary} />
            <Tile label="Consumo" value={summary.ac_load_power} unit="W" />
            <Tile label="Red" value={summary.grid_power} unit="W">
              {typeof summary.ac_input_connected === 'boolean' && (
                <span className="text-caption text-muted">
                  {summary.ac_input_connected ? 'Entrada de red conectada' : 'Entrada de red desconectada'}
                </span>
              )}
            </Tile>
          </div>
        )}
      </Section>

      <Section
        title="Historial"
        actions={
          <FilterChips
            label="Rango del historial"
            options={TIME_RANGES.map(r => ({ value: r.value, label: r.label }))}
            value={range}
            onChange={setRange}
          />
        }
      >
        <CardGrid>
          <MetricChart
            title="Carga de batería"
            history={soc.history}
            loading={soc.loading}
            error={soc.error}
            unit="%"
            digits={1}
            spanMs={rangeConfig.ms}
            domain={[0, 100]}
            reference={{ y: LOW_SOC, label: `carga baja (${LOW_SOC} %)` }}
            emptyText="Sin datos recientes en este rango."
          />
          <MetricChart
            title="Consumo"
            history={load.history}
            loading={load.loading}
            error={load.error}
            unit="W"
            digits={0}
            spanMs={rangeConfig.ms}
            domain={[0, 'auto']}
            emptyText="Sin datos recientes en este rango."
          />
        </CardGrid>
      </Section>

      <Section title="Energía de hoy" description="Acumulada desde la medianoche.">
        <CardGrid>
          <MetricChart
            title="Generación solar"
            history={solarToday.history}
            loading={solarToday.loading}
            error={solarToday.error}
            unit="kWh"
            digits={2}
            spanMs={DAY_MS}
            domain={[0, 'auto']}
            emptyText="Sin datos recientes hoy."
            headline={todayHeadline(solarToday.history, solarToday.totalKWh)}
          />
          <MetricChart
            title="Consumo"
            history={loadToday.history}
            loading={loadToday.loading}
            error={loadToday.error}
            unit="kWh"
            digits={2}
            spanMs={DAY_MS}
            domain={[0, 'auto']}
            emptyText="Sin datos recientes hoy."
            headline={todayHeadline(loadToday.history, loadToday.totalKWh)}
          />
        </CardGrid>
      </Section>

      <Section title="Red de hoy" description="Lo que sacas de la red y lo que inyectas, desde la medianoche.">
        <CardGrid>
          <GridHourlyChart grid={gridToday} />
          <GridCumulativeChart grid={gridToday} />
        </CardGrid>
      </Section>

      {hasEquipment && (
        <Section title="Equipos" className={cx(!isConnected && 'opacity-offline')}>
          <EquipmentGroup
            title="Baterías"
            entries={data.batteries}
            render={b => ({
              title: `Batería ${b.instance}`,
              items: [
                { label: 'Carga', value: unitOrEmpty(b.soc, '%') },
                { label: 'Voltaje', value: unitOrEmpty(b.voltage, 'V', 2) },
                { label: 'Corriente', value: unitOrEmpty(b.current, 'A', 1) },
                { label: 'Potencia', value: unitOrEmpty(b.power, 'W') },
                { label: 'Temperatura', value: unitOrEmpty(b.temperature, '°C', 1) },
              ],
            })}
          />
          <EquipmentGroup
            title="Cargadores solares"
            entries={data.solar_chargers}
            render={s => ({
              title: `Cargador solar ${s.instance}`,
              items: [
                { label: 'Potencia', value: unitOrEmpty(s.power, 'W') },
                { label: 'Voltaje', value: unitOrEmpty(s.voltage, 'V', 1) },
                { label: 'Corriente', value: unitOrEmpty(s.current, 'A', 1) },
              ],
            })}
          />
          <EquipmentGroup
            title="Cargas de corriente alterna"
            entries={data.ac_loads}
            render={l => ({
              title: `Carga ${l.instance}`,
              items: [
                { label: 'Potencia', value: unitOrEmpty(l.power, 'W') },
                { label: 'Voltaje', value: unitOrEmpty(l.voltage, 'V', 1) },
                { label: 'Corriente', value: unitOrEmpty(l.current, 'A', 1) },
              ],
            })}
          />
          <EquipmentGroup
            title="Inversor y cargador"
            entries={data.vebus}
            render={v => ({
              title: `Inversor ${v.instance}`,
              items: [
                { label: 'Potencia', value: unitOrEmpty(v.power, 'W') },
                { label: 'Estado', value: vebusState(v.state), mono: false },
              ],
            })}
          />
          <EquipmentGroup
            title="Temperaturas"
            entries={data.temperatures}
            render={t => ({
              title: `Sensor ${t.instance}`,
              items: [{ label: 'Temperatura', value: unitOrEmpty(t.temperature, '°C', 1) }],
            })}
          />
          <EquipmentGroup
            title="Estanques"
            entries={data.tanks}
            render={t => ({
              title: `Estanque ${t.instance}`,
              items: [{ label: 'Nivel', value: unitOrEmpty(t.level, '%') }],
            })}
          />
        </Section>
      )}
    </Page>
  )
}

export default VictronDashboard
