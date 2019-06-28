package decorate

import (
	"container/heap"
	"regexp"

	dto "github.com/prometheus/client_model/go"
)

// TopK is a decorator that removes metrics not in the top K by value
type TopK struct {
	K uint
	N string
}

// Decorate removes all but the top K metrics for a given metric name
func (t TopK) Decorate(mfs []*dto.MetricFamily) {
	var topk []*metricHeap
	var idx []int
	for i, fam := range mfs {
		if match, _ := regexp.MatchString(t.N, fam.GetName()); match {
			tk := metricHeap(fam.Metric)
			heap.Init(&tk)
			idx = append(idx, i)
			topk = append(topk, &tk)
		}
	}

	for i := range topk {
		mfs[idx[i]].Metric = topk[i].TopK(t.K)
	}
}

// Name is the name of this decorator
func (t TopK) Name() string {
	return "TopK"
}

type metricHeap []*dto.Metric

func (m metricHeap) Len() int { return len(m) }

func (m metricHeap) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

// invert less function to create max heap
func (m metricHeap) Less(i, j int) bool {
	switch {
	case m[i].Gauge != nil:
		return *m[i].Gauge.Value > *m[j].Gauge.Value
	case m[i].Counter != nil:
		return *m[i].Counter.Value > *m[j].Counter.Value
	case m[i].Summary != nil:
		return *m[i].Summary.SampleSum > *m[j].Summary.SampleSum
	case m[i].Untyped != nil:
		return *m[i].Untyped.Value > *m[j].Untyped.Value
	case m[i].Histogram != nil:
		return *m[i].Histogram.SampleSum > *m[j].Histogram.SampleSum
	default:
		return false
	}
}

func (m *metricHeap) Push(x interface{}) {
	*m = append(*m, x.(*dto.Metric))
}

func (m *metricHeap) Pop() interface{} {
	old := *m
	n := len(old)
	x := old[n-1]
	*m = old[:n-1]
	return x
}

func (m *metricHeap) TopK(k uint) []*dto.Metric {
	var topk []*dto.Metric
	if k > uint(len(*m)) {
		k = uint(len(*m))
	}
	for i := uint(0); i < k; i++ {
		topk = append(topk, heap.Pop(m).(*dto.Metric))
	}

	return topk
}
