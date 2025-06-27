package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudFlinkDeploymentFolders() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkDeploymentFoldersRead,
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
			"parent_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"folders": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"folder_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespace_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parent_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modified_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sub_folders": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"folder_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"parent_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAlicloudFlinkDeploymentFoldersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_name").(string)

	var folders []*aliyunFlinkAPI.DeploymentFolder

	// Parse input parameters
	parentId, hasParentId := d.GetOk("parent_id")
	idsInterface, hasIds := d.GetOk("ids")

	var requestedIds []string
	if hasIds {
		for _, id := range idsInterface.([]interface{}) {
			if id != nil {
				requestedIds = append(requestedIds, id.(string))
			}
		}
	}

	// Determine the logic based on input parameters
	switch {
	case hasIds && len(requestedIds) > 0 && hasParentId && parentId.(string) != "":
		// Case 4: parent_id + ids → Get parent's children, then filter by ids
		parentFolders, err := flinkService.GetDeploymentFoldersByParent(workspaceId, namespace, parentId.(string))
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_deployment_folders", "GetDeploymentFoldersByParent", AlibabaCloudSdkGoERROR)
		}

		// Create a map for faster lookup
		idsMap := make(map[string]bool)
		for _, id := range requestedIds {
			idsMap[id] = true
		}

		// Filter by ids
		for _, folder := range parentFolders {
			if idsMap[folder.FolderId] {
				folders = append(folders, folder)
			}
		}

	case hasParentId && parentId.(string) != "" && (!hasIds || len(requestedIds) == 0):
		// Case 3: Only parent_id → Get all children under parent
		folders, err = flinkService.GetDeploymentFoldersByParent(workspaceId, namespace, parentId.(string))
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_deployment_folders", "GetDeploymentFoldersByParent", AlibabaCloudSdkGoERROR)
		}

	case hasIds && len(requestedIds) > 0 && (!hasParentId || parentId.(string) == ""):
		// Case 2: Only ids → Get specific folders by ids
		for _, folderId := range requestedIds {
			folder, err := flinkService.GetDeploymentFolder(workspaceId, namespace, folderId)
			if err != nil {
				// Log warning but continue with other folders
				continue
			}
			folders = append(folders, folder)
		}

	default:
		// Case 1: Only workspace_id + namespace_name → Get root folder
		rootFolder, err := flinkService.GetDeploymentRootFolder(workspaceId, namespace)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_deployment_folders", "GetDeploymentRootFolder", AlibabaCloudSdkGoERROR)
		}
		if rootFolder != nil {
			folders = append(folders, rootFolder)
		}
	}

	// Convert to terraform schema format
	var folderMaps []map[string]interface{}
	var resultIds []string

	for _, folder := range folders {
		if folder == nil {
			continue
		}

		// Create composite ID
		compositeId := fmt.Sprintf("%s:%s:%s", workspaceId, namespace, folder.FolderId)

		// Convert sub folders
		var subFolders []map[string]interface{}
		if folder.SubFolders != nil {
			for _, subFolder := range folder.SubFolders {
				if subFolder != nil {
					subFolders = append(subFolders, map[string]interface{}{
						"folder_id": subFolder.FolderId,
						"name":      subFolder.Name,
						"parent_id": subFolder.ParentId,
					})
				}
			}
		}

		folderMap := map[string]interface{}{
			"id":             compositeId,
			"folder_id":      folder.FolderId,
			"name":           folder.Name,
			"namespace_name": folder.Namespace,
			"parent_id":      folder.ParentId,
			"created_at":     formatTimeInt64(folder.CreatedAt),
			"modified_at":    formatTimeInt64(folder.ModifiedAt),
			"sub_folders":    subFolders,
		}

		folderMaps = append(folderMaps, folderMap)
		resultIds = append(resultIds, compositeId)
	}

	// Set data source ID
	d.SetId(fmt.Sprintf("%s:%s:%d", workspaceId, namespace, time.Now().Unix()))

	// Set computed attributes
	if err := d.Set("ids", resultIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("folders", folderMaps); err != nil {
		return WrapError(err)
	}

	// Write to output file if specified
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), folderMaps); err != nil {
			return WrapError(err)
		}
	}

	return nil
}

// formatTimeInt64 formats int64 timestamp to string
func formatTimeInt64(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")
}
