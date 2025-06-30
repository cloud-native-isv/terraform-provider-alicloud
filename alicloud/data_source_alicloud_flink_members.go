package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudFlinkMembers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkMembersRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ID of the Flink workspace",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Name of the Flink namespace",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"workspace_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespace_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudFlinkMembersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	memberNamesMap := make(map[string]string)
	if v, ok := d.GetOk("names"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			memberNamesMap[vv.(string)] = vv.(string)
		}
	}

	addDebug("dataSourceAlicloudFlinkMembersRead", "ListMembers", map[string]interface{}{
		"workspaceId":       workspaceId,
		"namespaceName":     namespaceName,
		"memberNamesFilter": memberNamesMap,
	})

	// Get all members for the namespace
	members, err := flinkService.ListMembers(workspaceId, namespaceName)
	if err != nil {
		addDebug("dataSourceAlicloudFlinkMembersRead", "ListMembersError", err)
		return WrapError(err)
	}
	addDebug("dataSourceAlicloudFlinkMembersRead", "ListMembersResponse", len(members))

	// Filter and map results
	var memberMaps []map[string]interface{}
	var filteredNames []string

	for _, member := range members {
		memberName := member.Name
		memberId := fmt.Sprintf("%s/%s/%s", workspaceId, namespaceName, memberName)

		// Apply filters
		if len(memberNamesMap) > 0 {
			if _, ok := memberNamesMap[memberName]; !ok {
				continue
			}
		}

		memberMap := map[string]interface{}{
			"id":             memberId,
			"name":           memberName,
			"role":           member.Role,
			"workspace_id":   workspaceId,
			"namespace_name": namespaceName,
		}

		memberMaps = append(memberMaps, memberMap)
		filteredNames = append(filteredNames, memberName)
	}

	// Set the data source ID (required for Terraform data sources)
	d.SetId(fmt.Sprintf("%s/%s:%d", workspaceId, namespaceName, time.Now().Unix()))

	if err := d.Set("names", filteredNames); err != nil {
		return WrapError(err)
	}
	if err := d.Set("members", memberMaps); err != nil {
		return WrapError(err)
	}

	return nil
}
