package async_test

import (
	"context"

	"zensor-server/internal/infra/async"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Local Broker", func() {
	var broker *async.LocalBroker
	var topic async.BrokerTopicName
	var subscription async.Subscription
	var message async.BrokerMessage
	var ctx context.Context

	BeforeEach(func() {
		broker = async.NewLocalBroker()
		ctx = context.TODO()
	})

	Context("Subscribe", func() {
		When("add a new subscriber for a topic", func() {
			BeforeEach(func() {
				topic = "182efcc3-5b44-475f-a3d0-0a46c0311fb8"
			})

			It("should increase the number of subscriber", func() {
				subscription, _ = broker.Subscribe(topic)

				broker.Publish(ctx, topic, async.BrokerMessage{})

				Eventually(subscription.Receiver).Should(Receive(&async.BrokerMessage{}))
			})
		})

		When("multiple subscriptor", func() {
			var subscription2 async.Subscription
			BeforeEach(func() {
				topic = "182efcc3-5b44-475f-a3d0-0a46c0311fb8"
			})

			It("should increase the number of subscriber", func() {
				subscription, _ = broker.Subscribe(topic)
				subscription2, _ = broker.Subscribe(topic)

				broker.Publish(ctx, topic, async.BrokerMessage{})

				Eventually(subscription.Receiver).Should(Receive(&async.BrokerMessage{}))
				Eventually(subscription2.Receiver).Should(Receive(&async.BrokerMessage{}))
			})
		})

		When("a new message arrives", func() {
			BeforeEach(func() {
				topic = "806ce863-42d8-442f-8517-b8d2a9408767"
				subscription, _ = broker.Subscribe(topic)
				message = async.BrokerMessage{
					Event: "f20a4b57-95bc-4f2a-b3e6-7e36e05f1b23",
					Value: "a7024c11-3a52-4a29-9360-bc8e6a29ded1",
				}
			})

			It("should receive a message from channel", func() {
				broker.Publish(context.TODO(), topic, message)

				Eventually(subscription.Receiver).Should(Receive(And(
					HaveField("Event", "f20a4b57-95bc-4f2a-b3e6-7e36e05f1b23"),
					HaveField("Value", "a7024c11-3a52-4a29-9360-bc8e6a29ded1"),
				)))
			})
		})

		When("stop broker", func() {
			BeforeEach(func() {
				topic = "806ce863-42d8-442f-8517-b8d2a9408767"
				subscription, _ = broker.Subscribe(topic)
				message = async.BrokerMessage{
					Event: "f20a4b57-95bc-4f2a-b3e6-7e36e05f1b23",
					Value: "a7024c11-3a52-4a29-9360-bc8e6a29ded1",
				}
			})

			It("should receive a done message", func() {
				go broker.Stop()

				Eventually(subscription.Receiver).Should(BeClosed())
			})
		})

		When("topic doesn't exists", func() {
			BeforeEach(func() {
				topic = "806ce863-42d8-442f-8517-b8d2a9408767"
				message = async.BrokerMessage{
					Event: "f20a4b57-95bc-4f2a-b3e6-7e36e05f1b23",
					Value: "a7024c11-3a52-4a29-9360-bc8e6a29ded1",
				}
			})

			It("should receive a message from channel", func() {
				err := broker.Publish(context.TODO(), topic, message)

				Expect(err).ShouldNot(Succeed())
			})
		})
	})

	Context("Unsubscribe", func() {
		When("there is no subscriptor", func() {
			BeforeEach(func() {
				topic = "b6541d7c-f455-446c-bea0-1e11bf9c76fc"
				subscription = async.Subscription{
					ID: "2d582ce4-88e1-40a8-bc14-5cf0311943fd",
				}
			})

			It("should do nothing", func() {
				err := broker.Unsubscribe(topic, subscription)

				Expect(err).Should(MatchError(async.ErrTopicNotFound))
			})
		})
		When("subscriptor doesn't exists", func() {
			var subscription2 async.Subscription
			BeforeEach(func() {
				topic = "b6541d7c-f455-446c-bea0-1e11bf9c76fc"
				subscription, _ = broker.Subscribe(topic)
				subscription2 = async.Subscription{
					ID: "2d582ce4-88e1-40a8-bc14-5cf0311943fd",
				}
			})

			It("should do nothing", func() {
				err := broker.Unsubscribe(topic, subscription2)

				Expect(err).Should(MatchError(async.ErrSubscriptorNotFound))
			})
		})
		When("subscriptor does exists", func() {
			BeforeEach(func() {
				topic = "b6541d7c-f455-446c-bea0-1e11bf9c76fc"
				subscription, _ = broker.Subscribe(topic)
				broker.Unsubscribe(topic, subscription)
				message = async.BrokerMessage{
					Event: "f20a4b57-95bc-4f2a-b3e6-7e36e05f1b23",
					Value: "a7024c11-3a52-4a29-9360-bc8e6a29ded1",
				}
			})

			It("should not receive any message subscriptor", func() {
				broker.Publish(context.TODO(), topic, message)

				Eventually(subscription.Receiver).ShouldNot(Receive(And(
					HaveField("Event", "f20a4b57-95bc-4f2a-b3e6-7e36e05f1b23"),
					HaveField("Value", "a7024c11-3a52-4a29-9360-bc8e6a29ded1"),
				)))
			})
		})
		When("is called twice", func() {
			BeforeEach(func() {
				topic = "b6541d7c-f455-446c-bea0-1e11bf9c76fc"
				subscription, _ = broker.Subscribe(topic)
				broker.Unsubscribe(topic, subscription)

			})

			It("should remove subscriptor and don't panic", func() {
				err := broker.Unsubscribe(topic, subscription)

				Expect(err).Should(Succeed())
			})
		})
	})

	Context("Publish", func() {
		When("topic doesn't exists", func() {
			BeforeEach(func() {
				topic = "bcf9eba6-519e-4983-85b2-aa58d31c8a01"
			})

			It("should return an error", func() {
				err := broker.Publish(context.TODO(), topic, async.BrokerMessage{})

				Expect(err).ShouldNot(Succeed())
			})
		})

		When("there is no subscriptor", func() {
			BeforeEach(func() {
				topic = "bcf9eba6-519e-4983-85b2-aa58d31c8a01"
				subcription, _ := broker.Subscribe(topic)
				broker.Unsubscribe(topic, subcription)
			})

			It("should return no error", func() {
				err := broker.Publish(context.TODO(), topic, async.BrokerMessage{})

				Expect(err).Should(Succeed())
			})
		})

		When("there is at least one subscriptor", func() {
			BeforeEach(func() {
				topic = "bcf9eba6-519e-4983-85b2-aa58d31c8a01"
				subscription, _ = broker.Subscribe(topic)
			})

			It("should return no error", func() {
				err := broker.Publish(context.TODO(), topic, async.BrokerMessage{})

				Expect(err).Should(Succeed())
			})
		})
	})
})
