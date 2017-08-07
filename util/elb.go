package util

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

func MakeELBNamesFunc(tagName string, tagValue string, session *session.Session) func() ([]*string, error) {

	// get load balancer
	elbClient := elb.New(session)

	//
	return func() ([]*string, error) {
		loadBalancers, err := elbClient.DescribeLoadBalancers(nil)

		if err != nil {
			return nil, err
		}

		elbNames := make([]*string, 0)
		elbNamesInCluster := make([]*string, 0)

		for _, elbDesc := range loadBalancers.LoadBalancerDescriptions {
			elbNames = append(elbNames, elbDesc.LoadBalancerName)
		}

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
				return nil, err
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

		return elbNamesInCluster, nil
	}
}
