package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"zensor-server/internal/shared_kernel/httpapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WebPushController", func() {
	Context("getVAPIDPublicKey", func() {
		var (
			router   *http.ServeMux
			recorder *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			router = http.NewServeMux()
			recorder = httptest.NewRecorder()
		})

		When("a VAPID public key is configured", func() {
			BeforeEach(func() {
				controller := httpapi.NewWebPushController("test-public-key")
				controller.AddRoutes(router)
			})

			It("should return the public key", func() {
				request := httptest.NewRequest(http.MethodGet, "/v1/push/vapid-public-key", nil)
				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				var body map[string]string
				Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(Succeed())
				Expect(body["public_key"]).To(Equal("test-public-key"))
			})
		})

		When("no VAPID public key is configured", func() {
			BeforeEach(func() {
				controller := httpapi.NewWebPushController("")
				controller.AddRoutes(router)
			})

			It("should return 404", func() {
				request := httptest.NewRequest(http.MethodGet, "/v1/push/vapid-public-key", nil)
				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})
})
