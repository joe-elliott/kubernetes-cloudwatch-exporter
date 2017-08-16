package util

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func MakeMetricsFunc(session *session.Session) func(*string, *ELBMetric, *ELBSettings) ([]*cloudwatch.Datapoint, error) {

	cwClient := cloudwatch.New(session)

	return func(elbName *string, metric *ELBMetric, settings *ELBSettings) ([]*cloudwatch.Datapoint, error) {

		start := time.Now().Add(time.Duration(-settings.DelaySeconds+-settings.QuerySeconds) * time.Second)
		end := start.Add(time.Duration(settings.QuerySeconds) * time.Second)
		period := settings.PeriodSeconds
		namespace := "AWS/ELB"
		dimension := "LoadBalancerName"

		metricName := metric.Name
		statistics := metric.Statistics
		extendedStatistics := metric.ExtendedStatistics

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
			return nil, err
		}

		return metricStats.Datapoints, nil
	}
}
