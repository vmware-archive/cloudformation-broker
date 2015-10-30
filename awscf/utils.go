package awscf

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func BuilCloudFormationTags(tags map[string]string) []*cloudformation.Tag {
	var cfTags []*cloudformation.Tag

	for key, value := range tags {
		cfTags = append(cfTags, &cloudformation.Tag{Key: aws.String(key), Value: aws.String(value)})
	}

	return cfTags
}

func BuilCloudFormationParameters(tags map[string]string) []*cloudformation.Parameter {
	var cfParameters []*cloudformation.Parameter

	for key, value := range tags {
		cfParameters = append(cfParameters, &cloudformation.Parameter{ParameterKey: aws.String(key), ParameterValue: aws.String(value)})
	}

	return cfParameters
}
