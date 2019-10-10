package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-openapi/loads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/mfranczy/crd-rest-coverage/pkg/stats"
)

var _ = Describe("Swagger analysis", func() {

	Context("With swagger petstore", func() {

		DescribeTable("Should build a coverage structure", func(testFile string, filter string) {
			var expectedCoverage stats.Coverage

			content, err := ioutil.ReadFile(testFile)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(content, &expectedCoverage)
			Expect(err).NotTo(HaveOccurred())
			document, err := loads.JSONSpec(petStoreSwaggerPath)
			Expect(err).NotTo(HaveOccurred())

			coverage, err := AnalyzeSwagger(document, filter)
			Expect(err).NotTo(HaveOccurred(), "coverage structure should be initialized")

			Expect(coverage.Percent).To(Equal(expectedCoverage.Percent), "percent should be equal to 0")
			Expect(coverage.UniqueHits).To(Equal(expectedCoverage.UniqueHits), "uniqueHits should be equal")
			Expect(coverage.ExpectedUniqueHits).To(Equal(expectedCoverage.ExpectedUniqueHits), "expectedUniqueHits should be equal")
			Expect(len(coverage.Endpoints)).To(Equal(len(expectedCoverage.Endpoints)), "endpoints len should be equal")

			for path, methods := range coverage.Endpoints {

				Expect(expectedCoverage.Endpoints).To(HaveKey(path), fmt.Sprintf("path %s should exist in expectedCoverage structure", path))

				for method, endpoint := range methods {
					expectedEndpoint := expectedCoverage.Endpoints[path][method]

					By(fmt.Sprintf("Checking %s:%s endpoint", path, method))
					Expect(expectedCoverage.Endpoints[path]).To(HaveKey(method),
						fmt.Sprintf("method %s should exist in expectedCoverage structure", path))
					Expect(endpoint.MethodCalled).To(Equal(expectedEndpoint.MethodCalled), "should have the same methodCalled values")
					Expect(endpoint.UniqueHits).To(Equal(expectedEndpoint.UniqueHits), "should have the same uniqueHits values")
					Expect(endpoint.ExpectedUniqueHits).To(Equal(expectedEndpoint.ExpectedUniqueHits), "should have the same expectedUniqueHits values")

					By(fmt.Sprintf("Checking %s:%s endpoint's body params", path, method))
					Expect(endpoint.Params.Body.Height).To(Equal(expectedEndpoint.Params.Body.Height), "should have the same height values")
					Expect(endpoint.Params.Body.Size).To(Equal(expectedEndpoint.Params.Body.Size), "should have the same size values")
					Expect(endpoint.Params.Body.UniqueHits).To(Equal(expectedEndpoint.Params.Body.UniqueHits), "should have the same uniqueHits values")
					Expect(endpoint.Params.Body.ExpectedUniqueHits).To(Equal(expectedEndpoint.Params.Body.ExpectedUniqueHits), "should have the same expectedUniqueHits values")

					By(fmt.Sprintf("Checking %s:%s endpoint's query params", path, method))
					Expect(endpoint.Params.Query.Height).To(Equal(expectedEndpoint.Params.Query.Height), "should have the same height values")
					Expect(endpoint.Params.Query.Size).To(Equal(expectedEndpoint.Params.Query.Size), "should have the same size values")
					Expect(endpoint.Params.Query.UniqueHits).To(Equal(expectedEndpoint.Params.Query.UniqueHits), "should have the same uniqueHits values")
					Expect(endpoint.Params.Query.ExpectedUniqueHits).To(Equal(expectedEndpoint.Params.Query.ExpectedUniqueHits), "should have the same expectedUniqueHits values")
				}
			}
		},
			Entry("Without URI filter", "fixtures/test_output.json", ""),
			Entry("With URI filter", "fixtures/test_filter_output.json", "/pets/{name}"),
		)

	})
})
