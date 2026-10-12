const KINDS = [
  { kind: 'relay', match: /relay|valve/i, label: 'Válvula', unit: null },
  { kind: 'humidity', match: /humid|moisture/i, label: 'Humedad', unit: '%' },
  { kind: 'temperature', match: /temp/i, label: 'Temperatura', unit: '°C' },
  { kind: 'flow', match: /flow/i, label: 'Caudal', unit: 'L/min' },
  { kind: 'pressure', match: /pressure/i, label: 'Presión', unit: 'bar' },
]

const ORDER = ['humidity', 'temperature', 'flow', 'pressure', 'relay', 'other']

const readingValue = reading => reading?.Value ?? reading?.value
const readingIndex = reading => reading?.Index ?? reading?.index

export function classifySensor(sensorType) {
  return KINDS.find(entry => entry.match.test(sensorType)) || { kind: 'other', label: sensorType, unit: null }
}

export function sensorReadings(sensorData) {
  const data = sensorData?.data
  if (!data || typeof data !== 'object') return []

  const readings = Object.entries(data).flatMap(([sensorType, values]) => {
    if (!Array.isArray(values)) return []
    const { kind, label, unit } = classifySensor(sensorType)
    return values
      .filter(reading => readingValue(reading) !== undefined && readingValue(reading) !== null)
      .map(reading => {
        const index = readingIndex(reading)
        return {
          key: `${sensorType}-${index}`,
          kind,
          sensorType,
          index,
          label: values.length > 1 && index !== undefined ? `${label} #${index}` : label,
          value: readingValue(reading),
          unit,
        }
      })
  })

  return readings.sort((a, b) => ORDER.indexOf(a.kind) - ORDER.indexOf(b.kind) || (a.index ?? 0) - (b.index ?? 0))
}

export function isValveOpen(sensorData) {
  return sensorReadings(sensorData).some(reading => reading.kind === 'relay' && Number(reading.value) === 1)
}

export function hasValveReading(sensorData) {
  return sensorReadings(sensorData).some(reading => reading.kind === 'relay')
}
