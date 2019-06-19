package aggregate

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

const (
	metric      = "mysql_aggregate_me"
	table       = "table"
	bear        = "bear"
	donkey      = "donkey"
	lblOneName  = "label_one"
	lblOneValue = "label_one_value"
	lblTwoName  = "label_two"
	lblTwoValue = "label_two_value"
)

func TestAggregateOverTwoMetrics(t *testing.T) {
	metricName := metric
	metricType := dto.MetricType_GAUGE
	lblOneName := lblOneName
	lblOneValue := lblOneValue
	lblTwoName := lblTwoName
	lblTwoValue := lblTwoValue
	lblAggAwayName := table
	lblAggAwayValue := donkey
	lblAggAwayValueTwo := bear
	metricValue := 10.0
	metricValueTwo := 13.0

	metrics := make([]*dto.MetricFamily, 0)
	metrics = append(metrics, &dto.MetricFamily{Name: &metricName,
		Type: &metricType,
		Metric: []*dto.Metric{{
			Label: []*dto.LabelPair{
				{
					Name:  &lblOneName,
					Value: &lblOneValue,
				},
				{
					Name:  &lblTwoName,
					Value: &lblTwoValue,
				},
				{
					Name:  &lblAggAwayName,
					Value: &lblAggAwayValue,
				},
			},
			Gauge: &dto.Gauge{Value: &metricValue},
		}, {
			Label: []*dto.LabelPair{
				{
					Name:  &lblOneName,
					Value: &lblOneValue,
				},
				{
					Name:  &lblTwoName,
					Value: &lblTwoValue,
				},
				{
					Name:  &lblAggAwayName,
					Value: &lblAggAwayValueTwo,
				},
			},
			Gauge: &dto.Gauge{Value: &metricValueTwo},
		},
		}})

	aggregateSpec := map[string]string{metric: table, "other": "noeffect"}
	aggregated, err := Aggregate(metrics, aggregateSpec)
	require.NoError(t, err)
	require.Equal(t, 1, len(aggregated))
	require.Equal(t, 23.0, aggregated[0].Value)
	require.NotContains(t, aggregated[0].LFM, table)
	require.Contains(t, aggregated[0].LFM, lblOneName)
	require.Contains(t, aggregated[0].LFM, lblTwoName)
}

func TestAggregateDoesntAggregateMetricsWithDifferentLabels(t *testing.T) {
	metricName := metric
	metricType := dto.MetricType_GAUGE
	lblOneName := lblOneName
	lblOneValue := lblOneValue
	lblTwoName := lblTwoName
	lblTwoValue := lblTwoValue
	lblThreeName := "label_three"
	lblThreeValue := "label_three_value"
	lblAggAwayName := table
	lblAggAwayValue := donkey
	lblAggAwayValueTwo := bear
	metricValue := 10.0
	metricValueTwo := 13.0

	metrics := make([]*dto.MetricFamily, 0)
	metrics = append(metrics, &dto.MetricFamily{Name: &metricName,
		Type: &metricType,
		Metric: []*dto.Metric{{
			Label: []*dto.LabelPair{
				{
					Name:  &lblOneName,
					Value: &lblOneValue,
				},
				{
					Name:  &lblTwoName,
					Value: &lblTwoValue,
				},
				{
					Name:  &lblThreeName,
					Value: &lblThreeValue,
				},
				{
					Name:  &lblAggAwayName,
					Value: &lblAggAwayValue,
				},
			},
			Gauge: &dto.Gauge{Value: &metricValue},
		}, {
			Label: []*dto.LabelPair{
				{
					Name:  &lblOneName,
					Value: &lblOneValue,
				},
				{
					Name:  &lblTwoName,
					Value: &lblTwoValue,
				},
				{
					Name:  &lblAggAwayName,
					Value: &lblAggAwayValueTwo,
				},
			},
			Gauge: &dto.Gauge{Value: &metricValueTwo},
		},
		}})

	aggregateSpec := map[string]string{metric: table, "other": "noeffect"}
	aggregated, err := Aggregate(metrics, aggregateSpec)
	require.NoError(t, err)
	require.Equal(t, 2, len(aggregated))
	require.NotContains(t, aggregated[0].LFM, table)
	require.Contains(t, aggregated[0].LFM, lblOneName)
	require.Contains(t, aggregated[0].LFM, lblTwoName)
	require.NotContains(t, aggregated[1].LFM, table)
	require.Contains(t, aggregated[1].LFM, lblOneName)
	require.Contains(t, aggregated[1].LFM, lblTwoName)
	valueOne := aggregated[0].Value
	valueTwo := aggregated[1].Value
	if valueOne == 10.0 {
		require.Contains(t, aggregated[0].LFM, lblThreeName)
	} else {
		require.Contains(t, aggregated[1].LFM, lblThreeName)
	}
	require.True(t, valueOne == 10 || valueOne == 13)
	require.True(t, valueTwo == 10 || valueTwo == 13)
	require.Equal(t, 23.0, valueOne+valueTwo)
}

func TestAggregateNilAggregateSpecHasNoEffect(t *testing.T) {
	metricName := metric
	metricType := dto.MetricType_GAUGE
	lblOneName := lblOneName
	lblOneValue := lblOneValue
	lblTwoName := lblTwoName
	lblTwoValue := lblTwoValue
	lblAggAwayName := table
	lblAggAwayValue := donkey
	lblAggAwayValueTwo := bear
	metricValue := 10.0
	metricValueTwo := 13.0

	metrics := make([]*dto.MetricFamily, 0)
	metrics = append(metrics, &dto.MetricFamily{Name: &metricName,
		Type: &metricType,
		Metric: []*dto.Metric{{
			Label: []*dto.LabelPair{
				{
					Name:  &lblOneName,
					Value: &lblOneValue,
				},
				{
					Name:  &lblTwoName,
					Value: &lblTwoValue,
				},
				{
					Name:  &lblAggAwayName,
					Value: &lblAggAwayValue,
				},
			},
			Gauge: &dto.Gauge{Value: &metricValue},
		}, {
			Label: []*dto.LabelPair{
				{
					Name:  &lblOneName,
					Value: &lblOneValue,
				},
				{
					Name:  &lblTwoName,
					Value: &lblTwoValue,
				},
				{
					Name:  &lblAggAwayName,
					Value: &lblAggAwayValueTwo,
				},
			},
			Gauge: &dto.Gauge{Value: &metricValueTwo},
		},
		}})

	aggregated, err := Aggregate(metrics, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(aggregated))
	require.Contains(t, aggregated[0].LFM, table)
	require.Contains(t, aggregated[0].LFM, lblOneName)
	require.Contains(t, aggregated[0].LFM, lblTwoName)
	require.Contains(t, aggregated[1].LFM, table)
	require.Contains(t, aggregated[1].LFM, lblOneName)
	require.Contains(t, aggregated[1].LFM, lblTwoName)
	valueOne := aggregated[0].Value
	valueTwo := aggregated[1].Value
	require.True(t, valueOne == 10 || valueOne == 13)
	require.True(t, valueTwo == 10 || valueTwo == 13)
	require.Equal(t, 23.0, valueOne+valueTwo)
}
