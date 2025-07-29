package avro_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAvro(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Avro Suite")
}
