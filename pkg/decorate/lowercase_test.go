package decorate

import (
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestLowercaseNamesChangesLabels(t *testing.T) {
	d := LowercaseNames{}

	actual := "JKLKJSFDJKLjkasdfjklasdf"
	expected := strings.ToLower(actual)

	items := []*dto.MetricFamily{
		{Name: &actual},
	}
	d.Decorate(items)
	assert.Equal(t, expected, items[0].GetName())
}
