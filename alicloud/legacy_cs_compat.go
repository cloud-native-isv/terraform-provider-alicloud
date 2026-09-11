package alicloud

// TRANSITIONAL legacy compatibility shims.
//
// The cs kubernetes / node pool / addon resources and the clusters /
// cluster credential data sources were rewritten on top of the cws-lib-go
// ACK API layer (see resource_alicloud_cs_kubernetes.go and friends). The
// rewrite dropped several package-level helper symbols that the legacy cs
// resources/data sources scheduled for retirement (edge/managed/serverless
// kubernetes, addons/managed-clusters data sources, service_alicloud_cs.go)
// still reference. The helpers below are extracted VERBATIM from the
// pre-rewrite files so the package keeps compiling until the legacy
// retirement commit removes both the legacy files and this file.
// DO NOT use these symbols from new code.

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	roacs "github.com/alibabacloud-go/cs-20151215/v5/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/denverdino/aliyungo/cs"
	aliyungoecs "github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"gopkg.in/yaml.v2"
)

const ResourceAliCloudCSKubernetesAddon = "resourceAliCloudCSKubernetesAddon"

const (
	KubernetesClusterNetworkTypeFlannel = "flannel"
	KubernetesClusterNetworkTypeTerway  = "terway"

	KubernetesClusterLoggingTypeSLS = "SLS"

	KubernetesClusterRRSASupportedVersion = "1.22.3-aliyun.1"
)

var (
	KubernetesClusterNodeCIDRMasksByDefault = 24
)

func migrateCluster(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	csService := CsService{client}
	oldValue, newValue := d.GetChange("cluster_spec")
	o, ok := oldValue.(string)
	if ok != true {
		return WrapErrorf(fmt.Errorf("cluster_spec old value can not be parsed"), "parseError %d", oldValue)
	}
	n, ok := newValue.(string)
	if ok != true {
		return WrapErrorf(fmt.Errorf("cluster_pec new value can not be parsed"), "parseError %d", newValue)
	}

	// The field `cluster_spec` of some ack.standard managed cluster is "" since some historical reasons.
	// The logic here is to confirm whether the cluster is the above.
	// The interface error should not block the main process and errors will be output to the log.
	clusterInfo, err := csService.DescribeCsManagedKubernetes(d.Id())
	if err != nil || clusterInfo == nil {
		log.Printf("[DEBUG] Failed to DescribeCsManagedKubernetes cluster %s when migrate", d.Id())
	} else {
		o = clusterInfo.ClusterSpec
	}

	if (o == "ack.standard" || o == "") && strings.Contains(n, "pro") {
		err := migrateAliCloudManagedKubernetesCluster(d, meta)
		if err != nil {
			return WrapErrorf(err, ResponseCodeMsg, d.Id(), "MigrateCluster", AlibabaCloudSdkGoERROR)
		}
	}

	return nil
}

func modifyCluster(d *schema.ResourceData, meta interface{}, invoker *Invoker) error {
	updated := false
	request := &roacs.ModifyClusterRequest{}
	client := meta.(*connectivity.AliyunClient)
	csClient, err := client.NewRoaCsClient()
	csService := CsService{client}

	if !d.IsNewResource() && d.HasChange("resource_group_id") {
		request.SetResourceGroupId(d.Get("resource_group_id").(string))
		updated = true
	}

	if !d.IsNewResource() && d.HasChange("name") {
		var clusterName string
		if v, ok := d.GetOk("name"); ok {
			clusterName = v.(string)
			request.SetClusterName(clusterName)
			updated = true
		}
	}

	// modify cluster deletion protection
	if !d.IsNewResource() && d.HasChange("deletion_protection") {
		v := d.Get("deletion_protection")
		request.SetDeletionProtection(v.(bool))
		updated = true
	}

	// modify cluster maintenance window
	if !d.IsNewResource() && d.HasChange("maintenance_window") {
		if v := d.Get("maintenance_window").([]interface{}); len(v) > 0 {
			request.MaintenanceWindow = expandMaintenanceWindowConfigRoa(v)
			updated = true
		}
		d.SetPartial("maintenance_window")
	}

	// modify cluster maintenance window
	if !d.IsNewResource() && d.HasChange("operation_policy") {
		if v := d.Get("operation_policy").([]interface{}); len(v) > 0 {
			request.OperationPolicy = &roacs.ModifyClusterRequestOperationPolicy{}
			if vv := d.Get("operation_policy.0.cluster_auto_upgrade").([]interface{}); len(vv) > 0 {
				policy := vv[0].(map[string]interface{})
				request.OperationPolicy.ClusterAutoUpgrade = &roacs.ModifyClusterRequestOperationPolicyClusterAutoUpgrade{
					Enabled: tea.Bool(policy["enabled"].(bool)),
					Channel: tea.String(policy["channel"].(string)),
				}
			}
			updated = true
		}
	}

	// modify cluster rrsa policy
	if d.HasChange("enable_rrsa") {
		enableRRSA := false
		if v, ok := d.GetOk("enable_rrsa"); ok {
			enableRRSA = v.(bool)
		}
		// it's not allowed to disable rrsa
		if !enableRRSA {
			return fmt.Errorf("It's not supported to disable RRSA! " +
				"If your cluster has enabled this function, please manually modify your tf file and add the rrsa configuration to the file.")
		}

		// version check
		clusterVersion := d.Get("version").(string)
		if res, err := versionCompare(KubernetesClusterRRSASupportedVersion, clusterVersion); res < 0 || err != nil {
			return fmt.Errorf("RRSA is not supported in current version: %s", clusterVersion)
		}
		request.SetEnableRrsa(enableRRSA)
		updated = true
		d.SetPartial("enable_rrsa")
	}

	if d.HasChange("custom_san") {
		customSan := d.Get("custom_san").(string)
		request.SetApiServerCustomCertSans(
			&roacs.ModifyClusterRequestApiServerCustomCertSans{
				SubjectAlternativeNames: tea.StringSlice(strings.Split(customSan, ",")),
				Action:                  tea.String("overwrite"),
			},
		)
		updated = true
	}

	if d.HasChange("vswitch_ids") {
		vSwitchIds := expandStringList(d.Get("vswitch_ids").([]interface{}))
		request.SetVswitchIds(tea.StringSlice(vSwitchIds))
		updated = true
	}

	if d.HasChange("timezone") {
		request.SetTimezone(d.Get("timezone").(string))
		updated = true
	}

	if d.HasChange("security_group_id") {
		request.SetSecurityGroupId(d.Get("security_group_id").(string))
		updated = true
	}

	if updated == false {
		return nil
	}

	var resp *roacs.ModifyClusterResponse
	if err := invoker.Run(func() error {
		resp, err = csClient.ModifyCluster(tea.String(d.Id()), request)
		return err
	}); err != nil && !IsExpectedErrors(err, []string{"ClusterNameAlreadyExist", "ErrorModifyDeletionProtectionFailed"}) {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
	}

	taskId := tea.StringValue(resp.Body.TaskId)
	c := CsClient{client: csClient}
	stateConf := BuildStateConf([]string{}, []string{"success"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, c.DescribeTaskRefreshFunc(d, taskId, []string{"fail", "failed"}))
	if jobDetail, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, ResponseCodeMsg, d.Id(), "ModifyCluster", jobDetail)
	}

	stateConf = BuildStateConf([]string{"updating"}, []string{"running"}, d.Timeout(schema.TimeoutUpdate), 60*time.Second, csService.CsKubernetesInstanceStateRefreshFunc(d.Id(), []string{"deleting", "failed"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

func buildKubernetesArgs(d *schema.ResourceData, meta interface{}) (*cs.DelicatedKubernetesClusterCreationRequest, error) {
	client := meta.(*connectivity.AliyunClient)

	vpcService := VpcService{client}

	var vswitchID string
	list := make([]string, 0)
	if v, ok := d.GetOk("master_vswitch_ids"); ok {
		list = append(list, expandStringList(v.([]interface{}))...)
	}
	if v, ok := d.GetOk("worker_vswitch_ids"); ok {
		list = append(list, expandStringList(v.([]interface{}))...)
	}
	if len(list) > 0 {
		vswitchID = list[0]
	} else {
		vswitchID = ""
	}

	var vpcId string
	if vswitchID != "" {
		vsw, err := vpcService.DescribeVSwitch(vswitchID)
		if err != nil {
			return nil, err
		}
		vpcId = vsw.VpcId
	}

	var clusterName string
	if v, ok := d.GetOk("name"); ok {
		clusterName = v.(string)
	} else {
		clusterName = resource.PrefixedUniqueId(d.Get("name_prefix").(string))
	}

	addons := make([]cs.Addon, 0)
	if v, ok := d.GetOk("addons"); ok {
		all, ok := v.([]interface{})
		if ok {
			for _, a := range all {
				addon, ok := a.(map[string]interface{})
				if ok {
					addons = append(addons, cs.Addon{
						Name:     addon["name"].(string),
						Config:   addon["config"].(string),
						Version:  addon["version"].(string),
						Disabled: addon["disabled"].(bool),
					})
				}
			}
		}
	}

	var apiAudiences string
	if d.Get("api_audiences") != nil {
		if list := expandStringList(d.Get("api_audiences").([]interface{})); len(list) > 0 {
			apiAudiences = strings.Join(list, ",")
		}
	}

	creationArgs := &cs.DelicatedKubernetesClusterCreationRequest{
		ClusterArgs: cs.ClusterArgs{
			DisableRollback:    true,
			Name:               clusterName,
			DeletionProtection: d.Get("deletion_protection").(bool),
			VpcId:              vpcId,
			// the params below is ok to be empty
			KubernetesVersion:         d.Get("version").(string),
			NodeCidrMask:              strconv.Itoa(d.Get("node_cidr_mask").(int)),
			KeyPair:                   d.Get("key_name").(string),
			ServiceCidr:               d.Get("service_cidr").(string),
			CloudMonitorFlags:         d.Get("install_cloud_monitor").(bool),
			SecurityGroupId:           d.Get("security_group_id").(string),
			IsEnterpriseSecurityGroup: d.Get("is_enterprise_security_group").(bool),
			EndpointPublicAccess:      d.Get("slb_internet_enabled").(bool),
			SnatEntry:                 d.Get("new_nat_gateway").(bool),
			Addons:                    addons,
			ApiAudiences:              apiAudiences,
		},
	}

	if enableRRSA, ok := d.GetOk("enable_rrsa"); ok {
		creationArgs.EnableRRSA = enableRRSA.(bool)
	}

	if lbSpec, ok := d.GetOk("load_balancer_spec"); ok {
		creationArgs.LoadBalancerSpec = lbSpec.(string)
	}

	if osType, ok := d.GetOk("os_type"); ok {
		creationArgs.OsType = osType.(string)
	}

	if platform, ok := d.GetOk("platform"); ok {
		creationArgs.Platform = platform.(string)
	}

	if timezone, ok := d.GetOk("timezone"); ok {
		creationArgs.Timezone = timezone.(string)
	}

	if clusterDomain, ok := d.GetOk("cluster_domain"); ok {
		creationArgs.ClusterDomain = clusterDomain.(string)
	}

	if customSan, ok := d.GetOk("custom_san"); ok {
		creationArgs.CustomSAN = customSan.(string)
	}

	if imageId, ok := d.GetOk("image_id"); ok {
		creationArgs.ClusterArgs.ImageId = imageId.(string)
	}
	if nodeNameMode, ok := d.GetOk("node_name_mode"); ok {
		creationArgs.ClusterArgs.NodeNameMode = nodeNameMode.(string)
	}
	if saIssuer, ok := d.GetOk("service_account_issuer"); ok {
		creationArgs.ClusterArgs.ServiceAccountIssuer = saIssuer.(string)
	}
	if resourceGroupId, ok := d.GetOk("resource_group_id"); ok {
		creationArgs.ClusterArgs.ResourceGroupId = resourceGroupId.(string)
	}

	if v := d.Get("user_data").(string); v != "" {
		_, base64DecodeError := base64.StdEncoding.DecodeString(v)
		if base64DecodeError == nil {
			creationArgs.UserData = v
		} else {
			creationArgs.UserData = base64.StdEncoding.EncodeToString([]byte(v))
		}
	}

	if _, ok := d.GetOk("pod_vswitch_ids"); ok {
		creationArgs.PodVswitchIds = expandStringList(d.Get("pod_vswitch_ids").([]interface{}))
	} else {
		creationArgs.ContainerCidr = d.Get("pod_cidr").(string)
	}

	if password := d.Get("password").(string); password == "" {
		if v, ok := d.GetOk("kms_encrypted_password"); ok && v != "" {
			kmsService := KmsService{client}
			decryptResp, err := kmsService.Decrypt(v.(string), d.Get("kms_encryption_context").(map[string]interface{}))
			if err != nil {
				return nil, WrapError(err)
			}
			password = decryptResp
		}
		creationArgs.LoginPassword = password
	} else {
		creationArgs.LoginPassword = password
	}

	if tags, err := ConvertCsTags(d); err == nil {
		creationArgs.Tags = tags
	}
	// CA default is empty
	if userCa, ok := d.GetOk("user_ca"); ok {
		userCaContent, err := loadFileContent(userCa.(string))
		if err != nil {
			return nil, fmt.Errorf("reading user_ca file failed %s", err)
		}
		creationArgs.UserCa = string(userCaContent)
	}

	// set proxy mode and default is ipvs
	if proxyMode := d.Get("proxy_mode").(string); proxyMode != "" {
		creationArgs.ProxyMode = cs.ProxyMode(proxyMode)
	} else {
		creationArgs.ProxyMode = cs.ProxyMode(cs.IPVS)
	}

	// dedicated kubernetes must provide master_vswitch_ids
	if _, ok := d.GetOk("master_vswitch_ids"); ok {
		creationArgs.MasterArgs = cs.MasterArgs{
			MasterCount:              len(d.Get("master_vswitch_ids").([]interface{})),
			MasterVSwitchIds:         expandStringList(d.Get("master_vswitch_ids").([]interface{})),
			MasterInstanceTypes:      expandStringList(d.Get("master_instance_types").([]interface{})),
			MasterSystemDiskCategory: aliyungoecs.DiskCategory(d.Get("master_disk_category").(string)),
			MasterSystemDiskSize:     int64(d.Get("master_disk_size").(int)),
		}
	}

	if v, ok := d.GetOk("master_disk_snapshot_policy_id"); ok && v != "" {
		creationArgs.MasterArgs.MasterSnapshotPolicyId = v.(string)
	}

	if v, ok := d.GetOk("master_disk_performance_level"); ok && v != "" {
		creationArgs.MasterArgs.MasterSystemDiskPerformanceLevel = v.(string)
	}

	if v, ok := d.GetOk("master_instance_charge_type"); ok {
		creationArgs.MasterInstanceChargeType = v.(string)
		if creationArgs.MasterInstanceChargeType == string(PrePaid) {
			creationArgs.MasterAutoRenew = d.Get("master_auto_renew").(bool)
			creationArgs.MasterAutoRenewPeriod = d.Get("master_auto_renew_period").(int)
			creationArgs.MasterPeriod = d.Get("master_period").(int)
			creationArgs.MasterPeriodUnit = d.Get("master_period_unit").(string)
		}
	}

	var workerDiskSize int64
	if d.Get("worker_disk_size") != nil {
		workerDiskSize = int64(d.Get("worker_disk_size").(int))
	}

	if v, ok := d.GetOk("worker_vswitch_ids"); ok {
		creationArgs.WorkerArgs.WorkerVSwitchIds = expandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("worker_instance_types"); ok {
		creationArgs.WorkerArgs.WorkerInstanceTypes = expandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("worker_number"); ok {
		creationArgs.WorkerArgs.NumOfNodes = int64(v.(int))
	}
	if v, ok := d.GetOk("worker_disk_category"); ok {
		creationArgs.WorkerArgs.WorkerSystemDiskCategory = aliyungoecs.DiskCategory(v.(string))
	}
	if v, ok := d.GetOk("worker_disk_snapshot_policy_id"); ok && v != "" {
		creationArgs.WorkerArgs.WorkerSnapshotPolicyId = v.(string)
	}
	if v, ok := d.GetOk("worker_disk_performance_level"); ok && v != "" {
		creationArgs.WorkerArgs.WorkerSystemDiskPerformanceLevel = v.(string)
	}

	if dds, ok := d.GetOk("worker_data_disks"); ok {
		disks := dds.([]interface{})
		createDataDisks := make([]cs.DataDisk, 0, len(disks))
		for _, e := range disks {
			pack := e.(map[string]interface{})
			dataDisk := cs.DataDisk{
				Size:                 pack["size"].(string),
				DiskName:             pack["name"].(string),
				Category:             pack["category"].(string),
				Device:               pack["device"].(string),
				AutoSnapshotPolicyId: pack["auto_snapshot_policy_id"].(string),
				KMSKeyId:             pack["kms_key_id"].(string),
				Encrypted:            pack["encrypted"].(string),
				PerformanceLevel:     pack["performance_level"].(string),
			}
			createDataDisks = append(createDataDisks, dataDisk)
		}
		creationArgs.WorkerDataDisks = createDataDisks
	}
	if workerDiskSize != 0 {
		creationArgs.WorkerArgs.WorkerSystemDiskSize = workerDiskSize
	}

	if v, ok := d.GetOk("worker_instance_charge_type"); ok {
		creationArgs.WorkerInstanceChargeType = v.(string)
		if creationArgs.WorkerInstanceChargeType == string(PrePaid) {
			creationArgs.WorkerAutoRenew = d.Get("worker_auto_renew").(bool)
			creationArgs.WorkerAutoRenewPeriod = d.Get("worker_auto_renew_period").(int)
			creationArgs.WorkerPeriod = d.Get("worker_period").(int)
			creationArgs.WorkerPeriodUnit = d.Get("worker_period_unit").(string)
		}
	}

	if v, ok := d.GetOk("cluster_spec"); ok {
		creationArgs.ClusterSpec = v.(string)
	}

	if rdsInstances, ok := d.GetOk("rds_instances"); ok {
		creationArgs.RdsInstances = expandStringList(rdsInstances.([]interface{}))
	}

	if nodePortRange, ok := d.GetOk("node_port_range"); ok {
		creationArgs.NodePortRange = nodePortRange.(string)
	}

	if runtime, ok := d.GetOk("runtime"); ok {
		if v := runtime.(map[string]interface{}); len(v) > 0 {
			creationArgs.Runtime = expandKubernetesRuntimeConfig(v)
		}
	}

	if taints, ok := d.GetOk("taints"); ok {
		if v := taints.([]interface{}); len(v) > 0 {
			creationArgs.Taints = expandKubernetesTaintsConfig(v)
		}
	}

	// Cluster maintenance window. Effective only in the professional managed cluster
	if v, ok := d.GetOk("maintenance_window"); ok {
		creationArgs.MaintenanceWindow = expandMaintenanceWindowConfig(v.([]interface{}))
	}

	// Configure control plane log. Effective only in the professional managed cluster
	if v, ok := d.GetOk("control_plane_log_components"); ok {
		creationArgs.ControlplaneComponents = expandStringList(v.([]interface{}))
		// ttl default is 30 days
		creationArgs.ControlplaneLogTTL = "30"
	}
	if v, ok := d.GetOk("control_plane_log_ttl"); ok {
		creationArgs.ControlplaneLogTTL = v.(string)
	}
	if v, ok := d.GetOk("control_plane_log_project"); ok {
		creationArgs.ControlplaneLogProject = v.(string)
	}

	return creationArgs, nil
}

func expandKubernetesTaintsConfig(l []interface{}) []cs.Taint {
	config := []cs.Taint{}

	for _, v := range l {
		if m, ok := v.(map[string]interface{}); ok {
			config = append(config, cs.Taint{
				Key:    m["key"].(string),
				Value:  m["value"].(string),
				Effect: cs.Effect(m["effect"].(string)),
			})
		}
	}

	return config
}

func expandKubernetesRuntimeConfig(l map[string]interface{}) cs.Runtime {
	config := cs.Runtime{}

	if v, ok := l["name"]; ok && v != "" {
		config.Name = v.(string)
	}
	if v, ok := l["version"]; ok && v != "" {
		config.Version = v.(string)
	}

	return config
}

func flattenAliCloudCSCertificate(certificate *roacs.DescribeClusterUserKubeconfigResponseBody) map[string]string {
	if certificate == nil {
		return map[string]string{}
	}

	kubeConfig := make(map[string]interface{})
	_ = yaml.Unmarshal([]byte(tea.StringValue(certificate.Config)), &kubeConfig)

	m := make(map[string]string)
	m["cluster_cert"] = kubeConfig["clusters"].([]interface{})[0].(map[interface{}]interface{})["cluster"].(map[interface{}]interface{})["certificate-authority-data"].(string)
	m["client_cert"] = kubeConfig["users"].([]interface{})[0].(map[interface{}]interface{})["user"].(map[interface{}]interface{})["client-certificate-data"].(string)
	m["client_key"] = kubeConfig["users"].([]interface{})[0].(map[interface{}]interface{})["user"].(map[interface{}]interface{})["client-key-data"].(string)

	return m
}

// ACK pro maintenance window
func expandMaintenanceWindowConfig(l []interface{}) (config cs.MaintenanceWindow) {
	if len(l) == 0 || l[0] == nil {
		return
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["enable"]; ok {
		config.Enable = v.(bool)
	}
	if v, ok := m["maintenance_time"]; ok && v != "" {
		config.MaintenanceTime = cs.MaintenanceTime(v.(string))
	}
	if v, ok := m["duration"]; ok && v != "" {
		config.Duration = v.(string)
	}
	if v, ok := m["weekly_period"]; ok && v != "" {
		config.WeeklyPeriod = cs.WeeklyPeriod(v.(string))
	}

	return
}

func expandMaintenanceWindowConfigRoa(l []interface{}) *roacs.MaintenanceWindow {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	config := &roacs.MaintenanceWindow{}
	if v, ok := m["enable"]; ok {
		config.SetEnable(v.(bool))
	}
	if v, ok := m["maintenance_time"]; ok {
		config.SetMaintenanceTime(v.(string))
	}
	if v, ok := m["duration"]; ok {
		config.SetDuration(v.(string))
	}
	if v, ok := m["weekly_period"]; ok {
		config.SetWeeklyPeriod(v.(string))
	}

	return config
}

func flattenMaintenanceWindowConfigRoa(config *roacs.MaintenanceWindow) (m []map[string]interface{}) {
	if config == nil {
		return []map[string]interface{}{}
	}

	m = append(m, map[string]interface{}{
		"enable":           config.Enable,
		"maintenance_time": config.MaintenanceTime,
		"duration":         config.Duration,
		"weekly_period":    config.WeeklyPeriod,
	})

	return
}

// getApiServerSlbID gets cluster's API server SLB ID.
func getApiServerSlbID(d *schema.ResourceData, meta interface{}) (string, error) {
	rosClient, err := meta.(*connectivity.AliyunClient).NewRoaCsClient()
	if err != nil {
		return "", err
	}
	var clusterResources *roacs.DescribeClusterResourcesResponse
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		request := &roacs.DescribeClusterResourcesRequest{}
		clusterResources, err = rosClient.DescribeClusterResources(tea.String(d.Id()), request)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	for _, clusterResource := range clusterResources.Body {
		if tea.StringValue(clusterResource.ResourceType) == "SLB" || tea.StringValue(clusterResource.ResourceType) == "ALIYUN::SLB::LoadBalancer" {
			return tea.StringValue(clusterResource.InstanceId), nil
		}
	}

	return "", fmt.Errorf("cannot found api server SLB information for cluster: %s", d.Id())
}

func fetchClusterCapabilities(meta string) map[string]interface{} {
	metadata := make(map[string]interface{}, 0)
	capabilities := make(map[string]interface{}, 0)
	if meta != "" {
		err := json.Unmarshal([]byte(meta), &metadata)
		if err != nil {
			log.Printf("[DEBUG] Failed to unmarshal metadata due to %++v", err)
		}
	}
	if v, ok := metadata["Capabilities"]; ok {
		if IsEmpty(v) {
			return capabilities
		}
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return capabilities
}

type RRSAMetadata struct {
	Enabled      bool   `json:"enabled"`
	IssuerURL    string `json:"issuer"`
	ProviderName string `json:"oidc_name"`
	ProviderArn  string `json:"oidc_arn"`
}

func flattenRRSAMetadata(meta string) ([]map[string]interface{}, error) {
	meta = strings.TrimSpace(meta)
	if meta == "" {
		return nil, errors.New("invalid metadata")
	}
	metadata := struct {
		RRSAMetadata RRSAMetadata `json:"RRSAConfig"`
	}{}

	err := json.Unmarshal([]byte(meta), &metadata)
	if err != nil {
		log.Printf("[DEBUG] Failed to unmarshal metadata due to %++v", err)
		return nil, err
	}

	data := metadata.RRSAMetadata
	attributes := map[string]interface{}{
		"enabled":                data.Enabled,
		"rrsa_oidc_issuer_url":   "",
		"ram_oidc_provider_name": "",
		"ram_oidc_provider_arn":  "",
	}
	if !data.Enabled {
		return []map[string]interface{}{attributes}, nil
	}

	issuer := data.IssuerURL
	if strings.Contains(issuer, ",") {
		issuer = strings.Split(issuer, ",")[0]
	}
	attributes["rrsa_oidc_issuer_url"] = issuer
	attributes["ram_oidc_provider_name"] = data.ProviderName
	attributes["ram_oidc_provider_arn"] = data.ProviderArn

	return []map[string]interface{}{attributes}, nil
}

func flattenTags(config []*roacs.Tag) map[string]string {
	m := make(map[string]string, len(config))
	if len(config) < 0 {
		return m
	}

	for _, tag := range config {
		key := tea.StringValue(tag.Key)
		value := tea.StringValue(tag.Value)
		if key != DefaultClusterTag && key != CsPlayerAccountIdTag {
			m[key] = value
		}
	}

	return m
}

func fetchClusterMetaDataMap(meta string) map[string]interface{} {
	metadata := make(map[string]interface{}, 0)
	if meta != "" {
		err := json.Unmarshal([]byte(meta), &metadata)
		if err != nil {
			log.Printf("[DEBUG] Failed to unmarshal metadata due to %++v", err)
		}
	}

	return metadata
}

/* ConvertCsTags is extracted verbatim from the pre-rewrite resource_alicloud_cs_kubernetes_node_pool.go. */
func ConvertCsTags(d *schema.ResourceData) ([]cs.Tag, error) {
	tags := make([]cs.Tag, 0)
	tagsMap, ok := d.Get("tags").(map[string]interface{})
	if ok {
		for key, value := range tagsMap {
			if value != nil {
				if v, ok := value.(string); ok {
					tags = append(tags, cs.Tag{
						Key:   key,
						Value: v,
					})
				}
			}
		}
	}

	return tags, nil
}
