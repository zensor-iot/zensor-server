package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type fakeSessionResolver struct {
	sessions map[string]domain.Session
}

func (f *fakeSessionResolver) GetSession(_ context.Context, sessionID string) (domain.Session, error) {
	session, found := f.sessions[sessionID]
	if !found {
		return domain.Session{}, ErrNoSession
	}
	return session, nil
}

type fakeAPIKeyResolver struct {
	keys map[string]domain.APIKey
	err  error
}

func (f *fakeAPIKeyResolver) Validate(_ context.Context, rawKey string) (domain.APIKey, error) {
	if f.err != nil {
		return domain.APIKey{}, f.err
	}
	key, found := f.keys[rawKey]
	if !found {
		return domain.APIKey{}, errors.New("api key not found")
	}
	return key, nil
}

var _ = ginkgo.Describe("AuthMiddleware", func() {
	var (
		resolver    *fakeSessionResolver
		keyResolver *fakeAPIKeyResolver
		handler     http.Handler
		seenID      string
		seenName    string
		seenMail    string
	)

	ginkgo.BeforeEach(func() {
		resolver = &fakeSessionResolver{sessions: map[string]domain.Session{
			"valid-session": {
				ID:        "valid-session",
				UserID:    "user-1",
				Email:     "user@example.com",
				Name:      "User One",
				IsAdmin:   false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			"admin-session": {
				ID:        "admin-session",
				UserID:    "admin-1",
				Email:     "admin@example.com",
				Name:      "Admin One",
				IsAdmin:   true,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		}}

		keyResolver = &fakeAPIKeyResolver{keys: map[string]domain.APIKey{
			"zsk_valid": {
				ID:        "key-1",
				Name:      "grafana-sync",
				KeyPrefix: "zsk_valid",
			},
		}}

		seenID, seenName, seenMail = "", "", ""
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenID = r.Header.Get("X-User-ID")
			seenName = r.Header.Get("X-User-Name")
			seenMail = r.Header.Get("X-User-Email")
			w.WriteHeader(http.StatusOK)
		})
		handler = NewAuthMiddleware(resolver, keyResolver)(inner)
	})

	request := func(path, sessionID string, spoofHeaders bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if sessionID != "" {
			req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionID})
		}
		if spoofHeaders {
			req.Header.Set("X-User-ID", "spoofed-id")
			req.Header.Set("X-User-Name", "Spoofed")
			req.Header.Set("X-User-Email", "spoofed@example.com")
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	bearerRequest := func(path, authorization string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if authorization != "" {
			req.Header.Set("Authorization", authorization)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	ginkgo.Context("header stripping", func() {
		ginkgo.When("a request carries spoofed X-User headers on a public path", func() {
			ginkgo.It("should strip them before reaching the handler", func() {
				rec := request("/healthz", "", true)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(seenID).To(gomega.BeEmpty())
				gomega.Expect(seenName).To(gomega.BeEmpty())
				gomega.Expect(seenMail).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("a request with a valid session also carries spoofed headers", func() {
			ginkgo.It("should replace them with the session identity", func() {
				rec := request("/v1/tenants", "valid-session", true)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(seenID).To(gomega.Equal("user-1"))
				gomega.Expect(seenName).To(gomega.Equal("User One"))
				gomega.Expect(seenMail).To(gomega.Equal("user@example.com"))
			})
		})
	})

	ginkgo.Context("protected paths", func() {
		ginkgo.When("requesting /v1/ without a session", func() {
			ginkgo.It("should return 401", func() {
				rec := request("/v1/tenants", "", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("requesting /ws/ without a session", func() {
			ginkgo.It("should return 401", func() {
				rec := request("/ws/device-messages", "", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("the session cookie is invalid", func() {
			ginkgo.It("should return 401", func() {
				rec := request("/v1/tenants", "unknown-session", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("the session is valid", func() {
			ginkgo.It("should pass through /v1/ requests", func() {
				rec := request("/v1/tenants", "valid-session", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})

			ginkgo.It("should pass through /ws/ requests", func() {
				rec := request("/ws/device-messages", "valid-session", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("public paths", func() {
		ginkgo.When("requesting public routes without a session", func() {
			ginkgo.It("should allow /healthz", func() {
				gomega.Expect(request("/healthz", "", false).Code).To(gomega.Equal(http.StatusOK))
			})

			ginkgo.It("should allow /metrics", func() {
				gomega.Expect(request("/metrics", "", false).Code).To(gomega.Equal(http.StatusOK))
			})

			ginkgo.It("should allow /auth/login", func() {
				gomega.Expect(request("/auth/login", "", false).Code).To(gomega.Equal(http.StatusOK))
			})

			ginkgo.It("should allow /ui/ pages", func() {
				gomega.Expect(request("/ui/access-denied", "", false).Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("admin gate", func() {
		ginkgo.When("a non-admin session requests an admin route", func() {
			ginkgo.It("should return 403", func() {
				rec := request("/v1/admin/allowed-users", "valid-session", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusForbidden))
			})
		})

		ginkgo.When("an admin session requests an admin route", func() {
			ginkgo.It("should pass through", func() {
				rec := request("/v1/admin/allowed-users", "admin-session", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("no session requests an admin route", func() {
			ginkgo.It("should return 401", func() {
				rec := request("/v1/admin/allowed-users", "", false)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})
	})

	ginkgo.Context("api key authentication", func() {
		ginkgo.When("a valid bearer key requests a protected route", func() {
			ginkgo.It("should pass through with the key's synthetic identity", func() {
				rec := bearerRequest("/v1/tenants", "Bearer zsk_valid")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(seenID).To(gomega.Equal("key-1"))
				gomega.Expect(seenName).To(gomega.Equal("grafana-sync"))
				gomega.Expect(seenMail).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("a valid bearer key requests an admin route", func() {
			ginkgo.It("should return 403 unconditionally", func() {
				rec := bearerRequest("/v1/admin/allowed-users", "Bearer zsk_valid")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusForbidden))
			})
		})

		ginkgo.When("the bearer key is unknown", func() {
			ginkgo.It("should return 401", func() {
				rec := bearerRequest("/v1/tenants", "Bearer zsk_unknown")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("the authorization header is malformed", func() {
			ginkgo.It("should return 401 for a non-bearer scheme", func() {
				rec := bearerRequest("/v1/tenants", "Basic zsk_valid")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})

			ginkgo.It("should return 401 for an empty bearer token", func() {
				rec := bearerRequest("/v1/tenants", "Bearer ")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("the key resolver fails", func() {
			ginkgo.It("should fail closed with 401", func() {
				keyResolver.err = errors.New("database down")

				rec := bearerRequest("/v1/tenants", "Bearer zsk_valid")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("a request carries both a session cookie and a bearer key", func() {
			ginkgo.It("should prefer the session identity", func() {
				req := httptest.NewRequest(http.MethodGet, "/v1/tenants", nil)
				req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "valid-session"})
				req.Header.Set("Authorization", "Bearer zsk_valid")
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(seenID).To(gomega.Equal("user-1"))
				gomega.Expect(seenMail).To(gomega.Equal("user@example.com"))
			})
		})
	})
})
