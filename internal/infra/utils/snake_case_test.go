package utils_test

import (
	"zensor-server/internal/infra/utils"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ToSnakeCase", func() {
	ginkgo.Context("with camelCase strings", func() {
		ginkgo.It("should convert simple camelCase to snake_case", func() {
			result := utils.ToSnakeCase("camelCase")
			gomega.Expect(result).To(gomega.Equal("camel_case"))
		})

		ginkgo.It("should convert PascalCase to snake_case", func() {
			result := utils.ToSnakeCase("PascalCase")
			gomega.Expect(result).To(gomega.Equal("pascal_case"))
		})

		ginkgo.It("should convert multiple words", func() {
			result := utils.ToSnakeCase("thisIsALongCamelCaseString")
			gomega.Expect(result).To(gomega.Equal("this_is_a_long_camel_case_string"))
		})

		ginkgo.It("should handle single character", func() {
			result := utils.ToSnakeCase("A")
			gomega.Expect(result).To(gomega.Equal("a"))
		})

		ginkgo.It("should handle two characters", func() {
			result := utils.ToSnakeCase("AB")
			gomega.Expect(result).To(gomega.Equal("a_b"))
		})
	})

	ginkgo.Context("with already snake_case strings", func() {
		ginkgo.It("should leave snake_case unchanged", func() {
			result := utils.ToSnakeCase("snake_case")
			gomega.Expect(result).To(gomega.Equal("snake_case"))
		})

		ginkgo.It("should leave multiple underscores unchanged", func() {
			result := utils.ToSnakeCase("snake_case_with_multiple_words")
			gomega.Expect(result).To(gomega.Equal("snake_case_with_multiple_words"))
		})
	})

	ginkgo.Context("with mixed case strings", func() {
		ginkgo.It("should convert mixed camelCase and snake_case", func() {
			result := utils.ToSnakeCase("camelCase_with_snake")
			gomega.Expect(result).To(gomega.Equal("camel_case_with_snake"))
		})

		ginkgo.It("should handle numbers in camelCase", func() {
			result := utils.ToSnakeCase("version2Update")
			gomega.Expect(result).To(gomega.Equal("version2_update"))
		})
	})

	ginkgo.Context("with special cases", func() {
		ginkgo.It("should handle empty string", func() {
			result := utils.ToSnakeCase("")
			gomega.Expect(result).To(gomega.Equal(""))
		})

		ginkgo.It("should handle all lowercase", func() {
			result := utils.ToSnakeCase("lowercase")
			gomega.Expect(result).To(gomega.Equal("lowercase"))
		})

		ginkgo.It("should handle all uppercase", func() {
			result := utils.ToSnakeCase("UPPERCASE")
			gomega.Expect(result).To(gomega.Equal("u_p_p_e_r_c_a_s_e"))
		})

		ginkgo.It("should handle strings with numbers", func() {
			result := utils.ToSnakeCase("test123String")
			gomega.Expect(result).To(gomega.Equal("test123_string"))
		})

		ginkgo.It("should handle strings starting with numbers", func() {
			result := utils.ToSnakeCase("123testString")
			gomega.Expect(result).To(gomega.Equal("123test_string"))
		})
	})

	ginkgo.Context("with edge cases", func() {
		ginkgo.It("should handle single uppercase letter", func() {
			result := utils.ToSnakeCase("A")
			gomega.Expect(result).To(gomega.Equal("a"))
		})

		ginkgo.It("should handle single lowercase letter", func() {
			result := utils.ToSnakeCase("a")
			gomega.Expect(result).To(gomega.Equal("a"))
		})

		ginkgo.It("should handle strings with special characters", func() {
			result := utils.ToSnakeCase("test-string")
			gomega.Expect(result).To(gomega.Equal("test-string"))
		})

		ginkgo.It("should handle strings with spaces", func() {
			result := utils.ToSnakeCase("test string")
			gomega.Expect(result).To(gomega.Equal("test string"))
		})
	})

	ginkgo.Context("with real-world examples", func() {
		ginkgo.It("should convert HTTP method names", func() {
			result := utils.ToSnakeCase("GetUserById")
			gomega.Expect(result).To(gomega.Equal("get_user_by_id"))
		})

		ginkgo.It("should convert database field names", func() {
			result := utils.ToSnakeCase("CreatedAt")
			gomega.Expect(result).To(gomega.Equal("created_at"))
		})

		ginkgo.It("should convert API endpoint names", func() {
			result := utils.ToSnakeCase("CreateUserProfile")
			gomega.Expect(result).To(gomega.Equal("create_user_profile"))
		})

		ginkgo.It("should convert configuration keys", func() {
			result := utils.ToSnakeCase("MaxRetryAttempts")
			gomega.Expect(result).To(gomega.Equal("max_retry_attempts"))
		})
	})
})
