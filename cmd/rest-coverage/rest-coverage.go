package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	"github.com/mfranczy/crd-rest-coverage/pkg/report"
)

var (
	BuildVersion string = ""
)

func main() {
	var (
		auditLogPath          string
		swaggerPath           string
		outputJSONPath        string
		detailed              bool
		ignoreResourceVersion bool
		version               bool
	)

	// TODO: add filter param
	flag.StringVar(&swaggerPath, "swagger-path", "", "path to swagger file")
	flag.StringVar(&auditLogPath, "audit-log-path", "", "path to k8s audit log file")
	flag.StringVar(&outputJSONPath, "output-path", "", "destination path for report file")
	flag.BoolVar(&detailed, "detailed", false, "show report with coverage for each endpoint")
	flag.BoolVar(&ignoreResourceVersion, "ignore-resource-version", false, "ignore resource version")
	flag.BoolVar(&version, "version", false, "build version")
	flag.Parse()

	if version {
		fmt.Println(BuildVersion)
		return
	}

	// TODO: improve glog format
	if swaggerPath == "" || auditLogPath == "" {
		glog.Exitf("params --swagger-path and --audit-log-path are required")
	}

	coverage, err := report.Generate(auditLogPath, swaggerPath, "", ignoreResourceVersion)
	if err != nil {
		glog.Exit(err)
	}

	if outputJSONPath != "" {
		report.Dump(outputJSONPath, coverage)
	} else {
		report.Print(coverage, detailed)
	}
}
