package httpserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHttpserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Httpserver Suite")
}
