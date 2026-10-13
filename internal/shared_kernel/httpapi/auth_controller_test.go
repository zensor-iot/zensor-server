package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/httpapi"
	"zensor-server/internal/shared_kernel/usecases"

	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("AuthController", func() {
	var (
		ctrl    *gomock.Controller
		service *mockusecases.MockAuthService
		router  *http.ServeMux
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		service = mockusecases.NewMockAuthService(ctrl)
		router = http.NewServeMux()
		httpapi.NewAuthController(service, "zensor://auth/callback").AddRoutes(router)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	cookieByName := func(rec *httptest.ResponseRecorder, name string) *http.Cookie {
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == name {
				return cookie
			}
		}
		return nil
	}

	ginkgo.Context("Login", func() {
		ginkgo.When("starting a login", func() {
			ginkgo.It("should set a state cookie and redirect to the provider", func() {
				var capturedState string
				service.EXPECT().AuthCodeURL(gomock.Any()).DoAndReturn(func(state string) string {
					capturedState = state
					return "https://accounts.google.com/auth?state=" + state
				})

				req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusFound))
				gomega.Expect(rec.Header().Get("Location")).To(gomega.ContainSubstring("accounts.google.com"))

				stateCookie := cookieByName(rec, httpapi.StateCookieName)
				gomega.Expect(stateCookie).NotTo(gomega.BeNil())
				gomega.Expect(stateCookie.Value).To(gomega.Equal(capturedState))
				gomega.Expect(stateCookie.HttpOnly).To(gomega.BeTrue())
				gomega.Expect(capturedState).NotTo(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("Callback", func() {
		var session domain.Session

		ginkgo.BeforeEach(func() {
			session = domain.Session{
				ID:        "session-id",
				UserID:    "user-1",
				Email:     "user@example.com",
				Name:      "User",
				IsAdmin:   false,
				ExpiresAt: time.Now().Add(time.Hour),
			}
		})

		callbackRequest := func(queryState, cookieState, code string) *httptest.ResponseRecorder {
			target := "/auth/callback?" + url.Values{"state": {queryState}, "code": {code}}.Encode()
			req := httptest.NewRequest(http.MethodGet, target, nil)
			if cookieState != "" {
				req.AddCookie(&http.Cookie{Name: httpapi.StateCookieName, Value: cookieState})
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			return rec
		}

		ginkgo.When("the state matches and the email is allowed", func() {
			ginkgo.It("should create the session cookie and redirect to the app", func() {
				service.EXPECT().HandleCallback(gomock.Any(), "code-1").Return(session, nil)

				rec := callbackRequest("state-1", "state-1", "code-1")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusFound))
				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("/ui/"))

				sessionCookie := cookieByName(rec, httpserver.SessionCookieName)
				gomega.Expect(sessionCookie).NotTo(gomega.BeNil())
				gomega.Expect(sessionCookie.Value).To(gomega.Equal("session-id"))
				gomega.Expect(sessionCookie.HttpOnly).To(gomega.BeTrue())
				gomega.Expect(sessionCookie.SameSite).To(gomega.Equal(http.SameSiteLaxMode))
				gomega.Expect(sessionCookie.Path).To(gomega.Equal("/"))
			})
		})

		ginkgo.When("the state does not match", func() {
			ginkgo.It("should reject with 400 and no session cookie", func() {
				rec := callbackRequest("state-evil", "state-1", "code-1")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
				gomega.Expect(cookieByName(rec, httpserver.SessionCookieName)).To(gomega.BeNil())
			})
		})

		ginkgo.When("the state cookie is missing", func() {
			ginkgo.It("should reject with 400", func() {
				rec := callbackRequest("state-1", "", "code-1")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the email is not allowed", func() {
			ginkgo.It("should redirect to access denied without a session cookie", func() {
				service.EXPECT().HandleCallback(gomock.Any(), "code-1").Return(domain.Session{}, usecases.ErrEmailNotAllowed)

				rec := callbackRequest("state-1", "state-1", "code-1")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusFound))
				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("/ui/access-denied"))
				gomega.Expect(cookieByName(rec, httpserver.SessionCookieName)).To(gomega.BeNil())
			})
		})

		ginkgo.When("the email is not verified", func() {
			ginkgo.It("should redirect to access denied", func() {
				service.EXPECT().HandleCallback(gomock.Any(), "code-1").Return(domain.Session{}, usecases.ErrEmailNotVerified)

				rec := callbackRequest("state-1", "state-1", "code-1")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusFound))
				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("/ui/access-denied"))
			})
		})
	})

	ginkgo.Context("Mobile login", func() {
		mobileCallback := func(query url.Values, client, cookieState string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodGet, "/auth/callback?"+query.Encode(), nil)
			if cookieState != "" {
				req.AddCookie(&http.Cookie{Name: httpapi.StateCookieName, Value: cookieState})
			}
			if client != "" {
				req.AddCookie(&http.Cookie{Name: httpapi.ClientCookieName, Value: client})
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			return rec
		}

		ginkgo.When("starting a login with client=mobile", func() {
			ginkgo.It("should remember the client in a cookie and redirect to the provider", func() {
				service.EXPECT().AuthCodeURL(gomock.Any()).Return("https://accounts.google.com/auth")

				req := httptest.NewRequest(http.MethodGet, "/auth/login?client=mobile", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusFound))
				clientCookie := cookieByName(rec, httpapi.ClientCookieName)
				gomega.Expect(clientCookie).NotTo(gomega.BeNil())
				gomega.Expect(clientCookie.Value).To(gomega.Equal("mobile"))
				gomega.Expect(clientCookie.HttpOnly).To(gomega.BeTrue())
			})
		})

		ginkgo.When("mobile login is not configured", func() {
			ginkgo.It("should refuse to start it", func() {
				router = http.NewServeMux()
				httpapi.NewAuthController(service, "").AddRoutes(router)

				req := httptest.NewRequest(http.MethodGet, "/auth/login?client=mobile", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusServiceUnavailable))
			})
		})

		ginkgo.When("the callback completes a mobile login", func() {
			ginkgo.It("should redirect to the app with a handoff code and set no session cookie", func() {
				service.EXPECT().HandleMobileCallback(gomock.Any(), "code-1").Return("handoff-1", nil)

				rec := mobileCallback(url.Values{"state": {"s"}, "code": {"code-1"}}, "mobile", "s")

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusFound))
				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("zensor://auth/callback?code=handoff-1"))
				gomega.Expect(cookieByName(rec, httpserver.SessionCookieName)).To(gomega.BeNil())
			})
		})

		ginkgo.When("the mobile user is not on the allowlist", func() {
			ginkgo.It("should redirect to the app with access_denied", func() {
				service.EXPECT().HandleMobileCallback(gomock.Any(), "code-1").Return("", usecases.ErrEmailNotAllowed)

				rec := mobileCallback(url.Values{"state": {"s"}, "code": {"code-1"}}, "mobile", "s")

				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("zensor://auth/callback?error=access_denied"))
			})
		})

		ginkgo.When("the mobile callback state does not match", func() {
			ginkgo.It("should redirect to the app with invalid_state", func() {
				rec := mobileCallback(url.Values{"state": {"evil"}, "code": {"code-1"}}, "mobile", "s")

				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("zensor://auth/callback?error=invalid_state"))
			})
		})

		ginkgo.When("the person refuses consent on a mobile login", func() {
			ginkgo.It("should redirect to the app with cancelled", func() {
				rec := mobileCallback(url.Values{"state": {"s"}, "error": {"access_denied"}}, "mobile", "s")

				gomega.Expect(rec.Header().Get("Location")).To(gomega.Equal("zensor://auth/callback?error=cancelled"))
			})
		})

		exchange := func(body string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, "/auth/mobile/exchange", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			return rec
		}

		ginkgo.When("exchanging a valid handoff code", func() {
			ginkgo.It("should return the bearer token, its expiry and the user", func() {
				expiresAt := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
				service.EXPECT().ExchangeMobileCode(gomock.Any(), "handoff-1").Return(domain.Session{
					ID:        "session-id",
					UserID:    "user-1",
					Email:     "user@example.com",
					Name:      "User",
					ExpiresAt: expiresAt,
				}, nil)

				rec := exchange(`{"code":"handoff-1"}`)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				var body map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["token"]).To(gomega.Equal("session-id"))
				gomega.Expect(body["expires_at"]).To(gomega.Equal(expiresAt.Format(time.RFC3339)))
				gomega.Expect(body["user"]).To(gomega.HaveKeyWithValue("email", "user@example.com"))
			})
		})

		ginkgo.When("exchanging an unknown or used code", func() {
			ginkgo.It("should return 401", func() {
				service.EXPECT().ExchangeMobileCode(gomock.Any(), "used").Return(domain.Session{}, usecases.ErrAuthHandoffNotFound)

				gomega.Expect(exchange(`{"code":"used"}`).Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("the body is malformed", func() {
			ginkgo.It("should return 400", func() {
				gomega.Expect(exchange(`not json`).Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})
	})

	ginkgo.Context("Bearer session", func() {
		ginkgo.When("reading /v1/me with a bearer session token", func() {
			ginkgo.It("should return the current user", func() {
				service.EXPECT().GetSession(gomock.Any(), "session-id").Return(domain.Session{
					ID: "session-id", UserID: "user-1", Email: "user@example.com",
				}, nil)

				req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
				req.Header.Set("Authorization", "Bearer session-id")
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("logging out with a bearer session token", func() {
			ginkgo.It("should delete that session", func() {
				service.EXPECT().Logout(gomock.Any(), "session-id").Return(nil)

				req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
				req.Header.Set("Authorization", "Bearer session-id")
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNoContent))
			})
		})
	})

	ginkgo.Context("Logout", func() {
		ginkgo.When("logging out with a session cookie", func() {
			ginkgo.It("should delete the session and expire the cookie", func() {
				service.EXPECT().Logout(gomock.Any(), "session-id").Return(nil)

				req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
				req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: "session-id"})
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNoContent))

				sessionCookie := cookieByName(rec, httpserver.SessionCookieName)
				gomega.Expect(sessionCookie).NotTo(gomega.BeNil())
				gomega.Expect(sessionCookie.MaxAge).To(gomega.BeNumerically("<", 0))
			})
		})

		ginkgo.When("logging out without a session cookie", func() {
			ginkgo.It("should still succeed", func() {
				req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNoContent))
			})
		})
	})

	ginkgo.Context("Me", func() {
		ginkgo.When("the session is valid", func() {
			ginkgo.It("should return the current user", func() {
				session := domain.Session{ID: "session-id", UserID: "user-1", Email: "user@example.com", Name: "User", IsAdmin: true}
				service.EXPECT().GetSession(gomock.Any(), "session-id").Return(session, nil)

				req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
				req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: "session-id"})
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["user_id"]).To(gomega.Equal("user-1"))
				gomega.Expect(body["email"]).To(gomega.Equal("user@example.com"))
				gomega.Expect(body["name"]).To(gomega.Equal("User"))
				gomega.Expect(body["is_admin"]).To(gomega.BeTrue())
			})
		})

		ginkgo.When("there is no session cookie", func() {
			ginkgo.It("should return 401", func() {
				req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("the session is invalid", func() {
			ginkgo.It("should return 401", func() {
				service.EXPECT().GetSession(gomock.Any(), "stale").Return(domain.Session{}, usecases.ErrSessionNotFound)

				req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
				req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: "stale"})
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})
	})

	ginkgo.Context("AllowedUsers admin API", func() {
		ginkgo.When("listing allowed users", func() {
			ginkgo.It("should return them as JSON", func() {
				users := []domain.AllowedUser{
					{ID: "u1", Email: "a@example.com", IsAdmin: true},
					{ID: "u2", Email: "b@example.com"},
				}
				service.EXPECT().ListAllowedUsers(gomock.Any()).Return(users, nil)

				req := httptest.NewRequest(http.MethodGet, "/v1/admin/allowed-users", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body []map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body).To(gomega.HaveLen(2))
				gomega.Expect(body[0]["email"]).To(gomega.Equal("a@example.com"))
			})
		})

		ginkgo.When("adding an allowed user", func() {
			ginkgo.It("should return 201 with the created user", func() {
				created := domain.AllowedUser{ID: "u1", Email: "new@example.com", IsAdmin: true}
				service.EXPECT().AddAllowedUser(gomock.Any(), "new@example.com", true).Return(created, nil)

				req := httptest.NewRequest(http.MethodPost, "/v1/admin/allowed-users",
					strings.NewReader(`{"email":"new@example.com","is_admin":true}`))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusCreated))

				var body map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["email"]).To(gomega.Equal("new@example.com"))
			})
		})

		ginkgo.When("adding a duplicated email", func() {
			ginkgo.It("should return 409", func() {
				service.EXPECT().AddAllowedUser(gomock.Any(), "dup@example.com", false).
					Return(domain.AllowedUser{}, usecases.ErrAllowedUserDuplicated)

				req := httptest.NewRequest(http.MethodPost, "/v1/admin/allowed-users",
					strings.NewReader(`{"email":"dup@example.com"}`))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusConflict))
			})
		})

		ginkgo.When("updating an allowed user", func() {
			ginkgo.It("should return the updated user", func() {
				updated := domain.AllowedUser{ID: "u1", Email: "a@example.com", IsAdmin: true}
				service.EXPECT().UpdateAllowedUser(gomock.Any(), domain.ID("u1"), true).Return(updated, nil)

				req := httptest.NewRequest(http.MethodPut, "/v1/admin/allowed-users/u1",
					strings.NewReader(`{"is_admin":true}`))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body map[string]any
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["is_admin"]).To(gomega.BeTrue())
			})
		})

		ginkgo.When("updating a missing user", func() {
			ginkgo.It("should return 404", func() {
				service.EXPECT().UpdateAllowedUser(gomock.Any(), domain.ID("missing"), false).
					Return(domain.AllowedUser{}, usecases.ErrAllowedUserNotFound)

				req := httptest.NewRequest(http.MethodPut, "/v1/admin/allowed-users/missing",
					strings.NewReader(`{"is_admin":false}`))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
			})
		})

		ginkgo.When("removing an allowed user", func() {
			ginkgo.It("should return 204", func() {
				service.EXPECT().RemoveAllowedUser(gomock.Any(), domain.ID("u1")).Return(nil)

				req := httptest.NewRequest(http.MethodDelete, "/v1/admin/allowed-users/u1", nil)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNoContent))
			})
		})
	})
})
