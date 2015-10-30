package cfbroker

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type Catalog struct {
	Services []Service `json:"services,omitempty"`
}

type Service struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Bindable        bool             `json:"bindable,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	Metadata        *ServiceMetadata `json:"metadata,omitempty"`
	Requires        []string         `json:"requires,omitempty"`
	PlanUpdateable  bool             `json:"plan_updateable"`
	Plans           []ServicePlan    `json:"plans,omitempty"`
	DashboardClient *DashboardClient `json:"dashboard_client,omitempty"`
}

type ServiceMetadata struct {
	DisplayName         string `json:"displayName,omitempty"`
	ImageURL            string `json:"imageUrl,omitempty"`
	LongDescription     string `json:"longDescription,omitempty"`
	ProviderDisplayName string `json:"providerDisplayName,omitempty"`
	DocumentationURL    string `json:"documentationUrl,omitempty"`
	SupportURL          string `json:"supportUrl,omitempty"`
}

type ServicePlan struct {
	ID                       string                   `json:"id"`
	Name                     string                   `json:"name"`
	Description              string                   `json:"description"`
	Metadata                 *ServicePlanMetadata     `json:"metadata,omitempty"`
	Free                     bool                     `json:"free"`
	CloudFormationProperties CloudFormationProperties `json:"cloudformation_properties,omitempty"`
}

type ServicePlanMetadata struct {
	Bullets     []string `json:"bullets,omitempty"`
	Costs       []Cost   `json:"costs,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
}

type DashboardClient struct {
	ID          string `json:"id,omitempty"`
	Secret      string `json:"secret,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type Cost struct {
	Amount map[string]interface{} `json:"amount,omitempty"`
	Unit   string                 `json:"unit,omitempty"`
}

type CloudFormationProperties struct {
	Capabilities     []string          `json:"capabilities,omitempty"`
	DisableRollback  bool              `json:"disable_rollback,omitempty"`
	NotificationARNs []string          `json:"notification_arns,omitempty"`
	OnFailure        string            `json:"on_failure,omitempty"`
	Parameters       map[string]string `json:"parameters,omitempty"`
	ResourceTypes    []string          `json:"resource_types,omitempty"`
	StackPolicyURL   string            `json:"stack_policy_url,omitempty"`
	TemplateURL      string            `json:"template_url"`
	TimeoutInMinutes int64             `json:"timeout_in_minutes,omitempty"`
}

func (c Catalog) Validate() error {
	for _, service := range c.Services {
		if err := service.Validate(); err != nil {
			return fmt.Errorf("Validating Services configuration: %s", err)
		}
	}

	return nil
}

func (c Catalog) FindService(serviceID string) (service Service, found bool) {
	for _, service := range c.Services {
		if service.ID == serviceID {
			return service, true
		}
	}

	return service, false
}

func (c Catalog) FindServicePlan(planID string) (plan ServicePlan, found bool) {
	for _, service := range c.Services {
		for _, plan := range service.Plans {
			if plan.ID == planID {
				return plan, true
			}
		}
	}

	return plan, false
}

func (s Service) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("Must provide a non-empty ID (%+v)", s)
	}

	if s.Name == "" {
		return fmt.Errorf("Must provide a non-empty Name (%+v)", s)
	}

	if s.Description == "" {
		return fmt.Errorf("Must provide a non-empty Description (%+v)", s)
	}

	for _, servicePlan := range s.Plans {
		if err := servicePlan.Validate(); err != nil {
			return fmt.Errorf("Validating Plans configuration: %s", err)
		}
	}

	return nil
}

func (sp ServicePlan) Validate() error {
	if sp.ID == "" {
		return fmt.Errorf("Must provide a non-empty ID (%+v)", sp)
	}

	if sp.Name == "" {
		return fmt.Errorf("Must provide a non-empty Name (%+v)", sp)
	}

	if sp.Description == "" {
		return fmt.Errorf("Must provide a non-empty Description (%+v)", sp)
	}

	if err := sp.CloudFormationProperties.Validate(); err != nil {
		return fmt.Errorf("Validating CloudFormation Properties configuration: %s", err)
	}

	return nil
}

func (cp CloudFormationProperties) Validate() error {
	if cp.OnFailure != "" {
		switch strings.ToUpper(cp.OnFailure) {
		case cloudformation.OnFailureDelete, cloudformation.OnFailureDoNothing, cloudformation.OnFailureRollback:
		default:
			return fmt.Errorf("OnFailure '%s' not supported", cp.OnFailure)
		}
	}

	if cp.TemplateURL == "" {
		return fmt.Errorf("Must provide a non-empty TemplateURL (%+v)", cp)
	}

	return nil
}
