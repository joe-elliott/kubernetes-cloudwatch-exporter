# kubernetes-cloudwatch-exporter

Exports Kubernetes specific Cloudwatch metrics to Prometheus

### Usage

Deploy this application to your cluster to expose kubernetes cloudwatch metrics to prometheus.  It searches AWS Cloudwatch metrics for appropriately tagged resources and exposes configured metrics.  It also uses tags to nicely label prometheus metrics for easier visualization.

Currently it only supports ELB Metrics but suggestions are welcome.

#### Command line 

`./kubernetes-cloudwatch-exporter --settings-file <path to settings file>`

#### Sample Settings file

```
{
    "delaySeconds": 60,
    "periodSeconds": 60,
    "querySeconds": 60,
    "awsRegion": "us-east-1",
    "tagName": "KubernetesCluster",
    "tagValue": "k8s.example.com",
    "appTagName": "kubernetes.io/service-name",
    "requireAppName": false,
    "metrics": [
        {
            "name": "RequestCount",
            "statistics": ["Sum"],
            "default": 0
        },
        {
            "name": "Latency",
            "extendedStatistics": ["p50","p90","p99"]
        }
    ]
}
```

- `delay` - Time in the past to make cloudwatch requests.  Defaults to 60s to allow convergence.
- `period` - Cloudwatch period to make requests for.  Granularity/Bucket size.
- `queryRange` - Span of time to make for which to make Cloudwatch requests.
- `awsRegion` - For your sake I hope this isn't us-east-1. But, yeah, us-east-1.
- `tagName` - Tag name to use for cluster name.
- `tagValue` - Cluster name to search for.
- `appTagName` - Tag to use to extract application name.
- `requireAppName` - Flag to control whether ELBs with no app name will be included.  This can be used to exclude the API elb if that's desirable.

#### Sample Metrics

All cloudwatch metrics will be exposed as gauges with the following labels:

```
k8scw_elb_metric{k8sapp="k8s-appname",elb="a5c10cde971f831e7b7120ac23c20e11",metric="Latency",k8snamespace="k8s-namespace",statistic="p50"} 0.0025220091548062407
k8scw_elb_metric{k8sapp="k8s-appname",elb="a5c10cde971f831e7b7120ac23c20e11",metric="Latency",k8snamespace="k8s-namespace",statistic="p90"} 0.009827765151084543
k8scw_elb_metric{k8sapp="k8s-appname",elb="a5c10cde971f831e7b7120ac23c20e11",metric="Latency",k8snamespace="k8s-namespace",statistic="p99"} 0.09012280488075149
k8scw_elb_metric{k8sapp="k8s-appname",elb="a5c10cde971f831e7b7120ac23c20e11",metric="RequestCount",k8snamespace="k8s-namespace",statistic="Sum"} 125
```

The application also exposes an error counter:

```
k8scw_error_total 2
```

### Why does this exist?

This specialized exporter was created to export Kubernetes Cloudwatch metrics.  Advantages over a generic exporter:

- Uses standard Kubernetes tags to find only those ELBs belonging to your cluster.
- Uses standard Kubernetes tags to associate application names with metrics until this is resolved:  https://github.com/kubernetes/kubernetes/issues/29789
- Substitute in default values if cloudwatch doesn't return metrics until this is resolved:  https://github.com/prometheus/prometheus/issues/398