package tsclient

import (
	"fmt"
	"sort"
	"strings"
)

type labelDefType int

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
	name, ok := m["__name__"]
	if !ok {
		panic("missing __name__ key")
	}
	delete(m, "__name__")

	return NewDefinition(name, WithCommonLabels(m))
}

// GetLFM returns an lfm corresponding to a defitnition
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
