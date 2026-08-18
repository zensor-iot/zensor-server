package usecases_test

import (
	"context"
	"errors"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/victron/usecases"

	victrondto "zensor-server/internal/victron/dto"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func collectVictronGauge(reader *metric.ManualReader, ctx context.Context, name string) (float64, bool) {
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		return 0, false
	}

	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != name {
				continue
			}
			gauge, ok := m.Data.(metricdata.Gauge[float64])
			if !ok {
				continue
			}
			if len(gauge.DataPoints) == 0 {
				return 0, false
			}
			return gauge.DataPoints[0].Value, true
		}
	}
	return 0, false
}

func victronGaugeAttributes(reader *metric.ManualReader, ctx context.Context, name string) []attribute.KeyValue {
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		return nil
	}

	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != name {
				continue
			}
			gauge, ok := m.Data.(metricdata.Gauge[float64])
			if !ok {
				continue
			}
			if len(gauge.DataPoints) == 0 {
				return nil
			}
			return gauge.DataPoints[0].Attributes.ToSlice()
		}
	}
	return nil
}

var _ = ginkgo.Describe("VictronMetricWorker", func() {
	var (
		broker *async.LocalBroker
		worker *usecases.VictronMetricWorker
		reader *metric.ManualReader
		cancel context.CancelFunc
	)

	ginkgo.BeforeEach(func() {
		reader = metric.NewManualReader()
		provider := metric.NewMeterProvider(metric.WithReader(reader))
		otel.SetMeterProvider(provider)

		broker = async.NewLocalBroker()
		worker = usecases.NewVictronMetricWorker(broker)

		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		go worker.Run(ctx, func() {})

		gomega.Eventually(func() error {
			return broker.Publish(ctx, usecases.BrokerTopicVictronData, async.BrokerMessage{})
		}, 5*time.Second, 50*time.Millisecond).Should(gomega.Succeed())
	})

	ginkgo.AfterEach(func() {
		cancel()
	})

	ginkgo.Context("when telemetry points arrive on the victron topic", func() {
		ginkgo.When("a numeric battery voltage is published", func() {
			ginkgo.It("should record a gauge named after the service type and path", func() {
				gomega.Expect(broker.Publish(context.Background(), usecases.BrokerTopicVictronData, async.BrokerMessage{
					Value: victrondto.VictronTelemetry{
						PortalID:    "d41243b4e8e4",
						ServiceType: "battery",
						Instance:    512,
						Path:        "Dc/0/Voltage",
						Value:       victrondto.VictronValue{Value: 12.8, Numeric: true},
					},
				})).To(gomega.Succeed())

				gomega.Eventually(func() (float64, error) {
					value, ok := collectVictronGauge(reader, context.Background(), "zensor_server_victron_battery_dc_0_voltage")
					if !ok {
						return 0, errors.New("gauge not recorded yet")
					}
					return value, nil
				}, 5*time.Second, 50*time.Millisecond).Should(gomega.Equal(12.8))
			})

			ginkgo.It("should tag the gauge with portal id, service type, instance, and path", func() {
				gomega.Expect(broker.Publish(context.Background(), usecases.BrokerTopicVictronData, async.BrokerMessage{
					Value: victrondto.VictronTelemetry{
						PortalID:    "d41243b4e8e4",
						ServiceType: "battery",
						Instance:    512,
						Path:        "Soc",
						Value:       victrondto.VictronValue{Value: 87, Numeric: true},
					},
				})).To(gomega.Succeed())

				gomega.Eventually(func() []attribute.KeyValue {
					return victronGaugeAttributes(reader, context.Background(), "zensor_server_victron_battery_soc")
				}, 5*time.Second, 50*time.Millisecond).Should(gomega.ContainElements(
					attribute.String("portal_id", "d41243b4e8e4"),
					attribute.String("service_type", "battery"),
					attribute.Int("instance", 512),
					attribute.String("path", "Soc"),
				))
			})
		})

		ginkgo.When("a text-only value such as a serial number is published", func() {
			ginkgo.It("should not record a metric", func() {
				gomega.Expect(broker.Publish(context.Background(), usecases.BrokerTopicVictronData, async.BrokerMessage{
					Value: victrondto.VictronTelemetry{
						PortalID:    "d41243b4e8e4",
						ServiceType: "system",
						Instance:    0,
						Path:        "Serial",
						Value:       victrondto.VictronValue{Text: "d41243b4e8e4"},
					},
				})).To(gomega.Succeed())

				gomega.Consistently(func() bool {
					_, ok := collectVictronGauge(reader, context.Background(), "zensor_server_victron_system_serial")
					return ok
				}, 500*time.Millisecond, 100*time.Millisecond).Should(gomega.BeFalse())
			})
		})
	})
})
