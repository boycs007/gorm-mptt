package mptt

import "errors"

var (
	UnsupportedPositionError = errors.New("unsupported position error")
	ModelTypeError           = errors.New("tree node data should be a pointer")
)
