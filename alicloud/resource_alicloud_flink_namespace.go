package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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

	request := foasconsole.CreateCreateNamespaceRequest()
	request.WorkspaceId = d.Get("workspace_id").(string)
	request.Name = d.Get("name").(string)
	request.Description = d.Get("description").(string)

	response, err := foasService.CreateNamespace(request)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", request.WorkspaceId, response.Namespace))

	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 10*time.Second, foasService.FlinkNamespaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	workspaceId, namespace, err := parseFlinkNamespaceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	namespaceInfo, err := foasService.DescribeNamespace(workspaceId, namespace)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("name", namespaceInfo.Name)
	d.Set("description", namespaceInfo.Description)

	return nil
}

func resourceAliCloudFlinkNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	workspaceId, namespace, err := parseFlinkNamespaceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	request := foasconsole.CreateModifyNamespaceRequest()
	request.WorkspaceId = workspaceId
	request.Name = namespace

	if d.HasChange("description") {
		request.Description = d.Get("description").(string)
	}

	_, err = foasService.ModifyNamespace(request)
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{"MODIFYING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, foasService.FlinkNamespaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	workspaceId, namespace, err := parseFlinkNamespaceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	request := foasconsole.CreateDeleteNamespaceRequest()
	request.WorkspaceId = workspaceId
	request.Name = namespace

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := foasService.DeleteNamespace(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"IncorrectNamespaceStatus"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
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
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

func parseFlinkNamespaceId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ID format, expected <workspace_id>:<namespace>")
	}
	return parts[0], parts[1], nil
}

type FoasService struct {
	*connectivity.AliyunClient
}

func (s *FoasService) CreateNamespace(request *foasconsole.CreateNamespaceRequest) (*foasconsole.CreateNamespaceResponse, error) {
	response, err := s.FoasconsoleClient.CreateNamespace(request)
	if err != nil {
		log.Printf("[ERROR] CreateNamespace failed: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *FoasService) DescribeNamespace(workspaceId, namespace string) (*foasconsole.DescribeNamespacesResponseNamespace, error) {
	request := foasconsole.CreateDescribeNamespacesRequest()
	request.WorkspaceId = workspaceId
	response, err := s.FoasconsoleClient.DescribeNamespaces(request)
	if err != nil {
		return nil, err
	}
	for _, ns := range response.Namespaces {
		if ns.Name == namespace {
			return &ns, nil
		}
	}
	return nil, WrapErrorf(Error("Namespace %s not found", namespace))
}

func (s *FoasService) ModifyNamespace(request *foasconsole.ModifyNamespaceRequest) (*foasconsole.ModifyNamespaceResponse, error) {
	response, err := s.FoasconsoleClient.ModifyNamespace(request)
	if err != nil {
		log.Printf("[ERROR] ModifyNamespace failed: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *FoasService) DeleteNamespace(request *foasconsole.DeleteNamespaceRequest) (*foasconsole.DeleteNamespaceResponse, error) {
	response, err := s.FoasconsoleClient.DeleteNamespace(request)
	if err != nil {
		log.Printf("[ERROR] DeleteNamespace failed: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *FoasService) FlinkNamespaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspaceId, namespace, err := parseFlinkNamespaceId(id)
		if err != nil {
			return nil, "", err
		}
		obj, err := s.DescribeNamespace(workspaceId, namespace)
		if err != nil {
			if NotFoundError(err) {
				return nil, "DELETED", nil
			}
			return nil, "", err
		}
		return obj, obj.Status, nil
	}
}