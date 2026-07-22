package alicloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var flinkWorkspaceCapacityBootstrapPollInterval = 5 * time.Second

// resourceAliCloudFlinkWorkspaceCapacityBootstrap is a read-only graph
// barrier. It consumes the paid Workspace Create provenance through a required
// input, waits for the complete capacity tree, and exposes the same InstanceId
// for the allocation resource to depend on. Delete only detaches Terraform
// state and never mutates the Workspace or its capacity.
func resourceAliCloudFlinkWorkspaceCapacityBootstrap() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkWorkspaceCapacityBootstrapCreate,
		Read:   resourceAliCloudFlinkWorkspaceCapacityBootstrapRead,
		Delete: resourceAliCloudFlinkWorkspaceCapacityBootstrapDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"workspace_instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Workspace InstanceId passed through after the capacity tree becomes readable.",
			},
			"workspace_bootstrap_context": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Internal immutable handshake from alicloud_flink_workspace. STRICT_EXISTING grants no 404 grace and keeps strict identity handling for existing Workspaces.",
			},
			"observed_resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ResourceId observed and pinned by the read-only bootstrap barrier.",
			},
		},
	}
}

func resourceAliCloudFlinkWorkspaceCapacityBootstrapCreate(d *schema.ResourceData, meta interface{}) error {
	instanceID := d.Get("workspace_instance_id").(string)
	d.SetId(instanceID)
	return withFlinkWorkspaceCapacityAllocationLock(instanceID, func() error {
		service, err := newFlinkWorkspaceCapacityAllocationService(meta)
		if err != nil {
			return WrapError(err)
		}
		service, err = configureFlinkWorkspaceCapacityBootstrapService(
			service,
			instanceID,
			d.Get("workspace_bootstrap_context").(string),
			"",
			true,
			time.Now,
		)
		if err != nil {
			return WrapError(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
		defer cancel()
		if err := waitForFlinkWorkspaceCapacityBootstrapTree(ctx, service, instanceID, flinkWorkspaceCapacityBootstrapPollInterval); err != nil {
			return WrapError(err)
		}
		capacityService, ok := service.(*FlinkCapacityService)
		if !ok || capacityService.observedResourceID == "" {
			return WrapError(fmt.Errorf("Flink workspace %q capacity bootstrap completed without an observed ResourceId", instanceID))
		}
		return WrapError(d.Set("observed_resource_id", capacityService.observedResourceID))
	})
}

func resourceAliCloudFlinkWorkspaceCapacityBootstrapRead(d *schema.ResourceData, meta interface{}) error {
	instanceID := d.Id()
	configuredInstanceID := d.Get("workspace_instance_id").(string)
	if configuredInstanceID == "" || configuredInstanceID != instanceID {
		return WrapError(fmt.Errorf("Flink workspace capacity bootstrap state identity mismatch: state ID %q, workspace_instance_id %q", instanceID, configuredInstanceID))
	}
	return withFlinkWorkspaceCapacityAllocationLock(instanceID, func() error {
		service, err := newFlinkWorkspaceCapacityAllocationService(meta)
		if err != nil {
			return WrapError(err)
		}
		service, err = configureFlinkWorkspaceCapacityBootstrapService(
			service,
			instanceID,
			d.Get("workspace_bootstrap_context").(string),
			d.Get("observed_resource_id").(string),
			false,
			time.Now,
		)
		if err != nil {
			return WrapError(err)
		}
		_, workspaceAuthoritativelyAbsent, err := readFlinkWorkspaceCapacityAllocationTree(service, instanceID)
		if workspaceAuthoritativelyAbsent {
			d.SetId("")
			return nil
		}
		if err != nil {
			return WrapError(err)
		}
		capacityService, ok := service.(*FlinkCapacityService)
		if !ok || capacityService.observedResourceID == "" {
			return WrapError(fmt.Errorf("Flink workspace %q capacity bootstrap Read completed without an observed ResourceId", instanceID))
		}
		persistedResourceID, _ := d.Get("observed_resource_id").(string)
		if persistedResourceID != "" && persistedResourceID != capacityService.observedResourceID {
			return WrapError(fmt.Errorf("Flink workspace %q bootstrap ResourceId mismatch: observed %q, expected %q", instanceID, capacityService.observedResourceID, persistedResourceID))
		}
		return WrapError(d.Set("observed_resource_id", capacityService.observedResourceID))
	})
}

func resourceAliCloudFlinkWorkspaceCapacityBootstrapDelete(d *schema.ResourceData, _ interface{}) error {
	d.SetId("")
	return nil
}

func configureFlinkWorkspaceCapacityBootstrapService(service flinkcapacity.API, instanceID, rawContext, persistedResourceID string, allowInitialIdentityAbsence bool, now func() time.Time) (flinkcapacity.API, error) {
	if rawContext == "" || rawContext == flinkWorkspaceCapacityBootstrapContextStrictExisting {
		return service, nil
	}
	contextValue, err := parseFlinkWorkspaceCapacityBootstrapContext(rawContext)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace_bootstrap_context: %w", err)
	}
	if contextValue.ExpectedInstanceID != instanceID {
		return nil, fmt.Errorf("workspace_bootstrap_context expects InstanceId %q but workspace_instance_id is %q", contextValue.ExpectedInstanceID, instanceID)
	}
	if contextValue.ExpectedResourceID == "" && !allowInitialIdentityAbsence {
		if persistedResourceID == "" {
			return nil, fmt.Errorf("workspace_bootstrap_context requires persisted observed_resource_id outside fresh Create")
		}
		contextValue.ExpectedResourceID = persistedResourceID
	} else if persistedResourceID != "" && contextValue.ExpectedResourceID != "" && persistedResourceID != contextValue.ExpectedResourceID {
		return nil, fmt.Errorf("workspace_bootstrap_context ResourceId %q conflicts with persisted observed_resource_id %q", contextValue.ExpectedResourceID, persistedResourceID)
	}
	capacityService, ok := service.(*FlinkCapacityService)
	if !ok || capacityService == nil {
		return nil, fmt.Errorf("workspace_bootstrap_context requires the production Flink capacity service, got %T", service)
	}
	return capacityService.withCapacityBootstrapContext(contextValue, allowInitialIdentityAbsence, now), nil
}

func waitForFlinkWorkspaceCapacityBootstrapTree(ctx context.Context, service flinkcapacity.API, instanceID string, pollInterval time.Duration) error {
	if service == nil {
		return fmt.Errorf("Flink workspace capacity bootstrap service is nil")
	}
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := service.ReadTree(ctx, instanceID)
		if err == nil {
			return nil
		}
		if !flinkWorkspaceCapacityBootstrapErrorRetryable(err) {
			return err
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func flinkWorkspaceCapacityBootstrapErrorRetryable(err error) bool {
	var notReady *flinkcapacity.NotReadyError
	if errors.As(err, &notReady) {
		return true
	}
	var serviceErr *flink.FlinkServiceError
	if errors.As(err, &serviceErr) && serviceErr.GetErrorCode() == "404" {
		return false
	}
	return NeedRetry(err)
}
