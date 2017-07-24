package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

func main() {
	awsRegion := endpoints.UsEast1RegionID
	tagName := "KubernetesCluster"
	tagValue := "myCluster"

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))

	elbClient := elb.New(sess)

	loadBalancers, err := elbClient.DescribeLoadBalancers(nil)

	if err != nil {
		log.Fatalf("describeLoadBalancers %v", err)
	}

	elbNames := make([]*string, 0)

	for _, elbDesc := range loadBalancers.LoadBalancerDescriptions {
		elbNames = append(elbNames, elbDesc.LoadBalancerName)
	}

	loadBalancerTags, err := elbClient.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: elbNames,
	})

	if err != nil {
		log.Fatalf("describeTags %v", err)
	}

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
		}
	}
}
