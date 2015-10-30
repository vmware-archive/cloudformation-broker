package cfbroker_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/cloudformation-broker/cfbroker"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/cf-platform-eng/cloudformation-broker/awscf"
	cffake "github.com/cf-platform-eng/cloudformation-broker/awscf/fakes"
)

var _ = Describe("CloudFormation Broker", func() {
	var (
		cfProperties1 CloudFormationProperties
		cfProperties2 CloudFormationProperties
		plan1         ServicePlan
		plan2         ServicePlan
		service1      Service
		service2      Service
		catalog       Catalog

		config Config

		stack *cffake.FakeStack

		testSink *lagertest.TestSink
		logger   lager.Logger

		cfBroker *CloudFormationBroker

		allowUserProvisionParameters bool
		allowUserUpdateParameters    bool
		serviceBindable              bool
		planUpdateable               bool

		instanceID = "instance-id"
		bindingID  = "binding-id"
		stackName  = "cf-instance-id"
	)

	BeforeEach(func() {
		allowUserProvisionParameters = true
		allowUserUpdateParameters = true
		serviceBindable = true
		planUpdateable = true

		stack = &cffake.FakeStack{}

		cfProperties1 = CloudFormationProperties{}
		cfProperties2 = CloudFormationProperties{}
	})

	JustBeforeEach(func() {
		plan1 = ServicePlan{
			ID:                       "Plan-1",
			Name:                     "Plan 1",
			Description:              "This is the Plan 1",
			CloudFormationProperties: cfProperties1,
		}
		plan2 = ServicePlan{
			ID:                       "Plan-2",
			Name:                     "Plan 2",
			Description:              "This is the Plan 2",
			CloudFormationProperties: cfProperties2,
		}

		service1 = Service{
			ID:             "Service-1",
			Name:           "Service 1",
			Description:    "This is the Service 1",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []ServicePlan{plan1},
		}
		service2 = Service{
			ID:             "Service-2",
			Name:           "Service 2",
			Description:    "This is the Service 2",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []ServicePlan{plan2},
		}

		catalog = Catalog{
			Services: []Service{service1, service2},
		}

		config = Config{
			Region:                       "scloudformation-region",
			CloudFormationPrefix:         "cf",
			AllowUserProvisionParameters: allowUserProvisionParameters,
			AllowUserUpdateParameters:    allowUserUpdateParameters,
			Catalog:                      catalog,
		}

		logger = lager.NewLogger("cfbroker_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		cfBroker = New(config, stack, logger)
	})

	var _ = Describe("Services", func() {
		var (
			properCatalogResponse brokerapi.CatalogResponse
		)

		BeforeEach(func() {
			properCatalogResponse = brokerapi.CatalogResponse{
				Services: []brokerapi.Service{
					brokerapi.Service{
						ID:             "Service-1",
						Name:           "Service 1",
						Description:    "This is the Service 1",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-1",
								Name:        "Plan 1",
								Description: "This is the Plan 1",
							},
						},
					},
					brokerapi.Service{
						ID:             "Service-2",
						Name:           "Service 2",
						Description:    "This is the Service 2",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-2",
								Name:        "Plan 2",
								Description: "This is the Plan 2",
							},
						},
					},
				},
			}
		})

		It("returns the proper CatalogResponse", func() {
			brokerCatalog := cfBroker.Services()
			Expect(brokerCatalog).To(Equal(properCatalogResponse))
		})

	})

	var _ = Describe("Provision", func() {
		var (
			provisionDetails  brokerapi.ProvisionDetails
			acceptsIncomplete bool

			properProvisioningResponse brokerapi.ProvisioningResponse
		)

		BeforeEach(func() {
			provisionDetails = brokerapi.ProvisionDetails{
				OrganizationGUID: "organization-id",
				PlanID:           "Plan-1",
				ServiceID:        "Service-1",
				SpaceGUID:        "space-id",
				Parameters:       map[string]interface{}{},
			}
			acceptsIncomplete = true

			properProvisioningResponse = brokerapi.ProvisioningResponse{}
		})

		It("returns the proper response", func() {
			provisioningResponse, asynch, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
			Expect(provisioningResponse).To(Equal(properProvisioningResponse))
			Expect(asynch).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
			Expect(stack.CreateCalled).To(BeTrue())
			Expect(stack.CreateStackName).To(Equal(stackName))
			Expect(stack.CreateStackDetails.StackName).To(Equal(""))
			Expect(stack.CreateStackDetails.Capabilities).To(BeNil())
			Expect(stack.CreateStackDetails.DisableRollback).To(BeFalse())
			Expect(stack.CreateStackDetails.Description).To(Equal(""))
			Expect(stack.CreateStackDetails.NotificationARNs).To(BeNil())
			Expect(stack.CreateStackDetails.OnFailure).To(Equal(""))
			Expect(stack.CreateStackDetails.Outputs).To(BeNil())
			Expect(stack.CreateStackDetails.Parameters).To(BeEmpty())
			Expect(stack.CreateStackDetails.ResourceTypes).To(BeNil())
			Expect(stack.CreateStackDetails.StackID).To(Equal(""))
			Expect(stack.CreateStackDetails.StackPolicyURL).To(Equal(""))
			Expect(stack.CreateStackDetails.StackStatus).To(Equal(""))
			Expect(stack.CreateStackDetails.Tags["Owner"]).To(Equal("Cloud Foundry"))
			Expect(stack.CreateStackDetails.Tags["Created by"]).To(Equal("AWS CloudFormation Service Broker"))
			Expect(stack.CreateStackDetails.Tags).To(HaveKey("Created at"))
			Expect(stack.CreateStackDetails.Tags["Service ID"]).To(Equal("Service-1"))
			Expect(stack.CreateStackDetails.Tags["Plan ID"]).To(Equal("Plan-1"))
			Expect(stack.CreateStackDetails.Tags["Organization ID"]).To(Equal("organization-id"))
			Expect(stack.CreateStackDetails.Tags["Space ID"]).To(Equal("space-id"))
			Expect(stack.CreateStackDetails.TemplateURL).To(Equal(""))
			Expect(stack.CreateStackDetails.TimeoutInMinutes).To(Equal(int64(0)))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has Capabilities", func() {
			BeforeEach(func() {
				cfProperties1.Capabilities = []string{"test-capabilities"}
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.Capabilities).To(Equal([]string{"test-capabilities"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has DisableRollback", func() {
			BeforeEach(func() {
				cfProperties1.DisableRollback = true
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.DisableRollback).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has NotificationARNs", func() {
			BeforeEach(func() {
				cfProperties1.NotificationARNs = []string{"test-notification-arns"}
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.NotificationARNs).To(Equal([]string{"test-notification-arns"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has OnFailure", func() {
			BeforeEach(func() {
				cfProperties1.OnFailure = "test-on-failure"
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.OnFailure).To(Equal("test-on-failure"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Parameters", func() {
			BeforeEach(func() {
				cfProperties1.Parameters = map[string]string{"test-parameters-key": "test-parameters-value"}
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.Parameters).To(Equal(map[string]string{"test-parameters-key": "test-parameters-value"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has ResourceTypes", func() {
			BeforeEach(func() {
				cfProperties1.ResourceTypes = []string{"test-resource-types"}
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.ResourceTypes).To(Equal([]string{"test-resource-types"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has StackPolicyURL", func() {
			BeforeEach(func() {
				cfProperties1.StackPolicyURL = "test-stack-policy-url"
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.StackPolicyURL).To(Equal("test-stack-policy-url"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has TemplateURL", func() {
			BeforeEach(func() {
				cfProperties1.TemplateURL = "test-template-url"
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.TemplateURL).To(Equal("test-template-url"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has TimeoutInMinutes", func() {
			BeforeEach(func() {
				cfProperties1.TimeoutInMinutes = int64(1)
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.TimeoutInMinutes).To(Equal(int64(1)))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has user provision parameters", func() {
			BeforeEach(func() {
				provisionDetails.Parameters = map[string]interface{}{"test-key-1": "test-value-1", "test-key-2": "test-value-2"}
			})

			It("makes the proper calls", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(stack.CreateStackDetails.Parameters).To(Equal(map[string]string{"test-key-1": "test-value-1", "test-key-2": "test-value-2"}))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("but are not allowed", func() {
				BeforeEach(func() {
					allowUserProvisionParameters = false
				})

				It("makes the proper calls", func() {
					_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(stack.CreateStackDetails.Parameters).To(BeEmpty())
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("but are not valid", func() {
				BeforeEach(func() {
					provisionDetails.Parameters = map[string]interface{}{"invalid": true, "valid": "false"}
				})

				It("returns the proper error", func() {
					_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("'[invalid]' expected type 'string', got unconvertible type 'bool'"))
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserProvisionParameters = false
					})

					It("does not return an error", func() {
						_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when request does not accept incomplete", func() {
			BeforeEach(func() {
				acceptsIncomplete = false
			})

			It("returns the proper error", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrAsyncRequired))
			})
		})

		Context("when Service Plan is not found", func() {
			BeforeEach(func() {
				provisionDetails.PlanID = "unknown"
			})

			It("returns the proper error", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service Plan 'unknown' not found"))
			})
		})

		Context("when creating the Stack fails", func() {
			BeforeEach(func() {
				stack.CreateError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, _, err := cfBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})
	})

	var _ = Describe("Update", func() {
		var (
			updateDetails     brokerapi.UpdateDetails
			acceptsIncomplete bool
		)

		BeforeEach(func() {
			updateDetails = brokerapi.UpdateDetails{
				ServiceID:  "Service-2",
				PlanID:     "Plan-2",
				Parameters: map[string]interface{}{},
				PreviousValues: brokerapi.PreviousValues{
					PlanID:         "Plan-1",
					ServiceID:      "Service-1",
					OrganizationID: "organization-id",
					SpaceID:        "space-id",
				},
			}
			acceptsIncomplete = true
		})

		It("returns the proper response", func() {
			asynch, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
			Expect(asynch).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
			Expect(stack.ModifyCalled).To(BeTrue())
			Expect(stack.ModifyStackName).To(Equal(stackName))
			Expect(stack.ModifyStackDetails.StackName).To(Equal(""))
			Expect(stack.ModifyStackDetails.Capabilities).To(BeNil())
			Expect(stack.ModifyStackDetails.DisableRollback).To(BeFalse())
			Expect(stack.ModifyStackDetails.Description).To(Equal(""))
			Expect(stack.ModifyStackDetails.NotificationARNs).To(BeNil())
			Expect(stack.ModifyStackDetails.OnFailure).To(Equal(""))
			Expect(stack.ModifyStackDetails.Outputs).To(BeNil())
			Expect(stack.ModifyStackDetails.Parameters).To(BeEmpty())
			Expect(stack.ModifyStackDetails.ResourceTypes).To(BeNil())
			Expect(stack.ModifyStackDetails.StackID).To(Equal(""))
			Expect(stack.ModifyStackDetails.StackPolicyURL).To(Equal(""))
			Expect(stack.ModifyStackDetails.StackStatus).To(Equal(""))
			Expect(stack.ModifyStackDetails.TemplateURL).To(Equal(""))
			Expect(stack.ModifyStackDetails.TimeoutInMinutes).To(Equal(int64(0)))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has Capabilities", func() {
			BeforeEach(func() {
				cfProperties2.Capabilities = []string{"test-capabilities"}
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.Capabilities).To(Equal([]string{"test-capabilities"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has DisableRollback", func() {
			BeforeEach(func() {
				cfProperties2.DisableRollback = true
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.DisableRollback).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has NotificationARNs", func() {
			BeforeEach(func() {
				cfProperties2.NotificationARNs = []string{"test-notification-arns"}
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.NotificationARNs).To(Equal([]string{"test-notification-arns"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has OnFailure", func() {
			BeforeEach(func() {
				cfProperties2.OnFailure = "test-on-failure"
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.OnFailure).To(Equal("test-on-failure"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Parameters", func() {
			BeforeEach(func() {
				cfProperties2.Parameters = map[string]string{"test-parameters-key": "test-parameters-value"}
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.Parameters).To(Equal(map[string]string{"test-parameters-key": "test-parameters-value"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has ResourceTypes", func() {
			BeforeEach(func() {
				cfProperties2.ResourceTypes = []string{"test-resource-types"}
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.ResourceTypes).To(Equal([]string{"test-resource-types"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has StackPolicyURL", func() {
			BeforeEach(func() {
				cfProperties2.StackPolicyURL = "test-stack-policy-url"
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.StackPolicyURL).To(Equal("test-stack-policy-url"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has TemplateURL", func() {
			BeforeEach(func() {
				cfProperties2.TemplateURL = "test-template-url"
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.TemplateURL).To(Equal("test-template-url"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has TimeoutInMinutes", func() {
			BeforeEach(func() {
				cfProperties2.TimeoutInMinutes = int64(1)
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.TimeoutInMinutes).To(Equal(int64(1)))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when request does not accept incomplete", func() {
			BeforeEach(func() {
				acceptsIncomplete = false
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrAsyncRequired))
			})
		})

		Context("when has user provision parameters", func() {
			BeforeEach(func() {
				updateDetails.Parameters = map[string]interface{}{"test-key-1": "test-value-1", "test-key-2": "test-value-2"}
			})

			It("makes the proper calls", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(stack.ModifyStackDetails.Parameters).To(Equal(map[string]string{"test-key-1": "test-value-1", "test-key-2": "test-value-2"}))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("but are not allowed", func() {
				BeforeEach(func() {
					allowUserUpdateParameters = false
				})

				It("makes the proper calls", func() {
					_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(stack.ModifyStackDetails.Parameters).To(BeEmpty())
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("but are not valid", func() {
				BeforeEach(func() {
					updateDetails.Parameters = map[string]interface{}{"invalid": true, "valid": "false"}
				})

				It("returns the proper error", func() {
					_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("'[invalid]' expected type 'string', got unconvertible type 'bool'"))
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserUpdateParameters = false
					})

					It("does not return an error", func() {
						_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when Service is not found", func() {
			BeforeEach(func() {
				updateDetails.ServiceID = "unknown"
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service 'unknown' not found"))
			})
		})

		Context("when Plans is not updateable", func() {
			BeforeEach(func() {
				planUpdateable = false
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceNotUpdateable))
			})
		})

		Context("when Service Plan is not found", func() {
			BeforeEach(func() {
				updateDetails.PlanID = "unknown"
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service Plan 'unknown' not found"))
			})
		})

		Context("when modifying the Stack fails", func() {
			BeforeEach(func() {
				stack.ModifyError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("when the Stack does not exists", func() {
				BeforeEach(func() {
					stack.ModifyError = awscf.ErrStackDoesNotExist
				})

				It("returns the proper error", func() {
					_, err := cfBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Deprovision", func() {
		var (
			deprovisionDetails brokerapi.DeprovisionDetails
			acceptsIncomplete  bool
		)

		BeforeEach(func() {
			deprovisionDetails = brokerapi.DeprovisionDetails{
				ServiceID: "Service-1",
				PlanID:    "Plan-1",
			}
			acceptsIncomplete = true
		})

		It("returns the proper response", func() {
			asynch, err := cfBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
			Expect(asynch).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, err := cfBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
			Expect(stack.DeleteCalled).To(BeTrue())
			Expect(stack.DeleteStackName).To(Equal(stackName))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when request does not accept incomplete", func() {
			BeforeEach(func() {
				acceptsIncomplete = false
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrAsyncRequired))
			})
		})

		Context("when deleting the Stack fails", func() {
			BeforeEach(func() {
				stack.DeleteError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("when the Stack does not exists", func() {
				BeforeEach(func() {
					stack.DeleteError = awscf.ErrStackDoesNotExist
				})

				It("returns the proper error", func() {
					_, err := cfBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Bind", func() {
		var (
			bindDetails brokerapi.BindDetails
			credentials map[string]string
		)

		BeforeEach(func() {
			bindDetails = brokerapi.BindDetails{
				ServiceID:  "Service-1",
				PlanID:     "Plan-1",
				AppGUID:    "Application-1",
				Parameters: map[string]interface{}{},
			}

			credentials = map[string]string{
				"test-output-key-1": "test-output-key-1",
				"test-output-key-2": "test-output-key-2",
				"test-output-key-3": "test-output-key-3",
			}

			stack.DescribeStackDetails = awscf.StackDetails{
				StackName: stackName,
				Outputs:   credentials,
			}
		})

		It("returns the proper response", func() {
			bindingResponse, err := cfBroker.Bind(instanceID, bindingID, bindDetails)
			Expect(bindingResponse.Credentials).To(Equal(credentials))
			Expect(bindingResponse.SyslogDrainURL).To(BeEmpty())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, err := cfBroker.Bind(instanceID, bindingID, bindDetails)
			Expect(stack.DescribeCalled).To(BeTrue())
			Expect(stack.DescribeStackName).To(Equal(stackName))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when Service is not found", func() {
			BeforeEach(func() {
				bindDetails.ServiceID = "unknown"
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service 'unknown' not found"))
			})
		})

		Context("when Service is not bindable", func() {
			BeforeEach(func() {
				serviceBindable = false
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceNotBindable))
			})
		})

		Context("when describing the Stack fails", func() {
			BeforeEach(func() {
				stack.DescribeError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := cfBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("when the Stack does not exists", func() {
				BeforeEach(func() {
					stack.DescribeError = awscf.ErrStackDoesNotExist
				})

				It("returns the proper error", func() {
					_, err := cfBroker.Bind(instanceID, bindingID, bindDetails)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Unbind", func() {
		var (
			unbindDetails brokerapi.UnbindDetails
		)

		BeforeEach(func() {
			unbindDetails = brokerapi.UnbindDetails{
				ServiceID: "Service-1",
				PlanID:    "Plan-1",
			}
		})

		It("does not return error", func() {
			err := cfBroker.Unbind(instanceID, bindingID, unbindDetails)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	var _ = Describe("LastOperation", func() {
		var (
			stackStatus                 string
			lastOperationState          string
			properLastOperationResponse brokerapi.LastOperationResponse
		)

		JustBeforeEach(func() {
			stack.DescribeStackDetails = awscf.StackDetails{
				StackName:   stackName,
				StackStatus: stackStatus,
			}

			properLastOperationResponse = brokerapi.LastOperationResponse{
				State:       lastOperationState,
				Description: "Stack '" + stackName + "' status is '" + stackStatus + "'",
			}
		})

		Context("when describing the Stack fails", func() {
			BeforeEach(func() {
				stack.DescribeError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := cfBroker.LastOperation(instanceID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("when the Stack does not exists", func() {
				BeforeEach(func() {
					stack.DescribeError = awscf.ErrStackDoesNotExist
				})

				It("returns the proper error", func() {
					_, err := cfBroker.LastOperation(instanceID)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})

		Context("when last operation is still in progress", func() {
			BeforeEach(func() {
				stackStatus = awscf.StatusInProgress
				lastOperationState = brokerapi.LastOperationInProgress
			})

			It("returns the proper LastOperationResponse", func() {
				lastOperationResponse, err := cfBroker.LastOperation(instanceID)
				Expect(err).ToNot(HaveOccurred())
				Expect(lastOperationResponse).To(Equal(properLastOperationResponse))
			})
		})

		Context("when last operation failed", func() {
			BeforeEach(func() {
				stackStatus = awscf.StatusFailed
				lastOperationState = brokerapi.LastOperationFailed
			})

			It("returns the proper LastOperationResponse", func() {
				lastOperationResponse, err := cfBroker.LastOperation(instanceID)
				Expect(err).ToNot(HaveOccurred())
				Expect(lastOperationResponse).To(Equal(properLastOperationResponse))
			})
		})

		Context("when last operation succeeded", func() {
			BeforeEach(func() {
				stackStatus = awscf.StatusSucceeded
				lastOperationState = brokerapi.LastOperationSucceeded
			})

			It("returns the proper LastOperationResponse", func() {
				lastOperationResponse, err := cfBroker.LastOperation(instanceID)
				Expect(err).ToNot(HaveOccurred())
				Expect(lastOperationResponse).To(Equal(properLastOperationResponse))
			})
		})
	})
})
