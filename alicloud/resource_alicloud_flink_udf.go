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
				Default:  "JAR",
			},
			"dependency_jar_uris": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"udf_classes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"class_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"function_names": {
							Type:     schema.TypeList,
							Computed: true,
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
		Description:       d.Get("description").(string),
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
	d.Set("description", artifact.Description)
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
			Description:       d.Get("description").(string),
		}

		if _, err := flinkService.UpdateUdfArtifact(workspaceId, namespaceName, artifact); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_udf", "UpdateUdfArtifact", AlibabaCloudSdkGoERROR)
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
