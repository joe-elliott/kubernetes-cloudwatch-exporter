package util

import (
	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

type ELBSettings struct {
	DelaySeconds   int64
	PeriodSeconds  int64
	QuerySeconds   int64
	AWSRegion      string
	TagName        string
	TagValue       string
	AppTagName     string
	RequireAppName bool
	Metrics        []ELBMetric
}

type ELBMetric struct {
	Name               string
	Statistics         []*string
	ExtendedStatistics []*string
	Default            *float64
}

func NewSettings(filepath string) (*ELBSettings, error) {
	raw, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var settings ELBSettings
	err = json.Unmarshal(raw, &settings)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func (this *ELBSettings) UnmarshalJSON(data []byte) error {
	type Alias ELBSettings

	read := &Alias{
		DelaySeconds:   60,
		PeriodSeconds:  60,
		QuerySeconds:   60,
		AWSRegion:      endpoints.UsEast1RegionID,
		TagName:        "KubernetesCluster",
		TagValue:       "MyCluster",
		AppTagName:     "kubernetes.io/service-name",
		RequireAppName: false,
		Metrics:        nil,
	}

	if err := json.Unmarshal(data, &read); err != nil {
		return err
	}

	*this = ELBSettings(*read)

	return nil
}
