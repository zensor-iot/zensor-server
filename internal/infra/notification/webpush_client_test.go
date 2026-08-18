package notification_test

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"zensor-server/internal/infra/notification"

	webpush "github.com/SherClockHolmes/webpush-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WebPushClient", func() {
	var (
		client           *notification.WebPushClient
		pushService      *httptest.Server
		responseStatus   *atomic.Int64
		receivedRequests *atomic.Int64
		lastAuthHeader   *atomic.Value
		subscriptionJSON string
		request          notification.PushNotificationRequest
	)

	BeforeEach(func() {
		responseStatus = &atomic.Int64{}
		responseStatus.Store(http.StatusCreated)
		receivedRequests = &atomic.Int64{}
		lastAuthHeader = &atomic.Value{}
		pushService = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedRequests.Add(1)
			lastAuthHeader.Store(r.Header.Get("Authorization"))
			w.WriteHeader(int(responseStatus.Load()))
		}))
		DeferCleanup(pushService.Close)

		vapidPrivateKey, vapidPublicKey, err := webpush.GenerateVAPIDKeys()
		Expect(err).NotTo(HaveOccurred())

		client = notification.NewWebPushClient(notification.WebPushConfig{
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
			Subscriber:      "mailto:admin@zensor-iot.net",
		})

		browserKey, err := ecdh.P256().GenerateKey(rand.Reader)
		Expect(err).NotTo(HaveOccurred())
		authSecret := make([]byte, 16)
		_, err = rand.Read(authSecret)
		Expect(err).NotTo(HaveOccurred())

		subscription := map[string]any{
			"endpoint": pushService.URL,
			"keys": map[string]string{
				"p256dh": base64.RawURLEncoding.EncodeToString(browserKey.PublicKey().Bytes()),
				"auth":   base64.RawURLEncoding.EncodeToString(authSecret),
			},
		}
		subscriptionBytes, err := json.Marshal(subscription)
		Expect(err).NotTo(HaveOccurred())
		subscriptionJSON = string(subscriptionBytes)
	})

	Context("SendPushNotification", func() {
		BeforeEach(func() {
			request = notification.PushNotificationRequest{
				Token:    subscriptionJSON,
				Title:    "Execution Reminder",
				Body:     "Scheduled execution is due soon",
				DeepLink: "/maintenance/executions/abc",
			}
		})

		When("the push service accepts the notification", func() {
			It("should deliver a VAPID-authenticated request to the subscription endpoint", func() {
				err := client.SendPushNotification(context.Background(), request)
				Expect(err).NotTo(HaveOccurred())
				Expect(receivedRequests.Load()).To(Equal(int64(1)))
				authHeader, ok := lastAuthHeader.Load().(string)
				Expect(ok).To(BeTrue())
				Expect(authHeader).To(HavePrefix("vapid"))
			})
		})

		When("the token is not valid subscription JSON", func() {
			BeforeEach(func() {
				request.Token = "not-json"
			})

			It("should return an error without calling the push service", func() {
				err := client.SendPushNotification(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(receivedRequests.Load()).To(Equal(int64(0)))
			})
		})

		When("the push service returns 410 Gone", func() {
			BeforeEach(func() {
				responseStatus.Store(http.StatusGone)
			})

			It("should return ErrSubscriptionExpired", func() {
				err := client.SendPushNotification(context.Background(), request)
				Expect(err).To(MatchError(notification.ErrSubscriptionExpired))
			})
		})

		When("the push service returns 404 Not Found", func() {
			BeforeEach(func() {
				responseStatus.Store(http.StatusNotFound)
			})

			It("should return ErrSubscriptionExpired", func() {
				err := client.SendPushNotification(context.Background(), request)
				Expect(err).To(MatchError(notification.ErrSubscriptionExpired))
			})
		})

		When("the push service returns an unexpected error status", func() {
			BeforeEach(func() {
				responseStatus.Store(http.StatusInternalServerError)
			})

			It("should return an error that is not ErrSubscriptionExpired", func() {
				err := client.SendPushNotification(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(err).NotTo(MatchError(notification.ErrSubscriptionExpired))
			})
		})
	})

	Context("SendEmail", func() {
		It("should return an error", func() {
			err := client.SendEmail(context.Background(), notification.EmailRequest{To: "a@b.c"})
			Expect(err).To(HaveOccurred())
		})
	})
})
