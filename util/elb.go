package util

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	logging "github.com/op/go-logging"
)

type ELBDescription struct {
	Name         *string
	AppName      *string
	AppNamespace *string
}

var _log = logging.MustGetLogger("cloudwatch-exporter")

func MakeELBNamesFunc(tagName string, tagValue string, appTagName string, requireAppName bool, session *session.Session) func() ([]*ELBDescription, error) {

	// get load balancer
	elbClient := elb.New(session)

	//
	return func() ([]*ELBDescription, error) {
		loadBalancers, err := elbClient.DescribeLoadBalancers(nil)

		if err != nil {
			return nil, err
		}

		elbNames := make([]*string, 0)
		elbDescriptions := make([]*ELBDescription, 0)

		for _, elbDesc := range loadBalancers.LoadBalancerDescriptions {
			elbNames = append(elbNames, elbDesc.LoadBalancerName)
		}

		for i := 0; i < (len(elbNames)/20)+1; i++ {

			startSlice := i * 20
			endSlice := (i + 1) * 20

			if endSlice > len(elbNames) {
				endSlice = len(elbNames)
			}

			if startSlice == endSlice {
				continue
			}

			// get tags
			loadBalancerTags, err := elbClient.DescribeTags(&elb.DescribeTagsInput{
				LoadBalancerNames: elbNames[startSlice:endSlice],
			})

			if err != nil {
				return nil, err
			}

			// filter to only names that belong to the cluster
			_log.Debug("In Cluster:")

			for _, elbTags := range loadBalancerTags.TagDescriptions {
				inCluster := false
				appName := ""
				appNamespace := ""

				for _, kvp := range elbTags.Tags {
					if *kvp.Key == tagName && *kvp.Value == tagValue {
						inCluster = true
					}

					if *kvp.Key == appTagName {
						parts := strings.Split(*kvp.Value, "/")

						if len(parts) == 2 {
							appNamespace = parts[0]
							appName = parts[1]
						} else {
							appNamespace = ""
							appName = *kvp.Value
						}
					}
				}

				if requireAppName && appName == "" {
					continue
				}

				if inCluster {
					_log.Debugf("%v\n", *elbTags.LoadBalancerName)

					desc := &ELBDescription{
						Name:         elbTags.LoadBalancerName,
						AppName:      &appName,
						AppNamespace: &appNamespace,
					}

					elbDescriptions = append(elbDescriptions, desc)
				}
			}
		}

		return elbDescriptions, nil
	}
}
