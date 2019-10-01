package report

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/loads"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"

	"github.com/mfranczy/crd-rest-coverage/pkg/analysis"
	"github.com/mfranczy/crd-rest-coverage/pkg/stats"
)

// getSwaggerPath translates request path to generic swagger path, as an example,
// /apis/kubevirt.io/v1alpha3/namespaces/kubevirt-test-default/virtualmachineinstances/vm-name will be translated to
// /apis/kubevirt.io/v1alpha3/namespaces/{namespace}/virtualmachineinstances/{name}
func getSwaggerPath(path string, objectRef *auditv1.ObjectReference) string {
	if namespace := objectRef.Namespace; namespace != "" {
		path = strings.Replace(path, "namespaces/"+namespace, "namespaces/{namespace}", 1)
	}
	if name := objectRef.Name; name != "" {
		path = strings.Replace(path, objectRef.Resource+"/"+name, objectRef.Resource+"/{name}", 1)
	}
	return path
}

// getHTTPMethod translates k8s verbs from audit log into HTTP methods
// NOTE: audit log does not provide information about HTTP methods
func getHTTPMethod(verb string) string {
	switch verb {
	case "get", "list", "watch", "watchList":
		return "get"
	case "create":
		return "post"
	case "delete", "deletecollection":
		return "delete"
	case "update":
		return "put"
	case "patch":
		return "patch"
	default:
		return ""
	}
}

// matchQueryParams matches query params from request log to stats structure built based on swagger definition,
// as an example, by having stats.Endpoint{method: GET, Query: [param1: 0], ParamsHit: 0}
// and request GET /?param1=test1&param2=test2
// - it increments the parameter occurrence number, requestStats{method: GET, Query: [param1: 1], ParamsHit: 1}
// - it logs an error for not documented/invalid parameters, in this example parameter named "test2"
func matchQueryParams(values url.Values, endpoint *stats.Endpoint) {
	for k := range values {
		if hits, ok := endpoint.ParamsHitDetails.Query[k]; ok {
			if hits < 1 {
				// get only unique hits
				endpoint.UniqueHit++
			}
			endpoint.ParamsHitDetails.Query[k]++
		} else {
			glog.Errorf("Invalid query param: '%s' for '%s %s'", k, endpoint.Method, endpoint.Path)
		}
	}
}

// matchBodyParams matches body params from request log to stats structure built based on swagger definition,
// as an example, by having stats.Endpoint{method: POST, Body: {}, ParamsHit: 0} and request
// POST / {"param1": {"param2": "test"}}
// - it builds the parameter path and increase its occurrence number, stats.Endpoint{method: POST, Body: {param1.param2: 1}, ParamsHit: 1}
// - it returns an error if request provides body params but it is not defined in swagger definition
func matchBodyParams(requestObject *runtime.Unknown, endpoint *stats.Endpoint) error {
	if requestObject != nil && endpoint.ParamsHitDetails.Body != nil {
		var req interface{}
		err := json.Unmarshal(requestObject.Raw, &req)
		if err != nil {
			return err
		}
		switch r := req.(type) {
		case []interface{}:
			for _, v := range r {
				err = extractBodyParams(v, "", endpoint, 0)
				if err != nil {
					return fmt.Errorf("Invalid requestObject '%s' for '%s %s'", err, endpoint.Method, endpoint.Path)
				}
			}
		default:
			err = extractBodyParams(r, "", endpoint, 0)
			if err != nil {
				return fmt.Errorf("Invalid requestObject '%s' for '%s %s'", err, endpoint.Method, endpoint.Path)
			}
		}

	} else if requestObject != nil {
		glog.Warningf("Request '%s %s' should not contain body params", endpoint.Method, endpoint.Path)
	}

	return nil
}

// extractBodyParams builds a body parameter path from JSON structure and increase its occurence number, as an example,
// {param1: {param2: {param3a: value1, param3b: value2}}} will be extracted into paths:
// - param1.param2.param3a: 1
// - param1.param2.param3b: 1
func extractBodyParams(params interface{}, path string, endpoint *stats.Endpoint, level int) error {
	p, ok := params.(map[string]interface{})
	if !ok && level == 0 {
		return fmt.Errorf("%v", p)
	} else if !ok {
		return nil
	}
	level++

	pathCopy := path
	for k, v := range p {
		if level == 1 {
			path = k
		} else {
			path += "." + k
		}

		switch obj := v.(type) {
		case map[string]interface{}:
			extractBodyParams(obj, path, endpoint, level)
		case []interface{}:
			for _, v := range obj {
				extractBodyParams(v, path, endpoint, level)
			}
		default:
			if i, ok := endpoint.ParamsHitDetails.Body[path]; ok {
				if i < 1 {
					endpoint.UniqueHit++
				}
				endpoint.ParamsHitDetails.Body[path]++
			} else {
				glog.Errorf("Invalid body param: '%s' for '%s %s'", k, endpoint.Method, endpoint.Path)
			}
		}
		path = pathCopy
	}
	return nil
}

// calculateCoverage provides a total REST API and PATH:METHOD coverage number
func calculateCoverage(coverage *stats.Coverage) {
	for _, es := range coverage.Endpoints {
		for _, e := range es {
			if e.MethodCalled {
				e.UniqueHit++
			}
			// sometimes hit number is bigger than params number
			// for instance it might be caused by missing models definition
			// users have to make sure that their definitions are complete
			if e.UniqueHit > e.Sum {
				e.UniqueHit = e.Sum
			}

			if e.Sum > 0 {
				coverage.Sum += e.Sum
				coverage.UniqueHit += e.UniqueHit
				e.Percent = float64(e.UniqueHit) * 100 / float64(e.Sum)
			} else {
				e.Percent = 0
			}
		}
	}

	if coverage.Sum > 0 {
		coverage.Percent = float64(coverage.UniqueHit) * 100 / float64(coverage.Sum)
	} else {
		coverage.Percent = 0
	}
}

// Print shows a generated report, if detailed it will show coverage for each endpoint
func Print(coverage *stats.Coverage, detailed bool) error {
	fmt.Printf("\nREST API coverage report:\n\n")
	if detailed {
		for p, es := range coverage.Endpoints {
			fmt.Println(p)
			for _, e := range es {
				fmt.Printf("%s:%.2f%%\t", strings.ToUpper(e.Method), e.Percent)
			}
			fmt.Println("\n")
		}
	}
	fmt.Printf("\nTotal coverage: %.2f%%\n\n", coverage.Percent)
	return nil
}

// Dump saves a generated report into a file in JSON format
func Dump(path string, coverage *stats.Coverage) error {
	jsonCov, err := json.Marshal(coverage)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, jsonCov, 0644)
}

// Generate provides a full REST API coverage report based on k8s audit log and swagger definition,
// by passing param "filter" you can limit the report to specific resources, as an example,
// "/apis/kubevirt.io/v1alpha3/" limits to kubevirt v1alpha3; "" no limit
func Generate(auditLogsPath string, swaggerPath string, filter string) (*stats.Coverage, error) {
	start := time.Now()
	defer glog.Infof("REST API coverage execution time: %s", time.Since(start))

	auditLogs, err := os.Open(auditLogsPath)
	if err != nil {
		return nil, err
	}

	sDocument, err := loads.JSONSpec(swaggerPath)
	if err != nil {
		return nil, err
	}

	coverage, err := analysis.AnalyzeSwagger(sDocument, filter)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewReader(auditLogs)
	for {
		var event auditv1.Event
		b, err := scanner.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		err = json.Unmarshal(b, &event)
		if err != nil {
			return nil, err
		}

		uri, err := url.Parse(event.RequestURI)
		if err != nil {
			return nil, err
		}

		path := getSwaggerPath(uri.Path, event.ObjectRef)
		if _, ok := coverage.Endpoints[path]; !ok {
			glog.Errorf("Path '%s' not found in swagger", path)
			continue
		}

		method := getHTTPMethod(event.Verb)
		if method == "" {
			glog.Errorf("Method '%s' not found for '%s' path", method, path)
			continue
		}

		if _, ok := coverage.Endpoints[path][method]; !ok {
			glog.Errorf("Method '%s' not found for '%s' path", method, path)
			continue
		}

		coverage.Endpoints[path][method].MethodCalled = true
		matchQueryParams(uri.Query(), coverage.Endpoints[path][method])
		err = matchBodyParams(event.RequestObject, coverage.Endpoints[path][method])
		if err != nil {
			glog.Errorf("%s", err)
		}
	}

	calculateCoverage(coverage)
	return coverage, nil
}
