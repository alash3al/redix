package driver

// ScanOpts scanner options
type ScanOpts struct {
	Prefix        []byte
	Offset        []byte
	Scanner       Scanner
	IncludeOffset bool
	ReverseScan   bool
}

// Scanner a function that performs the scanning/filterig
type Scanner func([]byte, []byte) bool
