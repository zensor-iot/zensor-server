package dto_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDto(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dto Suite")
}
