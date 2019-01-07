package arm

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	armMsi "github.com/Azure/azure-sdk-for-go/services/preview/msi/mgmt/2015-08-31-preview/msi"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateUserAssignedManagedIdentity struct {
	client *AzureClient
	create func(ctx context.Context, resourceGroupName string, location string, tags map[string]*string) error
	say    func(message string)
	error  func(e error)
	exists func(ctx context.Context, resourceGroupName string) (bool, error)
}

func NewStepCreateUserAssignedManagedIdentity(client *AzureClient, ui packer.Ui) *StepCreateUserAssignedManagedIdentity {
	var step = &StepCreateUserAssignedManagedIdentity{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.create = step.createUserAssignedManagedIdentity
	return step
}

func (s *StepCreateUserAssignedManagedIdentity) createUserAssignedManagedIdentity(ctx context.Context, resourceGroupName string, location string, resourceName string, resourceRoles []string, map[string]*string) error {
	_, err := s.client.UserAssignedIdentitiesClient.CreateOrUpdate(ctx, resourceGroupName, armMsi.Identity{
		Location: &location,
		Tags:     tags,
	})

	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepCreateUserAssignedManagedIdentity) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Creating user assigned managed identity ...")
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var existingResourceGroup = state.Get(constants.ArmIsExistingResourceGroup).(bool)
	var resourceGroupCreated = state.Get(constants.ArmIsResourceGroupCreated).(bool)
	var managedUserIdentity = state.Get(constants.ArmManagedUserIdentity).(string)
	var managedUserIdentityRoles = state.Get(constants.ArmManagedUserIdentityRoles).([]string)
	var location = state.Get(constants.ArmLocation).(string)
	var tags = state.Get(constants.ArmTags).(map[string]*string)

	state.Put(constants.ArmIsManagedUserIdentityCreated, false)
	// If an existing resource group then just need to complete straight away
	if existingResourceGroup {
		state.Put(constants.ArmIsManagedUserIdentityCreated, true)
		return multistep.ActionContinue
	}

	// If the Resource Group has not been created need to let people know
	if !resourceGroupCreated {
		err := errors.New(fmt.Sprintf(" -> Resource Group '%s' not created", resourceGroupName))
		return processStepResult(err, s.error, state)
	}

	// Everything else is okay so lets create the Managed User Identity
	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ManagedUserIdentity : '%s'", managedUserIdentity))
	s.say(fmt.Sprintf(" -> ManagedUserIdentityRoles"))
	for _, id := range managedUserIdentityRoles {
		managedUserIdentityRole := fmt.Sprintf("%s", id)
		s.say(fmt.Sprintf("   -> '%s'", managedUserIdentityRole))
	}
	err := s.create(ctx, resourceGroupName, managedUserIdentity, managedUserIdentityRoles, location, tags)
	if err == nil {
		state.Put(constants.ArmIsManagedUserIdentityCreated, true)
	}
	return processStepResult(err, s.error, state)
}

func (s *StepCreateUserAssignedManagedIdentity) Cleanup(state multistep.StateBag) {
	return
}
