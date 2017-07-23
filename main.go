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
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))

	if sess != nil {
		fmt.Println("yay")
	}

	elbClient := elb.New(sess)

	loadBalancers, err := elbClient.DescribeLoadBalancers(nil)

	if err != nil {
		log.Fatalf("describeLoadBalancers %v", err)
	}

	for _, elbDesc := range loadBalancers.LoadBalancerDescriptions {
		fmt.Printf("%v\n", *elbDesc.LoadBalancerName)
	}
}
