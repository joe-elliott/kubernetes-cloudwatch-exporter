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
	"github.com/aws/aws-sdk-go/service/elb"
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

	// get load balancer
	elbClient := elb.New(sess)

	loadBalancers, err := elbClient.DescribeLoadBalancers(nil)

	if err != nil {
		log.Fatalf("describeLoadBalancers %v", err)
	}

	elbNames := make([]*string, 0)
	elbNamesInCluster := make([]*string, 0)

	for _, elbDesc := range loadBalancers.LoadBalancerDescriptions {
		elbNames = append(elbNames, elbDesc.LoadBalancerName)
	}

	var tagName = settings.TagName
	var tagValue = settings.TagValue

	for i := 0; i < (len(elbNames)/20)+1; i++ {

		startSlice := i * 20
		endSlice := (i + 1) * 20

		if endSlice > len(elbNames) {
			endSlice = len(elbNames)
		}

		// get tags
		loadBalancerTags, err := elbClient.DescribeTags(&elb.DescribeTagsInput{
			LoadBalancerNames: elbNames[startSlice:endSlice],
		})

		if err != nil {
			log.Fatalf("describeTags %v", err)
		}

		// filter to only names that belong to the cluster
		fmt.Println("In Cluster:")

		for _, elbTags := range loadBalancerTags.TagDescriptions {
			inCluster := false

			for _, kvp := range elbTags.Tags {
				if *kvp.Key == tagName && *kvp.Value == tagValue {
					inCluster = true
					break
				}
			}

			if inCluster {
				fmt.Printf("%v\n", *elbTags.LoadBalancerName)
				elbNamesInCluster = append(elbNamesInCluster, elbTags.LoadBalancerName)
			}
		}
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
