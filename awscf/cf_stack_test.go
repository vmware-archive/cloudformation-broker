package awscf_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/cloudformation-broker/awscf"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("CloudFormation Stack", func() {
	var (
		region    string
		stackName string

		cfsvc  *cloudformation.CloudFormation
		cfCall func(r *request.Request)

		testSink *lagertest.TestSink
		logger   lager.Logger

		stack Stack
	)

	BeforeEach(func() {
		region = "cloudformation-region"
		stackName = "cloudformation-stack"
	})

	JustBeforeEach(func() {
		cfsvc = cloudformation.New(nil)

		logger = lager.NewLogger("cfstack_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		stack = NewCloudFormationStack(region, cfsvc, logger)
	})

	var _ = Describe("Describe", func() {
		var (
			properStackDetails StackDetails

			describeStacks []*cloudformation.Stack
			describeStack  *cloudformation.Stack

			describeStacksInput *cloudformation.DescribeStacksInput
			describeStacksError error
		)

		BeforeEach(func() {
			properStackDetails = StackDetails{
				StackName:        stackName,
				Capabilities:     []string{"test-capability"},
				DisableRollback:  true,
				Description:      "test-stack-description",
				NotificationARNs: []string{"test-notification-arn"},
				StackID:          "test-stack-id",
				StackStatus:      StatusSucceeded,
				TimeoutInMinutes: int64(1),
			}

			describeStack = &cloudformation.Stack{
				StackName:        aws.String(stackName),
				Capabilities:     aws.StringSlice([]string{"test-capability"}),
				DisableRollback:  aws.Bool(true),
				Description:      aws.String("test-stack-description"),
				NotificationARNs: aws.StringSlice([]string{"test-notification-arn"}),
				StackId:          aws.String("test-stack-id"),
				StackStatus:      aws.String(cloudformation.StackStatusCreateComplete),
				TimeoutInMinutes: aws.Int64(int64(1)),
			}

			describeStacksInput = &cloudformation.DescribeStacksInput{
				StackName: aws.String(stackName),
			}
			describeStacksError = nil
		})

		JustBeforeEach(func() {
			describeStacks = []*cloudformation.Stack{describeStack}

			cfsvc.Handlers.Clear()

			cfCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("DescribeStacks"))
				Expect(r.Params).To(BeAssignableToTypeOf(&cloudformation.DescribeStacksInput{}))
				Expect(r.Params).To(Equal(describeStacksInput))
				data := r.Data.(*cloudformation.DescribeStacksOutput)
				data.Stacks = describeStacks
				r.Error = describeStacksError
			}
			cfsvc.Handlers.Send.PushBack(cfCall)
		})

		It("returns the proper Stack Details", func() {
			stackDetails, err := stack.Describe(stackName)
			Expect(err).ToNot(HaveOccurred())
			Expect(stackDetails).To(Equal(properStackDetails))
		})

		Context("when the Stack has Parameters", func() {
			BeforeEach(func() {
				describeStack.Parameters = []*cloudformation.Parameter{
					&cloudformation.Parameter{
						ParameterKey:   aws.String("test-parameter-key-1"),
						ParameterValue: aws.String("test-parameter-value-1"),
					},
				}

				properStackDetails.Parameters = map[string]string{
					"test-parameter-key-1": "test-parameter-value-1",
				}
			})

			It("returns the proper Stack Details", func() {
				stackDetails, err := stack.Describe(stackName)
				Expect(err).ToNot(HaveOccurred())
				Expect(stackDetails).To(Equal(properStackDetails))
			})
		})

		Context("when the Stack has Outputs", func() {
			BeforeEach(func() {
				describeStack.Outputs = []*cloudformation.Output{
					&cloudformation.Output{
						OutputKey:   aws.String("test-output-key-1"),
						OutputValue: aws.String("test-output-value-1"),
					},
				}

				properStackDetails.Outputs = map[string]string{
					"test-output-key-1": "test-output-value-1",
				}
			})

			It("returns the proper Stack Details", func() {
				stackDetails, err := stack.Describe(stackName)
				Expect(err).ToNot(HaveOccurred())
				Expect(stackDetails).To(Equal(properStackDetails))
			})
		})

		Context("when the Stack Status is in progress", func() {
			BeforeEach(func() {
				describeStack.StackStatus = aws.String(cloudformation.StackStatusCreateInProgress)
				properStackDetails.StackStatus = StatusInProgress
			})

			It("returns the proper Stack Details", func() {
				stackDetails, err := stack.Describe(stackName)
				Expect(err).ToNot(HaveOccurred())
				Expect(stackDetails).To(Equal(properStackDetails))
			})
		})

		Context("when the Stack Status is failed", func() {
			BeforeEach(func() {
				describeStack.StackStatus = aws.String(cloudformation.StackStatusCreateFailed)
				properStackDetails.StackStatus = StatusFailed
			})

			It("returns the proper Stack Details", func() {
				stackDetails, err := stack.Describe(stackName)
				Expect(err).ToNot(HaveOccurred())
				Expect(stackDetails).To(Equal(properStackDetails))
			})
		})

		Context("when the Stack does not exists", func() {
			JustBeforeEach(func() {
				describeStacksInput = &cloudformation.DescribeStacksInput{
					StackName: aws.String("unknown"),
				}
			})

			It("returns the proper error", func() {
				_, err := stack.Describe("unknown")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrStackDoesNotExist))
			})
		})

		Context("when describing the Stack fails", func() {
			BeforeEach(func() {
				describeStacksError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := stack.Describe(stackName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					describeStacksError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					_, err := stack.Describe(stackName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 400 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					describeStacksError = awserr.NewRequestFailure(awsError, 400, "request-id")
				})

				It("returns the proper error", func() {
					_, err := stack.Describe(stackName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrStackDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Create", func() {
		var (
			stackDetails StackDetails

			createStackInput *cloudformation.CreateStackInput
			createStackError error
		)

		BeforeEach(func() {
			stackDetails = StackDetails{
				StackName:   stackName,
				TemplateURL: "test-template-url",
			}

			createStackInput = &cloudformation.CreateStackInput{
				StackName:   aws.String(stackName),
				TemplateURL: aws.String("test-template-url"),
			}
			createStackError = nil
		})

		JustBeforeEach(func() {
			cfsvc.Handlers.Clear()

			cfCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("CreateStack"))
				Expect(r.Params).To(BeAssignableToTypeOf(&cloudformation.CreateStackInput{}))
				Expect(r.Params).To(Equal(createStackInput))
				r.Error = createStackError
			}
			cfsvc.Handlers.Send.PushBack(cfCall)
		})

		It("creates the Stack", func() {
			err := stack.Create(stackName, stackDetails)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has Capabilities", func() {
			BeforeEach(func() {
				stackDetails.Capabilities = []string{"test-capability"}
				createStackInput.Capabilities = aws.StringSlice([]string{"test-capability"})
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has DisableRollback", func() {
			BeforeEach(func() {
				stackDetails.DisableRollback = true
				createStackInput.DisableRollback = aws.Bool(true)
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has NotificationARNs", func() {
			BeforeEach(func() {
				stackDetails.NotificationARNs = []string{"test-notification-arn"}
				createStackInput.NotificationARNs = aws.StringSlice([]string{"test-notification-arn"})
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has OnFailure", func() {
			BeforeEach(func() {
				stackDetails.OnFailure = "test-on-failure"
				createStackInput.OnFailure = aws.String("test-on-failure")
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Parameters", func() {
			BeforeEach(func() {
				stackDetails.Parameters = map[string]string{"test-parameter-key-1": "test-parameter-value-1"}
				createStackInput.Parameters = []*cloudformation.Parameter{
					&cloudformation.Parameter{ParameterKey: aws.String("test-parameter-key-1"), ParameterValue: aws.String("test-parameter-value-1")},
				}
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has ResourceTypes", func() {
			BeforeEach(func() {
				stackDetails.ResourceTypes = []string{"test-resource-type"}
				createStackInput.ResourceTypes = aws.StringSlice([]string{"test-resource-type"})
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has StackPolicyURL", func() {
			BeforeEach(func() {
				stackDetails.StackPolicyURL = "test-stack-policy-url"
				createStackInput.StackPolicyURL = aws.String("test-stack-policy-url")
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Tags", func() {
			BeforeEach(func() {
				stackDetails.Tags = map[string]string{"test-tag-key-1": "test-tag-value-1"}
				createStackInput.Tags = []*cloudformation.Tag{
					&cloudformation.Tag{Key: aws.String("test-tag-key-1"), Value: aws.String("test-tag-value-1")},
				}
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has TimeoutInMinutes", func() {
			BeforeEach(func() {
				stackDetails.TimeoutInMinutes = int64(1)
				createStackInput.TimeoutInMinutes = aws.Int64(int64(1))
			})

			It("makes the proper call", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when creating the Stack fails", func() {
			BeforeEach(func() {
				createStackError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := stack.Create(stackName, stackDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					createStackError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := stack.Create(stackName, stackDetails)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})

	var _ = Describe("Modify", func() {
		var (
			stackDetails StackDetails

			updateStackInput *cloudformation.UpdateStackInput
			updateStackError error
		)

		BeforeEach(func() {
			stackDetails = StackDetails{
				StackName:   stackName,
				TemplateURL: "test-template-url",
			}

			updateStackInput = &cloudformation.UpdateStackInput{
				StackName:   aws.String(stackName),
				TemplateURL: aws.String("test-template-url"),
			}
			updateStackError = nil
		})

		JustBeforeEach(func() {
			cfsvc.Handlers.Clear()

			cfCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("UpdateStack"))
				Expect(r.Params).To(BeAssignableToTypeOf(&cloudformation.UpdateStackInput{}))
				Expect(r.Params).To(Equal(updateStackInput))
				r.Error = updateStackError
			}
			cfsvc.Handlers.Send.PushBack(cfCall)
		})

		It("creates the Stack", func() {
			err := stack.Modify(stackName, stackDetails)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has Capabilities", func() {
			BeforeEach(func() {
				stackDetails.Capabilities = []string{"test-capability"}
				updateStackInput.Capabilities = aws.StringSlice([]string{"test-capability"})
			})

			It("makes the proper call", func() {
				err := stack.Modify(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has NotificationARNs", func() {
			BeforeEach(func() {
				stackDetails.NotificationARNs = []string{"test-notification-arn"}
				updateStackInput.NotificationARNs = aws.StringSlice([]string{"test-notification-arn"})
			})

			It("makes the proper call", func() {
				err := stack.Modify(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Parameters", func() {
			BeforeEach(func() {
				stackDetails.Parameters = map[string]string{"test-parameter-key-1": "test-parameter-value-1"}
				updateStackInput.Parameters = []*cloudformation.Parameter{
					&cloudformation.Parameter{ParameterKey: aws.String("test-parameter-key-1"), ParameterValue: aws.String("test-parameter-value-1")},
				}
			})

			It("makes the proper call", func() {
				err := stack.Modify(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has ResourceTypes", func() {
			BeforeEach(func() {
				stackDetails.ResourceTypes = []string{"test-resource-type"}
				updateStackInput.ResourceTypes = aws.StringSlice([]string{"test-resource-type"})
			})

			It("makes the proper call", func() {
				err := stack.Modify(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has StackPolicyURL", func() {
			BeforeEach(func() {
				stackDetails.StackPolicyURL = "test-stack-policy-url"
				updateStackInput.StackPolicyURL = aws.String("test-stack-policy-url")
			})

			It("makes the proper call", func() {
				err := stack.Modify(stackName, stackDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when updating the Stack fails", func() {
			BeforeEach(func() {
				updateStackError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := stack.Modify(stackName, stackDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					updateStackError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := stack.Modify(stackName, stackDetails)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 400 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					updateStackError = awserr.NewRequestFailure(awsError, 400, "request-id")
				})

				It("returns the proper error", func() {
					err := stack.Modify(stackName, stackDetails)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrStackDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Delete", func() {
		var (
			deleteStackInput *cloudformation.DeleteStackInput
			deleteStackError error
		)

		BeforeEach(func() {
			deleteStackInput = &cloudformation.DeleteStackInput{
				StackName: aws.String(stackName),
			}
			deleteStackError = nil
		})

		JustBeforeEach(func() {
			cfsvc.Handlers.Clear()

			cfCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("DeleteStack"))
				Expect(r.Params).To(BeAssignableToTypeOf(&cloudformation.DeleteStackInput{}))
				Expect(r.Params).To(Equal(deleteStackInput))
				r.Error = deleteStackError
			}
			cfsvc.Handlers.Send.PushBack(cfCall)
		})

		It("does not return error", func() {
			err := stack.Delete(stackName)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when deleting the Stack fails", func() {
			BeforeEach(func() {
				deleteStackError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := stack.Delete(stackName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					deleteStackError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := stack.Delete(stackName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})
})
