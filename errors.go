package cronsun

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

	ErrSecurityInvalidCmd  = errors.New("Security error: the suffix of script file is not on the whitelist.")
	ErrSecurityInvalidUser = errors.New("Security error: the user is not on the whitelist.")
	ErrNilRule             = errors.New("invalid job rule, empty timer.")
)
