package tsclient

import (
	"fmt"
	"sort"
	"strings"
)

type labelDefType int

const metricNameLabel = "__name__"

const (
	dynamicLabel labelDefType = iota
	commonLabel
)

type definitionLabel struct {
	name      string
	labelType labelDefType

	// only when labelType is commonLabel
	commonValue string

	// only when labelType is dynamicLabel
	i int
}

// DefinitionOpts can be changed with opt setters
type DefinitionOpts struct {
	// CommonLabels is a set of static key-value labels
	CommonLabels map[string]string

	// MeasuredLabelKeys is a list of label keys whose values will be specified
	// at run-time
	MeasuredLabelKeys []string
}

// Definition holds the description of a metric.
type Definition struct {
	name         string
	sortedLabels []definitionLabel
}

// DefinitionOpt is an option initializer for metric registration.
type DefinitionOpt func(*DefinitionOpts)

// WithCommonLabels includes common labels
func WithCommonLabels(labels map[string]string) DefinitionOpt {
	return func(o *DefinitionOpts) {
		for k, v := range labels {
			o.CommonLabels[k] = v
		}
	}
}

// WithMeasuredLabels includes labels
func WithMeasuredLabels(labelKeys ...string) DefinitionOpt {
	return func(o *DefinitionOpts) {
		o.MeasuredLabelKeys = append(o.MeasuredLabelKeys, labelKeys...)
	}
}

// NewDefinition returns a new definition
func NewDefinition(name string, opts ...DefinitionOpt) *Definition {
	def := &DefinitionOpts{
		CommonLabels:      map[string]string{},
		MeasuredLabelKeys: []string{},
	}
	for _, opt := range opts {
		opt(def)
	}

	seen := map[string]bool{}
	sortedLabels := []definitionLabel{}
	for k, v := range def.CommonLabels {
		if _, ok := seen[k]; ok {
			panic(fmt.Sprintf("duplicate key: %q", k))
		}
		seen[k] = true
		sortedLabels = append(sortedLabels, definitionLabel{
			labelType:   commonLabel,
			name:        k,
			commonValue: v,
		})
	}
	for i, x := range def.MeasuredLabelKeys {
		if _, ok := seen[x]; ok {
			panic(fmt.Sprintf("duplicate key: %q", x))
		}
		seen[x] = true
		sortedLabels = append(sortedLabels, definitionLabel{
			labelType: dynamicLabel,
			name:      x,
			i:         i,
		})
	}
	sort.Slice(sortedLabels, func(i, j int) bool { return sortedLabels[i].name < sortedLabels[j].name })

	return &Definition{
		name:         name,
		sortedLabels: sortedLabels,
	}
}

// NewDefinitionFromMap returns a new definition with common labels for each given value, a "__name__" key must also be present
func NewDefinitionFromMap(m map[string]string) *Definition {
	name, ok := m[metricNameLabel]
	if !ok {
		panic("missing __name__ key")
	}
	delete(m, metricNameLabel)

	return NewDefinition(name, WithCommonLabels(m))
}

// GetLFM returns an lfm corresponding to a definition
func GetLFM(def *Definition, labels []string) (string, error) {
	lfm := []string{def.name}
	for _, x := range def.sortedLabels {
		lfm = append(lfm, x.name)
		switch x.labelType {
		case commonLabel:
			lfm = append(lfm, x.commonValue)
		case dynamicLabel:
			if x.i >= len(labels) {
				return "", ErrLabelMissmatch
			}
			lfm = append(lfm, labels[x.i])
		}
	}
	return strings.Join(lfm, "\x00"), nil
}

// ParseMetricDelimited parses a delimited message
func ParseMetricDelimited(s string) (map[string]string, error) {
	x := strings.Split(s, "\x00")
	if len(x)%2 != 1 {
		return nil, fmt.Errorf("incomplete delimited metric")
	}
	labels := map[string]string{}
	labels[metricNameLabel] = x[0]
	for i := 1; i < len(x); i += 2 {
		key := x[i]
		val := x[i+1]
		labels[key] = val
	}
	return labels, nil
}

// ConvertLFMMapToPrometheusEncodedName converts a metric in the form map[string]string{"__name__":"sonar_cpu","__source__":"user", "cpu":"cpu0"}
// to sonar_cpu{__source__="user",cpu="cpu0",host_id="899669",mode="iowait",user_id="208897"}
func ConvertLFMMapToPrometheusEncodedName(lfm map[string]string) string {
	name := lfm[metricNameLabel]

	labelStrings := make([]string, 0, len(lfm))

	for label, value := range lfm {
		if label == metricNameLabel {
			continue
		}
		lbl := fmt.Sprintf(`%s="%v"`, label, value)
		labelStrings = append(labelStrings, lbl)
	}
	sort.Strings(labelStrings)

	s := fmt.Sprintf(`%s{%s}`, name, strings.Join(labelStrings, ","))
	return s
}
