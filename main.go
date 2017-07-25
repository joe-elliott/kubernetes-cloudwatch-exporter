package main

import (
	"fmt"
	"log"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elb"
)

func main() {
	awsRegion := endpoints.UsEast1RegionID
	tagName := "KubernetesCluster"
	tagValue := "myCluster"

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))

	// get load balancer
	elbClient := elb.New(sess)

	loadBalancers, err := elbClient.DescribeLoadBalancers(nil)

	if err != nil {
		log.Fatalf("describeLoadBalancers %v", err)
	}

	elbNames := make([]*string, 0)

	for _, elbDesc := range loadBalancers.LoadBalancerDescriptions {
		elbNames = append(elbNames, elbDesc.LoadBalancerName)
	}

	// get tags
	loadBalancerTags, err := elbClient.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: elbNames,
	})

	if err != nil {
		log.Fatalf("describeTags %v", err)
	}

	// filter to only names that belong to the cluster
	elbNames = make([]*string, 0)
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
			elbNames = append(elbNames, elbTags.LoadBalancerName)
		}
	}

	// query metrics
	cwClient := cloudwatch.New(sess)

	now := time.Now()
	then := now.Add(-60 * time.Minute)
	metricName := "RequestCount"
	period := int64(60 * 60)
	statistic := "Sum"
	namespace := "AWS/ELB"
	dimension := "LoadBalancerName"

	for _, elbName := range elbNames {
		log.Printf("Getting stats for %v", *elbName)

		metricStats, err := cwClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
			Dimensions: []*cloudwatch.Dimension{&cloudwatch.Dimension{
				Name:  &dimension,
				Value: elbName,
			}},
			StartTime:          &then,
			EndTime:            &now,
			ExtendedStatistics: nil,
			MetricName:         &metricName,
			Namespace:          &namespace,
			Period:             &period,
			Statistics:         []*string{&statistic},
			Unit:               nil,
		})

		if err != nil {
			log.Fatalf("getMetricStatistics %v", err)
		}

		fmt.Printf("metricStats %v", *metricStats)
	}
}
