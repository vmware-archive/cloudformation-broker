package cfbroker

import (
	"errors"
	"fmt"
)

type Config struct {
	Region                       string  `json:"region"`
	CloudFormationPrefix         string  `json:"cloudformation_prefix"`
	AllowUserProvisionParameters bool    `json:"allow_user_provision_parameters"`
	AllowUserUpdateParameters    bool    `json:"allow_user_update_parameters"`
	Catalog                      Catalog `json:"catalog"`
}

func (c Config) Validate() error {
	if c.Region == "" {
		return errors.New("Must provide a non-empty Region")
	}

	if c.CloudFormationPrefix == "" {
		return errors.New("Must provide a non-empty CloudFormationPrefix")
	}

	if err := c.Catalog.Validate(); err != nil {
		return fmt.Errorf("Validating Catalog configuration: %s", err)
	}

	return nil
}
