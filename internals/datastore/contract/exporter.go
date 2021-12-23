package contract

import "io"

// Exporter export related actions
type Exporter interface {
	Export(io.Writer) error
}
