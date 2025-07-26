package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkDeploymentFolder() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkDeploymentFolderCreate,
		Read:   resourceAliCloudFlinkDeploymentFolderRead,
		Update: resourceAliCloudFlinkDeploymentFolderUpdate,
		Delete: resourceAliCloudFlinkDeploymentFolderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The workspace ID where the deployment folder is created.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The namespace where the deployment folder is created.",
			},
			"folder_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Description:  "The name of the deployment folder.",
			},
			"parent_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The parent folder ID. If not specified, the folder will be created in the root.",
			},
			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the deployment folder.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the deployment folder.",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last modification time of the deployment folder.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkDeploymentFolderCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize service
	service, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Extract parameters
	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	folderName := d.Get("folder_name").(string)
	parentId := d.Get("parent_id").(string)

	// Check if folder already exists before creating
	existingFolder, err := service.FindDeploymentFolderByName(workspaceId, namespaceName, folderName, parentId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_folder", "FindDeploymentFolderByName", AlibabaCloudSdkGoERROR)
	}

	if existingFolder != nil {
		// Folder already exists, set resource ID and proceed to read
		d.SetId(service.BuildDeploymentFolderId(workspaceId, namespaceName, existingFolder.FolderId))

		// Wait for folder to be available
		stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.DeploymentFolderStateRefreshFunc(workspaceId, namespaceName, existingFolder.FolderId, []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		return resourceAliCloudFlinkDeploymentFolderRead(d, meta)
	}

	// Create deployment folder
	folder, err := service.CreateDeploymentFolder(workspaceId, namespaceName, folderName, parentId)
	if err != nil {
		// Check if this is a duplicate folder error using the enhanced IsAlreadyExistError
		if IsAlreadyExistError(err) {
			// Try to find the existing folder again
			existingFolder, findErr := service.FindDeploymentFolderByName(workspaceId, namespaceName, folderName, parentId)
			if findErr != nil {
				return WrapErrorf(findErr, DefaultErrorMsg, "alicloud_flink_deployment_folder", "FindDeploymentFolderByName", AlibabaCloudSdkGoERROR)
			}

			if existingFolder != nil {
				// Set resource ID using the existing folder
				d.SetId(service.BuildDeploymentFolderId(workspaceId, namespaceName, existingFolder.FolderId))

				// Wait for folder to be available
				stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.DeploymentFolderStateRefreshFunc(workspaceId, namespaceName, existingFolder.FolderId, []string{}))
				if _, err := stateConf.WaitForState(); err != nil {
					return WrapErrorf(err, IdMsg, d.Id())
				}

				return resourceAliCloudFlinkDeploymentFolderRead(d, meta)
			}
		}

		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_folder", "CreateDeploymentFolder", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using service helper function
	d.SetId(service.BuildDeploymentFolderId(workspaceId, namespaceName, folder.FolderId))

	// Wait for folder to be available
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.DeploymentFolderStateRefreshFunc(workspaceId, namespaceName, folder.FolderId, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkDeploymentFolderRead(d, meta)
}

func resourceAliCloudFlinkDeploymentFolderRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize service
	service, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID using service helper function
	workspaceId, namespaceName, folderId, err := service.ParseDeploymentFolderId(d.Id())
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Get deployment folder
	folder, err := service.GetDeploymentFolder(workspaceId, namespaceName, folderId)
	if err != nil {
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetDeploymentFolder", AlibabaCloudSdkGoERROR)
	}

	// Set attributes
	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
	d.Set("folder_name", folder.Name)
	d.Set("parent_id", folder.ParentId)
	d.Set("folder_id", folder.FolderId)

	// Convert timestamps to strings
	if folder.CreatedAt > 0 {
		d.Set("created_at", time.Unix(folder.CreatedAt/1000, 0).Format(time.RFC3339))
	}
	if folder.ModifiedAt > 0 {
		d.Set("modified_at", time.Unix(folder.ModifiedAt/1000, 0).Format(time.RFC3339))
	}

	return nil
}

func resourceAliCloudFlinkDeploymentFolderUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize service
	service, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID using service helper function
	workspaceId, namespace, folderId, err := service.ParseDeploymentFolderId(d.Id())
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Check if folder_name has changed
	if d.HasChange("folder_name") {
		newName := d.Get("folder_name").(string)

		// Update deployment folder
		_, err := service.UpdateDeploymentFolder(workspaceId, namespace, folderId, newName)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateDeploymentFolder", AlibabaCloudSdkGoERROR)
		}

		// Wait for folder to be updated
		stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, service.DeploymentFolderStateRefreshFunc(workspaceId, namespace, folderId, []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkDeploymentFolderRead(d, meta)
}

func resourceAliCloudFlinkDeploymentFolderDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize service
	service, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID using service helper function
	workspaceId, namespace, folderId, err := service.ParseDeploymentFolderId(d.Id())
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Delete deployment folder
	err = service.DeleteDeploymentFolder(workspaceId, namespace, folderId)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidFolder.NotFound", "Forbidden.FolderNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeploymentFolder", AlibabaCloudSdkGoERROR)
	}

	return nil
}
