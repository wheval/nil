package storage

import "errors"

var (
	ErrTaskAlreadyExists   = errors.New("task with a given identifier already exists")
	ErrSerializationFailed = errors.New("failed to serialize/deserialize object")
	errNilTaskEntry        = errors.New("task entry cannot be nil")
)
