package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"kubernetes-cloudwatch-exporter/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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

	elbFunc := util.MakeELBNamesFunc(settings.TagName, settings.TagValue, sess)

	elbNamesInCluster, err := elbFunc()

	if err != nil {
		log.Fatalf("elbFunc %v", err)
	}

	fmt.Printf("Found %d load balancers\n", len(elbNamesInCluster))

	// query metrics
	cwClient := cloudwatch.New(sess)

	start := time.Now().Add(-settings.Delay + -settings.QueryRange)
	end := start.Add(settings.QueryRange)
	period := int64(settings.Period.Seconds())
	namespace := "AWS/ELB"
	dimension := "LoadBalancerName"

	for _, elbMetric := range settings.Metrics {

		log.Printf("Requesting Metrics %v", elbMetric)

		metricName := elbMetric.Name
		statistics := elbMetric.Statistics
		extendedStatistics := elbMetric.ExtendedStatistics

		for _, elbName := range elbNamesInCluster {
			log.Printf("Requesting for ELB %v", *elbName)

			metricStats, err := cwClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
				Dimensions: []*cloudwatch.Dimension{&cloudwatch.Dimension{
					Name:  &dimension,
					Value: elbName,
				}},
				StartTime:          &start,
				EndTime:            &end,
				ExtendedStatistics: extendedStatistics,
				MetricName:         &metricName,
				Namespace:          &namespace,
				Period:             &period,
				Statistics:         statistics,
				Unit:               nil,
			})

			if err != nil {
				log.Fatalf("getMetricStatistics %v", err)
			}

			fmt.Printf("metricStats %v", *metricStats)
		}
	}
}
