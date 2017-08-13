package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"kubernetes-cloudwatch-exporter/util"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	settingsFile = flag.String("settings-file", "./settings.json", "Path to load as the settings file")
	promMetrics  = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_cw_metric",
			Help: "Cloudwatch Metrics.",
		},
		[]string{"elb", "name", "statistic"},
	)
)

func init() {
	prometheus.MustRegister(promMetrics)
}

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

	go func() {
		for {
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

					observeDatapoints(datapoints, elbMetric, *elbName)
				}
			}

			time.Sleep(60 * time.Second)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func observeDatapoints(datapoints []*cloudwatch.Datapoint, elbMetric util.ELBMetric, elbName string) {

	for _, dp := range datapoints {
		for n, v := range getMetrics(dp) {
			promMetrics.WithLabelValues(elbName, elbMetric.Name, n).Set(v)
		}
	}
}

func getMetrics(dp *cloudwatch.Datapoint) map[string]float64 {

	metrics := make(map[string]float64)

	if dp.Average != nil {
		metrics["Average"] = *dp.Average
	}

	if dp.Maximum != nil {
		metrics["Maximum"] = *dp.Maximum
	}

	if dp.Minimum != nil {
		metrics["Minimum"] = *dp.Minimum
	}

	if dp.SampleCount != nil {
		metrics["SampleCount"] = *dp.SampleCount
	}

	if dp.Sum != nil {
		metrics["Sum"] = *dp.Sum
	}

	if dp.ExtendedStatistics != nil {
		for p, v := range dp.ExtendedStatistics {
			metrics[p] = *v
		}
	}

	return metrics
}
