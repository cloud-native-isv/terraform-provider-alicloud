package alicloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkNamespaceCreate,
		Read:   resourceAliCloudFlinkNamespaceRead,
		Update: resourceAliCloudFlinkNamespaceUpdate,
		Delete: resourceAliCloudFlinkNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"elastic_resource_spec": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"memory_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"guaranteed_resource_spec": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"memory_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"ha": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"observed_capacity": flinkObservedCapacitySchema(false),
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := newFlinkNamespaceCreateService(client)
	if err != nil {
		return WrapError(err)
	}
	return createFlinkNamespaceWithService(
		context.Background(),
		d,
		flinkService,
		flinkLegacyNamespaceReadinessTimeout(d, schema.TimeoutCreate),
		flinkLegacyNamespaceReadinessPollInterval,
	)
}

type flinkNamespaceCreateService interface {
	flinkLegacyNamespaceReadinessService
	CreateNamespace(string, *aliyunFlinkAPI.Namespace) (*aliyunFlinkAPI.Namespace, error)
}

var newFlinkNamespaceCreateService = func(client *connectivity.AliyunClient) (flinkNamespaceCreateService, error) {
	return NewFlinkService(client)
}

var flinkLegacyNamespaceReadinessPollInterval = 2 * time.Second

var flinkLegacyNamespaceReadinessTimeout = func(d *schema.ResourceData, timeoutKey string) time.Duration {
	return d.Timeout(timeoutKey)
}

func createFlinkNamespaceWithService(
	ctx context.Context,
	d *schema.ResourceData,
	flinkService flinkNamespaceCreateService,
	timeout time.Duration,
	pollInterval time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	namespace := &aliyunFlinkAPI.Namespace{
		Name: namespaceName,
	}

	// Handle HA configuration
	if v, ok := d.GetOk("ha"); ok {
		namespace.Ha = v.(bool)
	}

	// Handle elastic resource specification
	if elasticSpecList := d.Get("elastic_resource_spec").([]interface{}); len(elasticSpecList) > 0 {
		elasticSpecMap := elasticSpecList[0].(map[string]interface{})
		namespace.ElasticResourceSpec = &aliyunFlinkAPI.ResourceSpec{
			Cpu:      float64(elasticSpecMap["cpu"].(int)),
			MemoryGB: float64(elasticSpecMap["memory_gb"].(int)),
		}
	}

	// Handle guaranteed resource specification
	if guaranteedSpecList := d.Get("guaranteed_resource_spec").([]interface{}); len(guaranteedSpecList) > 0 {
		guaranteedSpecMap := guaranteedSpecList[0].(map[string]interface{})
		namespace.GuaranteedResourceSpec = &aliyunFlinkAPI.ResourceSpec{
			Cpu:      float64(guaranteedSpecMap["cpu"].(int)),
			MemoryGB: float64(guaranteedSpecMap["memory_gb"].(int)),
		}
	}

	// The parent Create returns immediately after recording its paid identity.
	// Wait on this dependent node until namespace operations are usable, so a
	// delayed control plane cannot turn into a failed child write or parent taint.
	if _, err := waitForFlinkLegacyNamespaceReadiness(
		ctx,
		flinkService,
		workspaceId,
		flinkLegacyBootstrapNamespaceSelection(),
		time.Until(deadline),
		pollInterval,
	); err != nil {
		return WrapErrorf(err, IdMsg, workspaceId+":"+namespaceName)
	}

	_, err := flinkService.CreateNamespace(workspaceId, namespace)
	if err != nil {
		var postCreateReadErr *aliyunFlinkAPI.FlinkPostCreateReadError
		if !errors.As(err, &postCreateReadErr) || postCreateReadErr.WorkspaceId != workspaceId || postCreateReadErr.Namespace != namespaceName {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "CreateNamespace", AlibabaCloudSdkGoERROR)
		}
	}

	d.SetId(workspaceId + ":" + namespaceName)

	remaining := time.Until(deadline)
	if remaining <= 0 {
		return WrapErrorf(fmt.Errorf("timed out waiting for Flink namespace %q readiness after CreateNamespace", namespaceName), IdMsg, d.Id())
	}
	ready, err := waitForFlinkLegacyNamespaceReadiness(
		ctx,
		flinkService,
		workspaceId,
		flinkLegacyExactNamespaceSelection(map[string]struct{}{namespaceName: {}}),
		remaining,
		pollInterval,
	)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return applyFlinkNamespaceReadState(d, workspaceId, ready[0])
}

func resourceAliCloudFlinkNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, err := parseNamespaceResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	namespace, err := flinkService.GetNamespace(workspaceId, namespaceName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_namespace GetNamespace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	return applyFlinkNamespaceReadState(d, workspaceId, *namespace)
}

func applyFlinkNamespaceReadState(d *schema.ResourceData, workspaceId string, namespace aliyunFlinkAPI.Namespace) error {
	for key, value := range map[string]interface{}{
		"workspace_id":      workspaceId,
		"namespace_name":    namespace.Name,
		"status":            namespace.Status,
		"ha":                namespace.Ha,
		"observed_capacity": flattenFlinkNamespaceObservedCapacity(&namespace),
	} {
		if err := d.Set(key, value); err != nil {
			return WrapError(err)
		}
	}

	// Set elastic resource specification
	if namespace.ElasticResourceSpec != nil {
		elasticSpec := map[string]interface{}{
			"cpu":       int(namespace.ElasticResourceSpec.Cpu),
			"memory_gb": int(namespace.ElasticResourceSpec.MemoryGB),
		}
		if err := d.Set("elastic_resource_spec", []interface{}{elasticSpec}); err != nil {
			return WrapError(err)
		}
	} else {
		if err := d.Set("elastic_resource_spec", nil); err != nil {
			return WrapError(err)
		}
	}

	// Set guaranteed resource specification
	if namespace.GuaranteedResourceSpec != nil {
		guaranteedSpec := map[string]interface{}{
			"cpu":       int(namespace.GuaranteedResourceSpec.Cpu),
			"memory_gb": int(namespace.GuaranteedResourceSpec.MemoryGB),
		}
		if err := d.Set("guaranteed_resource_spec", []interface{}{guaranteedSpec}); err != nil {
			return WrapError(err)
		}
	} else {
		if err := d.Set("guaranteed_resource_spec", nil); err != nil {
			return WrapError(err)
		}
	}

	return nil
}

func resourceAliCloudFlinkNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, err := parseNamespaceResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	capacityChanged := d.HasChange("elastic_resource_spec") || d.HasChange("guaranteed_resource_spec")
	if capacityChanged || d.HasChange("ha") {
		namespace := &aliyunFlinkAPI.Namespace{
			Name: namespaceName,
		}

		// Handle HA configuration
		if v, ok := d.GetOk("ha"); ok {
			namespace.Ha = v.(bool)
		}

		// Handle elastic resource specification
		if elasticSpecList := d.Get("elastic_resource_spec").([]interface{}); len(elasticSpecList) > 0 {
			elasticSpecMap := elasticSpecList[0].(map[string]interface{})
			namespace.ElasticResourceSpec = &aliyunFlinkAPI.ResourceSpec{
				Cpu:      float64(elasticSpecMap["cpu"].(int)),
				MemoryGB: float64(elasticSpecMap["memory_gb"].(int)),
			}
		}

		// Handle guaranteed resource specification
		if guaranteedSpecList := d.Get("guaranteed_resource_spec").([]interface{}); len(guaranteedSpecList) > 0 {
			guaranteedSpecMap := guaranteedSpecList[0].(map[string]interface{})
			namespace.GuaranteedResourceSpec = &aliyunFlinkAPI.ResourceSpec{
				Cpu:      float64(guaranteedSpecMap["cpu"].(int)),
				MemoryGB: float64(guaranteedSpecMap["memory_gb"].(int)),
			}
		}

		// Update namespace using the UpdateNamespace method
		_, err := flinkService.UpdateNamespace(workspaceId, namespace)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateNamespace", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := BuildStateConf([]string{"MODIFYING"}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, flinkService.FlinkNamespaceStateRefreshFunc(workspaceId, namespaceName, []string{"FAILED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, err := parseNamespaceResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.DeleteNamespace(workspaceId, namespaceName)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteNamespace", AlibabaCloudSdkGoERROR)
	}

	// Wait for namespace to be deleted
	stateConf := &resource.StateChangeConf{
		Pending: []string{"DELETING"},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			namespace, err := flinkService.GetNamespace(workspaceId, namespaceName)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return namespace, namespace.Status, nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper function to parse namespace resource ID
func parseNamespaceResourceId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid namespace resource ID format: %s, expected workspace_id:namespace_name", id)
	}
	return parts[0], parts[1], nil
}
