package utils

import "github.com/dustin/go-humanize"

// Bytes bytes for yaml decode
type Bytes uint64

// UnmarshalYAML custom decode bytes
func (bt *Bytes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	n, err := humanize.ParseBytes(str)
	if err != nil {
		return err
	}
	*bt = Bytes(n)
	return nil
}

// Bytes bytes count
func (bt *Bytes) Bytes() uint64 {
	return uint64(*bt)
}
