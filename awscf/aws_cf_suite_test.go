package awscf_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAWSCloudFormation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS CloudFormation Suite")
}
