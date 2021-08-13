package utils

import "github.com/dustin/go-humanize"

type Bytes uint64

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

func (bt *Bytes) Bytes() uint64 {
	return uint64(*bt)
}
