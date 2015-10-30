package awscf

import (
	"errors"
)

const StatusInProgress = "in progress"
const StatusFailed = "failed"
const StatusSucceeded = "succeeded"

type Stack interface {
	Describe(stackName string) (StackDetails, error)
	Create(stackName string, stackDetails StackDetails) error
	Modify(stackName string, stackDetails StackDetails) error
	Delete(stackName string) error
}

type StackDetails struct {
	StackName        string
	Capabilities     []string
	DisableRollback  bool
	Description      string
	NotificationARNs []string
	OnFailure        string
	Outputs          map[string]string
	Parameters       map[string]string
	ResourceTypes    []string
	StackID          string
	StackPolicyURL   string
	StackStatus      string
	Tags             map[string]string
	TemplateURL      string
	TimeoutInMinutes int64
}

var (
	ErrStackDoesNotExist = errors.New("cloudformation stack does not exist")
)
