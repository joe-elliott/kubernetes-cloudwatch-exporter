# kubernetes-cloudwatch-exporter

Exports Kubernetes specific Cloudwatch metrics to Prometheus

### Goals

This repo is a work in progress.  The ultimate goal is to export cloudwatch metrics specific to Kubernetes.  Initial progress will be made towards:

- ELB metrics only for ELBs with specific tags (i.e. KubernetesCluster=k8s.example.com)
- When no data is received pass 0 instead of nothing (this prevents issues with stale data https://github.com/prometheus/prometheus/issues/398)
- Pass ELB AWS tags as Prometheus labels due to Kubernetes ELB naming horribleness: https://github.com/kubernetes/kubernetes/issues/29789
