package report

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "math"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"

	"github.com/mfranczy/crd-rest-coverage/pkg/stats"
)

var _ = Describe("REST API coverage report", func() {

	Context("With pets audit log", func() {

		table.DescribeTable("Should generate a report", func(testFile string, filter string) {
			var expectedCoverage stats.Coverage

			content, err := ioutil.ReadFile(testFile)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(content, &expectedCoverage)
			Expect(err).NotTo(HaveOccurred())

			coverage, err := Generate(auditLogPath, petStoreSwaggerPath, filter, false)
			Expect(err).NotTo(HaveOccurred(), "coverage structure should be initialized")

			Expect(coverage.Percent).To(Equal(expectedCoverage.Percent), "percent should be equal")
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
			table.Entry("Without URI filter", "fixtures/test_output.json", ""),
			table.Entry("With URI filter", "fixtures/test_filter_output.json", "/pets/{name}"),
		)

		table.DescribeTable("Should return correct swagger path based on audit URL", func(URI string, objRef *auditv1.ObjectReference, swaggerPath string) {
			path := getSwaggerPath(URI, objRef, false)
			Expect(path).To(Equal(swaggerPath))
		},
			table.Entry(
				"With an empty namespace",
				"/pets/bite",
				&auditv1.ObjectReference{Name: "bite"},
				"/pets/{name}",
			),
			table.Entry(
				"With defined namespace",
				"/pets/namespace/default/bite",
				&auditv1.ObjectReference{Name: "bite"},
				"/pets/namespace/default/{name}",
			),
			table.Entry(
				"With empty object reference",
				"/pets",
				&auditv1.ObjectReference{},
				"/pets",
			),
			table.Entry(
				"With VMI list request",
				"/apis/kubevirt.io/v1alpha3/namespaces/kubevirt-test-default/virtualmachineinstances",
				&auditv1.ObjectReference{
					Resource:  "virtualmachineinstances",
					Namespace: "kubevirt-test-default",
				},
				"/apis/kubevirt.io/v1alpha3/namespaces/{namespace}/virtualmachineinstances",
			),
			table.Entry(
				"With VMI get request",
				"/apis/kubevirt.io/v1alpha3/namespaces/kubevirt/virtualmachineinstances/testvmi22gsnklt2flhqflcnp8jpmq6fkj72szv8h9sn26z2hdhkm6l",
				&auditv1.ObjectReference{
					Resource:  "virtualmachineinstances",
					Namespace: "kubevirt",
					Name:      "testvmi22gsnklt2flhqflcnp8jpmq6fkj72szv8h9sn26z2hdhkm6l",
				},
				"/apis/kubevirt.io/v1alpha3/namespaces/{namespace}/virtualmachineinstances/{name}",
			),
		)

		table.DescribeTable("Should translate k8s verb to HTTP method", func(verb string, httpMethod string) {
			Expect(getHTTPMethod(verb)).To(Equal(httpMethod), fmt.Sprintf("verb %s should be translated to %s", verb, httpMethod))
		},
			table.Entry("With get verb", "get", "get"),
			table.Entry("With list verb", "list", "get"),
			table.Entry("With watch verb", "watch", "get"),
			table.Entry("With watchList verb", "watchList", "get"),
			table.Entry("With create verb", "create", "post"),
			table.Entry("With delete verb", "delete", "delete"),
			table.Entry("With deletecollection verb", "deletecollection", "delete"),
			table.Entry("With update verb", "update", "put"),
			table.Entry("With patch verb", "patch", "patch"),
			table.Entry("With invalid verb", "invalid", ""),
			table.Entry("With empty verb", "", ""),
		)
	})
})
