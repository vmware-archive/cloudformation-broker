package awscf_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/cloudformation-broker/awscf"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var _ = Describe("AWS CloudFormation Utils", func() {
	var _ = Describe("BuilCloudFormationTags", func() {
		var (
			tags         map[string]string
			properCFTags []*cloudformation.Tag
		)

		BeforeEach(func() {
			tags = map[string]string{"test-tag-key-1": "test-tag-value-1"}
			properCFTags = []*cloudformation.Tag{
				&cloudformation.Tag{
					Key:   aws.String("test-tag-key-1"),
					Value: aws.String("test-tag-value-1"),
				},
			}
		})

		It("returns the proper CloudFormation Tags", func() {
			cfTags := BuilCloudFormationTags(tags)
			Expect(cfTags).To(Equal(properCFTags))
		})
	})

	var _ = Describe("BuilCloudFormationParameters", func() {
		var (
			parameters         map[string]string
			properCFParameters []*cloudformation.Parameter
		)

		BeforeEach(func() {
			parameters = map[string]string{"test-parameter-key-1": "test-parameter-value-1"}
			properCFParameters = []*cloudformation.Parameter{
				&cloudformation.Parameter{
					ParameterKey:   aws.String("test-parameter-key-1"),
					ParameterValue: aws.String("test-parameter-value-1"),
				},
			}
		})

		It("returns the proper CloudFormation Parameters", func() {
			cfParameters := BuilCloudFormationParameters(parameters)
			Expect(cfParameters).To(Equal(properCFParameters))
		})
	})
})
