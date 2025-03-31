package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink workspace",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Flink namespace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the namespace",
			},
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
	foasService := FoasService{client}

	request := foasconsole.CreateNamespaceRequest{}
	request.WorkspaceId = d.Get("workspace_id").(string)
	request.Name = d.Get("name").(string)
	if v, ok := d.GetOk("description"); ok {
		request.Description = v.(string)
	}

	response, err := foasService.CreateNamespace(request)
	if err != nil {
		return WrapError(err)
	}

	namespaceId := fmt.Sprintf("%s:%s", request.WorkspaceId, request.Name)
	d.SetId(namespaceId)

	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 10*time.Second, foasService.FlinkNamespaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, "Error waiting for Flink namespace (%s) to be created", d.Id())
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	workspaceId, namespaceName, err := parseNamespaceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	namespace, err := foasService.DescribeNamespace(workspaceId, namespaceName)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace_id", namespace.WorkspaceId)
	d.Set("name", namespace.Name)
	d.Set("description", namespace.Description)

	return nil
}

func resourceAliCloudFlinkNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	workspaceId, namespaceName, err := parseNamespaceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	request := foasconsole.CreateModifyNamespaceRequest()
	request.WorkspaceId = workspaceId
	request.Name = namespaceName

	if d.HasChange("description") {
		request.Description = d.Get("description").(string)
	}

	_, err = foasService.ModifyNamespace(request)
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{"MODIFYING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, foasService.FlinkNamespaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, "Error waiting for Flink namespace (%s) to be updated", d.Id())
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	workspaceId, namespaceName, err := parseNamespaceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	request := foasconsole.CreateDeleteNamespaceRequest()
	request.WorkspaceId = workspaceId
	request.Name = namespaceName

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, e := foasService.DeleteNamespace(request)
		if e != nil {
			if IsExpectedErrors(e, []string{"IncorrectNamespaceStatus"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}
		return nil
	})
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{"DELETING"}, []string{}, d.Timeout(schema.TimeoutDelete), 10*time.Second, foasService.FlinkNamespaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		if IsExpectedErrors(err, []string{"InvalidNamespace.NotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, "Error deleting Flink namespace: %s", d.Id())
	}

	d.SetId("")
	return nil
}

func parseNamespaceId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ID format, expected <workspace_id>:<namespace_name>")
	}
	return parts[0], parts[1], nil
}

// Add to foas_service.go
func (s *FoasService) CreateNamespace(request *foasconsole.CreateNamespaceRequest) (*foasconsole.CreateNamespaceResponse, error) {
	var response *foasconsole.CreateNamespaceResponse
	err := s.AliyunClient.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		var err error
		response, err = client.CreateNamespace(request)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *FoasService) DescribeNamespace(workspaceId, name string) (*foasconsole.DescribeNamespacesResponseNamespace, error) {
	request := foasconsole.CreateDescribeNamespacesRequest()
	request.WorkspaceId = workspaceId

	response, err := s.FoasconsoleClient.DescribeNamespaces(request)
	if err != nil {
		return nil, err
	}

	for _, ns := range response.Namespaces {
		if ns.Name == name {
			return &ns, nil
		}
	}
	return nil, WrapErrorf(fmt.Errorf("not found"), "Namespace %s not found", name)
}

func (s *FoasService) ModifyNamespace(request *foasconsole.ModifyNamespaceRequest) (*foasconsole.ModifyNamespaceResponse, error) {
	var response *foasconsole.ModifyNamespaceResponse
	err := s.AliyunClient.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		var err error
		response, err = client.ModifyNamespace(request)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *FoasService) DeleteNamespace(request *foasconsole.DeleteNamespaceRequest) (*foasconsole.DeleteNamespaceResponse, error) {
	var response *foasconsole.DeleteNamespaceResponse
	err := s.AliyunClient.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		var err error
		response, err = client.DeleteNamespace(request)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *FoasService) FlinkNamespaceStateRefreshFunc(namespaceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspaceId, name, err := parseNamespaceId(namespaceId)
		if err != nil {
			return nil, "", err
		}
		namespace, err := s.DescribeNamespace(workspaceId, name)
		if err != nil {
			if NotFoundError(err) {
				return nil, "DELETED", nil
			}
			return nil, "", err
		}
		return namespace, namespace.Status, nil
	}
}