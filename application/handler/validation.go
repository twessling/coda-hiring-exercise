package handler

import (
	"encoding/json"
	"errors"
)

func validateJson(input []byte) error {
	if !json.Valid(input) {
		return errors.New("input not valid json")
	}
	return nil
}
