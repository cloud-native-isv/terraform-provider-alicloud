package alicloud

import (
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const resourceAliCloudCSKubernetesName = "alicloud_cs_kubernetes"

// ackClusterProfileManaged and ackClusterProfileDefault distinguish managed
// (server-managed control plane) from proprietary (Docker/Default) clusters.
const (
	ackClusterProfileManaged = "Managed"
	ackClusterProfileDefault = "Default"
)

func resourceAliCloudCSKubernetes() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCSKubernetesCreate,
		Read:   resourceAliCloudCSKubernetesRead,
		Update: resourceAliCloudCSKubernetesUpdate,
		Delete: resourceAliCloudCSKubernetesDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster.",
			},
			"cluster_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "Kubernetes",
				Description: "The cluster type, default to `Kubernetes`.",
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     ackClusterProfileManaged,
				Description: "The cluster profile: `Managed` (ACK managed cluster) or `Default` (proprietary cluster), default to `Managed`.",
				ValidateFunc: validation.StringInSlice([]string{
					ackClusterProfileManaged, ackClusterProfileDefault,
				}, false),
			},
			"cluster_spec": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The cluster spec, e.g. `ack.pro.small` for a managed Pro cluster.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VPC where the cluster is deployed.",
			},
			"vswitch_ids": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Description: "The vSwitches where the cluster nodes are deployed.",
			},
			"container_cidr": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The CIDR block of the container network. Computed when omitted (service-assigned).",
			},
			"service_cidr": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The CIDR block of the service network. Computed when omitted (service-assigned).",
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The Kubernetes version. Computed when omitted (latest service default).",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The security group for the cluster. Computed when omitted (service-created).",
			},
			"node_cidr_mask": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      24,
				ValidateFunc: validation.IntBetween(24, 28),
				Description:  "The node CIDR mask for the flannel network, default to 24.",
			},
			"snat_entry": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Whether to create a SNAT entry for the cluster network, default to true.",
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to enable cluster deletion protection. Can be updated in place.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The resource group the cluster belongs to. Applied after creation via ModifyCluster (the CreateCluster API has no resource group parameter).",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The tags of the cluster.",
			},
			"worker_instance_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The instance type of the worker nodes (used with profile `Default`, or to seed an initial worker pool).",
			},
			"worker_system_disk_category": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The system disk category of worker nodes.",
			},
			"worker_system_disk_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The system disk size (GiB) of worker nodes.",
			},
			"num_of_workers": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The number of worker nodes in the initial worker pool.",
			},
			"retain_all_resources": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Delete semantics only: retain all cloud resources (SLB, NAT, ECS) when the cluster is destroyed.",
			},

			// Computed attributes
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The cluster state, e.g. `running`.",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current Kubernetes version of the cluster.",
			},
			"master_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API server endpoint of the cluster.",
			},
			"subnet_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The subnet CIDR of the cluster.",
			},
			"worker_ram_role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The RAM role attached to worker nodes.",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The zone of the cluster.",
			},
			"node_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of nodes in the cluster.",
			},
		},
	}
}

// buildAckClusterCreate maps the resource schema onto the cws-lib-go
// AckClusterCreate input. Exposed for offline table-driven tests.
func buildAckClusterCreate(d *schema.ResourceData, regionId string) *aliyunAckAPI.AckClusterCreate {
	create := &aliyunAckAPI.AckClusterCreate{
		Name:                     d.Get("name").(string),
		RegionId:                 regionId,
		ClusterType:              d.Get("cluster_type").(string),
		Profile:                  d.Get("profile").(string),
		ClusterSpec:              d.Get("cluster_spec").(string),
		VpcId:                    d.Get("vpc_id").(string),
		ContainerCidr:            d.Get("container_cidr").(string),
		ServiceCidr:              d.Get("service_cidr").(string),
		KubernetesVersion:        d.Get("kubernetes_version").(string),
		SecurityGroupId:          d.Get("security_group_id").(string),
		NodeCidrMask:             strconv.Itoa(d.Get("node_cidr_mask").(int)),
		SnatEntry:                d.Get("snat_entry").(bool),
		WorkerInstanceType:       d.Get("worker_instance_type").(string),
		WorkerSystemDiskCategory: d.Get("worker_system_disk_category").(string),
		WorkerSystemDiskSize:     int64(d.Get("worker_system_disk_size").(int)),
		NumOfWorkers:             int64(d.Get("num_of_workers").(int)),
		Tags:                     expandAckTags(d.Get("tags")),
	}
	for _, v := range d.Get("vswitch_ids").([]interface{}) {
		create.VSwitchIds = append(create.VSwitchIds, v.(string))
	}
	return create
}

// buildAckClusterModify maps the updatable schema fields (deletion
// protection, resource group) onto the AckClusterModify input.
func buildAckClusterModify(d *schema.ResourceData) *aliyunAckAPI.AckClusterModify {
	modify := &aliyunAckAPI.AckClusterModify{}
	if d.HasChange("deletion_protection") {
		deletionProtection := d.Get("deletion_protection").(bool)
		modify.DeletionProtection = &deletionProtection
	}
	if rg, ok := d.GetOk("resource_group_id"); ok {
		modify.ResourceGroupId = rg.(string)
	}
	return modify
}

// setAckClusterAttributes copies the DescribeClusterDetail result into the
// resource state. Exposed for offline table-driven tests.
func setAckClusterAttributes(d *schema.ResourceData, cluster *aliyunAckAPI.AckCluster) error {
	if cluster == nil {
		return nil
	}
	d.Set("name", cluster.Name)
	d.Set("cluster_type", cluster.ClusterType)
	d.Set("profile", cluster.Profile)
	d.Set("cluster_spec", cluster.ClusterSpec)
	d.Set("vpc_id", cluster.VpcId)
	d.Set("container_cidr", cluster.ContainerCidr)
	d.Set("service_cidr", cluster.ServiceCidr)
	d.Set("kubernetes_version", cluster.CurrentVersion)
	d.Set("security_group_id", cluster.SecurityGroupId)
	d.Set("resource_group_id", cluster.ResourceGroupId)
	d.Set("deletion_protection", cluster.DeletionProtection)
	d.Set("tags", flattenAckTags(cluster.Tags))
	d.Set("state", cluster.State)
	d.Set("version", cluster.CurrentVersion)
	d.Set("master_url", cluster.MasterUrl)
	d.Set("subnet_cidr", cluster.SubnetCidr)
	d.Set("worker_ram_role_name", cluster.WorkerRamRoleName)
	d.Set("zone_id", cluster.ZoneId)
	d.Set("node_count", int(cluster.Size))
	if len(cluster.VswitchIds) > 0 {
		vswitchIds := make([]interface{}, 0, len(cluster.VswitchIds))
		for _, v := range cluster.VswitchIds {
			vswitchIds = append(vswitchIds, v)
		}
		d.Set("vswitch_ids", vswitchIds)
	}
	return nil
}

func resourceAliCloudCSKubernetesCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	create := buildAckClusterCreate(d, client.RegionId)
	cluster, err := ackService.GetAPI().CreateCluster(create)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "CreateCluster", err)
	}
	d.SetId(cluster.ClusterId)

	// The CreateCluster API carries no resource group parameter; when the
	// user targets a resource group the cluster is moved there right after
	// creation, before the readiness wait.
	if resourceGroupId, ok := d.GetOk("resource_group_id"); ok {
		_, err = ackService.GetAPI().ModifyCluster(cluster.ClusterId, &aliyunAckAPI.AckClusterModify{
			ResourceGroupId: resourceGroupId.(string),
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "ModifyCluster", err)
		}
	}

	// Wait for the cluster to become running: pending initial/provisioning,
	// 30m timeout, 10s delay, 20s poll, failed is a terminal state.
	stateConf := buildAckStateConf(
		ackClusterCreatePendingStates, []string{"running"},
		d.Timeout(schema.TimeoutCreate), 10*time.Second, 20*time.Second,
		ackService.AckClusterStateRefreshFunc(cluster.ClusterId, ackClusterCreateFailStates))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudCSKubernetesRead(d, meta)
}

func resourceAliCloudCSKubernetesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	cluster, err := ackService.GetAPI().DescribeClusterDetail(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "DescribeClusterDetail", err)
	}

	return setAckClusterAttributes(d, cluster)
}

func resourceAliCloudCSKubernetesUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("deletion_protection") || d.HasChange("resource_group_id") {
		if _, err := ackService.GetAPI().ModifyCluster(d.Id(), buildAckClusterModify(d)); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "ModifyCluster", err)
		}
	}
	if d.HasChange("tags") {
		if _, err := ackService.GetAPI().ModifyClusterTags(d.Id(), expandAckTags(d.Get("tags"))); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "ModifyClusterTags", err)
		}
	}

	return resourceAliCloudCSKubernetesRead(d, meta)
}

func resourceAliCloudCSKubernetesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	// Deletion protection must be disabled before the cluster can be removed.
	if deletionProtection := d.Get("deletion_protection").(bool); deletionProtection {
		disabled := false
		if _, err := ackService.GetAPI().ModifyCluster(d.Id(), &aliyunAckAPI.AckClusterModify{
			DeletionProtection: &disabled,
		}); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "ModifyCluster", err)
		}
	}

	if _, err := ackService.GetAPI().DeleteCluster(d.Id(), d.Get("retain_all_resources").(bool)); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesName, "DeleteCluster", err)
	}

	stateConf := buildAckStateConf(
		nil, []string{ackClusterStateDeleted},
		d.Timeout(schema.TimeoutDelete), 10*time.Second, 20*time.Second,
		ackService.AckClusterDeleteStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return nil
}
