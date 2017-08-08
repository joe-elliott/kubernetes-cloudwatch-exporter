package main

import (
	"flag"
	"fmt"
	"log"

	"kubernetes-cloudwatch-exporter/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var settingsFile = flag.String("settings-file", "./settings.json", "Path to load as the settings file")

func main() {
	settings, err := util.NewSettings(*settingsFile)

	if err != nil {
		log.Fatalf("settings.NewSettings %v", err)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(settings.AWSRegion),
	}))

	getELBNamesFunc := util.MakeELBNamesFunc(settings.TagName, settings.TagValue, sess)
	getMetricsFunc := util.MakeMetricsFunc(sess)

	elbNamesInCluster, err := getELBNamesFunc()

	if err != nil {
		log.Fatalf("elbFunc %v", err)
	}

	fmt.Printf("Found %d load balancers\n", len(elbNamesInCluster))

	for _, elbName := range elbNamesInCluster {
		for _, elbMetric := range settings.Metrics {
			log.Printf("Requesting Metrics %v", elbMetric)
			log.Printf("Requesting for ELB %v", *elbName)

			datapoints, err := getMetricsFunc(elbName, &elbMetric, settings)

			if err != nil {
				log.Fatalf("metricsFunc %v", err)
			}

			log.Printf("Datapoints %v", datapoints)
		}
	}
}
