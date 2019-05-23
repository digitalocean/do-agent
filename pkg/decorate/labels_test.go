package decorate

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func sPtr(s string) *string {
	return &s
}

func TestAppendLabels(t *testing.T) {

	decorator := LabelAppender([]*dto.LabelPair{
		{
			Name:  sPtr("user_id"),
			Value: sPtr("1234"),
		},
		{
			Name:  sPtr("dbaas_uuid"),
			Value: sPtr("hello-world"),
		},
	})

	items := []*dto.MetricFamily{
		{
			Name: sPtr("sonar_cpu"),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value"),
						},
					},
				},
			},
		},
	}

	final := []*dto.LabelPair{
		{
			Name:  sPtr("some_name"),
			Value: sPtr("some_value"),
		},
		{
			Name:  sPtr("user_id"),
			Value: sPtr("1234"),
		},
		{
			Name:  sPtr("dbaas_uuid"),
			Value: sPtr("hello-world"),
		},
	}

	// Append new labels
	decorator.Decorate(items)
	require.Equal(t, final, items[0].Metric[0].Label)
}

func TestAppendLabelsToEmptyLabels(t *testing.T) {
	decorator := LabelAppender([]*dto.LabelPair{
		{
			Name:  sPtr("user_id"),
			Value: sPtr("1234"),
		},
		{
			Name:  sPtr("dbaas_uuid"),
			Value: sPtr("hello-world"),
		},
	})

	items := []*dto.MetricFamily{
		{
			Name: sPtr("sonar_cpu"),
			Metric: []*dto.Metric{
				{
					Label: nil,
				},
			},
		},
	}

	final := []*dto.LabelPair{
		{
			Name:  sPtr("user_id"),
			Value: sPtr("1234"),
		},
		{
			Name:  sPtr("dbaas_uuid"),
			Value: sPtr("hello-world"),
		},
	}

	// Append new labels
	decorator.Decorate(items)
	require.Equal(t, final, items[0].Metric[0].Label)
}

func TestAppendLabelsToEmptyMetric(t *testing.T) {
	decorator := LabelAppender([]*dto.LabelPair{
		{
			Name:  sPtr("user_id"),
			Value: sPtr("1234"),
		},
		{
			Name:  sPtr("dbaas_uuid"),
			Value: sPtr("hello-world"),
		},
	})

	items := []*dto.MetricFamily{
		{
			Name:   sPtr("sonar_cpu"),
			Metric: nil,
		},
	}

	// Append labels; expect nothing to be appended
	decorator.Decorate(items)
	require.Equal(t, 0, len(items[0].Metric))
}
