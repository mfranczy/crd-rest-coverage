package report

import (
	"path"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var petStoreSwaggerPath string
var auditLogPath string

func TestCoverage(t *testing.T) {
	_, p, _, ok := runtime.Caller(0)
	if !ok {
		panic("Not possible to get test file path")
	}
	fixturesPath := path.Join(path.Dir(p), "../../fixtures")
	petStoreSwaggerPath = path.Join(fixturesPath, "test_petstore.json")
	auditLogPath = path.Join(fixturesPath, "test_audit.log")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Coverage Suite")
}
