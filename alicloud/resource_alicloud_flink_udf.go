package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkUdf() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkUdfCreate,
		Read:   resourceAliCloudFlinkUdfRead,
		Update: resourceAliCloudFlinkUdfUpdate,
		Delete: resourceAliCloudFlinkUdfDelete,
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
			"udf_artifact_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"jar_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"artifact_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ARTIFACT_TYPE_JAVA",
				ValidateFunc: validation.StringInSlice([]string{
					"ARTIFACT_TYPE_JAVA",
					"ARTIFACT_TYPE_PYTHON",
					"ARTIFACT_TYPE_UNKNOWN",
				}, false),
			},
			"dependency_jar_uris": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"udf_classes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"class_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"function_names": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkUdfCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	udfArtifactName := d.Get("udf_artifact_name").(string)

	artifact := &flinkAPI.UdfArtifact{
		Name:              udfArtifactName,
		JarUrl:            d.Get("jar_url").(string),
		ArtifactType:      d.Get("artifact_type").(string),
		DependencyJarUris: expandStringList(d.Get("dependency_jar_uris").([]interface{})),
	}

	log.Printf("[DEBUG] Calling CreateUdfArtifact with workspaceId: %s, namespaceName: %s, artifact: %+v", workspaceId, namespaceName, artifact)
	if _, err := flinkService.CreateUdfArtifact(workspaceId, namespaceName, artifact); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_udf", "CreateUdfArtifact", AlibabaCloudSdkGoERROR)
	}
	log.Printf("[DEBUG] CreateUdfArtifact returned success")

	d.SetId(workspaceId + ":" + namespaceName + ":" + udfArtifactName)

	// Wait for artifact to be available
	stateConf := BuildStateConf([]string{"CREATING", ""}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkUdfArtifactStateRefreshFunc(workspaceId, namespaceName, udfArtifactName))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	if v, ok := d.GetOk("udf_classes"); ok {
		udfClasses := v.([]interface{})
		for _, c := range udfClasses {
			classMap := c.(map[string]interface{})
			className := classMap["class_name"].(string)
			if funcs, ok := classMap["function_names"].([]interface{}); ok {
				for _, f := range funcs {
					functionName := f.(string)
					function := &flinkAPI.UdfFunction{
						FunctionName:    functionName,
						ClassName:       className,
						UdfArtifactName: udfArtifactName,
					}
					if _, err := flinkService.RegisterUdfFunction(workspaceId, namespaceName, function); err != nil {
						return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_udf", "RegisterUdfFunction", AlibabaCloudSdkGoERROR)
					}
				}
			}
		}
	}

	return resourceAliCloudFlinkUdfRead(d, meta)
}

func resourceAliCloudFlinkUdfRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, udfArtifactName, err := parseUdfResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if udfArtifactName == "" {
		d.SetId("")
		return nil
	}

	artifact, err := flinkService.GetUdfArtifact(workspaceId, namespaceName, udfArtifactName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_udf GetUdfArtifact Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
	d.Set("udf_artifact_name", artifact.Name)
	d.Set("jar_url", artifact.JarUrl)
	d.Set("artifact_type", artifact.ArtifactType)
	d.Set("dependency_jar_uris", artifact.DependencyJarUris)
	d.Set("udf_classes", flattenUdfClasses(artifact.UdfClasses))

	return nil
}

func resourceAliCloudFlinkUdfUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, udfArtifactName, err := parseUdfResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if d.HasChanges("jar_url", "artifact_type", "dependency_jar_uris", "description") {
		artifact := &flinkAPI.UdfArtifact{
			Name:              udfArtifactName,
			JarUrl:            d.Get("jar_url").(string),
			ArtifactType:      d.Get("artifact_type").(string),
			DependencyJarUris: expandStringList(d.Get("dependency_jar_uris").([]interface{})),
		}

		if _, err := flinkService.UpdateUdfArtifact(workspaceId, namespaceName, artifact); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_udf", "UpdateUdfArtifact", AlibabaCloudSdkGoERROR)
		}
	}

	if d.HasChange("udf_classes") {
		o, n := d.GetChange("udf_classes")
		oldClasses := o.([]interface{})
		newClasses := n.([]interface{})

		// Helper to flatten classes to map[string]map[string]bool (className -> functionName -> true)
		flatten := func(classes []interface{}) map[string]map[string]bool {
			m := make(map[string]map[string]bool)
			for _, c := range classes {
				classMap := c.(map[string]interface{})
				className := classMap["class_name"].(string)
				if _, ok := m[className]; !ok {
					m[className] = make(map[string]bool)
				}
				if funcs, ok := classMap["function_names"].([]interface{}); ok {
					for _, f := range funcs {
						m[className][f.(string)] = true
					}
				}
			}
			return m
		}

		oldMap := flatten(oldClasses)
		newMap := flatten(newClasses)

		// Find functions to delete
		for className, funcs := range oldMap {
			for funcName := range funcs {
				if _, ok := newMap[className]; !ok || !newMap[className][funcName] {
					if err := flinkService.DeleteUdfFunction(workspaceId, namespaceName, funcName, udfArtifactName, className); err != nil {
						if !NotFoundError(err) {
							return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteUdfFunction", AlibabaCloudSdkGoERROR)
						}
					}
				}
			}
		}

		// Find functions to register
		for className, funcs := range newMap {
			for funcName := range funcs {
				if _, ok := oldMap[className]; !ok || !oldMap[className][funcName] {
					function := &flinkAPI.UdfFunction{
						FunctionName:    funcName,
						ClassName:       className,
						UdfArtifactName: udfArtifactName,
					}
					if _, err := flinkService.RegisterUdfFunction(workspaceId, namespaceName, function); err != nil {
						return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_udf", "RegisterUdfFunction", AlibabaCloudSdkGoERROR)
					}
				}
			}
		}
	}

	return resourceAliCloudFlinkUdfRead(d, meta)
}

func resourceAliCloudFlinkUdfDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, udfArtifactName, err := parseUdfResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if udfArtifactName == "" {
		return nil
	}

	if v, ok := d.GetOk("udf_classes"); ok {
		udfClasses := v.([]interface{})
		for _, c := range udfClasses {
			classMap := c.(map[string]interface{})
			className := classMap["class_name"].(string)
			if funcs, ok := classMap["function_names"].([]interface{}); ok {
				for _, f := range funcs {
					functionName := f.(string)
					if err := flinkService.DeleteUdfFunction(workspaceId, namespaceName, functionName, udfArtifactName, className); err != nil {
						if !NotFoundError(err) {
							return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteUdfFunction", AlibabaCloudSdkGoERROR)
						}
					}
				}
			}
		}
	}

	err = flinkService.DeleteUdfArtifact(workspaceId, namespaceName, udfArtifactName)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteUdfArtifact", AlibabaCloudSdkGoERROR)
	}

	// Wait for artifact to be deleted
	stateConf := &resource.StateChangeConf{
		Pending: []string{"DELETING"},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			artifact, err := flinkService.GetUdfArtifact(workspaceId, namespaceName, udfArtifactName)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return artifact, "DELETING", nil
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

func parseUdfResourceId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid udf resource ID format: %s, expected workspace_id:namespaceName:udf_artifact_name", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func flattenUdfClasses(classes []*flinkAPI.UdfClass) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(classes))
	for _, c := range classes {
		result = append(result, map[string]interface{}{
			"class_name":     c.ClassName,
			"function_names": c.FunctionNames,
		})
	}
	return result
}
