package compat

import (
	"fmt"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

const nodeExporterCPUName = "node_cpu_seconds_total"

func TestCPUChangesNames(t *testing.T) {
	const expected = "sonar_cpu"

	mfs := []*dto.MetricFamily{{Name: sptr(nodeExporterCPUName)}}
	CPU{}.Decorate(mfs)

	assert.Equal(t, expected, mfs[0].GetName())
}

func TestCPUChangesLabelValues(t *testing.T) {
	for i := 0; i < 4; i++ {
		dec := CPU{}

		cpu := i
		expected := fmt.Sprintf("cpu%d", cpu)
		t.Run(expected, func(t *testing.T) {
			v := 1.0
			metric := dto.Metric{
				Gauge: &dto.Gauge{Value: &v},
				Label: []*dto.LabelPair{
					{
						Name:  sptr("cpu"),
						Value: sptr(fmt.Sprint(cpu)),
					},
				},
			}
			mfs := []*dto.MetricFamily{
				{
					Type:   &counterMetricType,
					Name:   sptr(nodeExporterCPUName),
					Metric: []*dto.Metric{&metric},
				},
			}
			dec.Decorate(mfs)
			assert.EqualValues(t, expected, metric.Label[0].GetValue())
		})
	}
}

func TestCPUUpdatesAllLabels(t *testing.T) {
	mfs := []*dto.MetricFamily{}

	for i := 0; i < 4; i++ {
		v := 1.0
		m := dto.Metric{
			Gauge: &dto.Gauge{Value: &v},
			Label: []*dto.LabelPair{
				{
					Name:  sptr("cpu"),
					Value: sptr(fmt.Sprint(i)),
				},
			},
		}
		mfs = append(mfs, &dto.MetricFamily{
			Type:   &counterMetricType,
			Name:   sptr(nodeExporterCPUName),
			Metric: []*dto.Metric{&m},
		})
	}

	CPU{}.Decorate(mfs)

	for i, mf := range mfs {
		expected := fmt.Sprintf("cpu%d", i)
		assert.EqualValues(t, expected, mf.GetMetric()[0].Label[0].GetValue())
	}
}

func TestCPUDoesNotChangeOtherLabelValues(t *testing.T) {
	dec := CPU{}

	const expected = "0"

	v := 1.0
	metric := dto.Metric{
		Gauge: &dto.Gauge{Value: &v},
		Label: []*dto.LabelPair{
			{
				Name:  sptr("notcpu"),
				Value: sptr(expected),
			},
		},
	}
	mfs := []*dto.MetricFamily{
		{
			Type:   &counterMetricType,
			Name:   sptr(nodeExporterCPUName),
			Metric: []*dto.Metric{&metric},
		},
	}
	dec.Decorate(mfs)
	assert.EqualValues(t, expected, metric.Label[0].GetValue())
}

func TestCPUDoesNotChangeOtherMetrics(t *testing.T) {
	dec := CPU{}

	const expected = "0"

	v := 1.0
	metric := dto.Metric{
		Gauge: &dto.Gauge{Value: &v},
		Label: []*dto.LabelPair{
			{
				Name:  sptr("cpu"),
				Value: sptr(expected),
			},
		},
	}
	mfs := []*dto.MetricFamily{
		{
			Type:   &counterMetricType,
			Name:   sptr("something else"),
			Metric: []*dto.Metric{&metric},
		},
	}
	dec.Decorate(mfs)
	assert.EqualValues(t, expected, metric.Label[0].GetValue())
}

func TestCPUSkipsWhenFailsParsingCPUNumber(t *testing.T) {
	dec := CPU{}

	v := 1.0
	metric := dto.Metric{
		Gauge: &dto.Gauge{Value: &v},
		Label: []*dto.LabelPair{
			{
				Name: sptr("cpu"),
				// this should be a number
				Value: sptr("not a number"),
			},
			{
				Name:  sptr("cpu"),
				Value: sptr("1"),
			},
		},
	}
	mfs := []*dto.MetricFamily{
		{
			Type:   &counterMetricType,
			Name:   sptr(nodeExporterCPUName),
			Metric: []*dto.Metric{&metric},
		},
	}
	dec.Decorate(mfs)

	assert.EqualValues(t, "not a number", metric.Label[0].GetValue())
	assert.EqualValues(t, "cpu1", metric.Label[1].GetValue())
}

func TestCPUHasName(t *testing.T) {
	assert.Equal(t, "compat.CPU", CPU{}.Name())
}
