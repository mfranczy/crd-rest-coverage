package analysis_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAnalysis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Analysis Suite")
}
