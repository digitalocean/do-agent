package decorate

import dto "github.com/prometheus/client_model/go"

// Decorator decorates a list of metric families
type Decorator interface {
	Decorate([]*dto.MetricFamily)
	Name() string
}

// Chain of decorators to be applied to the metric family
type Chain []Decorator

// Decorate the metric family
func (c Chain) Decorate(mfs []*dto.MetricFamily) {
	for _, d := range c {
		d.Decorate(mfs)
	}
}

// Name is the name of the decorator
func (c Chain) Name() string {
	return "Chain"
}
