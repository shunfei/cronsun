package models

import "errors"

var (
	ErrNotFound        = errors.New("Record not found.")
	ErrValueMayChanged = errors.New("The value has been changed by others on this time.")

	ErrEmptyJobName        = errors.New("Name of job is empty.")
	ErrEmptyJobCommand     = errors.New("Command of job is empty.")
	ErrIllegalJobId        = errors.New("Invalid id that includes illegal characters such as '/'.")
	ErrIllegalJobGroupName = errors.New("Invalid job group name that includes illegal characters such as '/'.")

	ErrEmptyNodeGroupName = errors.New("Name of node group is empty.")
	ErrIllegalNodeGroupId = errors.New("Invalid node group id that includes illegal characters such as '/'.")

	InvalidJobErr = errors.New("invalid job")
)
