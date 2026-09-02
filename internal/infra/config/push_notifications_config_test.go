package config_test

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// shippedConfigPath is the real server.yaml, read from the package directory.
const shippedConfigPath = "../../../config/server.yaml"

type pushNotificationEntry struct {
	Name             string `yaml:"name"`
	Title            string `yaml:"title"`
	TitleTemplate    string `yaml:"title_template"`
	Body             string `yaml:"body"`
	BodyTemplate     string `yaml:"body_template"`
	DeepLink         string `yaml:"deeplink"`
	DeepLinkTemplate string `yaml:"deeplink_template"`
}

type shippedConfig struct {
	PushNotifications []pushNotificationEntry `yaml:"push_notifications"`
	Medicines         struct {
		PushNotifications []pushNotificationEntry `yaml:"push_notifications"`
	} `yaml:"medicines"`
}

// These specs assert the content of the shipped configuration rather than the
// loader. Both invariants below have already shipped broken once: a deep link
// without the SPA prefix sends the user to a 404, and neither failure is visible
// server side, so nothing else catches them.
var _ = ginkgo.Describe("shipped push notification configuration", func() {
	var entries []pushNotificationEntry

	ginkgo.BeforeEach(func() {
		raw, err := os.ReadFile(shippedConfigPath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		var parsed shippedConfig
		gomega.Expect(yaml.Unmarshal(raw, &parsed)).To(gomega.Succeed())

		entries = append(append([]pushNotificationEntry{}, parsed.PushNotifications...),
			parsed.Medicines.PushNotifications...)
		gomega.Expect(entries).NotTo(gomega.BeEmpty())
	})

	ginkgo.Context("deep links", func() {
		ginkgo.When("a notification carries the user into the SPA", func() {
			ginkgo.It("should start at the /ui mount point, which the service worker opens verbatim", func() {
				for _, entry := range entries {
					for field, value := range map[string]string{
						"deeplink":          entry.DeepLink,
						"deeplink_template": entry.DeepLinkTemplate,
					} {
						if value == "" {
							continue
						}
						gomega.Expect(value).To(gomega.HavePrefix("/ui/"),
							"%s of %q must start with /ui/", field, entry.Name)
					}
				}
			})
		})
	})

	ginkgo.Context("templates", func() {
		ginkgo.When("a template interpolates a value", func() {
			ginkgo.It("should use {{key}}, the only form the worker supports", func() {
				for _, entry := range entries {
					for field, value := range map[string]string{
						"title_template":    entry.TitleTemplate,
						"body_template":     entry.BodyTemplate,
						"deeplink_template": entry.DeepLinkTemplate,
					} {
						gomega.Expect(value).NotTo(gomega.ContainSubstring("%s"),
							"%s of %q uses an unsupported %%s placeholder", field, entry.Name)
					}
				}
			})
		})

		ginkgo.When("a template cannot be resolved", func() {
			ginkgo.It("should have a static value to fall back to", func() {
				for _, entry := range entries {
					if entry.TitleTemplate != "" {
						gomega.Expect(entry.Title).NotTo(gomega.BeEmpty(),
							"%q needs a static title", entry.Name)
					}
					if entry.BodyTemplate != "" {
						gomega.Expect(entry.Body).NotTo(gomega.BeEmpty(),
							"%q needs a static body", entry.Name)
					}
					if entry.DeepLinkTemplate != "" {
						gomega.Expect(entry.DeepLink).NotTo(gomega.BeEmpty(),
							"%q needs a static deeplink", entry.Name)
					}
				}
			})
		})

		ginkgo.When("a template names a placeholder", func() {
			ginkgo.It("should close every one it opens", func() {
				for _, entry := range entries {
					for _, value := range []string{entry.TitleTemplate, entry.BodyTemplate, entry.DeepLinkTemplate} {
						gomega.Expect(strings.Count(value, "{{")).To(gomega.Equal(strings.Count(value, "}}")),
							"unbalanced placeholder in %q: %s", entry.Name, value)
					}
				}
			})
		})
	})
})
