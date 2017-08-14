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
	_settingsFile = flag.String("settings-file", "./settings.json", "Path to load as the settings file")
	_elbMetrics   = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "k8scw",
			Subsystem: "elb",
			Name:      "metric",
			Help:      "Kubernetes ELB metrics",
		},
		[]string{"elb", "app", "namespace", "metric", "statistic"},
	)
)

func init() {
	prometheus.MustRegister(_elbMetrics)
}

func main() {
	settings, err := util.NewSettings(*_settingsFile)

	if err != nil {
		log.Fatalf("settings.NewSettings %v", err)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(settings.AWSRegion),
	}))

	getELBNamesFunc := util.MakeELBNamesFunc(settings.TagName, settings.TagValue, settings.AppTagName, settings.RequireAppName, sess)
	getMetricsFunc := util.MakeMetricsFunc(sess)

	go func() {
		for {
			elbDescriptions, err := getELBNamesFunc()

			if err != nil {
				log.Fatalf("elbFunc %v", err)
			}

			fmt.Printf("Found %d load balancers\n", len(elbDescriptions))

			for _, elbDesc := range elbDescriptions {
				for _, elbMetric := range settings.Metrics {

					log.Printf("Requesting Metrics %v", elbMetric)
					log.Printf("Requesting for ELB %v", *elbDesc.Name)

					datapoints, err := getMetricsFunc(elbDesc.Name, &elbMetric, settings)

					if err != nil {
						log.Fatalf("metricsFunc %v", err)
					}

					log.Printf("Datapoints %v", datapoints)

					observeDatapoints(datapoints, elbMetric, elbDesc)
				}
			}

			// sleep for the query range.  so if our cloudwatch queries covers 1 minute of data we request new data
			// once a minute.
			time.Sleep(settings.QueryRange)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func observeDatapoints(datapoints []*cloudwatch.Datapoint, elbMetric util.ELBMetric, elbDesc *util.ELBDescription) {

	if len(datapoints) > 0 {
		for _, dp := range datapoints {
			metrics := generateMetrics(dp)

			for n, v := range metrics {
				_elbMetrics.WithLabelValues(*elbDesc.Name, *elbDesc.AppName, *elbDesc.AppNamespace, elbMetric.Name, n).Set(v)
			}
		}
	} else if elbMetric.Default != nil {
		// there were no datapoints for this metric.  in this case let's default to 0s to avoid prometheus staleness issues:
		//  https://github.com/prometheus/prometheus/issues/398

		for _, metric := range elbMetric.Statistics {
			_elbMetrics.WithLabelValues(*elbDesc.Name, *elbDesc.AppName, *elbDesc.AppNamespace, elbMetric.Name, *metric).Set(*elbMetric.Default)
		}

		for _, metric := range elbMetric.ExtendedStatistics {
			_elbMetrics.WithLabelValues(*elbDesc.Name, *elbDesc.AppName, *elbDesc.AppNamespace, elbMetric.Name, *metric).Set(*elbMetric.Default)
		}
	}
}

func generateMetrics(dp *cloudwatch.Datapoint) map[string]float64 {

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
