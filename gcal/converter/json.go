package converter

import (
	"bytes"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

type JSONConverter struct{}

var jsoner = jsoniter.ConfigCompatibleWithStandardLibrary

// Pack the data package
func (*JSONConverter) Pack(data interface{}) ([]byte, error) {
	switch data.(type) {
	case string:
		res, ok := data.(string)
		if !ok {
			return nil, errors.New("json pack error: pack body to string")
		}
		return []byte(res), nil
	case []byte:
		res, ok := data.([]byte)
		if !ok {
			return nil, errors.New("json pack error: pack body to []byte")
		}
		return res, nil
	default:
		res, err := jsoner.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("json pack error: %s", err.Error())
		}
		return res, nil
	}
}

// UnPack the data package
func (*JSONConverter) UnPack(data interface{}, rsp interface{}) error {
	dec := jsoner.NewDecoder(bytes.NewReader(data.([]byte)))
	dec.UseNumber()
	err := dec.Decode(rsp)
	if err != nil {
		return fmt.Errorf("json unpack error: %s", err.Error())
	}
	return nil
}
