package models

import "errors"

var (
	ErrNotFound        = errors.New("Record not found.")
	ErrValueMayChanged = errors.New("The value has been changed by others on this time.")

	ErrEmptyJobName    = errors.New("Name of job is empty.")
	ErrEmptyJobCommand = errors.New("Command of job is empty.")

	ErrEmptyNodeGroupName = errors.New("Name of node group is empty.")
)
