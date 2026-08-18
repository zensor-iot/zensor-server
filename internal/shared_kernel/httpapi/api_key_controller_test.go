package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/httpapi"
	"zensor-server/internal/shared_kernel/usecases"

	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("APIKeyController", func() {
	var (
		ctrl    *gomock.Controller
		service *mockusecases.MockAPIKeyService
		router  *http.ServeMux
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		service = mockusecases.NewMockAPIKeyService(ctrl)
		router = http.NewServeMux()
		httpapi.NewAPIKeyController(service).AddRoutes(router)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("Create", func() {
		createRequest := func(body string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/api-keys", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "admin-1")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			return rec
		}

		ginkgo.When("the request is valid", func() {
			ginkgo.It("should return 201 with the plaintext key exactly once", func() {
				created := domain.APIKey{
					ID:        domain.ID("key-1"),
					Name:      "grafana-sync",
					KeyHash:   "hash",
					KeyPrefix: "zsk_ab12cd34",
					CreatedBy: domain.ID("admin-1"),
					CreatedAt: time.Now(),
				}
				service.EXPECT().Create(gomock.Any(), "grafana-sync", domain.ID("admin-1")).
					Return(created, "zsk_ab12cd34plaintext", nil)

				rec := createRequest(`{"name":"grafana-sync"}`)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusCreated))

				var body map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["id"]).To(gomega.Equal("key-1"))
				gomega.Expect(body["name"]).To(gomega.Equal("grafana-sync"))
				gomega.Expect(body["key"]).To(gomega.Equal("zsk_ab12cd34plaintext"))
				gomega.Expect(body["key_prefix"]).To(gomega.Equal("zsk_ab12cd34"))
				gomega.Expect(body).NotTo(gomega.HaveKey("key_hash"))
			})
		})

		ginkgo.When("the name is blank", func() {
			ginkgo.It("should return 400", func() {
				service.EXPECT().Create(gomock.Any(), "", domain.ID("admin-1")).
					Return(domain.APIKey{}, "", domain.ErrAPIKeyNameRequired)

				rec := createRequest(`{"name":""}`)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the body is not valid JSON", func() {
			ginkgo.It("should return 400", func() {
				rec := createRequest(`{invalid`)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the name already exists", func() {
			ginkgo.It("should return 409", func() {
				service.EXPECT().Create(gomock.Any(), "grafana-sync", domain.ID("admin-1")).
					Return(domain.APIKey{}, "", usecases.ErrAPIKeyDuplicated)

				rec := createRequest(`{"name":"grafana-sync"}`)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusConflict))
			})
		})
	})

	ginkgo.Context("List", func() {
		ginkgo.When("keys exist", func() {
			ginkgo.It("should return them without key material", func() {
				keys := []domain.APIKey{
					{
						ID:        domain.ID("key-1"),
						Name:      "grafana-sync",
						KeyHash:   "hash",
						KeyPrefix: "zsk_ab12cd34",
						CreatedAt: time.Now(),
					},
				}
				service.EXPECT().List(gomock.Any()).Return(keys, nil)

				req := httptest.NewRequest(http.MethodGet, "/v1/admin/api-keys", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body []map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body).To(gomega.HaveLen(1))
				gomega.Expect(body[0]["id"]).To(gomega.Equal("key-1"))
				gomega.Expect(body[0]["key_prefix"]).To(gomega.Equal("zsk_ab12cd34"))
				gomega.Expect(body[0]).NotTo(gomega.HaveKey("key"))
				gomega.Expect(body[0]).NotTo(gomega.HaveKey("key_hash"))
			})
		})
	})

	ginkgo.Context("Revoke", func() {
		ginkgo.When("the key exists", func() {
			ginkgo.It("should return 204", func() {
				service.EXPECT().Revoke(gomock.Any(), domain.ID("key-1")).Return(nil)

				req := httptest.NewRequest(http.MethodDelete, "/v1/admin/api-keys/key-1", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNoContent))
			})
		})

		ginkgo.When("the key does not exist", func() {
			ginkgo.It("should return 404", func() {
				service.EXPECT().Revoke(gomock.Any(), domain.ID("missing")).
					Return(usecases.ErrAPIKeyNotFound)

				req := httptest.NewRequest(http.MethodDelete, "/v1/admin/api-keys/missing", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
			})
		})
	})
})
