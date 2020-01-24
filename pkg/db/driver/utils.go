package driver

import "github.com/vmihailenco/msgpack/v4"

// EncodePair encodes the specified pairs to bytes
func EncodePair(pair Pair) ([]byte, error) {
	return msgpack.Marshal(pair)
}

// DecodePair decodes the specified bytes as pair
func DecodePair(data []byte) (pair *Pair, err error) {
	if data == nil {
		return
	}
	err = msgpack.Unmarshal(data, &pair)
	return
}
