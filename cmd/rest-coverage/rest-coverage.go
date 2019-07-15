package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/mfranczy/crd-rest-coverage/pkg/report"
)

func main() {
	var (
		auditLogPath   string
		swaggerPath    string
		outputJSONPath string
	)

	flag.StringVar(&swaggerPath, "swagger-path", "", "path to swagger file")
	flag.StringVar(&auditLogPath, "audit-log-path", "", "path to k8s audit log file")
	flag.StringVar(&outputJSONPath, "output-path", "", "destination path for report file")
	flag.Parse()

	// TODO: improve glog format
	if swaggerPath == "" || auditLogPath == "" {
		glog.Exitf("swagger-path and audit-log-path are required")
	}

	_, err := report.Generate(auditLogPath, swaggerPath, "", outputJSONPath, true)
	if err != nil {
		glog.Exit(err)
	}
}
