package driver

// ScanOpts scanner options
type ScanOpts struct {
	Prefix        []byte
	Offset        []byte
	Filter        ScanFilter
	IncludeOffset bool
	ReverseScan   bool
}

// ScanFilter a function that performs the scanning/filterig
type ScanFilter func([]byte, []byte) bool
