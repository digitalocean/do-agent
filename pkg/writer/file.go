package writer

import (
	"fmt"
	"io"
	"sync"

	dto "github.com/prometheus/client_model/go"
)

// File writes metrics to an io.Writer
type File struct {
	w io.Writer
	m *sync.Mutex
}

// NewFile creates a new File writer with the provided writer
func NewFile(w io.Writer) *File {
	return &File{
		w: w,
		m: new(sync.Mutex),
	}
}

// Write writes metrics to the file
func (w *File) Write(mets []*dto.MetricFamily) error {
	w.m.Lock()
	defer w.m.Unlock()
	for _, mf := range mets {
		for _, met := range mf.Metric {
			fmt.Fprintf(w.w, "[%s]: %s: %s\n", mf.GetType(), mf.GetName(), met.String())
		}
	}
	return nil
}

// Name is the name of this writer
func (w *File) Name() string {
	return "file"
}
