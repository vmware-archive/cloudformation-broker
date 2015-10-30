package cfbroker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/frodenas/brokerapi"
	"github.com/mitchellh/mapstructure"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/cloudformation-broker/awscf"
)

const instanceIDLogKey = "instance-id"
const bindingIDLogKey = "binding-id"
const detailsLogKey = "details"
const acceptsIncompleteLogKey = "acceptsIncomplete"

type ProvisionParameters map[string]string

type UpdateParameters map[string]string

type CloudFormationBroker struct {
	cloudformationPrefix         string
	allowUserProvisionParameters bool
	allowUserUpdateParameters    bool
	catalog                      Catalog
	stack                        awscf.Stack
	logger                       lager.Logger
}

func New(
	config Config,
	stack awscf.Stack,
	logger lager.Logger,
) *CloudFormationBroker {
	return &CloudFormationBroker{
		cloudformationPrefix:         config.CloudFormationPrefix,
		allowUserProvisionParameters: config.AllowUserProvisionParameters,
		allowUserUpdateParameters:    config.AllowUserUpdateParameters,
		catalog:                      config.Catalog,
		stack:                        stack,
		logger:                       logger.Session("broker"),
	}
}

func (b *CloudFormationBroker) Services() brokerapi.CatalogResponse {
	catalogResponse := brokerapi.CatalogResponse{}

	brokerCatalog, err := json.Marshal(b.catalog)
	if err != nil {
		b.logger.Error("marshal-error", err)
		return catalogResponse
	}

	apiCatalog := brokerapi.Catalog{}
	if err = json.Unmarshal(brokerCatalog, &apiCatalog); err != nil {
		b.logger.Error("unmarshal-error", err)
		return catalogResponse
	}

	catalogResponse.Services = apiCatalog.Services

	return catalogResponse
}

func (b *CloudFormationBroker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (brokerapi.ProvisioningResponse, bool, error) {
	b.logger.Debug("provision", lager.Data{
		instanceIDLogKey:        instanceID,
		detailsLogKey:           details,
		acceptsIncompleteLogKey: acceptsIncomplete,
	})

	provisioningResponse := brokerapi.ProvisioningResponse{}

	if !acceptsIncomplete {
		return provisioningResponse, true, brokerapi.ErrAsyncRequired
	}

	provisionParameters := ProvisionParameters{}
	if b.allowUserProvisionParameters {
		if err := mapstructure.Decode(details.Parameters, &provisionParameters); err != nil {
			return provisioningResponse, true, err
		}
	}

	servicePlan, ok := b.catalog.FindServicePlan(details.PlanID)
	if !ok {
		return provisioningResponse, true, fmt.Errorf("Service Plan '%s' not found", details.PlanID)
	}

	createStackDetails := b.createStackDetails(instanceID, servicePlan, provisionParameters, details)
	if err := b.stack.Create(b.stackName(instanceID), *createStackDetails); err != nil {
		return provisioningResponse, true, err
	}

	return provisioningResponse, true, nil
}

func (b *CloudFormationBroker) Update(instanceID string, details brokerapi.UpdateDetails, acceptsIncomplete bool) (bool, error) {
	b.logger.Debug("update", lager.Data{
		instanceIDLogKey:        instanceID,
		detailsLogKey:           details,
		acceptsIncompleteLogKey: acceptsIncomplete,
	})

	if !acceptsIncomplete {
		return true, brokerapi.ErrAsyncRequired
	}

	updateParameters := UpdateParameters{}
	if b.allowUserUpdateParameters {
		if err := mapstructure.Decode(details.Parameters, &updateParameters); err != nil {
			return true, err
		}
	}

	service, ok := b.catalog.FindService(details.ServiceID)
	if !ok {
		return true, fmt.Errorf("Service '%s' not found", details.ServiceID)
	}

	if !service.PlanUpdateable {
		return true, brokerapi.ErrInstanceNotUpdateable
	}

	servicePlan, ok := b.catalog.FindServicePlan(details.PlanID)
	if !ok {
		return true, fmt.Errorf("Service Plan '%s' not found", details.PlanID)
	}

	modifyStackDetails := b.modifyStackDetails(instanceID, servicePlan, updateParameters, details)
	if err := b.stack.Modify(b.stackName(instanceID), *modifyStackDetails); err != nil {
		if err == awscf.ErrStackDoesNotExist {
			return true, brokerapi.ErrInstanceDoesNotExist
		}
		return true, err
	}

	return true, nil
}

func (b *CloudFormationBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, acceptsIncomplete bool) (bool, error) {
	b.logger.Debug("deprovision", lager.Data{
		instanceIDLogKey:        instanceID,
		detailsLogKey:           details,
		acceptsIncompleteLogKey: acceptsIncomplete,
	})

	if !acceptsIncomplete {
		return true, brokerapi.ErrAsyncRequired
	}

	if err := b.stack.Delete(b.stackName(instanceID)); err != nil {
		if err == awscf.ErrStackDoesNotExist {
			return true, brokerapi.ErrInstanceDoesNotExist
		}
		return true, err
	}

	return true, nil
}

func (b *CloudFormationBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.BindingResponse, error) {
	b.logger.Debug("bind", lager.Data{
		instanceIDLogKey: instanceID,
		bindingIDLogKey:  bindingID,
		detailsLogKey:    details,
	})

	bindingResponse := brokerapi.BindingResponse{}

	service, ok := b.catalog.FindService(details.ServiceID)
	if !ok {
		return bindingResponse, fmt.Errorf("Service '%s' not found", details.ServiceID)
	}

	if !service.Bindable {
		return bindingResponse, brokerapi.ErrInstanceNotBindable
	}

	stackDetails, err := b.stack.Describe(b.stackName(instanceID))
	if err != nil {
		if err == awscf.ErrStackDoesNotExist {
			return bindingResponse, brokerapi.ErrInstanceDoesNotExist
		}
		return bindingResponse, err
	}

	credentials := make(map[string]string)
	for key, value := range stackDetails.Outputs {
		credentials[key] = value
	}
	bindingResponse.Credentials = credentials

	return bindingResponse, nil
}

func (b *CloudFormationBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	b.logger.Debug("unbind", lager.Data{
		instanceIDLogKey: instanceID,
		bindingIDLogKey:  bindingID,
		detailsLogKey:    details,
	})

	return nil
}

func (b *CloudFormationBroker) LastOperation(instanceID string) (brokerapi.LastOperationResponse, error) {
	b.logger.Debug("last-operation", lager.Data{
		instanceIDLogKey: instanceID,
	})

	lastOperationResponse := brokerapi.LastOperationResponse{State: brokerapi.LastOperationFailed}

	stackDetails, err := b.stack.Describe(b.stackName(instanceID))
	if err != nil {
		if err == awscf.ErrStackDoesNotExist {
			return lastOperationResponse, brokerapi.ErrInstanceDoesNotExist
		}
		return lastOperationResponse, err
	}

	lastOperationResponse.Description = fmt.Sprintf("Stack '%s' status is '%s'", b.stackName(instanceID), stackDetails.StackStatus)

	switch stackDetails.StackStatus {
	case awscf.StatusSucceeded:
		lastOperationResponse.State = brokerapi.LastOperationSucceeded
	case awscf.StatusInProgress:
		lastOperationResponse.State = brokerapi.LastOperationInProgress
	default:
		lastOperationResponse.State = brokerapi.LastOperationFailed
	}

	return lastOperationResponse, nil
}

func (b *CloudFormationBroker) stackName(instanceID string) string {
	return fmt.Sprintf("%s-%s", b.cloudformationPrefix, instanceID)
}

func (b *CloudFormationBroker) createStackDetails(instanceID string, servicePlan ServicePlan, provisionParameters ProvisionParameters, details brokerapi.ProvisionDetails) *awscf.StackDetails {
	stackDetails := b.stackDetailsFromPlan(servicePlan)

	if stackDetails.Parameters == nil {
		stackDetails.Parameters = make(map[string]string)
	}

	for key, value := range provisionParameters {
		stackDetails.Parameters[key] = value
	}

	stackDetails.Tags = b.stackTags("Created", details.ServiceID, details.PlanID, details.OrganizationGUID, details.SpaceGUID)

	return stackDetails
}

func (b *CloudFormationBroker) modifyStackDetails(instanceID string, servicePlan ServicePlan, updateParameters UpdateParameters, details brokerapi.UpdateDetails) *awscf.StackDetails {
	stackDetails := b.stackDetailsFromPlan(servicePlan)

	if stackDetails.Parameters == nil {
		stackDetails.Parameters = make(map[string]string)
	}

	for key, value := range updateParameters {
		stackDetails.Parameters[key] = value
	}

	return stackDetails
}

func (b *CloudFormationBroker) stackDetailsFromPlan(servicePlan ServicePlan) *awscf.StackDetails {
	stackDetails := &awscf.StackDetails{
		Capabilities:     servicePlan.CloudFormationProperties.Capabilities,
		DisableRollback:  servicePlan.CloudFormationProperties.DisableRollback,
		NotificationARNs: servicePlan.CloudFormationProperties.NotificationARNs,
		OnFailure:        servicePlan.CloudFormationProperties.OnFailure,
		Parameters:       servicePlan.CloudFormationProperties.Parameters,
		ResourceTypes:    servicePlan.CloudFormationProperties.ResourceTypes,
		StackPolicyURL:   servicePlan.CloudFormationProperties.StackPolicyURL,
		TemplateURL:      servicePlan.CloudFormationProperties.TemplateURL,
		TimeoutInMinutes: servicePlan.CloudFormationProperties.TimeoutInMinutes,
	}

	return stackDetails
}

func (b *CloudFormationBroker) stackTags(action, serviceID, planID, organizationID, spaceID string) map[string]string {
	tags := make(map[string]string)

	tags["Owner"] = "Cloud Foundry"

	tags[action+" by"] = "AWS CloudFormation Service Broker"

	tags[action+" at"] = time.Now().Format(time.RFC822Z)

	if serviceID != "" {
		tags["Service ID"] = serviceID
	}

	if planID != "" {
		tags["Plan ID"] = planID
	}

	if organizationID != "" {
		tags["Organization ID"] = organizationID
	}

	if spaceID != "" {
		tags["Space ID"] = spaceID
	}

	return tags
}
