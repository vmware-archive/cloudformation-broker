package fakes

import (
	"github.com/cf-platform-eng/cloudformation-broker/awscf"
)

type FakeStack struct {
	DescribeCalled       bool
	DescribeStackName    string
	DescribeStackDetails awscf.StackDetails
	DescribeError        error

	CreateCalled       bool
	CreateStackName    string
	CreateStackDetails awscf.StackDetails
	CreateError        error

	ModifyCalled       bool
	ModifyStackName    string
	ModifyStackDetails awscf.StackDetails
	ModifyError        error

	DeleteCalled    bool
	DeleteStackName string
	DeleteError     error
}

func (f *FakeStack) Describe(stackName string) (awscf.StackDetails, error) {
	f.DescribeCalled = true
	f.DescribeStackName = stackName

	return f.DescribeStackDetails, f.DescribeError
}

func (f *FakeStack) Create(stackName string, stackDetails awscf.StackDetails) error {
	f.CreateCalled = true
	f.CreateStackName = stackName
	f.CreateStackDetails = stackDetails

	return f.CreateError
}

func (f *FakeStack) Modify(stackName string, stackDetails awscf.StackDetails) error {
	f.ModifyCalled = true
	f.ModifyStackName = stackName
	f.ModifyStackDetails = stackDetails

	return f.ModifyError
}

func (f *FakeStack) Delete(stackName string) error {
	f.DeleteCalled = true
	f.DeleteStackName = stackName

	return f.DeleteError
}
