package mptt

import "errors"

var (
    UnsupportedPositionError = errors.New("unsupported position error")
    NodeNotExistsError       = errors.New("node not exists error")
    ModelTypeError           = errors.New("tree node data should be a pointer")
)
