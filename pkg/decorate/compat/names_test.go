package compat

import (
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestCompatConvertsLabels(t *testing.T) {
	for old, new := range nameConversions {
		t.Run(old, func(t *testing.T) {
			mfs := []*dto.MetricFamily{
				{Name: &old},
			}
			Names{}.Decorate(mfs)

			assert.Equal(t, new, mfs[0].GetName())
		})
	}
}

func TestCompatIsCaseInsensitive(t *testing.T) {
	for old, new := range nameConversions {
		t.Run(old, func(t *testing.T) {
			mfs := []*dto.MetricFamily{
				{Name: sptr(strings.ToUpper(old))},
			}
			Names{}.Decorate(mfs)

			assert.Equal(t, new, mfs[0].GetName())
		})
	}
}

func TestNamesHasName(t *testing.T) {
	assert.Equal(t, "compat.Names", Names{}.Name())
}
