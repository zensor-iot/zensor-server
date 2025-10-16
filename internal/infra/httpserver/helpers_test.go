package httpserver

import (
	"net/http"
	"net/url"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Helpers", func() {
	ginkgo.Context("ExtractPaginationParams", func() {
		var (
			req      *http.Request
			query    string
			expected PaginationParams
		)

		ginkgo.When("extracting pagination parameters", func() {
			ginkgo.It("should return default values when no query params", func() {
				query = ""
				expected = PaginationParams{Page: 1, Limit: 10}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should return valid page and limit", func() {
				query = "page=2&limit=20"
				expected = PaginationParams{Page: 2, Limit: 20}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should default to 1 when page is invalid", func() {
				query = "page=0&limit=10"
				expected = PaginationParams{Page: 1, Limit: 10}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should default to 10 when limit is invalid", func() {
				query = "page=1&limit=0"
				expected = PaginationParams{Page: 1, Limit: 10}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should cap to 100 when limit is too high", func() {
				query = "page=1&limit=150"
				expected = PaginationParams{Page: 1, Limit: 100}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle only page parameter", func() {
				query = "page=3"
				expected = PaginationParams{Page: 3, Limit: 10}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle only limit parameter", func() {
				query = "limit=25"
				expected = PaginationParams{Page: 1, Limit: 25}

				req = &http.Request{
					URL: &url.URL{},
				}
				if query != "" {
					req.URL.RawQuery = query
				}

				result := ExtractPaginationParams(req)
				gomega.Expect(result).To(gomega.Equal(expected))
			})
		})
	})

	ginkgo.Context("DefaultPaginationParams", func() {
		ginkgo.When("getting default pagination parameters", func() {
			ginkgo.It("should return default values", func() {
				params := DefaultPaginationParams()
				expected := PaginationParams{Page: 1, Limit: 10}

				gomega.Expect(params).To(gomega.Equal(expected))
			})
		})
	})
})
