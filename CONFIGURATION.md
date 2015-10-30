# Configuration

A sample configuration file can be found at [config-sample.json](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/config-sample.json).

Also, a sample AWS CloudFormation template file can be found at [sample-s3-cftemplate.json](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/sample-s3-cftemplate.json).

## General Configuration

| Option                | Required | Type   | Description
|:----------------------|:--------:|:------ |:-----------
| log_level             | Y        | String | Broker Log Level (DEBUG, INFO, ERROR, FATAL)
| username              | Y        | String | Broker Auth Username
| password              | Y        | String | Broker Auth Password
| cloudformation_config | Y        | Hash   | [CloudFormation Broker configuration](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md#cloudformation-broker-configuration)

## CloudFormation Broker Configuration

| Option                         | Required | Type    | Description
|:-------------------------------|:--------:|:------- |:-----------
| region                         | Y        | String  | CloudFormation Region
| cloudformation_prefix          | Y        | String  | Prefix to add to CloudFormation Stack Names
| allow_user_provision_parameters| N        | Boolean | Allow users to send arbitrary parameters on provision calls (defaults to `false`)
| allow_user_update_parameters   | N        | Boolean | Allow users to send arbitrary parameters on update calls (defaults to `false`)
| catalog                        | Y        | Hash    | [CloudFormation Broker catalog](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md#cloudformation-broker-catalog)

## CloudFormation Broker catalog

Please refer to the [Catalog Documentation](https://docs.cloudfoundry.org/services/api.html#catalog-mgmt) for more details about these properties.

### Catalog

| Option   | Required | Type      | Description
|:---------|:--------:|:--------- |:-----------
| services | N        | []Service | A list of [Services](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md#service)

### Service

| Option                        | Required | Type          | Description
|:------------------------------|:--------:|:------------- |:-----------
| id                            | Y        | String        | An identifier used to correlate this service in future requests to the catalog
| name                          | Y        | String        | The CLI-friendly name of the service that will appear in the catalog. All lowercase, no spaces
| description                   | Y        | String        | A short description of the service that will appear in the catalog
| bindable                      | N        | Boolean       | Whether the service can be bound to applications
| tags                          | N        | []String      | A list of service tags
| metadata.displayName          | N        | String        | The name of the service to be displayed in graphical clients
| metadata.imageUrl             | N        | String        | The URL to an image
| metadata.longDescription      | N        | String        | Long description
| metadata.providerDisplayName  | N        | String        | The name of the upstream entity providing the actual service
| metadata.documentationUrl     | N        | String        | Link to documentation page for service
| metadata.supportUrl           | N        | String        | Link to support for the service
| requires                      | N        | []String      | A list of permissions that the user would have to give the service, if they provision it (only `syslog_drain` is supported)
| plan_updateable               | N        | Boolean       | Whether the service supports upgrade/downgrade for some plans
| plans                         | N        | []ServicePlan | A list of [Plans](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md#service-plan) for this service
| dashboard_client.id           | N        | String        | The id of the Oauth2 client that the service intends to use
| dashboard_client.secret       | N        | String        | A secret for the dashboard client
| dashboard_client.redirect_uri | N        | String        | A domain for the service dashboard that will be whitelisted by the UAA to enable SSO

### Service Plan

| Option                    | Required | Type                     | Description
|:--------------------------|:--------:|:------------------------ |:-----------
| id                        | Y        | String                   | An identifier used to correlate this plan in future requests to the catalog
| name                      | Y        | String                   | The CLI-friendly name of the plan that will appear in the catalog. All lowercase, no spaces
| description               | Y        | String                   | A short description of the plan that will appear in the catalog
| metadata.bullets          | N        | []String                 | Features of this plan, to be displayed in a bulleted-list
| metadata.costs            | N        | Cost Object              | An array-of-objects that describes the costs of a service, in what currency, and the unit of measure
| metadata.displayName      | N        | String                   | Name of the plan to be display in graphical clients
| free                      | N        | Boolean                  | This field allows the plan to be limited by the non_basic_services_allowed field in a Cloud Foundry Quota
| cloudformation_properties | Y        | CloudFormationProperties | [CloudFormation Properties](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md#cloudformation-properties)

## CloudFormation Properties

Please refer to the [Amazon CloudFormation Documentation](https://aws.amazon.com/documentation/cloudformation/) for more details about these properties.

| Option             | Required | Type          | Description
|:-------------------|:--------:|:------------- |:-----------
| capabilities       | N        | Array<String> | A list of capabilities that you must specify before AWS CloudFormation can create or update certain stacks
| disable_rollback   | N        | Boolean       | Set to true to disable rollback of the stack if stack creation failed
| notification_arns  | N        | Array<String> | The Simple Notification Service (SNS) topic ARNs to publish stack related events
| on_failure         | N        | String        | Determines what action will be taken if stack creation fails (`DO_NOTHING`, `ROLLBACK` or `DELETE`)
| parameters         | N        | Hash          | A list of Parameters that specify input parameters for the stack
| resource_types     | N        | Array<String> | The template resource types that you have permissions to work with for this create stack action
| stack_policy_url   | N        | String        | Location of a file containing the stack policy
| template_url       | Y        | String        | Location of file containing the template body
| timeout_in_minutes | N        | Integer       | The amount of time that can pass before the stack status becomes failed



