package stateinfo

import (
	"encoding/json"
	"fmt"
)

const maxInt = int64(^(uint64(0)) >> 1)

type StateInfoV2AndUp struct {
	Version          uint64 `json:"version"`
	TerraformVersion string `json:"terraform_version"`
	Serial           uint64 `json:"serial"`
	Lineage          string `json:"lineage"`
}

func (s StateInfoV2AndUp) SerialValueInt64() int64 {
	// OK because of the sanity check during Read
	return int64(s.Serial)
}

func Read(src []byte) (*StateInfoV2AndUp, error) {
	result := &StateInfoV2AndUp{}
	err := json.Unmarshal(src, result)

	if err != nil {
		return nil, fmt.Errorf("could not parse state contents: %w", err)
	}

	if result.Serial > uint64(maxInt) {
		// The SDK only supports up to range of int64
		return nil, fmt.Errorf("serial too high; state cannot be uploaded by SDK")
	}

	return result, nil
}
