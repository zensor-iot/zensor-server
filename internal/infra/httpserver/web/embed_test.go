package web_test

import (
	"net/http"
	"net/http/httptest"
	"zensor-server/internal/infra/httpserver/web"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("SPAHandler", func() {
	var recorder *httptest.ResponseRecorder

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
	})

	Context("known asset", func() {
		When("the asset exists in dist/assets", func() {
			It("should serve it with the correct content type", func() {
				request := httptest.NewRequest(http.MethodGet, "/assets/test-fixture.css", nil)

				web.SPAHandler().ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.HavePrefix("text/css"))
			})
		})

		When("the asset does not exist", func() {
			It("should return 404, not the SPA fallback", func() {
				request := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)

				web.SPAHandler().ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
				gomega.Expect(recorder.Body.String()).NotTo(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})
	})

	Context("client-side route", func() {
		When("the path is not a known asset or file", func() {
			It("should fall back to index.html", func() {
				request := httptest.NewRequest(http.MethodGet, "/tenants/example/devices", nil)

				web.SPAHandler().ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Body.String()).To(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})
	})
})
