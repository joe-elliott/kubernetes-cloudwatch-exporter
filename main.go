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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	settingsFile = flag.String("settings-file", "./settings.json", "Path to load as the settings file")
	promMetrics  = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_cw_metric",
			Help: "Cloudwatch Metrics.",
		},
		[]string{"elb"},
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

					// todo: find and choose correct property.
					//       report all values
					//       read docs and figure out how to actually do this all correctly
					if len(datapoints) > 0 && datapoints[0].Average != nil {
						promMetrics.WithLabelValues(*elbName).Observe(*datapoints[0].Average)
					} else {
						promMetrics.WithLabelValues(*elbName).Observe(0)
					}
				}
			}

			time.Sleep(60 * time.Second)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
