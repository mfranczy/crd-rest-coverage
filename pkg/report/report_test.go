package report

import (
	"fmt"
	"math"
	"net/url"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"

	"github.com/mfranczy/crd-rest-coverage/pkg/stats"
)

var _ = Describe("REST API coverage report", func() {

	Context("With pets audit log", func() {

		It("Should generate a coverage report", func() {
			By("Generating a coverage report")
			coverage, err := Generate(auditLogPath, petStoreSwaggerPath, "")
			Expect(err).NotTo(HaveOccurred(), "report should be generated")

			By("Checking generated report")
			Expect(math.Round(coverage.Percent*100)/100).To(Equal(47.37), "coverage percent should be equal to expected output")
		})

		table.DescribeTable("Should return correct swagger path based on audit URL", func(URI string, objRef *auditv1.ObjectReference, swaggerPath string) {
			path := getSwaggerPath(URI, objRef)
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

		Context("With matching query params", func() {

			It("Should match query params", func() {
				endpoint := stats.Endpoint{
					ParamsHitsDetails: stats.ParamsHitsDetails{
						Query: map[string]int{
							"limit": 0,
							"tags":  0,
						},
					},
					UniqueHits: 0,
				}
				vals := url.Values{
					"limit": []string{"100"},
					"tags":  []string{"color"},
				}
				By("Matching the first time it should increase UniqueHits")
				matchQueryParams(vals, &endpoint)
				Expect(endpoint.UniqueHits).To(Equal(2), "unique params hit should be equal to provided params number")
				Expect(endpoint.Query["limit"]).To(Equal(1), "'limit' parameter should increase its occurrence number")
				Expect(endpoint.Query["tags"]).To(Equal(1), "'tags' parameter should increase its occurrence number")

				By("Matching the second time it should not increase UniqueHits")
				matchQueryParams(vals, &endpoint)
				Expect(endpoint.UniqueHits).To(Equal(2), "unique params hit should stay the same")
				Expect(endpoint.Query["limit"]).To(Equal(2), "'limit' parameter should increase its occurrence number")
				Expect(endpoint.Query["tags"]).To(Equal(2), "'tags' parameter should increase its occurrence number")
			})

			It("Should not increase hits for undefined query params", func() {
				endpoint := stats.Endpoint{
					ParamsHitsDetails: stats.ParamsHitsDetails{
						Query: map[string]int{
							"limit": 0,
						},
					},
					UniqueHits: 0,
				}
				vals := url.Values{"unknown": []string{"test"}}
				matchQueryParams(vals, &endpoint)
				Expect(endpoint.UniqueHits).To(Equal(0), "uniqueHits should not be increased for unknown parameters")
				Expect(endpoint.ParamsHitsDetails.Query["limit"]).To(Equal(0), "hits counter should not be increased for unknown parameters")
				_, exists := endpoint.ParamsHitsDetails.Query["unknown"]
				Expect(exists).To(BeFalse(), "'unkown' parameter should not exist")
			})

		})

		Context("With matching and extracting body params", func() {

			It("Should match body params", func() {
				reqObject := runtime.Unknown{
					Raw: []byte(
						`{
							"pet": {
								"name": "bite",
								"kind": {
									"color": "red",
									"origin": {
										"country": "unknown",
										"region": "west"
									},
									"profile": {
										"size": "small"
									}
								}
							}
						}`,
					),
				}

				endpoint := stats.Endpoint{
					UniqueHits: 0,
					ParamsHitsDetails: stats.ParamsHitsDetails{
						Body: map[string]int{
							"pet.name":                0,
							"pet.tag":                 0,
							"pet.kind.color":          0,
							"pet.kind.origin.country": 0,
							"pet.kind.origin.region":  0,
							"pet.kind.profile.size":   0,
						},
					},
				}

				By("Matching the first time it should increase ParamsHit")
				err := matchBodyParams(&reqObject, &endpoint)
				Expect(err).NotTo(HaveOccurred())
				Expect(endpoint.UniqueHits).To(Equal(5), "uniqueHits should be equal to provided params number")
				for path, hits := range endpoint.ParamsHitsDetails.Body {
					if path == "pet.tag" {
						Expect(hits).To(Equal(0), fmt.Sprintf("'%s' parameter should not increase hits number", path))
					} else {
						Expect(hits).To(Equal(1), fmt.Sprintf("'%s' parameter should increase hits number", path))
					}
				}

				By("Matching the second time it should not increase ParamsHit")
				matchBodyParams(&reqObject, &endpoint)
				Expect(endpoint.UniqueHits).To(Equal(5), "uniqueHits should stay the same")
				for path, hits := range endpoint.ParamsHitsDetails.Body {
					if path == "pet.tag" {
						Expect(hits).To(Equal(0), fmt.Sprintf("'%s' parameter should not increase hits number", path))
					} else {
						Expect(hits).To(Equal(2), fmt.Sprintf("'%s' parameter should increase hits number", path))
					}
				}
			})

			It("Should return an error for invalid body params", func() {
				reqObject := runtime.Unknown{
					Raw: []byte(
						`{
							"pet: {
								name: bite,
							}
						}`,
					),
				}
				endpoint := stats.Endpoint{
					ParamsHitsDetails: stats.ParamsHitsDetails{
						Body: map[string]int{},
					},
				}

				By("Passing invalid JSON object")
				err := matchBodyParams(&reqObject, &endpoint)
				Expect(err).To(HaveOccurred(), "matchBodyParams should return an error for invalid JSON object")
			})
		})

		Context("With coverage calculation", func() {

			It("Should calculate REST coverage", func() {
				coverage := &stats.Coverage{
					UniqueHits:         0,
					ExpectedUniqueHits: 0,
					Percent:            0,
					Endpoints: map[string]map[string]*stats.Endpoint{
						"/pets": map[string]*stats.Endpoint{
							"post": &stats.Endpoint{
								UniqueHits:         1,
								ExpectedUniqueHits: 6,
								MethodCalled:       true,
								Percent:            0,
							},
						},
						"/pets/{name}": map[string]*stats.Endpoint{
							"patch": &stats.Endpoint{
								UniqueHits:         10,
								ExpectedUniqueHits: 6,
								MethodCalled:       true,
								Percent:            0,
							},
						},
					},
				}
				calculateCoverage(coverage)

				By("Checking total coverage")
				Expect(coverage.ExpectedUniqueHits).To(Equal(12), "total expectedUniqueHits should match expected value")
				Expect(coverage.UniqueHits).To(Equal(8), "total uniqueHits should match expected value")
				Expect(math.Round(coverage.Percent*100)/100).To(Equal(66.67), "total percent should match expected value")

				By("Checking endpoints coverage")
				Expect(math.Round(coverage.Endpoints["/pets"]["post"].Percent*100)/100).To(Equal(33.33),
					"pets:post percent should match expected value")
				Expect(math.Round(coverage.Endpoints["/pets/{name}"]["patch"].Percent*100)/100).To(Equal(100.0),
					"pets/{name}:patch percent should match 100%, UniqueHits > ExpectedUniqueHits")

			})

		})
	})

})
