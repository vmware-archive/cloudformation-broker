package awscf

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-golang/lager"
)

type CloudFormationStack struct {
	region string
	cfsvc  *cloudformation.CloudFormation
	logger lager.Logger
}

func NewCloudFormationStack(
	region string,
	cfsvc *cloudformation.CloudFormation,
	logger lager.Logger,
) *CloudFormationStack {
	return &CloudFormationStack{
		region: region,
		cfsvc:  cfsvc,
		logger: logger.Session("cloudformation-stack"),
	}
}

func (s *CloudFormationStack) Describe(stackName string) (StackDetails, error) {
	stackDetails := StackDetails{}

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	s.logger.Debug("describe-stacks", lager.Data{"input": describeStacksInput})

	stack, err := s.cfsvc.DescribeStacks(describeStacksInput)
	if err != nil {
		s.logger.Error("aws-cloudformation-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS CloudFormation returns a 400 if Stack is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return stackDetails, ErrStackDoesNotExist
				}
			}
			return stackDetails, errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return stackDetails, err
	}

	for _, stack := range stack.Stacks {
		if aws.StringValue(stack.StackName) == stackName {
			s.logger.Debug("describe-stacks", lager.Data{"stack": stack})
			return s.buildStackDetails(stack), nil
		}
	}

	return stackDetails, ErrStackDoesNotExist
}

func (s *CloudFormationStack) Create(stackName string, stackDetails StackDetails) error {
	createStackInput := s.buildCreateStackInput(stackName, stackDetails)
	s.logger.Debug("create-stack", lager.Data{"input": createStackInput})

	createStackOutput, err := s.cfsvc.CreateStack(createStackInput)
	if err != nil {
		s.logger.Error("aws-cloudformation-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("create-stack", lager.Data{"output": createStackOutput})

	return nil
}

func (s *CloudFormationStack) Modify(stackName string, stackDetails StackDetails) error {
	updateStackInput := s.buildUpdateStackInput(stackName, stackDetails)
	s.logger.Debug("update-stack", lager.Data{"input": updateStackInput})

	updateStackOutput, err := s.cfsvc.UpdateStack(updateStackInput)
	if err != nil {
		s.logger.Error("aws-cloudformation-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS CloudFormation returns a 400 if Stack is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return ErrStackDoesNotExist
				}
			}
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("update-stack", lager.Data{"output": updateStackOutput})

	return nil
}

func (s *CloudFormationStack) Delete(stackName string) error {
	deleteStackInput := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	}
	s.logger.Debug("delete-stack", lager.Data{"input": deleteStackInput})

	deleteStackOutput, err := s.cfsvc.DeleteStack(deleteStackInput)
	if err != nil {
		s.logger.Error("aws-cloudformation-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("delete-stack", lager.Data{"output": deleteStackOutput})

	return nil
}

func (s *CloudFormationStack) buildStackDetails(stack *cloudformation.Stack) StackDetails {
	stackDetails := StackDetails{
		StackName:        aws.StringValue(stack.StackName),
		Capabilities:     aws.StringValueSlice(stack.Capabilities),
		DisableRollback:  aws.BoolValue(stack.DisableRollback),
		Description:      aws.StringValue(stack.Description),
		NotificationARNs: aws.StringValueSlice(stack.NotificationARNs),
		StackID:          aws.StringValue(stack.StackId),
		StackStatus:      s.stackStatus(aws.StringValue(stack.StackStatus)),
		TimeoutInMinutes: aws.Int64Value(stack.TimeoutInMinutes),
	}

	if stack.Parameters != nil && len(stack.Parameters) > 0 {
		stackDetails.Parameters = make(map[string]string)
		for _, parameter := range stack.Parameters {
			stackDetails.Parameters[aws.StringValue(parameter.ParameterKey)] = aws.StringValue(parameter.ParameterValue)
		}
	}

	if stack.Outputs != nil && len(stack.Outputs) > 0 {
		stackDetails.Outputs = make(map[string]string)
		for _, output := range stack.Outputs {
			stackDetails.Outputs[aws.StringValue(output.OutputKey)] = aws.StringValue(output.OutputValue)
		}
	}

	return stackDetails
}

func (s *CloudFormationStack) buildCreateStackInput(stackName string, stackDetails StackDetails) *cloudformation.CreateStackInput {
	createStackInput := &cloudformation.CreateStackInput{
		StackName:   aws.String(stackName),
		TemplateURL: aws.String(stackDetails.TemplateURL),
	}

	if len(stackDetails.Capabilities) > 0 {
		createStackInput.Capabilities = aws.StringSlice(stackDetails.Capabilities)
	}

	if stackDetails.DisableRollback {
		createStackInput.DisableRollback = aws.Bool(stackDetails.DisableRollback)
	}

	if len(stackDetails.NotificationARNs) > 0 {
		createStackInput.NotificationARNs = aws.StringSlice(stackDetails.NotificationARNs)
	}

	if stackDetails.OnFailure != "" {
		createStackInput.OnFailure = aws.String(stackDetails.OnFailure)
	}

	if len(stackDetails.Parameters) > 0 {
		createStackInput.Parameters = BuilCloudFormationParameters(stackDetails.Parameters)
	}

	if len(stackDetails.ResourceTypes) > 0 {
		createStackInput.ResourceTypes = aws.StringSlice(stackDetails.ResourceTypes)
	}

	if stackDetails.StackPolicyURL != "" {
		createStackInput.StackPolicyURL = aws.String(stackDetails.StackPolicyURL)
	}

	if len(stackDetails.Tags) > 0 {
		createStackInput.Tags = BuilCloudFormationTags(stackDetails.Tags)
	}

	if stackDetails.TimeoutInMinutes > 0 {
		createStackInput.TimeoutInMinutes = aws.Int64(stackDetails.TimeoutInMinutes)
	}

	return createStackInput
}

func (s *CloudFormationStack) buildUpdateStackInput(stackName string, stackDetails StackDetails) *cloudformation.UpdateStackInput {
	updateStackInput := &cloudformation.UpdateStackInput{
		StackName:   aws.String(stackName),
		TemplateURL: aws.String(stackDetails.TemplateURL),
	}

	if len(stackDetails.Capabilities) > 0 {
		updateStackInput.Capabilities = aws.StringSlice(stackDetails.Capabilities)
	}

	if len(stackDetails.NotificationARNs) > 0 {
		updateStackInput.NotificationARNs = aws.StringSlice(stackDetails.NotificationARNs)
	}

	if len(stackDetails.Parameters) > 0 {
		updateStackInput.Parameters = BuilCloudFormationParameters(stackDetails.Parameters)
	}

	if len(stackDetails.ResourceTypes) > 0 {
		updateStackInput.ResourceTypes = aws.StringSlice(stackDetails.ResourceTypes)
	}

	if stackDetails.StackPolicyURL != "" {
		updateStackInput.StackPolicyURL = aws.String(stackDetails.StackPolicyURL)
	}

	return updateStackInput
}

func (s *CloudFormationStack) stackStatus(status string) string {
	switch status {
	case cloudformation.StackStatusCreateComplete:
		return StatusSucceeded
	case cloudformation.StackStatusDeleteComplete:
		return StatusSucceeded
	case cloudformation.StackStatusUpdateComplete:
		return StatusSucceeded
	case cloudformation.StackStatusCreateInProgress:
		return StatusInProgress
	case cloudformation.StackStatusDeleteInProgress:
		return StatusInProgress
	case cloudformation.StackStatusUpdateInProgress:
		return StatusInProgress
	case cloudformation.StackStatusRollbackInProgress:
		return StatusInProgress
	case cloudformation.StackStatusUpdateCompleteCleanupInProgress:
		return StatusInProgress
	case cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress:
		return StatusInProgress
	case cloudformation.StackStatusUpdateRollbackInProgress:
		return StatusInProgress
	default:
		return StatusFailed
	}
}
