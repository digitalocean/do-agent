package writer

import (
	"fmt"
	"io"
	"sync"

	"github.com/digitalocean/do-agent/pkg/aggregate"
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
func (w *File) Write(mets []aggregate.MetricWithValue) error {
	w.m.Lock()
	defer w.m.Unlock()
	for _, met := range mets {
		fmt.Fprintf(w.w, "[%s]: %v: %v\n", met.LFM["__name__"], met.LFM, met.Value)
	}
	return nil
}

// Name is the name of this writer
func (w *File) Name() string {
	return "file"
}
