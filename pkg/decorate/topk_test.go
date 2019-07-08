package decorate

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func floatPtr(f float64) *float64 { return &f }

func TestTopK(t *testing.T) {

	decorator := TopK{
		K: 2, // Top 2 metrics
		N: "sonar_",
	}

	items := []*dto.MetricFamily{
		{
			Name: sPtr("sonar_cpu"),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value1"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(75.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value2"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(160.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value3"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(79.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value4"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(1.30),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value5"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(0.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value6"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(0.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("user_id"),
							Value: sPtr("1234"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(80.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("dbaas_uuid"),
							Value: sPtr("hello-world"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(90.00),
					},
				},
			},
		},
	}

	final := []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  sPtr("some_name"),
					Value: sPtr("some_value2"),
				},
			},
			Counter: &dto.Counter{
				Value: floatPtr(160),
			},
		},
		{
			Label: []*dto.LabelPair{
				{
					Name:  sPtr("dbaas_uuid"),
					Value: sPtr("hello-world"),
				},
			},
			Counter: &dto.Counter{
				Value: floatPtr(90.00),
			},
		},
	}

	// Return only the top K metrics
	decorator.Decorate(items)
	require.Equal(t, final, items[0].Metric)
}

func TestTopKNoMatch(t *testing.T) {

	decorator := TopK{
		K: 2, // Top 2 metrics
		N: "doesnt_match",
	}

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
					Counter: &dto.Counter{
						Value: floatPtr(0.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("user_id"),
							Value: sPtr("1234"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(80.00),
					},
				},
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("dbaas_uuid"),
							Value: sPtr("hello-world"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(90.00),
					},
				},
			},
		},
	}

	final := []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  sPtr("some_name"),
					Value: sPtr("some_value"),
				},
			},
			Counter: &dto.Counter{
				Value: floatPtr(0.00),
			},
		},
		{
			Label: []*dto.LabelPair{
				{
					Name:  sPtr("user_id"),
					Value: sPtr("1234"),
				},
			},
			Counter: &dto.Counter{
				Value: floatPtr(80.00),
			},
		},
		{
			Label: []*dto.LabelPair{
				{
					Name:  sPtr("dbaas_uuid"),
					Value: sPtr("hello-world"),
				},
			},
			Counter: &dto.Counter{
				Value: floatPtr(90.00),
			},
		},
	}

	// Return only the top K metrics
	decorator.Decorate(items)
	require.Equal(t, final, items[0].Metric)
}

func TestTopKTooMany(t *testing.T) {

	decorator := TopK{
		K: 30, // Top 30 metrics
		N: "sonar_",
	}

	items := []*dto.MetricFamily{
		{
			Name: sPtr("sonar_cpu"),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  sPtr("some_name"),
							Value: sPtr("some_value1"),
						},
					},
					Counter: &dto.Counter{
						Value: floatPtr(75.00),
					},
				},
			},
		},
	}

	final := []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  sPtr("some_name"),
					Value: sPtr("some_value1"),
				},
			},
			Counter: &dto.Counter{
				Value: floatPtr(75.00),
			},
		},
	}

	decorator.Decorate(items)
	require.Equal(t, final, items[0].Metric)
}
