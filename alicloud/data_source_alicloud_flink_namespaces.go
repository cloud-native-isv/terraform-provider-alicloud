package alicloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudFlinkNamespaces() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkNamespacesRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ID of the Flink workspace",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"namespaces": {
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
						"workspace_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ha": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"cpu": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudFlinkNamespacesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := newFlinkNamespacesReadinessService(client)
	if err != nil {
		return WrapError(err)
	}
	return readFlinkNamespacesWithService(
		context.Background(),
		d,
		flinkService,
		flinkLegacyNamespaceReadinessTimeout(d, schema.TimeoutRead),
		flinkLegacyNamespaceReadinessPollInterval,
	)
}

var newFlinkNamespacesReadinessService = func(client *connectivity.AliyunClient) (flinkLegacyNamespaceReadinessService, error) {
	return NewFlinkService(client)
}

func readFlinkNamespacesWithService(
	ctx context.Context,
	d *schema.ResourceData,
	flinkService flinkLegacyNamespaceReadinessService,
	timeout time.Duration,
	pollInterval time.Duration,
) error {
	workspace := d.Get("workspace_id").(string)
	idsMap := make(map[string]string)
	requiredNames := make(map[string]struct{})
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			id := vv.(string)
			prefix := workspace + "/"
			if !strings.HasPrefix(id, prefix) || strings.TrimPrefix(id, prefix) == "" {
				return fmt.Errorf("Flink namespace ID %q must use workspace/name for workspace %q", id, workspace)
			}
			idsMap[id] = id
			requiredNames[strings.TrimPrefix(id, prefix)] = struct{}{}
		}
	}

	namesMap := make(map[string]string)
	if v, ok := d.GetOk("names"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			name := vv.(string)
			namesMap[name] = name
			requiredNames[name] = struct{}{}
		}
	}

	addDebug("dataSourceAliCloudFlinkNamespacesRead", "ListNamespaces", map[string]interface{}{
		"workspaceId": workspace,
		"idsFilter":   idsMap,
		"namesFilter": namesMap,
	})

	// A paid Workspace Create intentionally returns as soon as its real ID is
	// known. Dependent reads therefore own their readiness failures and wait for
	// the namespace/default-queue control plane without tainting the parent.
	selection := flinkLegacyUnfilteredNamespaceSelection()
	if len(requiredNames) > 0 {
		selection = flinkLegacyExactNamespaceSelection(requiredNames)
	}
	namespaces, err := waitForFlinkLegacyNamespaceReadiness(ctx, flinkService, workspace, selection, timeout, pollInterval)
	if err != nil {
		addDebug("dataSourceAliCloudFlinkNamespacesRead", "NamespaceReadinessError", err)
		return WrapError(err)
	}
	addDebug("dataSourceAliCloudFlinkNamespacesRead", "NamespaceReadinessResponse", len(namespaces))

	// Filter and map results
	var namespaceMaps []map[string]interface{}
	var filteredIds []string
	var filteredNames []string

	for _, namespace := range namespaces {
		namespaceName := namespace.Name
		namespaceId := fmt.Sprintf("%s/%s", workspace, namespaceName)

		// Apply filters
		if len(idsMap) > 0 {
			if _, ok := idsMap[namespaceId]; !ok {
				continue
			}
		}

		if len(namesMap) > 0 {
			if _, ok := namesMap[namespaceName]; !ok {
				continue
			}
		}

		namespaceMap := map[string]interface{}{
			"id":           namespaceId,
			"name":         namespaceName,
			"workspace_id": workspace,
			"status":       namespace.Status,
			"ha":           namespace.Ha,
		}

		// Set resource specifications if available
		if namespace.ResourceSpec != nil {
			namespaceMap["cpu"] = int(namespace.ResourceSpec.Cpu)
			namespaceMap["memory"] = int(namespace.ResourceSpec.MemoryGB)
		}

		namespaceMaps = append(namespaceMaps, namespaceMap)
		filteredIds = append(filteredIds, namespaceId)
		filteredNames = append(filteredNames, namespaceName)
	}

	// Set the data source ID (required for Terraform data sources)
	d.SetId(fmt.Sprintf("%s:%d", workspace, time.Now().Unix()))

	if err := d.Set("ids", filteredIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", filteredNames); err != nil {
		return WrapError(err)
	}
	if err := d.Set("namespaces", namespaceMaps); err != nil {
		return WrapError(err)
	}

	// Output to file if specified
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), namespaceMaps); err != nil {
			return WrapError(err)
		}
	}

	return nil
}
