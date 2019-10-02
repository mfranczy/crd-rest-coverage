package analysis

import (
	"path"
	"runtime"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mfranczy/crd-rest-coverage/pkg/stats"
)

var _ = Describe("Swagger analysis", func() {

	_, p, _, ok := runtime.Caller(0)
	if !ok {
		panic("Unable to get the test file path")
	}
	fixturesPath := path.Join(path.Dir(p), "../../fixtures")

	Context("With swagger petstore", func() {

		petStoreSwaggerPath := path.Join(fixturesPath, "petstore.json")
		document, err := loads.JSONSpec(petStoreSwaggerPath)
		if err != nil {
			panic(err)
		}

		Context("Without URI filter", func() {

			It("Should build a full Coverage structure", func() {
				expectedCoverage := &stats.Coverage{
					Percent:            0,
					UniqueHits:         0,
					ExpectedUniqueHits: 0,
					Endpoints: map[string]map[string]*stats.Endpoint{
						"/pets": map[string]*stats.Endpoint{
							"post": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body: map[string]int{
										"pet.kind.profile.size":   0,
										"pet.name":                0,
										"pet.tag":                 0,
										"pet.kind.color":          0,
										"pet.kind.origin.country": 0,
										"pet.kind.origin.region":  0,
									},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 7,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets",
								Method:             "post",
							},
							"get": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body: map[string]int{},
									Query: map[string]int{
										"tags":  0,
										"limit": 0,
									},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 3,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets",
								Method:             "get",
							},
						},
						"/pets/{name}": map[string]*stats.Endpoint{
							"get": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body:  map[string]int{},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 1,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets/{name}",
								Method:             "get",
							},
							"delete": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body:  map[string]int{},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 1,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets/{name}",
								Method:             "delete",
							},
							"patch": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body: map[string]int{
										"pet.kind.origin.region":  0,
										"pet.kind.profile.size":   0,
										"pet.kind.color":          0,
										"pet.name":                0,
										"pet.tag":                 0,
										"pet.kind.origin.country": 0,
									},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 7,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets/{name}",
								Method:             "patch",
							},
						},
					},
				}

				coverage, err := AnalyzeSwagger(document, "")
				Expect(err).NotTo(HaveOccurred(), "coverage structure should be initialized")
				Expect(coverage).To(Equal(expectedCoverage), "coverage values should be equal to expected values")
			})

		})

		Context("With URI filter", func() {

			It("Should build filtered REST API structure", func() {
				expectedCoverage := &stats.Coverage{
					Percent:            0,
					UniqueHits:         0,
					ExpectedUniqueHits: 0,
					Endpoints: map[string]map[string]*stats.Endpoint{
						"/pets/{name}": map[string]*stats.Endpoint{
							"get": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body:  map[string]int{},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 1,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets/{name}",
								Method:             "get",
							},
							"delete": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body:  map[string]int{},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 1,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets/{name}",
								Method:             "delete",
							},
							"patch": &stats.Endpoint{
								ParamsHitsDetails: stats.ParamsHitsDetails{
									Body: map[string]int{
										"pet.kind.origin.region":  0,
										"pet.kind.origin.country": 0,
										"pet.kind.profile.size":   0,
										"pet.kind.color":          0,
										"pet.name":                0,
										"pet.tag":                 0,
									},
									Query: map[string]int{},
								},
								UniqueHits:         0,
								ExpectedUniqueHits: 7,
								Percent:            0,
								MethodCalled:       false,
								Path:               "/pets/{name}",
								Method:             "patch",
							},
						},
					},
				}

				coverage, err := AnalyzeSwagger(document, "/pets/{name}")
				Expect(err).NotTo(HaveOccurred(), "coverage structure should be initialized")
				Expect(coverage).To(Equal(expectedCoverage), "coverage values should be equal to expected values")

			})
		})

		It("Should add swagger params", func() {
			By("Resolving referenced definitions")
			expectedEndpoint := stats.Endpoint{
				ParamsHitsDetails: stats.ParamsHitsDetails{
					Body: map[string]int{
						"pet.name":                0,
						"pet.tag":                 0,
						"pet.kind.color":          0,
						"pet.kind.origin.country": 0,
						"pet.kind.origin.region":  0,
						"pet.kind.profile.size":   0,
					},
					Query: map[string]int{},
				},
				UniqueHits:         0,
				ExpectedUniqueHits: 7,
				Percent:            0,
				MethodCalled:       false,
				Path:               "/pets",
				Method:             "post",
			}

			initializedEndpoint := stats.Endpoint{
				ParamsHitsDetails: stats.ParamsHitsDetails{
					Body:  map[string]int{},
					Query: map[string]int{},
				},
				UniqueHits:         0,
				ExpectedUniqueHits: 1,
				Percent:            0,
				MethodCalled:       false,
				Path:               "/pets",
				Method:             "post",
			}

			addSwaggerParams(&initializedEndpoint, document.Analyzer.ParamsFor("POST", "/pets"), document.Spec().Definitions)
			Expect(initializedEndpoint).To(Equal(expectedEndpoint), "endpoint values should be equal to expected values")

			By("Not resolving referenced definitions")
			expectedEndpoint = stats.Endpoint{
				ParamsHitsDetails: stats.ParamsHitsDetails{
					Body: map[string]int{},
					Query: map[string]int{
						"tags":  0,
						"limit": 0,
					},
				},
				UniqueHits:         0,
				ExpectedUniqueHits: 3,
				Percent:            0,
				MethodCalled:       false,
				Path:               "/pets",
				Method:             "get",
			}

			initializedEndpoint = stats.Endpoint{
				ParamsHitsDetails: stats.ParamsHitsDetails{
					Body:  map[string]int{},
					Query: map[string]int{},
				},
				UniqueHits:         0,
				ExpectedUniqueHits: 1,
				Percent:            0,
				MethodCalled:       false,
				Path:               "/pets",
				Method:             "get",
			}

			addSwaggerParams(&initializedEndpoint, document.Analyzer.ParamsFor("GET", "/pets"), document.Spec().Definitions)
			Expect(initializedEndpoint).To(Equal(expectedEndpoint), "endpoint values should be equal to expected values")
		})

		It("Should count params from referenced models", func() {
			document, err := loads.JSONSpec(petStoreSwaggerPath)
			Expect(err).NotTo(HaveOccurred(), "swagger json file should be open")

			params := document.Analyzer.ParamsFor("POST", "/pets")

			paths := make(map[string]int)
			pCnt := extractDefParams(params["body#Pet"].Schema, document.Spec().Definitions, "test", paths)
			Expect(pCnt).To(Equal(6), "reference params number should be equal to expected value")

			paths = make(map[string]int)
			pCnt = extractDefParams(params["body#Pet"].Schema, spec.Definitions{}, "", paths)
			Expect(pCnt).To(Equal(0), "reference params number for non-existence definition should not be increased")
		})
	})
})
