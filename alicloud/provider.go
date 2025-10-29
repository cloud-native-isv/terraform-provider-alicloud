package alicloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/aliyun/credentials-go/credentials/providers"

	"github.com/aliyun/credentials-go/credentials"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/mitchellh/go-homedir"
)

// Provider returns a schema.Provider for alicloud
func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_ACCESS_KEY", "ALIBABA_CLOUD_ACCESS_KEY_ID", "ALIBABACLOUD_ACCESS_KEY_ID"}, nil),
				Description: descriptions["access_key"],
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_SECRET_KEY", "ALIBABA_CLOUD_ACCESS_KEY_SECRET", "ALIBABACLOUD_ACCESS_KEY_SECRET"}, nil),
				Description: descriptions["secret_key"],
			},
			"security_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_SECURITY_TOKEN", "ALIBABA_CLOUD_SECURITY_TOKEN", "ALIBABACLOUD_SECURITY_TOKEN"}, nil),
				Description: descriptions["security_token"],
			},
			"ecs_role_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_ECS_ROLE_NAME", "ALIBABA_CLOUD_ECS_METADATA"}, nil),
				Description: descriptions["ecs_role_name"],
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_REGION", "ALIBABA_CLOUD_REGION"}, nil),
				Description: descriptions["region"],
			},
			"ots_instance_name": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Field 'ots_instance_name' has been deprecated from provider version 1.10.0. New field 'instance_name' of resource 'alicloud_ots_table' instead.",
			},
			"log_endpoint": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Field 'log_endpoint' has been deprecated from provider version 1.28.0. New field 'log' which in nested endpoints instead.",
			},
			"mns_endpoint": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Field 'mns_endpoint' has been deprecated from provider version 1.28.0. New field 'mns' which in nested endpoints instead.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_ACCOUNT_ID", "ALIBABA_CLOUD_ACCOUNT_ID"}, nil),
				Description: descriptions["account_id"],
			},
			"account_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: StringInSlice([]string{"Domestic", "International"}, true),
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ALIBABA_CLOUD_ACCOUNT_TYPE"}, nil),
			},
			"assume_role":           assumeRoleSchema(),
			"sign_version":          signVersionSchema(),
			"assume_role_with_oidc": assumeRoleWithOidcSchema(),
			"fc": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Field 'fc' has been deprecated from provider version 1.28.0. New field 'fc' which in nested endpoints instead.",
			},
			"endpoints": endpointsSchema(),
			"shared_credentials_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["shared_credentials_file"],
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_SHARED_CREDENTIALS_FILE", "ALIBABA_CLOUD_CREDENTIALS_FILE"}, nil),
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["profile"],
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_PROFILE", "ALIBABA_CLOUD_PROFILE"}, nil),
			},
			"skip_region_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: descriptions["skip_region_validation"],
			},
			"configuration_source": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  descriptions["configuration_source"],
				ValidateFunc: validation.StringLenBetween(0, 128),
				DefaultFunc:  schema.EnvDefaultFunc("TF_APPEND_USER_AGENT", ""),
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "HTTPS",
				Description:  descriptions["protocol"],
				ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, false),
			},
			"client_read_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLIENT_READ_TIMEOUT", 60000),
				Description: descriptions["client_read_timeout"],
			},
			"client_connect_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLIENT_CONNECT_TIMEOUT", 60000),
				Description: descriptions["client_connect_timeout"],
			},
			"source_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_SOURCE_IP", "ALIBABA_CLOUD_SOURCE_IP"}, nil),
				Description: descriptions["source_ip"],
			},
			"security_transport": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_SECURITY_TRANSPORT", "ALIBABA_CLOUD_SECURITY_TRANSPORT"}, nil),
				//Deprecated:  "It has been deprecated from version 1.136.0 and using new field secure_transport instead.",
			},
			"secure_transport": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_SECURE_TRANSPORT", "ALIBABA_CLOUD_SECURE_TRANSPORT"}, nil),
				Description: descriptions["secure_transport"],
			},
			"credentials_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_CREDENTIALS_URI", "ALIBABA_CLOUD_CREDENTIALS_URI"}, nil),
				Description: descriptions["credentials_uri"],
			},
			"max_retry_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MAX_RETRY_TIMEOUT", 0),
				Description: descriptions["max_retry_timeout"],
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"alicloud_account":     dataSourceAliCloudAccount(),
			"alicloud_ack_service": dataSourceAliCloudAckService(),
			"alicloud_actiontrail_global_events_storage_region":         dataSourceAliCloudActiontrailGlobalEventsStorageRegion(),
			"alicloud_actiontrail_history_delivery_jobs":                dataSourceAliCloudActiontrailHistoryDeliveryJobs(),
			"alicloud_actiontrail_trails":                               dataSourceAliCloudActiontrailTrails(),
			"alicloud_actiontrails":                                     dataSourceAliCloudActiontrailTrails(),
			"alicloud_adb_clusters":                                     dataSourceAliCloudAdbDbClusters(),
			"alicloud_adb_db_cluster_lake_versions":                     dataSourceAliCloudAdbDbClusterLakeVersions(),
			"alicloud_adb_db_clusters":                                  dataSourceAliCloudAdbDbClusters(),
			"alicloud_adb_resource_groups":                              dataSourceAliCloudAdbResourceGroups(),
			"alicloud_adb_zones":                                        dataSourceAliCloudAdbZones(),
			"alicloud_alb_acls":                                         dataSourceAliCloudAlbAcls(),
			"alicloud_alb_ascripts":                                     dataSourceAliCloudAlbAscripts(),
			"alicloud_alb_health_check_templates":                       dataSourceAliCloudAlbHealthCheckTemplates(),
			"alicloud_alb_listeners":                                    dataSourceAliCloudAlbListeners(),
			"alicloud_alb_load_balancers":                               dataSourceAliCloudAlbLoadBalancers(),
			"alicloud_alb_rules":                                        dataSourceAliCloudAlbRules(),
			"alicloud_alb_security_policies":                            dataSourceAliCloudAlbSecurityPolicies(),
			"alicloud_alb_server_groups":                                dataSourceAliCloudAlbServerGroups(),
			"alicloud_alb_system_security_policies":                     dataSourceAliCloudAlbSystemSecurityPolicies(),
			"alicloud_alb_zones":                                        dataSourceAliCloudAlbZones(),
			"alicloud_alidns_access_strategies":                         dataSourceAliCloudAlidnsAccessStrategies(),
			"alicloud_alidns_address_pools":                             dataSourceAliCloudAlidnsAddressPools(),
			"alicloud_alidns_custom_lines":                              dataSourceAliCloudAlidnsCustomLines(),
			"alicloud_alidns_domain_groups":                             dataSourceAliCloudAlidnsDomainGroups(),
			"alicloud_alidns_domains":                                   dataSourceAliCloudAlidnsDomains(),
			"alicloud_alidns_gtm_instances":                             dataSourceAliCloudAlidnsGtmInstances(),
			"alicloud_alidns_instances":                                 dataSourceAliCloudAlidnsInstances(),
			"alicloud_alidns_records":                                   dataSourceAliCloudAlidnsRecords(),
			"alicloud_alikafka_consumer_groups":                         dataSourceAliCloudAlikafkaConsumerGroups(),
			"alicloud_alikafka_instances":                               dataSourceAliCloudAlikafkaInstances(),
			"alicloud_alikafka_sasl_acls":                               dataSourceAliCloudAlikafkaSaslAcls(),
			"alicloud_alikafka_sasl_users":                              dataSourceAliCloudAlikafkaSaslUsers(),
			"alicloud_alikafka_topics":                                  dataSourceAliCloudAlikafkaTopics(),
			"alicloud_amqp_bindings":                                    dataSourceAliCloudAmqpBindings(),
			"alicloud_amqp_exchanges":                                   dataSourceAliCloudAmqpExchanges(),
			"alicloud_amqp_instances":                                   dataSourceAliCloudAmqpInstances(),
			"alicloud_amqp_queues":                                      dataSourceAliCloudAmqpQueues(),
			"alicloud_amqp_static_accounts":                             dataSourceAliCloudAmqpStaticAccounts(),
			"alicloud_amqp_virtual_hosts":                               dataSourceAliCloudAmqpVirtualHosts(),
			"alicloud_api_gateway_apis":                                 dataSourceAliCloudApiGatewayApis(),
			"alicloud_api_gateway_apps":                                 dataSourceAliCloudApiGatewayApps(),
			"alicloud_api_gateway_backends":                             dataSourceAliCloudApiGatewayBackends(),
			"alicloud_api_gateway_groups":                               dataSourceAliCloudApiGatewayGroups(),
			"alicloud_api_gateway_log_configs":                          dataSourceAliCloudApiGatewayLogConfigs(),
			"alicloud_api_gateway_models":                               dataSourceAliCloudApiGatewayModels(),
			"alicloud_api_gateway_plugins":                              dataSourceAliCloudApiGatewayPlugins(),
			"alicloud_api_gateway_service":                              dataSourceAliCloudApiGatewayService(),
			"alicloud_arms_alert_activities":                            dataSourceAliCloudArmsAlertActivities(),
			"alicloud_arms_alert_alarms":                                dataSourceAliCloudArmsAlertAlarms(),
			"alicloud_arms_alert_contact_groups":                        dataSourceAliCloudArmsAlertContactGroups(),
			"alicloud_arms_alert_contacts":                              dataSourceAliCloudArmsAlertContacts(),
			"alicloud_arms_alert_events":                                dataSourceAliCloudArmsAlertEvents(),
			"alicloud_arms_alert_integrations":                          dataSourceAliCloudArmsAlertIntegrations(),
			"alicloud_arms_alert_items":                                 dataSourceAliCloudArmsAlertItems(),
			"alicloud_arms_alert_notification_policies":                 dataSourceAliCloudArmsAlertNotificationPolicies(),
			"alicloud_arms_alert_robots":                                dataSourceAliCloudArmsAlertRobots(),
			"alicloud_arms_alert_rules":                                 dataSourceAliCloudArmsAlertRules(),
			"alicloud_arms_alert_silence_policies":                      dataSourceAliCloudArmsAlertSilencePolicies(),
			"alicloud_arms_alert_oncall_schedules":                      dataSourceAliCloudArmsAlertOncallSchedules(),
			"alicloud_arms_prometheus_remote_writes":                    dataSourceAliCloudArmsPrometheusRemoteWrites(),
			"alicloud_arms_grafana_workspaces":                          dataSourceAliCloudArmsGrafanaWorkspaces(),
			"alicloud_bastionhost_host_accounts":                        dataSourceAliCloudBastionhostHostAccounts(),
			"alicloud_bastionhost_host_groups":                          dataSourceAliCloudBastionhostHostGroups(),
			"alicloud_bastionhost_host_share_keys":                      dataSourceAliCloudBastionhostHostShareKeys(),
			"alicloud_bastionhost_hosts":                                dataSourceAliCloudBastionhostHosts(),
			"alicloud_bastionhost_instances":                            dataSourceAliCloudBastionhostInstances(),
			"alicloud_bastionhost_user_groups":                          dataSourceAliCloudBastionhostUserGroups(),
			"alicloud_bastionhost_users":                                dataSourceAliCloudBastionhostUsers(),
			"alicloud_bp_studio_applications":                           dataSourceAliCloudBpStudioApplications(),
			"alicloud_brain_industrial_pid_loops":                       dataSourceAliCloudBrainIndustrialPidLoops(),
			"alicloud_brain_industrial_pid_organizations":               dataSourceAliCloudBrainIndustrialPidOrganizations(),
			"alicloud_brain_industrial_pid_projects":                    dataSourceAliCloudBrainIndustrialPidProjects(),
			"alicloud_brain_industrial_service":                         dataSourceAliCloudBrainIndustrialService(),
			"alicloud_bss_open_api_pricing_modules":                     dataSourceAliCloudBssOpenApiPricingModules(),
			"alicloud_bss_open_api_products":                            dataSourceAliCloudBssOpenApiProducts(),
			"alicloud_caller_identity":                                  dataSourceAliCloudCallerIdentity(),
			"alicloud_cas_certificates":                                 dataSourceAliCloudSslCertificatesServiceCertificates(),
			"alicloud_cassandra_backup_plans":                           dataSourceAliCloudCassandraBackupPlans(),
			"alicloud_cassandra_clusters":                               dataSourceAliCloudCassandraClusters(),
			"alicloud_cassandra_data_centers":                           dataSourceAliCloudCassandraDataCenters(),
			"alicloud_cassandra_zones":                                  dataSourceAliCloudCassandraZones(),
			"alicloud_cddc_dedicated_host_accounts":                     dataSourceAliCloudCddcDedicatedHostAccounts(),
			"alicloud_cddc_dedicated_host_groups":                       dataSourceAliCloudCddcDedicatedHostGroups(),
			"alicloud_cddc_dedicated_hosts":                             dataSourceAliCloudCddcDedicatedHosts(),
			"alicloud_cddc_host_ecs_level_infos":                        dataSourceAliCloudCddcHostEcsLevelInfos(),
			"alicloud_cddc_zones":                                       dataSourceAliCloudCddcZones(),
			"alicloud_cdn_blocked_regions":                              dataSourceAliCloudCdnBlockedRegions(),
			"alicloud_cdn_ip_info":                                      dataSourceAliCloudCdnIpInfo(),
			"alicloud_cdn_real_time_log_deliveries":                     dataSourceAliCloudCdnRealTimeLogDeliveries(),
			"alicloud_cdn_service":                                      dataSourceAliCloudCdnService(),
			"alicloud_cen_bandwidth_limits":                             dataSourceAliCloudCenBandwidthLimits(),
			"alicloud_cen_bandwidth_packages":                           dataSourceAliCloudCenBandwidthPackages(),
			"alicloud_cen_child_instance_route_entry_to_attachments":    dataSourceAliCloudCenChildInstanceRouteEntryToAttachments(),
			"alicloud_cen_flowlogs":                                     dataSourceAliCloudCenFlowLogs(),
			"alicloud_cen_instance_attachments":                         dataSourceAliCloudCenInstanceAttachments(),
			"alicloud_cen_instances":                                    dataSourceAliCloudCenInstances(),
			"alicloud_cen_inter_region_traffic_qos_policies":            dataSourceAliCloudCenInterRegionTrafficQosPolicies(),
			"alicloud_cen_inter_region_traffic_qos_queues":              dataSourceAliCloudCenInterRegionTrafficQosQueues(),
			"alicloud_cen_private_zones":                                dataSourceAliCloudCenPrivateZones(),
			"alicloud_cen_region_route_entries":                         dataSourceAliCloudCenRegionRouteEntries(),
			"alicloud_cen_route_entries":                                dataSourceAliCloudCenRouteEntries(),
			"alicloud_cen_route_maps":                                   dataSourceAliCloudCenRouteMaps(),
			"alicloud_cen_route_services":                               dataSourceAliCloudCenRouteServices(),
			"alicloud_cen_traffic_marking_policies":                     dataSourceAliCloudCenTrafficMarkingPolicies(),
			"alicloud_cen_transit_route_table_aggregations":             dataSourceAliCloudCenTransitRouteTableAggregations(),
			"alicloud_cen_transit_router_available_resources":           dataSourceAliCloudCenTransitRouterAvailableResources(),
			"alicloud_cen_transit_router_cidrs":                         dataSourceAliCloudCenTransitRouterCidrs(),
			"alicloud_cen_transit_router_multicast_domain_associations": dataSourceAliCloudCenTransitRouterMulticastDomainAssociations(),
			"alicloud_cen_transit_router_multicast_domain_members":      dataSourceAliCloudCenTransitRouterMulticastDomainMembers(),
			"alicloud_cen_transit_router_multicast_domain_peer_members": dataSourceAliCloudCenTransitRouterMulticastDomainPeerMembers(),
			"alicloud_cen_transit_router_multicast_domain_sources":      dataSourceAliCloudCenTransitRouterMulticastDomainSources(),
			"alicloud_cen_transit_router_multicast_domains":             dataSourceAliCloudCenTransitRouterMulticastDomains(),
			"alicloud_cen_transit_router_peer_attachments":              dataSourceAliCloudCenTransitRouterPeerAttachments(),
			"alicloud_cen_transit_router_prefix_list_associations":      dataSourceAliCloudCenTransitRouterPrefixListAssociations(),
			"alicloud_cen_transit_router_route_entries":                 dataSourceAliCloudCenTransitRouterRouteEntries(),
			"alicloud_cen_transit_router_route_table_associations":      dataSourceAliCloudCenTransitRouterRouteTableAssociations(),
			"alicloud_cen_transit_router_route_table_propagations":      dataSourceAliCloudCenTransitRouterRouteTablePropagations(),
			"alicloud_cen_transit_router_route_tables":                  dataSourceAliCloudCenTransitRouterRouteTables(),
			"alicloud_cen_transit_router_service":                       dataSourceAliCloudCenTransitRouterService(),
			"alicloud_cen_transit_router_vbr_attachments":               dataSourceAliCloudCenTransitRouterVbrAttachments(),
			"alicloud_cen_transit_router_vpc_attachments":               dataSourceAliCloudCenTransitRouterVpcAttachments(),
			"alicloud_cen_transit_router_vpn_attachments":               dataSourceAliCloudCenTransitRouterVpnAttachments(),
			"alicloud_cen_transit_routers":                              dataSourceAliCloudCenTransitRouters(),
			"alicloud_cen_vbr_health_checks":                            dataSourceAliCloudCenVbrHealthChecks(),
			"alicloud_chatbot_agents":                                   dataSourceAliCloudChatbotAgents(),
			"alicloud_click_house_accounts":                             dataSourceAliCloudClickHouseAccounts(),
			"alicloud_click_house_backup_policies":                      dataSourceAliCloudClickHouseBackupPolicies(),
			"alicloud_click_house_db_clusters":                          dataSourceAliCloudClickHouseDbClusters(),
			"alicloud_click_house_regions":                              dataSourceAliCloudClickHouseRegions(),
			"alicloud_cloud_connect_networks":                           dataSourceAliCloudCloudConnectNetworks(),
			"alicloud_cloud_control_prices":                             dataSourceAliCloudCloudControlPrices(),
			"alicloud_cloud_control_products":                           dataSourceAliCloudCloudControlProducts(),
			"alicloud_cloud_control_resource_types":                     dataSourceAliCloudCloudControlResourceTypes(),
			"alicloud_cloud_firewall_address_books":                     dataSourceAliCloudCloudFirewallAddressBooks(),
			"alicloud_cloud_firewall_control_policies":                  dataSourceAliCloudCloudFirewallControlPolicies(),
			"alicloud_cloud_firewall_instance_members":                  dataSourceAliCloudCloudFirewallInstanceMembers(),
			"alicloud_cloud_firewall_instances":                         dataSourceAliCloudCloudFirewallInstances(),
			"alicloud_cloud_firewall_nat_firewalls":                     dataSourceAliCloudCloudFirewallNatFirewalls(),
			"alicloud_cloud_firewall_vpc_cen_tr_firewalls":              dataSourceAliCloudCloudFirewallVpcCenTrFirewalls(),
			"alicloud_cloud_firewall_vpc_firewall_cens":                 dataSourceAliCloudCloudFirewallVpcFirewallCens(),
			"alicloud_cloud_firewall_vpc_firewall_control_policies":     dataSourceAliCloudCloudFirewallVpcFirewallControlPolicies(),
			"alicloud_cloud_firewall_vpc_firewalls":                     dataSourceAliCloudCloudFirewallVpcFirewalls(),
			"alicloud_cloud_monitor_service_hybrid_double_writes":       dataSourceAliCloudCloudMonitorServiceHybridDoubleWrites(),
			"alicloud_cloud_sso_access_assignments":                     dataSourceAliCloudCloudSsoAccessAssignments(),
			"alicloud_cloud_sso_access_configurations":                  dataSourceAliCloudCloudSsoAccessConfigurations(),
			"alicloud_cloud_sso_directories":                            dataSourceAliCloudCloudSsoDirectories(),
			"alicloud_cloud_sso_groups":                                 dataSourceAliCloudCloudSsoGroups(),
			"alicloud_cloud_sso_scim_server_credentials":                dataSourceAliCloudCloudSsoScimServerCredentials(),
			"alicloud_cloud_sso_service":                                dataSourceAliCloudCloudSsoService(),
			"alicloud_cloud_sso_users":                                  dataSourceAliCloudCloudSsoUsers(),
			"alicloud_cloud_storage_gateway_express_syncs":              dataSourceAliCloudCloudStorageGatewayExpressSyncs(),
			"alicloud_cloud_storage_gateway_gateway_block_volumes":      dataSourceAliCloudCloudStorageGatewayGatewayBlockVolumes(),
			"alicloud_cloud_storage_gateway_gateway_cache_disks":        dataSourceAliCloudCloudStorageGatewayGatewayCacheDisks(),
			"alicloud_cloud_storage_gateway_gateway_file_shares":        dataSourceAliCloudCloudStorageGatewayGatewayFileShares(),
			"alicloud_cloud_storage_gateway_gateway_smb_users":          dataSourceAliCloudCloudStorageGatewayGatewaySmbUsers(),
			"alicloud_cloud_storage_gateway_gateways":                   dataSourceAliCloudCloudStorageGatewayGateways(),
			"alicloud_cloud_storage_gateway_service":                    dataSourceAliCloudCloudStorageGatewayService(),
			"alicloud_cloud_storage_gateway_stocks":                     dataSourceAliCloudCloudStorageGatewayStocks(),
			"alicloud_cloud_storage_gateway_storage_bundles":            dataSourceAliCloudCloudStorageGatewayStorageBundles(),
			"alicloud_cloudauth_face_configs":                           dataSourceAliCloudCloudauthFaceConfigs(),
			"alicloud_cms_alarm_contact_groups":                         dataSourceAliCloudCmsAlarmContactGroups(),
			"alicloud_cms_alarm_contacts":                               dataSourceAliCloudCmsAlarmContacts(),
			"alicloud_cms_dynamic_tag_groups":                           dataSourceAliCloudCmsDynamicTagGroups(),
			"alicloud_cms_event_rules":                                  dataSourceAliCloudCmsEventRules(),
			"alicloud_cms_group_metric_rules":                           dataSourceAliCloudCmsGroupMetricRules(),
			"alicloud_cms_hybrid_monitor_datas":                         dataSourceAliCloudCmsHybridMonitorDatas(),
			"alicloud_cms_hybrid_monitor_fc_tasks":                      dataSourceAliCloudCmsHybridMonitorFcTasks(),
			"alicloud_cms_hybrid_monitor_sls_tasks":                     dataSourceAliCloudCmsHybridMonitorSlsTasks(),
			"alicloud_cms_metric_rule_black_lists":                      dataSourceAliCloudCmsMetricRuleBlackLists(),
			"alicloud_cms_metric_rule_templates":                        dataSourceAliCloudCmsMetricRuleTemplates(),
			"alicloud_cms_monitor_group_instances":                      dataSourceAliCloudCmsMonitorGroupInstances(),
			"alicloud_cms_monitor_group_instanceses":                    dataSourceAliCloudCmsMonitorGroupInstances(),
			"alicloud_cms_monitor_groups":                               dataSourceAliCloudCmsMonitorGroups(),
			"alicloud_cms_namespaces":                                   dataSourceAliCloudCmsNamespaces(),
			"alicloud_cms_service":                                      dataSourceAliCloudCmsService(),
			"alicloud_cms_site_monitors":                                dataSourceAliCloudCloudMonitorServiceSiteMonitors(),
			"alicloud_cms_sls_groups":                                   dataSourceAliCloudCmsSlsGroups(),
			"alicloud_common_bandwidth_packages":                        dataSourceAliCloudCommonBandwidthPackages(),
			"alicloud_compute_nest_service_instances":                   dataSourceAliCloudComputeNestServiceInstances(),
			"alicloud_config_aggregate_compliance_packs":                dataSourceAliCloudConfigAggregateCompliancePacks(),
			"alicloud_config_aggregate_config_rules":                    dataSourceAliCloudConfigAggregateConfigRules(),
			"alicloud_config_aggregate_deliveries":                      dataSourceAliCloudConfigAggregateDeliveries(),
			"alicloud_config_aggregators":                               dataSourceAliCloudConfigAggregators(),
			"alicloud_config_compliance_packs":                          dataSourceAliCloudConfigCompliancePacks(),
			"alicloud_config_configuration_recorders":                   dataSourceAliCloudConfigConfigurationRecorders(),
			"alicloud_config_deliveries":                                dataSourceAliCloudConfigDeliveries(),
			"alicloud_config_delivery_channels":                         dataSourceAliCloudConfigDeliveryChannels(),
			"alicloud_config_rules":                                     dataSourceAliCloudConfigRules(),
			"alicloud_cr_chains":                                        dataSourceAliCloudCrChains(),
			"alicloud_cr_chart_namespaces":                              dataSourceAliCloudCrChartNamespaces(),
			"alicloud_cr_chart_repositories":                            dataSourceAliCloudCrChartRepositories(),
			"alicloud_cr_ee_instances":                                  dataSourceAliCloudCrEEInstances(),
			"alicloud_cr_ee_namespaces":                                 dataSourceAliCloudCrEENamespaces(),
			"alicloud_cr_ee_repos":                                      dataSourceAliCloudCrEERepos(),
			"alicloud_cr_ee_sync_rules":                                 dataSourceAliCloudCrEESyncRules(),
			"alicloud_cr_endpoint_acl_policies":                         dataSourceAliCloudCrEndpointAclPolicies(),
			"alicloud_cr_endpoint_acl_service":                          dataSourceAliCloudCrEndpointAclService(),
			"alicloud_cr_namespaces":                                    dataSourceAliCloudCRNamespaces(),
			"alicloud_cr_repos":                                         dataSourceAliCloudCRRepos(),
			"alicloud_cr_service":                                       dataSourceAliCloudCrService(),
			"alicloud_cr_vpc_endpoint_linked_vpcs":                      dataSourceAliCloudCrVpcEndpointLinkedVpcs(),
			"alicloud_cs_cluster_credential":                            dataSourceAliCloudCSClusterCredential(),
			"alicloud_cs_edge_kubernetes_clusters":                      dataSourceAliCloudCSEdgeKubernetesClusters(),
			"alicloud_cs_kubernetes_addon_metadata":                     dataSourceAliCloudCSKubernetesAddonMetadata(),
			"alicloud_cs_kubernetes_addons":                             dataSourceAliCloudCSKubernetesAddons(),
			"alicloud_cs_kubernetes_clusters":                           dataSourceAliCloudCSKubernetesClusters(),
			"alicloud_cs_kubernetes_node_pools":                         dataSourceAliCloudAckNodepools(),
			"alicloud_cs_kubernetes_permissions":                        dataSourceAliCloudCSKubernetesPermissions(),
			"alicloud_cs_kubernetes_version":                            dataSourceAliCloudCSKubernetesVersion(),
			"alicloud_cs_managed_kubernetes_clusters":                   dataSourceAliCloudCSManagerKubernetesClusters(),
			"alicloud_cs_serverless_kubernetes_clusters":                dataSourceAliCloudCSServerlessKubernetesClusters(),
			"alicloud_data_works_folders":                               dataSourceAliCloudDataWorksFolders(),
			"alicloud_data_works_service":                               dataSourceAliCloudDataWorksService(),
			"alicloud_database_gateway_gateways":                        dataSourceAliCloudDatabaseGatewayGateways(),
			"alicloud_datahub_service":                                  dataSourceAliCloudDatahubService(),
			"alicloud_dataworks_service":                                dataSourceAliCloudDataWorksService(),
			"alicloud_db_instance_class_infos":                          dataSourceAliCloudDBInstanceClassInfos(),
			"alicloud_db_instance_classes":                              dataSourceAliCloudDBInstanceClasses(),
			"alicloud_db_instance_engines":                              dataSourceAliCloudDBInstanceEngines(),
			"alicloud_db_instances":                                     dataSourceAliCloudDBInstances(),
			"alicloud_db_zones":                                         dataSourceAliCloudDBZones(),
			"alicloud_dbfs_auto_snap_shot_policies":                     dataSourceAliCloudDbfsAutoSnapShotPolicies(),
			"alicloud_dbfs_instances":                                   dataSourceAliCloudDbfsInstances(),
			"alicloud_dbfs_snapshots":                                   dataSourceAliCloudDbfsSnapshots(),
			"alicloud_dbs_backup_plans":                                 dataSourceAliCloudDbsBackupPlans(),
			"alicloud_dcdn_domains":                                     dataSourceAliCloudDcdnDomains(),
			"alicloud_dcdn_ipa_domains":                                 dataSourceAliCloudDcdnIpaDomains(),
			"alicloud_dcdn_kv_account":                                  dataSourceAliCloudDcdnKvAccount(),
			"alicloud_dcdn_service":                                     dataSourceAliCloudDcdnService(),
			"alicloud_dcdn_waf_domains":                                 dataSourceAliCloudDcdnWafDomains(),
			"alicloud_dcdn_waf_policies":                                dataSourceAliCloudDcdnWafPolicies(),
			"alicloud_dcdn_waf_rules":                                   dataSourceAliCloudDcdnWafRules(),
			"alicloud_ddosbgp_instances":                                dataSourceAliCloudDdosbgpInstances(),
			"alicloud_ddosbgp_ips":                                      dataSourceAliCloudDdosbgpIps(),
			"alicloud_ddoscoo_domain_resources":                         dataSourceAliCloudDdoscooDomainResources(),
			"alicloud_ddoscoo_instances":                                dataSourceAliCloudDdoscooInstances(),
			"alicloud_ddoscoo_ports":                                    dataSourceAliCloudDdoscooPorts(),
			"alicloud_dfs_access_groups":                                dataSourceAliCloudDfsAccessGroups(),
			"alicloud_dfs_access_rules":                                 dataSourceAliCloudDfsAccessRules(),
			"alicloud_dfs_file_systems":                                 dataSourceAliCloudDfsFileSystems(),
			"alicloud_dfs_mount_points":                                 dataSourceAliCloudDfsMountPoints(),
			"alicloud_dfs_zones":                                        dataSourceAliCloudDfsZones(),
			"alicloud_direct_mail_domains":                              dataSourceAliCloudDirectMailDomains(),
			"alicloud_direct_mail_mail_addresses":                       dataSourceAliCloudDirectMailMailAddresses(),
			"alicloud_direct_mail_receiverses":                          dataSourceAliCloudDirectMailReceiverses(),
			"alicloud_direct_mail_tags":                                 dataSourceAliCloudDirectMailTags(),
			"alicloud_disks":                                            dataSourceAliCloudEcsDisks(),
			"alicloud_dms_enterprise_databases":                         dataSourceAliCloudDmsEnterpriseDatabases(),
			"alicloud_dms_enterprise_instances":                         dataSourceAliCloudDmsEnterpriseInstances(),
			"alicloud_dms_enterprise_logic_databases":                   dataSourceAliCloudDmsEnterpriseLogicDatabases(),
			"alicloud_dms_enterprise_proxies":                           dataSourceAliCloudDmsEnterpriseProxies(),
			"alicloud_dms_enterprise_proxy_accesses":                    dataSourceAliCloudDmsEnterpriseProxyAccesses(),
			"alicloud_dms_enterprise_users":                             dataSourceAliCloudDmsEnterpriseUsers(),
			"alicloud_dms_user_tenants":                                 dataSourceAliCloudDmsUserTenants(),
			"alicloud_dns_domain_groups":                                dataSourceAliCloudDnsGroups(),
			"alicloud_dns_domain_records":                               dataSourceAliCloudDnsRecords(),
			"alicloud_dns_domain_txt_guid":                              dataSourceAliCloudDnsDomainTxtGuid(),
			"alicloud_dns_domains":                                      dataSourceAliCloudAlidnsDomains(),
			"alicloud_dns_groups":                                       dataSourceAliCloudDnsGroups(),
			"alicloud_dns_instances":                                    dataSourceAliCloudAlidnsInstances(),
			"alicloud_dns_records":                                      dataSourceAliCloudDnsRecords(),
			"alicloud_dns_resolution_lines":                             dataSourceAliCloudDnsResolutionLines(),
			"alicloud_drds_instances":                                   dataSourceAliCloudDRDSInstances(),
			"alicloud_dts_consumer_channels":                            dataSourceAliCloudDtsConsumerChannels(),
			"alicloud_dts_instances":                                    dataSourceAliCloudDtsInstances(),
			"alicloud_dts_migration_jobs":                               dataSourceAliCloudDtsMigrationJobs(),
			"alicloud_dts_subscription_jobs":                            dataSourceAliCloudDtsSubscriptionJobs(),
			"alicloud_dts_synchronization_jobs":                         dataSourceAliCloudDtsSynchronizationJobs(),
			"alicloud_eais_instances":                                   dataSourceAliCloudEaisInstances(),
			"alicloud_ebs_dedicated_block_storage_clusters":             dataSourceAliCloudEbsDedicatedBlockStorageClusters(),
			"alicloud_ebs_disk_replica_groups":                          dataSourceAliCloudEbsDiskReplicaGroups(),
			"alicloud_ebs_disk_replica_pairs":                           dataSourceAliCloudEbsDiskReplicaPairs(),
			"alicloud_ebs_regions":                                      dataSourceAliCloudEbsRegions(),
			"alicloud_ecd_ad_connector_directories":                     dataSourceAliCloudEcdAdConnectorDirectories(),
			"alicloud_ecd_ad_connector_office_sites":                    dataSourceAliCloudEcdAdConnectorOfficeSites(),
			"alicloud_ecd_bundles":                                      dataSourceAliCloudEcdBundles(),
			"alicloud_ecd_commands":                                     dataSourceAliCloudEcdCommands(),
			"alicloud_ecd_custom_properties":                            dataSourceAliCloudEcdCustomProperties(),
			"alicloud_ecd_desktop_types":                                dataSourceAliCloudEcdDesktopTypes(),
			"alicloud_ecd_desktops":                                     dataSourceAliCloudEcdDesktops(),
			"alicloud_ecd_images":                                       dataSourceAliCloudEcdImages(),
			"alicloud_ecd_nas_file_systems":                             dataSourceAliCloudEcdNasFileSystems(),
			"alicloud_ecd_network_packages":                             dataSourceAliCloudEcdNetworkPackages(),
			"alicloud_ecd_policy_groups":                                dataSourceAliCloudEcdPolicyGroups(),
			"alicloud_ecd_ram_directories":                              dataSourceAliCloudEcdRamDirectories(),
			"alicloud_ecd_simple_office_sites":                          dataSourceAliCloudEcdSimpleOfficeSites(),
			"alicloud_ecd_snapshots":                                    dataSourceAliCloudEcdSnapshots(),
			"alicloud_ecd_users":                                        dataSourceAliCloudEcdUsers(),
			"alicloud_ecd_zones":                                        dataSourceAliCloudEcdZones(),
			"alicloud_eci_container_groups":                             dataSourceAliCloudEciContainerGroups(),
			"alicloud_eci_image_caches":                                 dataSourceAliCloudEciImageCaches(),
			"alicloud_eci_virtual_nodes":                                dataSourceAliCloudEciVirtualNodes(),
			"alicloud_eci_zones":                                        dataSourceAliCloudEciZones(),
			"alicloud_ecp_instance_types":                               dataSourceAliCloudEcpInstanceTypes(),
			"alicloud_ecp_instances":                                    dataSourceAliCloudEcpInstances(),
			"alicloud_ecp_key_pairs":                                    dataSourceAliCloudEcpKeyPairs(),
			"alicloud_ecp_zones":                                        dataSourceAliCloudEcpZones(),
			"alicloud_ecs_activations":                                  dataSourceAliCloudEcsActivations(),
			"alicloud_ecs_auto_snapshot_policies":                       dataSourceAliCloudEcsAutoSnapshotPolicies(),
			"alicloud_ecs_capacity_reservations":                        dataSourceAliCloudEcsCapacityReservations(),
			"alicloud_ecs_commands":                                     dataSourceAliCloudEcsCommands(),
			"alicloud_ecs_dedicated_host_clusters":                      dataSourceAliCloudEcsDedicatedHostClusters(),
			"alicloud_ecs_dedicated_hosts":                              dataSourceAliCloudEcsDedicatedHosts(),
			"alicloud_ecs_deployment_sets":                              dataSourceAliCloudEcsDeploymentSets(),
			"alicloud_ecs_disks":                                        dataSourceAliCloudEcsDisks(),
			"alicloud_ecs_elasticity_assurances":                        dataSourceAliCloudEcsElasticityAssurances(),
			"alicloud_ecs_hpc_clusters":                                 dataSourceAliCloudEcsHpcClusters(),
			"alicloud_ecs_image_components":                             dataSourceAliCloudEcsImageComponents(),
			"alicloud_ecs_image_pipelines":                              dataSourceAliCloudEcsImagePipelines(),
			"alicloud_ecs_invocations":                                  dataSourceAliCloudEcsInvocations(),
			"alicloud_ecs_key_pairs":                                    dataSourceAliCloudEcsKeyPairs(),
			"alicloud_ecs_launch_templates":                             dataSourceAliCloudEcsLaunchTemplates(),
			"alicloud_ecs_network_interface_permissions":                dataSourceAliCloudEcsNetworkInterfacePermissions(),
			"alicloud_ecs_network_interfaces":                           dataSourceAliCloudEcsNetworkInterfaces(),
			"alicloud_ecs_prefix_lists":                                 dataSourceAliCloudEcsPrefixLists(),
			"alicloud_ecs_snapshot_groups":                              dataSourceAliCloudEcsSnapshotGroups(),
			"alicloud_ecs_snapshots":                                    dataSourceAliCloudEcsSnapshots(),
			"alicloud_ecs_storage_capacity_units":                       dataSourceAliCloudEcsStorageCapacityUnits(),
			"alicloud_edas_applications":                                dataSourceAliCloudEdasApplications(),
			"alicloud_edas_clusters":                                    dataSourceAliCloudEdasClusters(),
			"alicloud_edas_deploy_groups":                               dataSourceAliCloudEdasDeployGroups(),
			"alicloud_edas_namespaces":                                  dataSourceAliCloudEdasNamespaces(),
			"alicloud_edas_service":                                     dataSourceAliCloudEdasService(),
			"alicloud_eflo_subnets":                                     dataSourceAliCloudEfloSubnets(),
			"alicloud_eflo_vpds":                                        dataSourceAliCloudEfloVpds(),
			"alicloud_ehpc_clusters":                                    dataSourceAliCloudEhpcClusters(),
			"alicloud_ehpc_job_templates":                               dataSourceAliCloudEhpcJobTemplates(),
			"alicloud_eip_addresses":                                    dataSourceAliCloudEipAddresses(),
			"alicloud_eipanycast_anycast_eip_addresses":                 dataSourceAliCloudEipanycastAnycastEipAddresses(),
			"alicloud_eips":                                             dataSourceAliCloudEipAddresses(),
			"alicloud_elasticsearch_instances":                          dataSourceAliCloudElasticsearch(),
			"alicloud_elasticsearch_zones":                              dataSourceAliCloudElaticsearchZones(),
			"alicloud_emr_clusters":                                     dataSourceAliCloudEmrClusters(),
			"alicloud_emr_disk_types":                                   dataSourceAliCloudEmrDiskTypes(),
			"alicloud_emr_instance_types":                               dataSourceAliCloudEmrInstanceTypes(),
			"alicloud_emr_main_versions":                                dataSourceAliCloudEmrMainVersions(),
			"alicloud_emrv2_cluster_instances":                          dataSourceAliCloudEmrV2ClusterInstances(),
			"alicloud_emrv2_clusters":                                   dataSourceAliCloudEmrV2Clusters(),
			"alicloud_enhanced_nat_available_zones":                     dataSourceAliCloudEnhancedNatAvailableZones(),
			"alicloud_ens_key_pairs":                                    dataSourceAliCloudEnsKeyPairs(),
			"alicloud_esa_sites":                                        dataSourceAliCloudEsaSites(),
			"alicloud_ess_alarms":                                       dataSourceAliCloudEssAlarms(),
			"alicloud_ess_lifecycle_hooks":                              dataSourceAliCloudEssLifecycleHooks(),
			"alicloud_ess_notifications":                                dataSourceAliCloudEssNotifications(),
			"alicloud_ess_scaling_configurations":                       dataSourceAliCloudEssScalingConfigurations(),
			"alicloud_ess_scaling_groups":                               dataSourceAliCloudEssScalingGroups(),
			"alicloud_ess_scaling_rules":                                dataSourceAliCloudEssScalingRules(),
			"alicloud_ess_scheduled_tasks":                              dataSourceAliCloudEssScheduledTasks(),
			"alicloud_event_bridge_event_buses":                         dataSourceAliCloudEventBridgeEventBuses(),
			"alicloud_event_bridge_event_sources":                       dataSourceAliCloudEventBridgeEventSources(),
			"alicloud_event_bridge_rules":                               dataSourceAliCloudEventBridgeRules(),
			"alicloud_event_bridge_service":                             dataSourceAliCloudEventBridgeService(),
			"alicloud_express_connect_access_points":                    dataSourceAliCloudExpressConnectAccessPoints(),
			"alicloud_express_connect_grant_rule_to_cens":               dataSourceAliCloudExpressConnectGrantRuleToCens(),
			"alicloud_express_connect_physical_connection_service":      dataSourceAliCloudExpressConnectPhysicalConnectionService(),
			"alicloud_express_connect_physical_connections":             dataSourceAliCloudExpressConnectPhysicalConnections(),
			"alicloud_express_connect_router_interfaces":                dataSourceAliCloudExpressConnectRouterInterfaces(),
			"alicloud_express_connect_vbr_pconn_associations":           dataSourceAliCloudExpressConnectVbrPconnAssociations(),
			"alicloud_express_connect_virtual_border_routers":           dataSourceAliCloudExpressConnectVirtualBorderRouters(),
			"alicloud_express_connect_virtual_physical_connections":     dataSourceAliCloudExpressConnectVirtualPhysicalConnections(),
			"alicloud_fc_custom_domains":                                dataSourceAliCloudFCCustomDomains(),
			"alicloud_fc_functions":                                     dataSourceAliCloudFCFunctions(),
			"alicloud_fc_zones":                                         dataSourceAliCloudFCZones(),
			"alicloud_fc_triggers":                                      dataSourceAliCloudFCTriggers(),
			"alicloud_file_crc64_checksum":                              dataSourceAliCloudFileCRC64Checksum(),
			"alicloud_flink_deployment_folders":                         dataSourceAliCloudFlinkDeploymentFolders(),
			"alicloud_flink_deployment_targets":                         dataSourceAliCloudFlinkDeploymentTargets(),
			"alicloud_flink_deployments":                                dataSourceAliCloudFlinkDeployments(),
			"alicloud_flink_engines":                                    dataSourceAliCloudFlinkEngines(),
			"alicloud_flink_members":                                    dataSourceAliCloudFlinkMembers(),
			"alicloud_flink_namespaces":                                 dataSourceAliCloudFlinkNamespaces(),
			"alicloud_flink_variables":                                  dataSourceAliCloudFlinkVariables(),
			"alicloud_flink_workspaces":                                 dataSourceAliCloudFlinkWorkspaces(),
			"alicloud_flink_zones":                                      dataSourceAliCloudFlinkZones(),
			"alicloud_fnf_executions":                                   dataSourceAliCloudFnFExecutions(),
			"alicloud_fnf_flows":                                        dataSourceAliCloudFnfFlows(),
			"alicloud_fnf_schedules":                                    dataSourceAliCloudFnfSchedules(),
			"alicloud_fnf_service":                                      dataSourceAliCloudFnfService(),
			"alicloud_forward_entries":                                  dataSourceAliCloudForwardEntries(),
			"alicloud_ga_accelerator_spare_ip_attachments":              dataSourceAliCloudGaAcceleratorSpareIpAttachments(),
			"alicloud_ga_accelerators":                                  dataSourceAliCloudGaAccelerators(),
			"alicloud_ga_acls":                                          dataSourceAliCloudGaAcls(),
			"alicloud_ga_additional_certificates":                       dataSourceAliCloudGaAdditionalCertificates(),
			"alicloud_ga_bandwidth_packages":                            dataSourceAliCloudGaBandwidthPackages(),
			"alicloud_ga_basic_accelerate_ip_endpoint_relations":        dataSourceAliCloudGaBasicAccelerateIpEndpointRelations(),
			"alicloud_ga_basic_accelerate_ips":                          dataSourceAliCloudGaBasicAccelerateIps(),
			"alicloud_ga_basic_accelerators":                            dataSourceAliCloudGaBasicAccelerators(),
			"alicloud_ga_basic_endpoints":                               dataSourceAliCloudGaBasicEndpoints(),
			"alicloud_ga_custom_routing_endpoint_group_destinations":    dataSourceAliCloudGaCustomRoutingEndpointGroupDestinations(),
			"alicloud_ga_custom_routing_endpoint_groups":                dataSourceAliCloudGaCustomRoutingEndpointGroups(),
			"alicloud_ga_custom_routing_endpoint_traffic_policies":      dataSourceAliCloudGaCustomRoutingEndpointTrafficPolicies(),
			"alicloud_ga_custom_routing_endpoints":                      dataSourceAliCloudGaCustomRoutingEndpoints(),
			"alicloud_ga_custom_routing_port_mappings":                  dataSourceAliCloudGaCustomRoutingPortMappings(),
			"alicloud_ga_domains":                                       dataSourceAliCloudGaDomains(),
			"alicloud_ga_endpoint_group_ip_address_cidr_blocks":         dataSourceAliCloudGaEndpointGroupIpAddressCidrBlocks(),
			"alicloud_ga_endpoint_groups":                               dataSourceAliCloudGaEndpointGroups(),
			"alicloud_ga_forwarding_rules":                              dataSourceAliCloudGaForwardingRules(),
			"alicloud_ga_ip_sets":                                       dataSourceAliCloudGaIpSets(),
			"alicloud_ga_listeners":                                     dataSourceAliCloudGaListeners(),
			"alicloud_governance_baselines":                             dataSourceAliCloudGovernanceBaselines(),
			"alicloud_gpdb_accounts":                                    dataSourceAliCloudGpdbAccounts(),
			"alicloud_gpdb_data_backups":                                dataSourceAliCloudGpdbDataBackups(),
			"alicloud_gpdb_db_instance_plans":                           dataSourceAliCloudGpdbDbInstancePlans(),
			"alicloud_gpdb_instances":                                   dataSourceAliCloudGpdbInstances(),
			"alicloud_gpdb_log_backups":                                 dataSourceAliCloudGpdbLogbackups(),
			"alicloud_gpdb_zones":                                       dataSourceAliCloudGpdbZones(),
			"alicloud_graph_database_db_instances":                      dataSourceAliCloudGraphDatabaseDbInstances(),
			"alicloud_gwlb_zones":                                       dataSourceAliCloudGwlbZones(),
			"alicloud_havips":                                           dataSourceAliCloudHavips(),
			"alicloud_hbase_instance_types":                             dataSourceAliCloudHBaseInstanceTypes(),
			"alicloud_hbase_instances":                                  dataSourceAliCloudHBaseInstances(),
			"alicloud_hbase_zones":                                      dataSourceAliCloudHBaseZones(),
			"alicloud_hbr_backup_jobs":                                  dataSourceAliCloudHbrBackupJobs(),
			"alicloud_hbr_ecs_backup_clients":                           dataSourceAliCloudHbrEcsBackupClients(),
			"alicloud_hbr_ecs_backup_plans":                             dataSourceAliCloudHbrEcsBackupPlans(),
			"alicloud_hbr_hana_backup_clients":                          dataSourceAliCloudHbrHanaBackupClients(),
			"alicloud_hbr_hana_backup_plans":                            dataSourceAliCloudHbrHanaBackupPlans(),
			"alicloud_hbr_hana_instances":                               dataSourceAliCloudHbrHanaInstances(),
			"alicloud_hbr_nas_backup_plans":                             dataSourceAliCloudHbrNasBackupPlans(),
			"alicloud_hbr_oss_backup_plans":                             dataSourceAliCloudHbrOssBackupPlans(),
			"alicloud_hbr_ots_backup_plans":                             dataSourceAliCloudHbrOtsBackupPlans(),
			"alicloud_hbr_ots_snapshots":                                dataSourceAliCloudHbrOtsSnapshots(),
			"alicloud_hbr_replication_vault_regions":                    dataSourceAliCloudHbrReplicationVaultRegions(),
			"alicloud_hbr_restore_jobs":                                 dataSourceAliCloudHbrRestoreJobs(),
			"alicloud_hbr_server_backup_plans":                          dataSourceAliCloudHbrServerBackupPlans(),
			"alicloud_hbr_service":                                      dataSourceAliCloudHbrService(),
			"alicloud_hbr_snapshots":                                    dataSourceAliCloudHbrSnapshots(),
			"alicloud_hbr_vaults":                                       dataSourceAliCloudHbrVaults(),
			"alicloud_images":                                           dataSourceAliCloudImages(),
			"alicloud_imm_projects":                                     dataSourceAliCloudImmProjects(),
			"alicloud_imp_app_templates":                                dataSourceAliCloudImpAppTemplates(),
			"alicloud_ims_oidc_providers":                               dataSourceAliCloudImsOidcProviders(),
			"alicloud_instance_keywords":                                dataSourceAliCloudInstanceKeywords(),
			"alicloud_instance_type_families":                           dataSourceAliCloudInstanceTypeFamilies(),
			"alicloud_instance_types":                                   dataSourceAliCloudInstanceTypes(),
			"alicloud_instances":                                        dataSourceAliCloudInstances(),
			"alicloud_iot_device_groups":                                dataSourceAliCloudIotDeviceGroups(),
			"alicloud_iot_service":                                      dataSourceAliCloudIotService(),
			"alicloud_key_pairs":                                        dataSourceAliCloudEcsKeyPairs(),
			"alicloud_kms_aliases":                                      dataSourceAliCloudKmsAliases(),
			"alicloud_kms_ciphertext":                                   dataSourceAliCloudKmsCiphertext(),
			"alicloud_kms_instances":                                    dataSourceAliCloudKmsInstances(),
			"alicloud_kms_key_versions":                                 dataSourceAliCloudKmsKeyVersions(),
			"alicloud_kms_keys":                                         dataSourceAliCloudKmsKeys(),
			"alicloud_kms_plaintext":                                    dataSourceAliCloudKmsPlaintext(),
			"alicloud_kms_secret_versions":                              dataSourceAliCloudKmsSecretVersions(),
			"alicloud_kms_secrets":                                      dataSourceAliCloudKmsSecrets(),
			"alicloud_kms_service":                                      dataSourceAliCloudKmsService(),
			"alicloud_kvstore_accounts":                                 dataSourceAliCloudKvstoreAccounts(),
			"alicloud_kvstore_connections":                              dataSourceAliCloudKvstoreConnections(),
			"alicloud_kvstore_instance_classes":                         dataSourceAliCloudKVStoreInstanceClasses(),
			"alicloud_kvstore_instance_engines":                         dataSourceAliCloudKVStoreInstanceEngines(),
			"alicloud_kvstore_instances":                                dataSourceAliCloudKvstoreInstances(),
			"alicloud_kvstore_permission":                               dataSourceAliCloudKVStorePermission(),
			"alicloud_kvstore_zones":                                    dataSourceAliCloudKVStoreZones(),
			"alicloud_lindorm_instances":                                dataSourceAliCloudLindormInstances(),
			"alicloud_log_alert_resource":                               dataSourceAliCloudLogAlertResource(),
			"alicloud_log_logtail_configs":                              dataSourceAliCloudLogLogtailConfigs(),
			"alicloud_log_machine_groups":                               dataSourceAliCloudLogMachineGroups(),
			"alicloud_log_projects":                                     dataSourceAliCloudLogProjects(),
			"alicloud_log_query":                                        dataSourceAliCloudLogQuery(),
			"alicloud_log_service":                                      dataSourceAliCloudLogService(),
			"alicloud_log_store_indexes":                                dataSourceAliCloudLogStoreIndexes(),
			"alicloud_log_stores":                                       dataSourceAliCloudLogStores(),
			"alicloud_market_product":                                   dataSourceAliCloudProduct(),
			"alicloud_market_products":                                  dataSourceAliCloudProducts(),
			"alicloud_maxcompute_projects":                              dataSourceAliCloudMaxComputeProjects(),
			"alicloud_maxcompute_service":                               dataSourceAliCloudMaxcomputeService(),
			"alicloud_message_service_queues":                           dataSourceAliCloudMessageServiceQueues(),
			"alicloud_message_service_subscriptions":                    dataSourceAliCloudMessageServiceSubscriptions(),
			"alicloud_message_service_topics":                           dataSourceAliCloudMessageServiceTopics(),
			"alicloud_mhub_apps":                                        dataSourceAliCloudMhubApps(),
			"alicloud_mhub_products":                                    dataSourceAliCloudMhubProducts(),
			"alicloud_mns_queues":                                       dataSourceAliCloudMNSQueues(),
			"alicloud_mns_service":                                      dataSourceAliCloudMnsService(),
			"alicloud_mns_topic_subscriptions":                          dataSourceAliCloudMNSTopicSubscriptions(),
			"alicloud_mns_topics":                                       dataSourceAliCloudMNSTopics(),
			"alicloud_mongo_instances":                                  dataSourceAliCloudMongoDBInstances(),
			"alicloud_mongodb_accounts":                                 dataSourceAliCloudMongodbAccounts(),
			"alicloud_mongodb_audit_policies":                           dataSourceAliCloudMongodbAuditPolicies(),
			"alicloud_mongodb_instances":                                dataSourceAliCloudMongoDBInstances(),
			"alicloud_mongodb_serverless_instances":                     dataSourceAliCloudMongodbServerlessInstances(),
			"alicloud_mongodb_sharding_network_private_addresses":       dataSourceAliCloudMongodbShardingNetworkPrivateAddresses(),
			"alicloud_mongodb_sharding_network_public_addresses":        dataSourceAliCloudMongodbShardingNetworkPublicAddresses(),
			"alicloud_mongodb_zones":                                    dataSourceAliCloudMongoDBZones(),
			"alicloud_msc_sub_contact_verification_message":             dataSourceAliCloudMscSubContactVerificationMessage(),
			"alicloud_msc_sub_contacts":                                 dataSourceAliCloudMscSubContacts(),
			"alicloud_msc_sub_subscriptions":                            dataSourceAliCloudMscSubSubscriptions(),
			"alicloud_msc_sub_webhooks":                                 dataSourceAliCloudMscSubWebhooks(),
			"alicloud_mse_clusters":                                     dataSourceAliCloudMseClusters(),
			"alicloud_mse_engine_namespaces":                            dataSourceAliCloudMseEngineNamespaces(),
			"alicloud_mse_gateways":                                     dataSourceAliCloudMseGateways(),
			"alicloud_mse_nacos_configs":                                dataSourceAliCloudMseNacosConfigs(),
			"alicloud_mse_znodes":                                       dataSourceAliCloudMseZnodes(),
			"alicloud_nas_access_groups":                                dataSourceAliCloudNasAccessGroups(),
			"alicloud_nas_access_rules":                                 dataSourceAliCloudAccessRules(),
			"alicloud_nas_auto_snapshot_policies":                       dataSourceAliCloudNasAutoSnapshotPolicies(),
			"alicloud_nas_data_flows":                                   dataSourceAliCloudNasDataFlows(),
			"alicloud_nas_file_systems":                                 dataSourceAliCloudFileSystems(),
			"alicloud_nas_filesets":                                     dataSourceAliCloudNasFilesets(),
			"alicloud_nas_lifecycle_policies":                           dataSourceAliCloudNasLifecyclePolicies(),
			"alicloud_nas_mount_targets":                                dataSourceAliCloudNasMountTargets(),
			"alicloud_nas_protocols":                                    dataSourceAliCloudNasProtocols(),
			"alicloud_nas_service":                                      dataSourceAliCloudNasService(),
			"alicloud_nas_snapshots":                                    dataSourceAliCloudNasSnapshots(),
			"alicloud_nas_zones":                                        dataSourceAliCloudNasZones(),
			"alicloud_nat_gateways":                                     dataSourceAliCloudNatGateways(),
			"alicloud_network_acls":                                     dataSourceAliCloudNetworkAcls(),
			"alicloud_network_interfaces":                               dataSourceAliCloudEcsNetworkInterfaces(),
			"alicloud_nlb_listeners":                                    dataSourceAliCloudNlbListeners(),
			"alicloud_nlb_load_balancers":                               dataSourceAliCloudNlbLoadBalancers(),
			"alicloud_nlb_security_policies":                            dataSourceAliCloudNlbSecurityPolicies(),
			"alicloud_nlb_server_group_server_attachments":              dataSourceAliCloudNlbServerGroupServerAttachments(),
			"alicloud_nlb_server_groups":                                dataSourceAliCloudNlbServerGroups(),
			"alicloud_nlb_zones":                                        dataSourceAliCloudNlbZones(),
			"alicloud_ocean_base_instances":                             dataSourceAliCloudOceanBaseInstances(),
			"alicloud_ons_groups":                                       dataSourceAliCloudOnsGroups(),
			"alicloud_ons_instances":                                    dataSourceAliCloudOnsInstances(),
			"alicloud_ons_service":                                      dataSourceAliCloudOnsService(),
			"alicloud_ons_topics":                                       dataSourceAliCloudOnsTopics(),
			"alicloud_oos_application_groups":                           dataSourceAliCloudOosApplicationGroups(),
			"alicloud_oos_applications":                                 dataSourceAliCloudOosApplications(),
			"alicloud_oos_executions":                                   dataSourceAliCloudOosExecutions(),
			"alicloud_oos_parameters":                                   dataSourceAliCloudOosParameters(),
			"alicloud_oos_patch_baselines":                              dataSourceAliCloudOosPatchBaselines(),
			"alicloud_oos_secret_parameters":                            dataSourceAliCloudOosSecretParameters(),
			"alicloud_oos_state_configurations":                         dataSourceAliCloudOosStateConfigurations(),
			"alicloud_oos_templates":                                    dataSourceAliCloudOosTemplates(),
			"alicloud_open_search_app_groups":                           dataSourceAliCloudOpenSearchAppGroups(),
			"alicloud_oss_bucket_objects":                               dataSourceAliCloudOssBucketObjects(),
			"alicloud_oss_buckets":                                      dataSourceAliCloudOssBuckets(),
			"alicloud_oss_service":                                      dataSourceAliCloudOssService(),
			"alicloud_ots_indexes":                                      dataSourceAliCloudOtsIndexes(),
			"alicloud_ots_instance_attachments":                         dataSourceAliCloudOtsInstanceAttachments(),
			"alicloud_ots_instances":                                    dataSourceAliCloudOtsInstances(),
			"alicloud_ots_search_indexes":                               dataSourceAliCloudOtsSearchIndexes(),
			"alicloud_ots_service":                                      dataSourceAliCloudOtsService(),
			"alicloud_ots_tables":                                       dataSourceAliCloudOtsTables(),
			"alicloud_polardb_accounts":                                 dataSourceAliCloudPolarDBAccounts(),
			"alicloud_polardb_clusters":                                 dataSourceAliCloudPolarDBClusters(),
			"alicloud_polardb_databases":                                dataSourceAliCloudPolarDBDatabases(),
			"alicloud_polardb_endpoints":                                dataSourceAliCloudPolarDBEndpoints(),
			"alicloud_polardb_global_database_networks":                 dataSourceAliCloudPolarDBGlobalDatabaseNetworks(),
			"alicloud_polardb_node_classes":                             dataSourceAliCloudPolarDBNodeClasses(),
			"alicloud_polardb_parameter_groups":                         dataSourceAliCloudPolarDBParameterGroups(),
			"alicloud_polardb_zones":                                    dataSourceAliCloudPolarDBZones(),
			"alicloud_privatelink_service":                              dataSourceAliCloudPrivateLinkService(),
			"alicloud_privatelink_vpc_endpoint_connections":             dataSourceAliCloudPrivatelinkVpcEndpointConnections(),
			"alicloud_privatelink_vpc_endpoint_service_resources":       dataSourceAliCloudPrivatelinkVpcEndpointServiceResources(),
			"alicloud_privatelink_vpc_endpoint_service_users":           dataSourceAliCloudPrivatelinkVpcEndpointServiceUsers(),
			"alicloud_privatelink_vpc_endpoint_services":                dataSourceAliCloudPrivatelinkVpcEndpointServices(),
			"alicloud_privatelink_vpc_endpoint_zones":                   dataSourceAliCloudPrivatelinkVpcEndpointZones(),
			"alicloud_privatelink_vpc_endpoints":                        dataSourceAliCloudPrivatelinkVpcEndpoints(),
			"alicloud_pvtz_endpoints":                                   dataSourceAliCloudPvtzEndpoints(),
			"alicloud_pvtz_resolver_zones":                              dataSourceAliCloudPvtzResolverZones(),
			"alicloud_pvtz_rules":                                       dataSourceAliCloudPvtzRules(),
			"alicloud_pvtz_service":                                     dataSourceAliCloudPvtzService(),
			"alicloud_pvtz_zone_records":                                dataSourceAliCloudPvtzZoneRecords(),
			"alicloud_pvtz_zones":                                       dataSourceAliCloudPvtzZones(),
			"alicloud_quick_bi_users":                                   dataSourceAliCloudQuickBiUsers(),
			"alicloud_quotas_application_infos":                         dataSourceAliCloudQuotasQuotaApplications(),
			"alicloud_quotas_quota_alarms":                              dataSourceAliCloudQuotasQuotaAlarms(),
			"alicloud_quotas_quota_applications":                        dataSourceAliCloudQuotasQuotaApplications(),
			"alicloud_quotas_quotas":                                    dataSourceAliCloudQuotasQuotas(),
			"alicloud_quotas_template_applications":                     dataSourceAliCloudQuotasTemplateApplications(),
			"alicloud_ram_account_alias":                                dataSourceAliCloudRamAccountAlias(),
			"alicloud_ram_account_aliases":                              dataSourceAliCloudRamAccountAlias(),
			"alicloud_ram_groups":                                       dataSourceAliCloudRamGroups(),
			"alicloud_ram_policies":                                     dataSourceAliCloudRamPolicies(),
			"alicloud_ram_policy_document":                              dataSourceAliCloudRamPolicyDocument(),
			"alicloud_ram_role_policy_attachments":                      dataSourceAliCloudRamRolePolicyAttachments(),
			"alicloud_ram_roles":                                        dataSourceAliCloudRamRoles(),
			"alicloud_ram_saml_providers":                               dataSourceAliCloudRamSamlProviders(),
			"alicloud_ram_system_policys":                               dataSourceAliCloudRamSystemPolicys(),
			"alicloud_ram_users":                                        dataSourceAliCloudRamUsers(),
			"alicloud_rdc_organizations":                                dataSourceAliCloudRdcOrganizations(),
			"alicloud_rds_accounts":                                     dataSourceAliCloudRdsAccounts(),
			"alicloud_rds_backups":                                      dataSourceAliCloudRdsBackups(),
			"alicloud_rds_character_set_names":                          dataSourceAliCloudRdsCharacterSetNames(),
			"alicloud_rds_class_details":                                dataSourceAliCloudRdsClassDetails(),
			"alicloud_rds_collation_time_zones":                         dataSourceAliCloudRdsCollationTimeZones(),
			"alicloud_rds_cross_region_backups":                         dataSourceAliCloudRdsCrossRegionBackups(),
			"alicloud_rds_cross_regions":                                dataSourceAliCloudRdsCrossRegions(),
			"alicloud_rds_modify_parameter_logs":                        dataSourceAliCloudRdsModifyParameterLogs(),
			"alicloud_rds_parameter_groups":                             dataSourceAliCloudRdsParameterGroups(),
			"alicloud_rds_slots":                                        dataSourceAliCloudRdsSlots(),
			"alicloud_regions":                                          dataSourceAliCloudRegions(),
			"alicloud_resource_manager_account_deletion_check_task":     dataSourceAliCloudResourceManagerAccountDeletionCheckTask(),
			"alicloud_resource_manager_accounts":                        dataSourceAliCloudResourceManagerAccounts(),
			"alicloud_resource_manager_control_policies":                dataSourceAliCloudResourceManagerControlPolicies(),
			"alicloud_resource_manager_control_policy_attachments":      dataSourceAliCloudResourceManagerControlPolicyAttachments(),
			"alicloud_resource_manager_delegated_administrators":        dataSourceAliCloudResourceManagerDelegatedAdministrators(),
			"alicloud_resource_manager_folders":                         dataSourceAliCloudResourceManagerFolders(),
			"alicloud_resource_manager_handshakes":                      dataSourceAliCloudResourceManagerHandshakes(),
			"alicloud_resource_manager_policies":                        dataSourceAliCloudResourceManagerPolicies(),
			"alicloud_resource_manager_policy_attachments":              dataSourceAliCloudResourceManagerPolicyAttachments(),
			"alicloud_resource_manager_policy_versions":                 dataSourceAliCloudResourceManagerPolicyVersions(),
			"alicloud_resource_manager_resource_directories":            dataSourceAliCloudResourceManagerResourceDirectories(),
			"alicloud_resource_manager_resource_groups":                 dataSourceAliCloudResourceManagerResourceGroups(),
			"alicloud_resource_manager_resource_shares":                 dataSourceAliCloudResourceManagerResourceShares(),
			"alicloud_resource_manager_roles":                           dataSourceAliCloudResourceManagerRoles(),
			"alicloud_resource_manager_shared_resources":                dataSourceAliCloudResourceManagerSharedResources(),
			"alicloud_resource_manager_shared_targets":                  dataSourceAliCloudResourceManagerSharedTargets(),
			"alicloud_ros_change_sets":                                  dataSourceAliCloudRosChangeSets(),
			"alicloud_ros_regions":                                      dataSourceAliCloudRosRegions(),
			"alicloud_ros_stack_groups":                                 dataSourceAliCloudRosStackGroups(),
			"alicloud_ros_stack_instances":                              dataSourceAliCloudRosStackInstances(),
			"alicloud_ros_stacks":                                       dataSourceAliCloudRosStacks(),
			"alicloud_ros_template_scratches":                           dataSourceAliCloudRosTemplateScratches(),
			"alicloud_ros_templates":                                    dataSourceAliCloudRosTemplates(),
			"alicloud_route_entries":                                    dataSourceAliCloudRouteEntries(),
			"alicloud_route_tables":                                     dataSourceAliCloudRouteTables(),
			"alicloud_router_interfaces":                                dataSourceAliCloudRouterInterfaces(),
			"alicloud_sae_application_scaling_rules":                    dataSourceAliCloudSaeApplicationScalingRules(),
			"alicloud_sae_applications":                                 dataSourceAliCloudSaeApplications(),
			"alicloud_sae_config_maps":                                  dataSourceAliCloudSaeConfigMaps(),
			"alicloud_sae_grey_tag_routes":                              dataSourceAliCloudSaeGreyTagRoutes(),
			"alicloud_sae_ingresses":                                    dataSourceAliCloudSaeIngresses(),
			"alicloud_sae_instance_specifications":                      dataSourceAliCloudSaeInstanceSpecifications(),
			"alicloud_sae_namespaces":                                   dataSourceAliCloudSaeNamespaces(),
			"alicloud_sae_service":                                      dataSourceAliCloudSaeService(),
			"alicloud_sag_acls":                                         dataSourceAliCloudSagAcls(),
			"alicloud_scdn_domains":                                     dataSourceAliCloudScdnDomains(),
			"alicloud_schedulerx_namespaces":                            dataSourceAliCloudSchedulerxNamespaces(),
			"alicloud_sddp_configs":                                     dataSourceAliCloudSddpConfigs(),
			"alicloud_sddp_data_limits":                                 dataSourceAliCloudSddpDataLimits(),
			"alicloud_sddp_instances":                                   dataSourceAliCloudSddpInstances(),
			"alicloud_sddp_rules":                                       dataSourceAliCloudSddpRules(),
			"alicloud_security_center_groups":                           dataSourceAliCloudSecurityCenterGroups(),
			"alicloud_security_group_rules":                             dataSourceAliCloudSecurityGroupRules(),
			"alicloud_security_groups":                                  dataSourceAliCloudSecurityGroups(),
			"alicloud_selectdb_clusters":                                dataSourceAliCloudSelectDBClusters(),
			"alicloud_selectdb_instance_classes":                        dataSourceAliCloudSelectDBInstanceClasses(),
			"alicloud_selectdb_instances":                               dataSourceAliCloudSelectDBInstances(),
			"alicloud_service_catalog_end_user_products":                dataSourceAliCloudServiceCatalogEndUserProducts(),
			"alicloud_service_catalog_launch_options":                   dataSourceAliCloudServiceCatalogLaunchOptions(),
			"alicloud_service_catalog_portfolios":                       dataSourceAliCloudServiceCatalogPortfolios(),
			"alicloud_service_catalog_product_as_end_users":             dataSourceAliCloudServiceCatalogProductAsEndUsers(),
			"alicloud_service_catalog_product_versions":                 dataSourceAliCloudServiceCatalogProductVersions(),
			"alicloud_service_catalog_provisioned_products":             dataSourceAliCloudServiceCatalogProvisionedProducts(),
			"alicloud_service_mesh_extension_providers":                 dataSourceAliCloudServiceMeshExtensionProviders(),
			"alicloud_service_mesh_service_meshes":                      dataSourceAliCloudServiceMeshServiceMeshes(),
			"alicloud_service_mesh_versions":                            dataSourceAliCloudServiceMeshVersions(),
			"alicloud_simple_application_server_custom_images":          dataSourceAliCloudSimpleApplicationServerCustomImages(),
			"alicloud_simple_application_server_disks":                  dataSourceAliCloudSimpleApplicationServerDisks(),
			"alicloud_simple_application_server_firewall_rules":         dataSourceAliCloudSimpleApplicationServerFirewallRules(),
			"alicloud_simple_application_server_images":                 dataSourceAliCloudSimpleApplicationServerImages(),
			"alicloud_simple_application_server_instances":              dataSourceAliCloudSimpleApplicationServerInstances(),
			"alicloud_simple_application_server_plans":                  dataSourceAliCloudSimpleApplicationServerPlans(),
			"alicloud_simple_application_server_snapshots":              dataSourceAliCloudSimpleApplicationServerSnapshots(),
			"alicloud_slb_acls":                                         dataSourceAliCloudSlbAcls(),
			"alicloud_slb_attachments":                                  dataSourceAliCloudSlbAttachments(),
			"alicloud_slb_backend_servers":                              dataSourceAliCloudSlbBackendServers(),
			"alicloud_slb_ca_certificates":                              dataSourceAliCloudSlbCaCertificates(),
			"alicloud_slb_domain_extensions":                            dataSourceAliCloudSlbDomainExtensions(),
			"alicloud_slb_listeners":                                    dataSourceAliCloudSlbListeners(),
			"alicloud_slb_load_balancers":                               dataSourceAliCloudSlbLoadBalancers(),
			"alicloud_slb_master_slave_server_groups":                   dataSourceAliCloudSlbMasterSlaveServerGroups(),
			"alicloud_slb_rules":                                        dataSourceAliCloudSlbRules(),
			"alicloud_slb_server_certificates":                          dataSourceAliCloudSlbServerCertificates(),
			"alicloud_slb_server_groups":                                dataSourceAliCloudSlbServerGroups(),
			"alicloud_slb_tls_cipher_policies":                          dataSourceAliCloudSlbTlsCipherPolicies(),
			"alicloud_slb_zones":                                        dataSourceAliCloudSlbZones(),
			"alicloud_slbs":                                             dataSourceAliCloudSlbLoadBalancers(),
			"alicloud_sls_alerts":                                       dataSourceAliCloudSlsAlerts(),
			"alicloud_smartag_flow_logs":                                dataSourceAliCloudSmartagFlowLogs(),
			"alicloud_snapshots":                                        dataSourceAliCloudEcsSnapshots(),
			"alicloud_snat_entries":                                     dataSourceAliCloudSnatEntries(),
			"alicloud_ssl_certificates_service_certificates":            dataSourceAliCloudSslCertificatesServiceCertificates(),
			"alicloud_ssl_vpn_client_certs":                             dataSourceAliCloudSslVpnClientCerts(),
			"alicloud_ssl_vpn_servers":                                  dataSourceAliCloudSslVpnServers(),
			"alicloud_tag_meta_tags":                                    dataSourceAliCloudTagMetaTags(),
			"alicloud_threat_detection_anti_brute_force_rules":          dataSourceAliCloudThreatDetectionAntiBruteForceRules(),
			"alicloud_threat_detection_assets":                          dataSourceAliCloudThreatDetectionAssets(),
			"alicloud_threat_detection_backup_policies":                 dataSourceAliCloudThreatDetectionBackupPolicies(),
			"alicloud_threat_detection_baseline_strategies":             dataSourceAliCloudThreatDetectionBaselineStrategies(),
			"alicloud_threat_detection_honey_pots":                      dataSourceAliCloudThreatDetectionHoneyPots(),
			"alicloud_threat_detection_honeypot_images":                 dataSourceAliCloudThreatDetectionHoneypotImages(),
			"alicloud_threat_detection_honeypot_nodes":                  dataSourceAliCloudThreatDetectionHoneypotNodes(),
			"alicloud_threat_detection_honeypot_presets":                dataSourceAliCloudThreatDetectionHoneypotPresets(),
			"alicloud_threat_detection_honeypot_probes":                 dataSourceAliCloudThreatDetectionHoneypotProbes(),
			"alicloud_threat_detection_instances":                       dataSourceAliCloudThreatDetectionInstances(),
			"alicloud_threat_detection_log_shipper":                     dataSourceAliCloudThreatDetectionLogShipper(),
			"alicloud_threat_detection_vul_whitelists":                  dataSourceAliCloudThreatDetectionVulWhitelists(),
			"alicloud_threat_detection_web_lock_configs":                dataSourceAliCloudThreatDetectionWebLockConfigs(),
			"alicloud_tsdb_instances":                                   dataSourceAliCloudTsdbInstances(),
			"alicloud_tsdb_zones":                                       dataSourceAliCloudTsdbZones(),
			"alicloud_video_surveillance_system_groups":                 dataSourceAliCloudVideoSurveillanceSystemGroups(),
			"alicloud_vod_domains":                                      dataSourceAliCloudVodDomains(),
			"alicloud_vpc_bgp_groups":                                   dataSourceAliCloudVpcBgpGroups(),
			"alicloud_vpc_bgp_networks":                                 dataSourceAliCloudVpcBgpNetworks(),
			"alicloud_vpc_bgp_peers":                                    dataSourceAliCloudVpcBgpPeers(),
			"alicloud_vpc_dhcp_options_sets":                            dataSourceAliCloudVpcDhcpOptionsSets(),
			"alicloud_vpc_flow_log_service":                             dataSourceAliCloudVpcFlowLogService(),
			"alicloud_vpc_flow_logs":                                    dataSourceAliCloudVpcFlowLogs(),
			"alicloud_vpc_ipam_ipam_pool_allocations":                   dataSourceAliCloudVpcIpamIpamPoolAllocations(),
			"alicloud_vpc_ipam_ipam_pool_cidrs":                         dataSourceAliCloudVpcIpamIpamPoolCidrs(),
			"alicloud_vpc_ipam_ipam_pools":                              dataSourceAliCloudVpcIpamIpamPools(),
			"alicloud_vpc_ipam_ipam_scopes":                             dataSourceAliCloudVpcIpamIpamScopes(),
			"alicloud_vpc_ipam_ipams":                                   dataSourceAliCloudVpcIpamIpams(),
			"alicloud_vpc_ipv4_gateways":                                dataSourceAliCloudVpcIpv4Gateways(),
			"alicloud_vpc_ipv6_addresses":                               dataSourceAliCloudVpcIpv6Addresses(),
			"alicloud_vpc_ipv6_egress_rules":                            dataSourceAliCloudVpcIpv6EgressRules(),
			"alicloud_vpc_ipv6_gateways":                                dataSourceAliCloudVpcIpv6Gateways(),
			"alicloud_vpc_ipv6_internet_bandwidths":                     dataSourceAliCloudVpcIpv6InternetBandwidths(),
			"alicloud_vpc_nat_ip_cidrs":                                 dataSourceAliCloudVpcNatIpCidrs(),
			"alicloud_vpc_nat_ips":                                      dataSourceAliCloudVpcNatIps(),
			"alicloud_vpc_peer_connections":                             dataSourceAliCloudVpcPeerConnections(),
			"alicloud_vpc_prefix_lists":                                 dataSourceAliCloudVpcPrefixLists(),
			"alicloud_vpc_public_ip_address_pool_cidr_blocks":           dataSourceAliCloudVpcPublicIpAddressPoolCidrBlocks(),
			"alicloud_vpc_public_ip_address_pools":                      dataSourceAliCloudVpcPublicIpAddressPools(),
			"alicloud_vpc_traffic_mirror_filter_egress_rules":           dataSourceAliCloudVpcTrafficMirrorFilterEgressRules(),
			"alicloud_vpc_traffic_mirror_filter_ingress_rules":          dataSourceAliCloudVpcTrafficMirrorFilterIngressRules(),
			"alicloud_vpc_traffic_mirror_filters":                       dataSourceAliCloudVpcTrafficMirrorFilters(),
			"alicloud_vpc_traffic_mirror_service":                       dataSourceAliCloudVpcTrafficMirrorService(),
			"alicloud_vpc_traffic_mirror_sessions":                      dataSourceAliCloudVpcTrafficMirrorSessions(),
			"alicloud_vpcs":                                             dataSourceAliCloudVpcs(),
			"alicloud_vpn_connections":                                  dataSourceAliCloudVpnConnections(),
			"alicloud_vpn_customer_gateways":                            dataSourceAliCloudVpnCustomerGateways(),
			"alicloud_vpn_gateway_vco_routes":                           dataSourceAliCloudVpnGatewayVcoRoutes(),
			"alicloud_vpn_gateway_vpn_attachments":                      dataSourceAliCloudVpnGatewayVpnAttachments(),
			"alicloud_vpn_gateway_zones":                                dataSourceAliCloudVPNGatewayZones(),
			"alicloud_vpn_gateways":                                     dataSourceAliCloudVpnGateways(),
			"alicloud_vpn_ipsec_servers":                                dataSourceAliCloudVpnIpsecServers(),
			"alicloud_vpn_pbr_route_entries":                            dataSourceAliCloudVpnPbrRouteEntries(),
			"alicloud_vs_service":                                       dataSourceAliCloudVsService(),
			"alicloud_vswitches":                                        dataSourceAliCloudVswitches(),
			"alicloud_waf_certificates":                                 dataSourceAliCloudWafCertificates(),
			"alicloud_waf_domains":                                      dataSourceAliCloudWafDomains(),
			"alicloud_waf_instances":                                    dataSourceAliCloudWafInstances(),
			"alicloud_wafv3_domains":                                    dataSourceAliCloudWafv3Domains(),
			"alicloud_wafv3_instances":                                  dataSourceAliCloudWafv3Instances(),
			"alicloud_yundun_bastionhost_instances":                     dataSourceAliCloudBastionhostInstances(),
			"alicloud_yundun_dbaudit_instance":                          dataSourceAliCloudDbauditInstances(),
			"alicloud_zones":                                            dataSourceAliCloudZones(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"alicloud_ack_one_cluster":                                       resourceAliCloudAckOneCluster(),
			"alicloud_ack_one_membership_attachment":                         resourceAliCloudAckOneMembershipAttachment(),
			"alicloud_actiontrail":                                           resourceAliCloudActiontrailTrail(),
			"alicloud_actiontrail_global_events_storage_region":              resourceAliCloudActiontrailGlobalEventsStorageRegion(),
			"alicloud_actiontrail_history_delivery_job":                      resourceAliCloudActiontrailHistoryDeliveryJob(),
			"alicloud_actiontrail_trail":                                     resourceAliCloudActiontrailTrail(),
			"alicloud_adb_account":                                           resourceAliCloudAdbAccount(),
			"alicloud_adb_backup_policy":                                     resourceAliCloudAdbBackupPolicy(),
			"alicloud_adb_cluster":                                           resourceAliCloudAdbDbCluster(),
			"alicloud_adb_connection":                                        resourceAliCloudAdbConnection(),
			"alicloud_adb_db_cluster":                                        resourceAliCloudAdbDbCluster(),
			"alicloud_adb_db_cluster_lake_version":                           resourceAliCloudAdbDbClusterLakeVersion(),
			"alicloud_adb_lake_account":                                      resourceAliCloudAdbLakeAccount(),
			"alicloud_adb_resource_group":                                    resourceAliCloudAdbResourceGroup(),
			"alicloud_alb_acl":                                               resourceAliCloudAlbAcl(),
			"alicloud_alb_acl_entry_attachment":                              resourceAliCloudAlbAclEntryAttachment(),
			"alicloud_alb_ascript":                                           resourceAliCloudAlbAScript(),
			"alicloud_alb_health_check_template":                             resourceAliCloudAlbHealthCheckTemplate(),
			"alicloud_alb_listener":                                          resourceAliCloudAlbListener(),
			"alicloud_alb_listener_acl_attachment":                           resourceAliCloudAlbListenerAclAttachment(),
			"alicloud_alb_listener_additional_certificate_attachment":        resourceAliCloudAlbListenerAdditionalCertificateAttachment(),
			"alicloud_alb_load_balancer":                                     resourceAliCloudAlbLoadBalancer(),
			"alicloud_alb_load_balancer_access_log_config_attachment":        resourceAliCloudAlbLoadBalancerAccessLogConfigAttachment(),
			"alicloud_alb_load_balancer_common_bandwidth_package_attachment": resourceAliCloudAlbLoadBalancerCommonBandwidthPackageAttachment(),
			"alicloud_alb_load_balancer_security_group_attachment":           resourceAliCloudAlbLoadBalancerSecurityGroupAttachment(),
			"alicloud_alb_load_balancer_zone_shifted_attachment":             resourceAliCloudAlbLoadBalancerZoneShiftedAttachment(),
			"alicloud_alb_rule":                                              resourceAliCloudAlbRule(),
			"alicloud_alb_security_policy":                                   resourceAliCloudAlbSecurityPolicy(),
			"alicloud_alb_server_group":                                      resourceAliCloudAlbServerGroup(),
			"alicloud_alidns_access_strategy":                                resourceAliCloudAlidnsAccessStrategy(),
			"alicloud_alidns_address_pool":                                   resourceAliCloudAlidnsAddressPool(),
			"alicloud_alidns_custom_line":                                    resourceAliCloudAlidnsCustomLine(),
			"alicloud_alidns_domain":                                         resourceAliCloudAlidnsDomain(),
			"alicloud_alidns_domain_attachment":                              resourceAliCloudAlidnsDomainAttachment(),
			"alicloud_alidns_domain_group":                                   resourceAliCloudAlidnsDomainGroup(),
			"alicloud_alidns_gtm_instance":                                   resourceAliCloudAlidnsGtmInstance(),
			"alicloud_alidns_instance":                                       resourceAliCloudAlidnsInstance(),
			"alicloud_alidns_monitor_config":                                 resourceAliCloudAlidnsMonitorConfig(),
			"alicloud_alidns_record":                                         resourceAliCloudAlidnsRecord(),
			"alicloud_aligreen_audit_callback":                               resourceAliCloudAligreenAuditCallback(),
			"alicloud_aligreen_biz_type":                                     resourceAliCloudAligreenBizType(),
			"alicloud_aligreen_callback":                                     resourceAliCloudAligreenCallback(),
			"alicloud_aligreen_image_lib":                                    resourceAliCloudAligreenImageLib(),
			"alicloud_aligreen_keyword_lib":                                  resourceAliCloudAligreenKeywordLib(),
			"alicloud_aligreen_oss_stock_task":                               resourceAliCloudAligreenOssStockTask(),
			"alicloud_alikafka_consumer_group":                               resourceAliCloudAlikafkaConsumerGroup(),
			"alicloud_alikafka_instance":                                     resourceAliCloudAlikafkaInstance(),
			"alicloud_alikafka_instance_allowed_ip_attachment":               resourceAliCloudAliKafkaInstanceAllowedIpAttachment(),
			"alicloud_alikafka_sasl_acl":                                     resourceAliCloudAlikafkaSaslAcl(),
			"alicloud_alikafka_sasl_user":                                    resourceAliCloudAlikafkaSaslUser(),
			"alicloud_alikafka_topic":                                        resourceAliCloudAlikafkaTopic(),
			"alicloud_amqp_binding":                                          resourceAliCloudAmqpBinding(),
			"alicloud_amqp_exchange":                                         resourceAliCloudAmqpExchange(),
			"alicloud_amqp_instance":                                         resourceAliCloudAmqpInstance(),
			"alicloud_amqp_queue":                                            resourceAliCloudAmqpQueue(),
			"alicloud_amqp_static_account":                                   resourceAliCloudAmqpStaticAccount(),
			"alicloud_amqp_virtual_host":                                     resourceAliCloudAmqpVirtualHost(),
			"alicloud_api_gateway_access_control_list":                       resourceAliCloudApiGatewayAccessControlList(),
			"alicloud_api_gateway_acl_entry_attachment":                      resourceAliCloudApiGatewayAclEntryAttachment(),
			"alicloud_api_gateway_api":                                       resourceAliyunApigatewayApi(),
			"alicloud_api_gateway_app":                                       resourceAliyunApigatewayApp(),
			"alicloud_api_gateway_app_attachment":                            resourceAliyunApigatewayAppAttachment(),
			"alicloud_api_gateway_backend":                                   resourceAliCloudApiGatewayBackend(),
			"alicloud_api_gateway_group":                                     resourceAliyunApigatewayGroup(),
			"alicloud_api_gateway_instance":                                  resourceAliCloudApiGatewayInstance(),
			"alicloud_api_gateway_instance_acl_attachment":                   resourceAliCloudApiGatewayInstanceAclAttachment(),
			"alicloud_api_gateway_log_config":                                resourceAliCloudApiGatewayLogConfig(),
			"alicloud_api_gateway_model":                                     resourceAliCloudApiGatewayModel(),
			"alicloud_api_gateway_plugin":                                    resourceAliCloudApiGatewayPlugin(),
			"alicloud_api_gateway_plugin_attachment":                         resourceAliCloudApiGatewayPluginAttachment(),
			"alicloud_api_gateway_vpc_access":                                resourceAliyunApigatewayVpc(),
			"alicloud_apig_environment":                                      resourceAliCloudApigEnvironment(),
			"alicloud_apig_gateway":                                          resourceAliCloudApigGateway(),
			"alicloud_apig_http_api":                                         resourceAliCloudApigHttpApi(),
			"alicloud_arms_alert_contact":                                    resourceAliCloudArmsAlertContact(),
			"alicloud_arms_alert_contact_group":                              resourceAliCloudArmsAlertContactGroup(),
			"alicloud_arms_alert_integration":                                resourceAliCloudArmsAlertIntegration(),
			"alicloud_arms_alert_integration_field_redefine_rule":            resourceAliCloudArmsAlertIntegrationFieldRedefineRule(),
			"alicloud_arms_alert_notification_policy":                        resourceAliCloudArmsAlertNotificationPolicy(),
			"alicloud_arms_alert_robot":                                      resourceAliCloudArmsAlertRobot(),
			"alicloud_arms_alert_rule":                                       resourceAliCloudArmsAlertRule(),
			"alicloud_arms_alert_silence_policy":                             resourceAliCloudArmsAlertSilencePolicy(),
			"alicloud_arms_grafana_workspace":                                resourceAliCloudArmsGrafanaWorkspace(),
			"alicloud_arms_prometheus_feature":                               resourceAliCloudArmsPrometheusFeature(),
			"alicloud_arms_prometheus_environment":                           resourceAliCloudArmsPrometheusEnvironment(),
			"alicloud_arms_prometheus_remote_write":                          resourceAliCloudArmsPrometheusRemoteWrite(),
			"alicloud_auto_provisioning_group":                               resourceAliCloudAutoProvisioningGroup(),
			"alicloud_bastionhost_host":                                      resourceAliCloudBastionhostHost(),
			"alicloud_bastionhost_host_account":                              resourceAliCloudBastionhostHostAccount(),
			"alicloud_bastionhost_host_account_share_key_attachment":         resourceAliCloudBastionhostHostAccountShareKeyAttachment(),
			"alicloud_bastionhost_host_account_user_attachment":              resourceAliCloudBastionhostHostAccountUserAttachment(),
			"alicloud_bastionhost_host_account_user_group_attachment":        resourceAliCloudBastionhostHostAccountUserGroupAttachment(),
			"alicloud_bastionhost_host_attachment":                           resourceAliCloudBastionhostHostAttachment(),
			"alicloud_bastionhost_host_group":                                resourceAliCloudBastionhostHostGroup(),
			"alicloud_bastionhost_host_group_account_user_attachment":        resourceAliCloudBastionhostHostGroupAccountUserAttachment(),
			"alicloud_bastionhost_host_group_account_user_group_attachment":  resourceAliCloudBastionhostHostGroupAccountUserGroupAttachment(),
			"alicloud_bastionhost_host_share_key":                            resourceAliCloudBastionhostHostShareKey(),
			"alicloud_bastionhost_instance":                                  resourceAliCloudBastionhostInstance(),
			"alicloud_bastionhost_user":                                      resourceAliCloudBastionhostUser(),
			"alicloud_bastionhost_user_attachment":                           resourceAliCloudBastionhostUserAttachment(),
			"alicloud_bastionhost_user_group":                                resourceAliCloudBastionhostUserGroup(),
			"alicloud_bp_studio_application":                                 resourceAliCloudBpStudioApplication(),
			"alicloud_brain_industrial_pid_loop":                             resourceAliCloudBrainIndustrialPidLoop(),
			"alicloud_brain_industrial_pid_organization":                     resourceAliCloudBrainIndustrialPidOrganization(),
			"alicloud_brain_industrial_pid_project":                          resourceAliCloudBrainIndustrialPidProject(),
			"alicloud_cas_certificate":                                       resourceAliCloudSslCertificatesServiceCertificate(),
			"alicloud_cassandra_backup_plan":                                 resourceAliCloudCassandraBackupPlan(),
			"alicloud_cassandra_cluster":                                     resourceAliCloudCassandraCluster(),
			"alicloud_cassandra_data_center":                                 resourceAliCloudCassandraDataCenter(),
			"alicloud_cddc_dedicated_host":                                   resourceAliCloudCddcDedicatedHost(),
			"alicloud_cddc_dedicated_host_account":                           resourceAliCloudCddcDedicatedHostAccount(),
			"alicloud_cddc_dedicated_host_group":                             resourceAliCloudCddcDedicatedHostGroup(),
			"alicloud_cddc_dedicated_propre_host":                            resourceAliCloudCddcDedicatedPropreHost(),
			"alicloud_cdn_domain":                                            resourceAliCloudCdnDomain(),
			"alicloud_cdn_domain_config":                                     resourceAliCloudCdnDomainConfig(),
			"alicloud_cdn_fc_trigger":                                        resourceAliCloudCdnFcTrigger(),
			"alicloud_cdn_real_time_log_delivery":                            resourceAliCloudCdnRealTimeLogDelivery(),
			"alicloud_cen_bandwidth_limit":                                   resourceAliCloudCenBandwidthLimit(),
			"alicloud_cen_bandwidth_package":                                 resourceAliCloudCenBandwidthPackage(),
			"alicloud_cen_bandwidth_package_attachment":                      resourceAliCloudCenBandwidthPackageAttachment(),
			"alicloud_cen_child_instance_route_entry_to_attachment":          resourceAliCloudCenChildInstanceRouteEntryToAttachment(),
			"alicloud_cen_flowlog":                                           resourceAliCloudCenFlowLog(),
			"alicloud_cen_instance":                                          resourceAliCloudCenCenInstance(),
			"alicloud_cen_instance_attachment":                               resourceAliCloudCenInstanceAttachment(),
			"alicloud_cen_instance_grant":                                    resourceAliCloudCenInstanceGrant(),
			"alicloud_cen_inter_region_traffic_qos_policy":                   resourceAliCloudCenInterRegionTrafficQosPolicy(),
			"alicloud_cen_inter_region_traffic_qos_queue":                    resourceAliCloudCenInterRegionTrafficQosQueue(),
			"alicloud_cen_private_zone":                                      resourceAliCloudCenPrivateZone(),
			"alicloud_cen_route_entry":                                       resourceAliCloudCenRouteEntry(),
			"alicloud_cen_route_map":                                         resourceAliCloudCenRouteMap(),
			"alicloud_cen_route_service":                                     resourceAliCloudCenRouteService(),
			"alicloud_cen_traffic_marking_policy":                            resourceAliCloudCenTrafficMarkingPolicy(),
			"alicloud_cen_transit_route_table_aggregation":                   resourceAliCloudCenTransitRouteTableAggregation(),
			"alicloud_cen_transit_router":                                    resourceAliCloudCenTransitRouter(),
			"alicloud_cen_transit_router_cidr":                               resourceAliCloudCenTransitRouterCidr(),
			"alicloud_cen_transit_router_ecr_attachment":                     resourceAliCloudCenTransitRouterEcrAttachment(),
			"alicloud_cen_transit_router_grant_attachment":                   resourceAliCloudCenTransitRouterGrantAttachment(),
			"alicloud_cen_transit_router_multicast_domain":                   resourceAliCloudCenTransitRouterMulticastDomain(),
			"alicloud_cen_transit_router_multicast_domain_association":       resourceAliCloudCenTransitRouterMulticastDomainAssociation(),
			"alicloud_cen_transit_router_multicast_domain_member":            resourceAliCloudCenTransitRouterMulticastDomainMember(),
			"alicloud_cen_transit_router_multicast_domain_peer_member":       resourceAliCloudCenTransitRouterMulticastDomainPeerMember(),
			"alicloud_cen_transit_router_multicast_domain_source":            resourceAliCloudCenTransitRouterMulticastDomainSource(),
			"alicloud_cen_transit_router_peer_attachment":                    resourceAliCloudCenTransitRouterPeerAttachment(),
			"alicloud_cen_transit_router_prefix_list_association":            resourceAliCloudCenTransitRouterPrefixListAssociation(),
			"alicloud_cen_transit_router_route_entry":                        resourceAliCloudCenTransitRouterRouteEntry(),
			"alicloud_cen_transit_router_route_table":                        resourceAliCloudCenTransitRouterRouteTable(),
			"alicloud_cen_transit_router_route_table_association":            resourceAliCloudCenTransitRouterRouteTableAssociation(),
			"alicloud_cen_transit_router_route_table_propagation":            resourceAliCloudCenTransitRouterRouteTablePropagation(),
			"alicloud_cen_transit_router_vbr_attachment":                     resourceAliCloudCenTransitRouterVbrAttachment(),
			"alicloud_cen_transit_router_vpc_attachment":                     resourceAliCloudCenTransitRouterVpcAttachment(),
			"alicloud_cen_transit_router_vpn_attachment":                     resourceAliCloudCenTransitRouterVpnAttachment(),
			"alicloud_cen_vbr_health_check":                                  resourceAliCloudCenVbrHealthCheck(),
			"alicloud_chatbot_publish_task":                                  resourceAliCloudChatbotPublishTask(),
			"alicloud_click_house_account":                                   resourceAliCloudClickHouseAccount(),
			"alicloud_click_house_backup_policy":                             resourceAliCloudClickHouseBackupPolicy(),
			"alicloud_click_house_db_cluster":                                resourceAliCloudClickHouseDbCluster(),
			"alicloud_click_house_enterprise_db_cluster":                     resourceAliCloudClickHouseEnterpriseDBCluster(),
			"alicloud_click_house_enterprise_db_cluster_account":             resourceAliCloudClickHouseEnterpriseDBClusterAccount(),
			"alicloud_click_house_enterprise_db_cluster_backup_policy":       resourceAliCloudClickHouseEnterpriseDbClusterBackupPolicy(),
			"alicloud_click_house_enterprise_db_cluster_public_endpoint":     resourceAliCloudClickHouseEnterpriseDbClusterPublicEndpoint(),
			"alicloud_click_house_enterprise_db_cluster_security_ip":         resourceAliCloudClickHouseEnterpriseDbClusterSecurityIP(),
			"alicloud_cloud_connect_network":                                 resourceAliCloudCloudConnectNetwork(),
			"alicloud_cloud_connect_network_attachment":                      resourceAliCloudCloudConnectNetworkAttachment(),
			"alicloud_cloud_connect_network_grant":                           resourceAliCloudCloudConnectNetworkGrant(),
			"alicloud_cloud_control_resource":                                resourceAliCloudCloudControlResource(),
			"alicloud_cloud_firewall_address_book":                           resourceAliCloudCloudFirewallAddressBook(),
			"alicloud_cloud_firewall_control_policy":                         resourceAliCloudCloudFirewallControlPolicy(),
			"alicloud_cloud_firewall_control_policy_order":                   resourceAliCloudCloudFirewallControlPolicyOrder(),
			"alicloud_cloud_firewall_instance":                               resourceAliCloudCloudFirewallInstance(),
			"alicloud_cloud_firewall_instance_member":                        resourceAliCloudCloudFirewallInstanceMember(),
			"alicloud_cloud_firewall_ips_config":                             resourceAliCloudCloudFirewallIPSConfig(),
			"alicloud_cloud_firewall_nat_firewall":                           resourceAliCloudCloudFirewallNatFirewall(),
			"alicloud_cloud_firewall_nat_firewall_control_policy":            resourceAliCloudCloudFirewallNatFirewallControlPolicy(),
			"alicloud_cloud_firewall_vpc_cen_tr_firewall":                    resourceAliCloudCloudFirewallVpcCenTrFirewall(),
			"alicloud_cloud_firewall_vpc_firewall":                           resourceAliCloudCloudFirewallVpcFirewall(),
			"alicloud_cloud_firewall_vpc_firewall_cen":                       resourceAliCloudCloudFirewallVpcFirewallCen(),
			"alicloud_cloud_firewall_vpc_firewall_control_policy":            resourceAliCloudCloudFirewallVpcFirewallControlPolicy(),
			"alicloud_cloud_monitor_service_basic_public":                    resourceAliCloudCloudMonitorServiceBasicPublic(),
			"alicloud_cloud_monitor_service_enterprise_public":               resourceAliCloudCloudMonitorServiceEnterprisePublic(),
			"alicloud_cloud_monitor_service_group_monitoring_agent_process":  resourceAliCloudCloudMonitorServiceGroupMonitoringAgentProcess(),
			"alicloud_cloud_monitor_service_hybrid_double_write":             resourceAliCloudCloudMonitorServiceHybridDoubleWrite(),
			"alicloud_cloud_monitor_service_monitoring_agent_process":        resourceAliCloudCloudMonitorServiceMonitoringAgentProcess(),
			"alicloud_cloud_phone_image":                                     resourceAliCloudCloudPhoneImage(),
			"alicloud_cloud_phone_instance":                                  resourceAliCloudCloudPhoneInstance(),
			"alicloud_cloud_phone_instance_group":                            resourceAliCloudCloudPhoneInstanceGroup(),
			"alicloud_cloud_phone_key_pair":                                  resourceAliCloudCloudPhoneKeyPair(),
			"alicloud_cloud_phone_policy":                                    resourceAliCloudCloudPhonePolicy(),
			"alicloud_cloud_sso_access_assignment":                           resourceAliCloudCloudSsoAccessAssignment(),
			"alicloud_cloud_sso_access_configuration":                        resourceAliCloudCloudSsoAccessConfiguration(),
			"alicloud_cloud_sso_access_configuration_provisioning":           resourceAliCloudCloudSsoAccessConfigurationProvisioning(),
			"alicloud_cloud_sso_directory":                                   resourceAliCloudCloudSSODirectory(),
			"alicloud_cloud_sso_group":                                       resourceAliCloudCloudSsoGroup(),
			"alicloud_cloud_sso_scim_server_credential":                      resourceAliCloudCloudSsoScimServerCredential(),
			"alicloud_cloud_sso_user":                                        resourceAliCloudCloudSsoUser(),
			"alicloud_cloud_sso_user_attachment":                             resourceAliCloudCloudSsoUserAttachment(),
			"alicloud_cloud_storage_gateway_express_sync":                    resourceAliCloudCloudStorageGatewayExpressSync(),
			"alicloud_cloud_storage_gateway_express_sync_share_attachment":   resourceAliCloudCloudStorageGatewayExpressSyncShareAttachment(),
			"alicloud_cloud_storage_gateway_gateway":                         resourceAliCloudCloudStorageGatewayGateway(),
			"alicloud_cloud_storage_gateway_gateway_block_volume":            resourceAliCloudCloudStorageGatewayGatewayBlockVolume(),
			"alicloud_cloud_storage_gateway_gateway_cache_disk":              resourceAliCloudCloudStorageGatewayGatewayCacheDisk(),
			"alicloud_cloud_storage_gateway_gateway_file_share":              resourceAliCloudCloudStorageGatewayGatewayFileShare(),
			"alicloud_cloud_storage_gateway_gateway_logging":                 resourceAliCloudCloudStorageGatewayGatewayLogging(),
			"alicloud_cloud_storage_gateway_gateway_smb_user":                resourceAliCloudCloudStorageGatewayGatewaySmbUser(),
			"alicloud_cloud_storage_gateway_storage_bundle":                  resourceAliCloudCloudStorageGatewayStorageBundle(),
			"alicloud_cloudauth_face_config":                                 resourceAliCloudCloudauthFaceConfig(),
			"alicloud_cms_alarm":                                             resourceAliCloudCmsAlarm(),
			"alicloud_cms_alarm_contact":                                     resourceAliCloudCmsAlarmContact(),
			"alicloud_cms_alarm_contact_group":                               resourceAliCloudCmsAlarmContactGroup(),
			"alicloud_cms_dynamic_tag_group":                                 resourceAliCloudCmsDynamicTagGroup(),
			"alicloud_cms_event_rule":                                        resourceAliCloudCloudMonitorServiceEventRule(),
			"alicloud_cms_group_metric_rule":                                 resourceAliCloudCmsGroupMetricRule(),
			"alicloud_cms_hybrid_monitor_fc_task":                            resourceAliCloudCmsHybridMonitorFcTask(),
			"alicloud_cms_hybrid_monitor_sls_task":                           resourceAliCloudCmsHybridMonitorSlsTask(),
			"alicloud_cms_metric_rule_black_list":                            resourceAliCloudCmsMetricRuleBlackList(),
			"alicloud_cms_metric_rule_template":                              resourceAliCloudCmsMetricRuleTemplate(),
			"alicloud_cms_monitor_group":                                     resourceAliCloudCmsMonitorGroup(),
			"alicloud_cms_monitor_group_instances":                           resourceAliCloudCmsMonitorGroupInstances(),
			"alicloud_cms_namespace":                                         resourceAliCloudCmsNamespace(),
			"alicloud_cms_site_monitor":                                      resourceAliCloudCmsSiteMonitor(),
			"alicloud_cms_sls_group":                                         resourceAliCloudCmsSlsGroup(),
			"alicloud_common_bandwidth_package":                              resourceAliCloudCbwpCommonBandwidthPackage(),
			"alicloud_common_bandwidth_package_attachment":                   resourceAliCloudCbwpCommonBandwidthPackageAttachment(),
			"alicloud_compute_nest_service_instance":                         resourceAliCloudComputeNestServiceInstance(),
			"alicloud_config_aggregate_compliance_pack":                      resourceAliCloudConfigAggregateCompliancePack(),
			"alicloud_config_aggregate_config_rule":                          resourceAliCloudConfigAggregateConfigRule(),
			"alicloud_config_aggregate_delivery":                             resourceAliCloudConfigAggregateDelivery(),
			"alicloud_config_aggregator":                                     resourceAliCloudConfigAggregator(),
			"alicloud_config_compliance_pack":                                resourceAliCloudConfigCompliancePack(),
			"alicloud_config_configuration_recorder":                         resourceAliCloudConfigConfigurationRecorder(),
			"alicloud_config_delivery":                                       resourceAliCloudConfigDelivery(),
			"alicloud_config_delivery_channel":                               resourceAliCloudConfigDeliveryChannel(),
			"alicloud_config_remediation":                                    resourceAliCloudConfigRemediation(),
			"alicloud_config_rule":                                           resourceAliCloudConfigRule(),
			"alicloud_container_cluster":                                     resourceAliCloudCSSwarm(),
			"alicloud_copy_image":                                            resourceAliCloudImageCopy(),
			"alicloud_cr_chain":                                              resourceAliCloudCrChain(),
			"alicloud_cr_chart_namespace":                                    resourceAliCloudCrChartNamespace(),
			"alicloud_cr_chart_repository":                                   resourceAliCloudCrChartRepository(),
			"alicloud_cr_ee_instance":                                        resourceAliCloudCrInstance(),
			"alicloud_cr_ee_namespace":                                       resourceAliCloudCrEENamespace(),
			"alicloud_cr_ee_repo":                                            resourceAliCloudCrEERepo(),
			"alicloud_cr_ee_sync_rule":                                       resourceAliCloudCrRepoSyncRule(),
			"alicloud_cr_endpoint_acl_policy":                                resourceAliCloudCrEndpointAclPolicy(),
			"alicloud_cr_namespace":                                          resourceAliCloudCRNamespace(),
			"alicloud_cr_repo":                                               resourceAliCloudCRRepo(),
			"alicloud_cr_vpc_endpoint_linked_vpc":                            resourceAliCloudCrVpcEndpointLinkedVpc(),
			"alicloud_cs_application":                                        resourceAliCloudCSApplication(),
			"alicloud_cs_autoscaling_config":                                 resourceAliCloudCSAutoscalingConfig(),
			"alicloud_cs_edge_kubernetes":                                    resourceAliCloudCSEdgeKubernetes(),
			"alicloud_cs_kubernetes":                                         resourceAliCloudCSKubernetes(),
			"alicloud_cs_kubernetes_addon":                                   resourceAliCloudCSKubernetesAddon(),
			"alicloud_cs_kubernetes_autoscaler":                              resourceAliCloudCSKubernetesAutoscaler(),
			"alicloud_cs_kubernetes_node_pool":                               resourceAliCloudAckNodepool(),
			"alicloud_cs_kubernetes_permissions":                             resourceAliCloudCSKubernetesPermissions(),
			"alicloud_cs_managed_kubernetes":                                 resourceAliCloudCSManagedKubernetes(),
			"alicloud_cs_serverless_kubernetes":                              resourceAliCloudCSServerlessKubernetes(),
			"alicloud_cs_swarm":                                              resourceAliCloudCSSwarm(),
			"alicloud_das_switch_das_pro":                                    resourceAliCloudDasSwitchDasPro(),
			"alicloud_data_works_data_source":                                resourceAliCloudDataWorksDataSource(),
			"alicloud_data_works_data_source_shared_rule":                    resourceAliCloudDataWorksDataSourceSharedRule(),
			"alicloud_data_works_di_alarm_rule":                              resourceAliCloudDataWorksDiAlarmRule(),
			"alicloud_data_works_di_job":                                     resourceAliCloudDataWorksDiJob(),
			"alicloud_data_works_dw_resource_group":                          resourceAliCloudDataWorksDwResourceGroup(),
			"alicloud_data_works_folder":                                     resourceAliCloudDataWorksFolder(),
			"alicloud_data_works_network":                                    resourceAliCloudDataWorksNetwork(),
			"alicloud_data_works_project":                                    resourceAliCloudDataWorksProject(),
			"alicloud_data_works_project_member":                             resourceAliCloudDataWorksProjectMember(),
			"alicloud_database_gateway_gateway":                              resourceAliCloudDatabaseGatewayGateway(),
			"alicloud_datahub_project":                                       resourceAliCloudDatahubProject(),
			"alicloud_datahub_subscription":                                  resourceAliCloudDatahubSubscription(),
			"alicloud_datahub_topic":                                         resourceAliCloudDatahubTopic(),
			"alicloud_db_account":                                            resourceAliCloudRdsAccount(),
			"alicloud_db_account_privilege":                                  resourceAliCloudDBAccountPrivilege(),
			"alicloud_db_backup_policy":                                      resourceAliCloudDBBackupPolicy(),
			"alicloud_db_connection":                                         resourceAliCloudDBConnection(),
			"alicloud_db_database":                                           resourceAliCloudDBDatabase(),
			"alicloud_db_instance":                                           resourceAliCloudDBInstance(),
			"alicloud_db_read_write_splitting_connection":                    resourceAliCloudDBReadWriteSplittingConnection(),
			"alicloud_db_readonly_instance":                                  resourceAliCloudDBReadonlyInstance(),
			"alicloud_dbfs_auto_snap_shot_policy":                            resourceAliCloudDbfsAutoSnapShotPolicy(),
			"alicloud_dbfs_instance":                                         resourceAliCloudDbfsDbfsInstance(),
			"alicloud_dbfs_instance_attachment":                              resourceAliCloudDbfsInstanceAttachment(),
			"alicloud_dbfs_service_linked_role":                              resourceAliCloudDbfsServiceLinkedRole(),
			"alicloud_dbfs_snapshot":                                         resourceAliCloudDbfsSnapshot(),
			"alicloud_dbs_backup_plan":                                       resourceAliCloudDbsBackupPlan(),
			"alicloud_dcdn_domain":                                           resourceAliCloudDcdnDomain(),
			"alicloud_dcdn_domain_config":                                    resourceAliCloudDcdnDomainConfig(),
			"alicloud_dcdn_er":                                               resourceAliCloudDcdnEr(),
			"alicloud_dcdn_ipa_domain":                                       resourceAliCloudDcdnIpaDomain(),
			"alicloud_dcdn_kv":                                               resourceAliCloudDcdnKv(),
			"alicloud_dcdn_kv_namespace":                                     resourceAliCloudDcdnKvNamespace(),
			"alicloud_dcdn_waf_domain":                                       resourceAliCloudDcdnWafDomain(),
			"alicloud_dcdn_waf_policy":                                       resourceAliCloudDcdnWafPolicy(),
			"alicloud_dcdn_waf_policy_domain_attachment":                     resourceAliCloudDcdnWafPolicyDomainAttachment(),
			"alicloud_dcdn_waf_rule":                                         resourceAliCloudDcdnWafRule(),
			"alicloud_ddos_basic_defense_threshold":                          resourceAliCloudDdosBasicDefenseThreshold(),
			"alicloud_ddos_basic_threshold":                                  resourceAliCloudDdosBasicThreshold(),
			"alicloud_ddos_bgp_policy":                                       resourceAliCloudDdosBgpPolicy(),
			"alicloud_ddosbgp_instance":                                      resourceAliCloudDdosbgpInstance(),
			"alicloud_ddosbgp_ip":                                            resourceAliCloudDdosbgpIp(),
			"alicloud_ddoscoo_domain_resource":                               resourceAliCloudDdosCooDomainResource(),
			"alicloud_ddoscoo_instance":                                      resourceAliCloudDdoscooInstance(),
			"alicloud_ddoscoo_port":                                          resourceAliCloudDdosCooPort(),
			"alicloud_ddoscoo_scheduler_rule":                                resourceAliCloudDdoscooSchedulerRule(),
			"alicloud_dfs_access_group":                                      resourceAliCloudDfsAccessGroup(),
			"alicloud_dfs_access_rule":                                       resourceAliCloudDfsAccessRule(),
			"alicloud_dfs_file_system":                                       resourceAliCloudDfsFileSystem(),
			"alicloud_dfs_mount_point":                                       resourceAliCloudDfsMountPoint(),
			"alicloud_dfs_vsc_mount_point":                                   resourceAliCloudDfsVscMountPoint(),
			"alicloud_direct_mail_domain":                                    resourceAliCloudDirectMailDomain(),
			"alicloud_direct_mail_mail_address":                              resourceAliCloudDirectMailMailAddress(),
			"alicloud_direct_mail_receivers":                                 resourceAliCloudDirectMailReceivers(),
			"alicloud_direct_mail_tag":                                       resourceAliCloudDirectMailTag(),
			"alicloud_disk":                                                  resourceAliCloudEcsDisk(),
			"alicloud_disk_attachment":                                       resourceAliCloudEcsDiskAttachment(),
			"alicloud_dms_enterprise_authority_template":                     resourceAliCloudDMSEnterpriseAuthorityTemplate(),
			"alicloud_dms_enterprise_instance":                               resourceAliCloudDmsEnterpriseInstance(),
			"alicloud_dms_enterprise_logic_database":                         resourceAliCloudDmsEnterpriseLogicDatabase(),
			"alicloud_dms_enterprise_proxy":                                  resourceAliCloudDmsEnterpriseProxy(),
			"alicloud_dms_enterprise_proxy_access":                           resourceAliCloudDmsEnterpriseProxyAccess(),
			"alicloud_dms_enterprise_user":                                   resourceAliCloudDmsEnterpriseUser(),
			"alicloud_dns":                                                   resourceAliCloudDns(),
			"alicloud_dns_domain":                                            resourceAliCloudAlidnsDomain(),
			"alicloud_dns_domain_attachment":                                 resourceAliCloudAlidnsDomainAttachment(),
			"alicloud_dns_group":                                             resourceAliCloudDnsGroup(),
			"alicloud_dns_instance":                                          resourceAliCloudAlidnsInstance(),
			"alicloud_dns_record":                                            resourceAliCloudDnsRecord(),
			"alicloud_drds_instance":                                         resourceAliCloudDRDSInstance(),
			"alicloud_drds_polardbx_instance":                                resourceAliCloudDrdsPolardbxInstance(),
			"alicloud_dts_consumer_channel":                                  resourceAliCloudDtsConsumerChannel(),
			"alicloud_dts_instance":                                          resourceAliCloudDtsInstance(),
			"alicloud_dts_job_monitor_rule":                                  resourceAliCloudDtsJobMonitorRule(),
			"alicloud_dts_migration_instance":                                resourceAliCloudDtsMigrationInstance(),
			"alicloud_dts_migration_job":                                     resourceAliCloudDtsMigrationJob(),
			"alicloud_dts_subscription_job":                                  resourceAliCloudDtsSubscriptionJob(),
			"alicloud_dts_synchronization_instance":                          resourceAliCloudDtsSynchronizationInstance(),
			"alicloud_dts_synchronization_job":                               resourceAliCloudDtsSynchronizationJob(),
			"alicloud_eais_client_instance_attachment":                       resourceAliCloudEaisClientInstanceAttachment(),
			"alicloud_eais_instance":                                         resourceAliCloudEaisInstance(),
			"alicloud_ebs_dedicated_block_storage_cluster":                   resourceAliCloudEbsDedicatedBlockStorageCluster(),
			"alicloud_ebs_disk_replica_group":                                resourceAliCloudEbsDiskReplicaGroup(),
			"alicloud_ebs_disk_replica_pair":                                 resourceAliCloudEbsDiskReplicaPair(),
			"alicloud_ebs_enterprise_snapshot_policy":                        resourceAliCloudEbsEnterpriseSnapshotPolicy(),
			"alicloud_ebs_enterprise_snapshot_policy_attachment":             resourceAliCloudEbsEnterpriseSnapshotPolicyAttachment(),
			"alicloud_ebs_replica_group_drill":                               resourceAliCloudEbsReplicaGroupDrill(),
			"alicloud_ebs_replica_pair_drill":                                resourceAliCloudEbsReplicaPairDrill(),
			"alicloud_ebs_solution_instance":                                 resourceAliCloudEbsSolutionInstance(),
			"alicloud_ecd_ad_connector_directory":                            resourceAliCloudEcdAdConnectorDirectory(),
			"alicloud_ecd_ad_connector_office_site":                          resourceAliCloudEcdAdConnectorOfficeSite(),
			"alicloud_ecd_bundle":                                            resourceAliCloudEcdBundle(),
			"alicloud_ecd_command":                                           resourceAliCloudEcdCommand(),
			"alicloud_ecd_custom_property":                                   resourceAliCloudEcdCustomProperty(),
			"alicloud_ecd_desktop":                                           resourceAliCloudEcdDesktop(),
			"alicloud_ecd_image":                                             resourceAliCloudEcdImage(),
			"alicloud_ecd_nas_file_system":                                   resourceAliCloudEcdNasFileSystem(),
			"alicloud_ecd_network_package":                                   resourceAliCloudEcdNetworkPackage(),
			"alicloud_ecd_policy_group":                                      resourceAliCloudEcdPolicyGroup(),
			"alicloud_ecd_ram_directory":                                     resourceAliCloudEcdRamDirectory(),
			"alicloud_ecd_simple_office_site":                                resourceAliCloudEcdSimpleOfficeSite(),
			"alicloud_ecd_snapshot":                                          resourceAliCloudEcdSnapshot(),
			"alicloud_ecd_user":                                              resourceAliCloudEcdUser(),
			"alicloud_eci_container_group":                                   resourceAliCloudEciContainerGroup(),
			"alicloud_eci_image_cache":                                       resourceAliCloudEciImageCache(),
			"alicloud_eci_openapi_image_cache":                               resourceAliCloudEciImageCache(),
			"alicloud_eci_virtual_node":                                      resourceAliCloudEciVirtualNode(),
			"alicloud_ecp_instance":                                          resourceAliCloudEcpInstance(),
			"alicloud_ecp_key_pair":                                          resourceAliCloudEcpKeyPair(),
			"alicloud_ecs_activation":                                        resourceAliCloudEcsActivation(),
			"alicloud_ecs_auto_snapshot_policy":                              resourceAliCloudEcsAutoSnapshotPolicy(),
			"alicloud_ecs_auto_snapshot_policy_attachment":                   resourceAliCloudEcsAutoSnapshotPolicyAttachment(),
			"alicloud_ecs_capacity_reservation":                              resourceAliCloudEcsCapacityReservation(),
			"alicloud_ecs_command":                                           resourceAliCloudEcsCommand(),
			"alicloud_ecs_dedicated_host":                                    resourceAliCloudEcsDedicatedHost(),
			"alicloud_ecs_dedicated_host_cluster":                            resourceAliCloudEcsDedicatedHostCluster(),
			"alicloud_ecs_deployment_set":                                    resourceAliCloudEcsDeploymentSet(),
			"alicloud_ecs_disk":                                              resourceAliCloudEcsDisk(),
			"alicloud_ecs_disk_attachment":                                   resourceAliCloudEcsDiskAttachment(),
			"alicloud_ecs_elasticity_assurance":                              resourceAliCloudEcsElasticityAssurance(),
			"alicloud_ecs_hpc_cluster":                                       resourceAliCloudEcsHpcCluster(),
			"alicloud_ecs_image_component":                                   resourceAliCloudEcsImageComponent(),
			"alicloud_ecs_image_pipeline":                                    resourceAliCloudEcsImagePipeline(),
			"alicloud_ecs_image_pipeline_execution":                          resourceAliCloudEcsImagePipelineExecution(),
			"alicloud_ecs_instance_set":                                      resourceAliCloudEcsInstanceSet(),
			"alicloud_ecs_invocation":                                        resourceAliCloudEcsInvocation(),
			"alicloud_ecs_key_pair":                                          resourceAliCloudEcsKeyPair(),
			"alicloud_ecs_key_pair_attachment":                               resourceAliCloudEcsKeyPairAttachment(),
			"alicloud_ecs_launch_template":                                   resourceAliCloudEcsLaunchTemplate(),
			"alicloud_ecs_network_interface":                                 resourceAliCloudEcsNetworkInterface(),
			"alicloud_ecs_network_interface_attachment":                      resourceAliCloudEcsNetworkInterfaceAttachment(),
			"alicloud_ecs_network_interface_permission":                      resourceAliCloudEcsNetworkInterfacePermission(),
			"alicloud_ecs_prefix_list":                                       resourceAliCloudEcsPrefixList(),
			"alicloud_ecs_ram_role_attachment":                               resourceAliCloudEcsRamRoleAttachment(),
			"alicloud_ecs_session_manager_status":                            resourceAliCloudEcsSessionManagerStatus(),
			"alicloud_ecs_snapshot":                                          resourceAliCloudEcsSnapshot(),
			"alicloud_ecs_snapshot_group":                                    resourceAliCloudEcsSnapshotGroup(),
			"alicloud_ecs_storage_capacity_unit":                             resourceAliCloudEcsStorageCapacityUnit(),
			"alicloud_edas_application":                                      resourceAliCloudEdasApplication(),
			"alicloud_edas_application_deployment":                           resourceAliCloudEdasApplicationPackageAttachment(),
			"alicloud_edas_application_scale":                                resourceAliCloudEdasApplicationScale(),
			"alicloud_edas_cluster":                                          resourceAliCloudEdasCluster(),
			"alicloud_edas_deploy_group":                                     resourceAliCloudEdasDeployGroup(),
			"alicloud_edas_instance_cluster_attachment":                      resourceAliCloudEdasInstanceClusterAttachment(),
			"alicloud_edas_k8s_application":                                  resourceAliCloudEdasK8sApplication(),
			"alicloud_edas_k8s_cluster":                                      resourceAliCloudEdasK8sCluster(),
			"alicloud_edas_k8s_slb_attachment":                               resourceAliCloudEdasK8sSlbAttachment(),
			"alicloud_edas_namespace":                                        resourceAliCloudEdasNamespace(),
			"alicloud_edas_slb_attachment":                                   resourceAliCloudEdasSlbAttachment(),
			"alicloud_eflo_cluster":                                          resourceAliCloudEfloCluster(),
			"alicloud_eflo_experiment_plan":                                  resourceAliCloudEfloExperimentPlan(),
			"alicloud_eflo_experiment_plan_template":                         resourceAliCloudEfloExperimentPlanTemplate(),
			"alicloud_eflo_invocation":                                       resourceAliCloudEfloInvocation(),
			"alicloud_eflo_node":                                             resourceAliCloudEfloNode(),
			"alicloud_eflo_node_group":                                       resourceAliCloudEfloNodeGroup(),
			"alicloud_eflo_resource":                                         resourceAliCloudEfloResource(),
			"alicloud_eflo_subnet":                                           resourceAliCloudEfloSubnet(),
			"alicloud_eflo_vpd":                                              resourceAliCloudEfloVpd(),
			"alicloud_eflo_vsc":                                              resourceAliCloudEfloVsc(),
			"alicloud_ehpc_cluster":                                          resourceAliCloudEhpcCluster(),
			"alicloud_ehpc_job_template":                                     resourceAliCloudEhpcJobTemplate(),
			"alicloud_eip":                                                   resourceAliCloudEipAddress(),
			"alicloud_eip_address":                                           resourceAliCloudEipAddress(),
			"alicloud_eip_association":                                       resourceAliCloudEipAssociation(),
			"alicloud_eip_segment_address":                                   resourceAliCloudEipSegmentAddress(),
			"alicloud_eipanycast_anycast_eip_address":                        resourceAliCloudEipanycastAnycastEipAddress(),
			"alicloud_eipanycast_anycast_eip_address_attachment":             resourceAliCloudEipanycastAnycastEipAddressAttachment(),
			"alicloud_elasticsearch_instance":                                resourceAliCloudElasticsearch(),
			"alicloud_emr_cluster":                                           resourceAliCloudEmrCluster(),
			"alicloud_emrv2_cluster":                                         resourceAliCloudEmrV2Cluster(),
			"alicloud_ens_disk":                                              resourceAliCloudEnsDisk(),
			"alicloud_ens_disk_instance_attachment":                          resourceAliCloudEnsDiskInstanceAttachment(),
			"alicloud_ens_eip":                                               resourceAliCloudEnsEip(),
			"alicloud_ens_eip_instance_attachment":                           resourceAliCloudEnsEipInstanceAttachment(),
			"alicloud_ens_image":                                             resourceAliCloudEnsImage(),
			"alicloud_ens_instance":                                          resourceAliCloudEnsInstance(),
			"alicloud_ens_instance_security_group_attachment":                resourceAliCloudEnsInstanceSecurityGroupAttachment(),
			"alicloud_ens_key_pair":                                          resourceAliCloudEnsKeyPair(),
			"alicloud_ens_load_balancer":                                     resourceAliCloudEnsLoadBalancer(),
			"alicloud_ens_nat_gateway":                                       resourceAliCloudEnsNatGateway(),
			"alicloud_ens_network":                                           resourceAliCloudEnsNetwork(),
			"alicloud_ens_security_group":                                    resourceAliCloudEnsSecurityGroup(),
			"alicloud_ens_snapshot":                                          resourceAliCloudEnsSnapshot(),
			"alicloud_ens_vswitch":                                           resourceAliCloudEnsVswitch(),
			"alicloud_esa_cache_reserve_instance":                            resourceAliCloudEsaCacheReserveInstance(),
			"alicloud_esa_cache_rule":                                        resourceAliCloudEsaCacheRule(),
			"alicloud_esa_certificate":                                       resourceAliCloudEsaCertificate(),
			"alicloud_esa_client_ca_certificate":                             resourceAliCloudEsaClientCaCertificate(),
			"alicloud_esa_client_certificate":                                resourceAliCloudEsaClientCertificate(),
			"alicloud_esa_compression_rule":                                  resourceAliCloudEsaCompressionRule(),
			"alicloud_esa_edge_container_app":                                resourceAliCloudEsaEdgeContainerApp(),
			"alicloud_esa_edge_container_app_record":                         resourceAliCloudEsaEdgeContainerAppRecord(),
			"alicloud_esa_http_request_header_modification_rule":             resourceAliCloudEsaHttpRequestHeaderModificationRule(),
			"alicloud_esa_http_response_header_modification_rule":            resourceAliCloudEsaHttpResponseHeaderModificationRule(),
			"alicloud_esa_https_application_configuration":                   resourceAliCloudEsaHttpsApplicationConfiguration(),
			"alicloud_esa_https_basic_configuration":                         resourceAliCloudEsaHttpsBasicConfiguration(),
			"alicloud_esa_image_transform":                                   resourceAliCloudEsaImageTransform(),
			"alicloud_esa_kv":                                                resourceAliCloudEsaKv(),
			"alicloud_esa_kv_namespace":                                      resourceAliCloudEsaKvNamespace(),
			"alicloud_esa_list":                                              resourceAliCloudEsaList(),
			"alicloud_esa_network_optimization":                              resourceAliCloudEsaNetworkOptimization(),
			"alicloud_esa_origin_pool":                                       resourceAliCloudEsaOriginPool(),
			"alicloud_esa_origin_rule":                                       resourceAliCloudEsaOriginRule(),
			"alicloud_esa_page":                                              resourceAliCloudEsaPage(),
			"alicloud_esa_rate_plan_instance":                                resourceAliCloudEsaRatePlanInstance(),
			"alicloud_esa_record":                                            resourceAliCloudEsaRecord(),
			"alicloud_esa_redirect_rule":                                     resourceAliCloudEsaRedirectRule(),
			"alicloud_esa_rewrite_url_rule":                                  resourceAliCloudEsaRewriteUrlRule(),
			"alicloud_esa_routine":                                           resourceAliCloudEsaRoutine(),
			"alicloud_esa_routine_route":                                     resourceAliCloudEsaRoutineRoute(),
			"alicloud_esa_scheduled_preload_execution":                       resourceAliCloudEsaScheduledPreloadExecution(),
			"alicloud_esa_scheduled_preload_job":                             resourceAliCloudEsaScheduledPreloadJob(),
			"alicloud_esa_site":                                              resourceAliCloudEsaSite(),
			"alicloud_esa_site_delivery_task":                                resourceAliCloudEsaSiteDeliveryTask(),
			"alicloud_esa_video_processing":                                  resourceAliCloudEsaVideoProcessing(),
			"alicloud_esa_waiting_room":                                      resourceAliCloudEsaWaitingRoom(),
			"alicloud_esa_waiting_room_event":                                resourceAliCloudEsaWaitingRoomEvent(),
			"alicloud_esa_waiting_room_rule":                                 resourceAliCloudEsaWaitingRoomRule(),
			"alicloud_ess_alarm":                                             resourceAliCloudEssAlarm(),
			"alicloud_ess_alb_server_group_attachment":                       resourceAliCloudEssAlbServerGroupAttachment(),
			"alicloud_ess_attachment":                                        resourceAliCloudEssAttachment(),
			"alicloud_ess_eci_scaling_configuration":                         resourceAliCloudEssEciScalingConfiguration(),
			"alicloud_ess_lifecycle_hook":                                    resourceAliCloudEssLifecycleHook(),
			"alicloud_ess_notification":                                      resourceAliCloudEssNotification(),
			"alicloud_ess_scaling_configuration":                             resourceAliCloudEssScalingConfiguration(),
			"alicloud_ess_scaling_group":                                     resourceAliCloudEssScalingGroup(),
			"alicloud_ess_scaling_rule":                                      resourceAliCloudEssScalingRule(),
			"alicloud_ess_scalinggroup_vserver_groups":                       resourceAliCloudEssScalingGroupVserverGroups(),
			"alicloud_ess_schedule":                                          resourceAliCloudEssScheduledTask(),
			"alicloud_ess_scheduled_task":                                    resourceAliCloudEssScheduledTask(),
			"alicloud_ess_server_group_attachment":                           resourceAliCloudEssServerGroupAttachment(),
			"alicloud_ess_suspend_process":                                   resourceAliCloudEssSuspendProcess(),
			"alicloud_event_bridge_api_destination":                          resourceAliCloudEventBridgeApiDestination(),
			"alicloud_event_bridge_connection":                               resourceAliCloudEventBridgeConnection(),
			"alicloud_event_bridge_event_bus":                                resourceAliCloudEventBridgeEventBus(),
			"alicloud_event_bridge_event_source":                             resourceAliCloudEventBridgeEventSource(),
			"alicloud_event_bridge_rule":                                     resourceAliCloudEventBridgeRule(),
			"alicloud_event_bridge_service_linked_role":                      resourceAliCloudEventBridgeServiceLinkedRole(),
			"alicloud_event_bridge_slr":                                      resourceAliCloudEventBridgeServiceLinkedRole(),
			"alicloud_express_connect_ec_failover_test_job":                  resourceAliCloudExpressConnectEcFailoverTestJob(),
			"alicloud_express_connect_grant_rule_to_cen":                     resourceAliCloudExpressConnectGrantRuleToCen(),
			"alicloud_express_connect_physical_connection":                   resourceAliCloudExpressConnectPhysicalConnection(),
			"alicloud_express_connect_router_express_connect_router":         resourceAliCloudExpressConnectRouterExpressConnectRouter(),
			"alicloud_express_connect_router_grant_association":              resourceAliCloudExpressConnectRouterGrantAssociation(),
			"alicloud_express_connect_router_interface":                      resourceAliCloudExpressConnectRouterInterface(),
			"alicloud_express_connect_router_tr_association":                 resourceAliCloudExpressConnectRouterExpressConnectRouterTrAssociation(),
			"alicloud_express_connect_router_vbr_child_instance":             resourceAliCloudExpressConnectRouterExpressConnectRouterVbrChildInstance(),
			"alicloud_express_connect_router_vpc_association":                resourceAliCloudExpressConnectRouterExpressConnectRouterVpcAssociation(),
			"alicloud_express_connect_traffic_qos":                           resourceAliCloudExpressConnectTrafficQos(),
			"alicloud_express_connect_traffic_qos_association":               resourceAliCloudExpressConnectTrafficQosAssociation(),
			"alicloud_express_connect_traffic_qos_queue":                     resourceAliCloudExpressConnectTrafficQosQueue(),
			"alicloud_express_connect_traffic_qos_rule":                      resourceAliCloudExpressConnectTrafficQosRule(),
			"alicloud_express_connect_vbr_pconn_association":                 resourceAliCloudExpressConnectVbrPconnAssociation(),
			"alicloud_express_connect_virtual_border_router":                 resourceAliCloudExpressConnectVirtualBorderRouter(),
			"alicloud_express_connect_virtual_physical_connection":           resourceAliCloudExpressConnectVirtualPhysicalConnection(),
			"alicloud_fc_alias":                                              resourceAliCloudFCAlias(),
			"alicloud_fc_async_invoke_config":                                resourceAliCloudFCAsyncInvokeConfig(),
			"alicloud_fc_concurrency_config":                                 resourceAliCloudFCConcurrencyConfig(),
			"alicloud_fc_custom_domain":                                      resourceAliCloudFCCustomDomain(),
			"alicloud_fc_function":                                           resourceAliCloudFCFunction(),
			"alicloud_fc_function_version":                                   resourceAliCloudFCFunctionVersion(),
			"alicloud_fc_layer_version":                                      resourceAliCloudFCLayerVersion(),
			"alicloud_fc_provision_config":                                   resourceAliCloudFCProvisionConfig(),
			"alicloud_fc_trigger":                                            resourceAliCloudFCTrigger(),
			"alicloud_fc_vpc_binding":                                        resourceAliCloudFCVpcBinding(),
			"alicloud_flink_deployment":                                      resourceAliCloudFlinkDeployment(),
			"alicloud_flink_deployment_draft":                                resourceAliCloudFlinkDeploymentDraft(),
			"alicloud_flink_deployment_folder":                               resourceAliCloudFlinkDeploymentFolder(),
			"alicloud_flink_deployment_target":                               resourceAliCloudFlinkDeploymentTarget(),
			"alicloud_flink_job":                                             resourceAliCloudFlinkJob(),
			"alicloud_flink_member":                                          resourceAliCloudFlinkMember(),
			"alicloud_flink_namespace":                                       resourceAliCloudFlinkNamespace(),
			"alicloud_flink_session_cluster":                                 resourceAliCloudFlinkSessionCluster(),
			"alicloud_flink_variable":                                        resourceAliCloudFlinkVariable(),
			"alicloud_flink_workspace":                                       resourceAliCloudFlinkWorkspace(),
			"alicloud_flink_connector":                                       resourceAliCloudFlinkConnector(),
			"alicloud_fnf_execution":                                         resourceAliCloudFnFExecution(),
			"alicloud_fnf_flow":                                              resourceAliCloudFnfFlow(),
			"alicloud_fnf_schedule":                                          resourceAliCloudFnfSchedule(),
			"alicloud_forward_entry":                                         resourceAliCloudForwardEntry(),
			"alicloud_ga_accelerator":                                        resourceAliCloudGaAccelerator(),
			"alicloud_ga_accelerator_spare_ip_attachment":                    resourceAliCloudGaAcceleratorSpareIpAttachment(),
			"alicloud_ga_access_log":                                         resourceAliCloudGaAccessLog(),
			"alicloud_ga_acl":                                                resourceAliCloudGaAcl(),
			"alicloud_ga_acl_attachment":                                     resourceAliCloudGaAclAttachment(),
			"alicloud_ga_acl_entry_attachment":                               resourceAliCloudGaAclEntryAttachment(),
			"alicloud_ga_additional_certificate":                             resourceAliCloudGaAdditionalCertificate(),
			"alicloud_ga_bandwidth_package":                                  resourceAliCloudGaBandwidthPackage(),
			"alicloud_ga_bandwidth_package_attachment":                       resourceAliCloudGaBandwidthPackageAttachment(),
			"alicloud_ga_basic_accelerate_ip":                                resourceAliCloudGaBasicAccelerateIp(),
			"alicloud_ga_basic_accelerate_ip_endpoint_relation":              resourceAliCloudGaBasicAccelerateIpEndpointRelation(),
			"alicloud_ga_basic_accelerator":                                  resourceAliCloudGaBasicAccelerator(),
			"alicloud_ga_basic_endpoint":                                     resourceAliCloudGaBasicEndpoint(),
			"alicloud_ga_basic_endpoint_group":                               resourceAliCloudGaBasicEndpointGroup(),
			"alicloud_ga_basic_ip_set":                                       resourceAliCloudGaBasicIpSet(),
			"alicloud_ga_custom_routing_endpoint":                            resourceAliCloudGaCustomRoutingEndpoint(),
			"alicloud_ga_custom_routing_endpoint_group":                      resourceAliCloudGaCustomRoutingEndpointGroup(),
			"alicloud_ga_custom_routing_endpoint_group_destination":          resourceAliCloudGaCustomRoutingEndpointGroupDestination(),
			"alicloud_ga_custom_routing_endpoint_traffic_policy":             resourceAliCloudGaCustomRoutingEndpointTrafficPolicy(),
			"alicloud_ga_domain":                                             resourceAliCloudGaDomain(),
			"alicloud_ga_endpoint_group":                                     resourceAliCloudGaEndpointGroup(),
			"alicloud_ga_forwarding_rule":                                    resourceAliCloudGaForwardingRule(),
			"alicloud_ga_ip_set":                                             resourceAliCloudGaIpSet(),
			"alicloud_ga_listener":                                           resourceAliCloudGaListener(),
			"alicloud_governance_account":                                    resourceAliCloudGovernanceAccount(),
			"alicloud_governance_baseline":                                   resourceAliCloudGovernanceBaseline(),
			"alicloud_gpdb_account":                                          resourceAliCloudGpdbAccount(),
			"alicloud_gpdb_backup_policy":                                    resourceAliCloudGpdbBackupPolicy(),
			"alicloud_gpdb_connection":                                       resourceAliCloudGpdbConnection(),
			"alicloud_gpdb_database":                                         resourceAliCloudGpdbDatabase(),
			"alicloud_gpdb_db_instance_ip_array":                             resourceAliCloudGpdbDBInstanceIPArray(),
			"alicloud_gpdb_db_instance_plan":                                 resourceAliCloudGpdbDbInstancePlan(),
			"alicloud_gpdb_db_resource_group":                                resourceAliCloudGpdbDbResourceGroup(),
			"alicloud_gpdb_elastic_instance":                                 resourceAliCloudGpdbElasticInstance(),
			"alicloud_gpdb_external_data_service":                            resourceAliCloudGpdbExternalDataService(),
			"alicloud_gpdb_hadoop_data_source":                               resourceAliCloudGpdbHadoopDataSource(),
			"alicloud_gpdb_instance":                                         resourceAliCloudGpdbInstance(),
			"alicloud_gpdb_jdbc_data_source":                                 resourceAliCloudGpdbJdbcDataSource(),
			"alicloud_gpdb_remote_adb_data_source":                           resourceAliCloudGpdbRemoteADBDataSource(),
			"alicloud_gpdb_streaming_data_service":                           resourceAliCloudGpdbStreamingDataService(),
			"alicloud_gpdb_streaming_data_source":                            resourceAliCloudGpdbStreamingDataSource(),
			"alicloud_gpdb_streaming_job":                                    resourceAliCloudGpdbStreamingJob(),
			"alicloud_graph_database_db_instance":                            resourceAliCloudGraphDatabaseDbInstance(),
			"alicloud_gwlb_listener":                                         resourceAliCloudGwlbListener(),
			"alicloud_gwlb_load_balancer":                                    resourceAliCloudGwlbLoadBalancer(),
			"alicloud_gwlb_server_group":                                     resourceAliCloudGwlbServerGroup(),
			"alicloud_havip":                                                 resourceAliCloudVpcHaVip(),
			"alicloud_havip_attachment":                                      resourceAliCloudVpcHaVipAttachment(),
			"alicloud_hbase_instance":                                        resourceAliCloudHBaseInstance(),
			"alicloud_hbr_cross_account":                                     resourceAliCloudHbrCrossAccount(),
			"alicloud_hbr_ecs_backup_client":                                 resourceAliCloudHbrEcsBackupClient(),
			"alicloud_hbr_ecs_backup_plan":                                   resourceAliCloudHbrEcsBackupPlan(),
			"alicloud_hbr_hana_backup_client":                                resourceAliCloudHbrHanaBackupClient(),
			"alicloud_hbr_hana_backup_plan":                                  resourceAliCloudHbrHanaBackupPlan(),
			"alicloud_hbr_hana_instance":                                     resourceAliCloudHbrHanaInstance(),
			"alicloud_hbr_nas_backup_plan":                                   resourceAliCloudHbrNasBackupPlan(),
			"alicloud_hbr_oss_backup_plan":                                   resourceAliCloudHbrOssBackupPlan(),
			"alicloud_hbr_ots_backup_plan":                                   resourceAliCloudHbrOtsBackupPlan(),
			"alicloud_hbr_policy":                                            resourceAliCloudHbrPolicy(),
			"alicloud_hbr_policy_binding":                                    resourceAliCloudHbrPolicyBinding(),
			"alicloud_hbr_replication_vault":                                 resourceAliCloudHbrReplicationVault(),
			"alicloud_hbr_restore_job":                                       resourceAliCloudHbrRestoreJob(),
			"alicloud_hbr_server_backup_plan":                                resourceAliCloudHbrServerBackupPlan(),
			"alicloud_hbr_vault":                                             resourceAliCloudHbrVault(),
			"alicloud_hologram_instance":                                     resourceAliCloudHologramInstance(),
			"alicloud_image":                                                 resourceAliCloudEcsImage(),
			"alicloud_image_copy":                                            resourceAliCloudImageCopy(),
			"alicloud_image_export":                                          resourceAliCloudImageExport(),
			"alicloud_image_import":                                          resourceAliCloudImageImport(),
			"alicloud_image_share_permission":                                resourceAliCloudImageSharePermission(),
			"alicloud_imm_project":                                           resourceAliCloudImmProject(),
			"alicloud_imp_app_template":                                      resourceAliCloudImpAppTemplate(),
			"alicloud_ims_oidc_provider":                                     resourceAliCloudImsOidcProvider(),
			"alicloud_instance":                                              resourceAliCloudInstance(),
			"alicloud_iot_device_group":                                      resourceAliCloudIotDeviceGroup(),
			"alicloud_key_pair":                                              resourceAliCloudEcsKeyPair(),
			"alicloud_key_pair_attachment":                                   resourceAliCloudEcsKeyPairAttachment(),
			"alicloud_kms_alias":                                             resourceAliCloudKmsAlias(),
			"alicloud_kms_application_access_point":                          resourceAliCloudKmsApplicationAccessPoint(),
			"alicloud_kms_ciphertext":                                        resourceAliCloudKmsCiphertext(),
			"alicloud_kms_client_key":                                        resourceAliCloudKmsClientKey(),
			"alicloud_kms_instance":                                          resourceAliCloudKmsInstance(),
			"alicloud_kms_key":                                               resourceAliCloudKmsKey(),
			"alicloud_kms_key_version":                                       resourceAliCloudKmsKeyVersion(),
			"alicloud_kms_network_rule":                                      resourceAliCloudKmsNetworkRule(),
			"alicloud_kms_policy":                                            resourceAliCloudKmsPolicy(),
			"alicloud_kms_secret":                                            resourceAliCloudKmsSecret(),
			"alicloud_kvstore_account":                                       resourceAliCloudKvstoreAccount(),
			"alicloud_kvstore_audit_log_config":                              resourceAliCloudKvstoreAuditLogConfig(),
			"alicloud_kvstore_backup_policy":                                 resourceAliCloudKvStoreBackupPolicy(),
			"alicloud_kvstore_connection":                                    resourceAliCloudKvstoreConnection(),
			"alicloud_kvstore_instance":                                      resourceAliCloudKvstoreInstance(),
			"alicloud_launch_template":                                       resourceAliCloudEcsLaunchTemplate(),
			"alicloud_lindorm_instance":                                      resourceAliCloudLindormInstance(),
			"alicloud_lindorm_public_network":                                resourceAliCloudLindormPublicNetwork(),
			"alicloud_live_caster":                                           resourceAliCloudLiveCaster(),
			"alicloud_log_alert":                                             resourceAliCloudLogAlert(),
			"alicloud_log_alert_resource":                                    resourceAliCloudLogAlertResource(),
			"alicloud_log_dashboard":                                         resourceAliCloudLogDashboard(),
			"alicloud_log_etl":                                               resourceAliCloudLogETL(),
			"alicloud_log_machine_group":                                     resourceAliCloudLogMachineGroup(),
			"alicloud_log_oss_export":                                        resourceAliCloudLogOssExport(),
			"alicloud_log_oss_ingestion":                                     resourceAliCloudLogOssIngestion(),
			"alicloud_log_project":                                           resourceAliCloudLogProject(),
			"alicloud_log_project_logging":                                   resourceAliCloudLogProjectLogging(),
			"alicloud_log_store":                                             resourceAliCloudLogStore(),
			"alicloud_log_store_index":                                       resourceAliCloudLogStoreIndex(),
			"alicloud_logtail_attachment":                                    resourceAliCloudLogtailAttachment(),
			"alicloud_logtail_config":                                        resourceAliCloudLogtailConfig(),
			"alicloud_market_order":                                          resourceAliCloudMarketOrder(),
			"alicloud_max_compute_quota":                                     resourceAliCloudMaxComputeQuota(),
			"alicloud_max_compute_quota_plan":                                resourceAliCloudMaxComputeQuotaPlan(),
			"alicloud_max_compute_quota_schedule":                            resourceAliCloudMaxComputeQuotaSchedule(),
			"alicloud_max_compute_role":                                      resourceAliCloudMaxComputeRole(),
			"alicloud_max_compute_role_user_attachment":                      resourceAliCloudMaxComputeRoleUserAttachment(),
			"alicloud_max_compute_tunnel_quota_timer":                        resourceAliCloudMaxComputeTunnelQuotaTimer(),
			"alicloud_maxcompute_project":                                    resourceAliCloudMaxComputeProject(),
			"alicloud_message_service_endpoint":                              resourceAliCloudMessageServiceEndpoint(),
			"alicloud_message_service_endpoint_acl":                          resourceAliCloudMessageServiceEndpointAcl(),
			"alicloud_message_service_queue":                                 resourceAliCloudMessageServiceQueue(),
			"alicloud_message_service_service":                               resourceAliCloudMessageServiceService(),
			"alicloud_message_service_subscription":                          resourceAliCloudMessageServiceSubscription(),
			"alicloud_message_service_topic":                                 resourceAliCloudMessageServiceTopic(),
			"alicloud_mhub_app":                                              resourceAliCloudMhubApp(),
			"alicloud_mhub_product":                                          resourceAliCloudMhubProduct(),
			"alicloud_mns_queue":                                             resourceAliCloudMNSQueue(),
			"alicloud_mns_topic":                                             resourceAliCloudMNSTopic(),
			"alicloud_mns_topic_subscription":                                resourceAliCloudMNSSubscription(),
			"alicloud_mongodb_account":                                       resourceAliCloudMongodbAccount(),
			"alicloud_mongodb_audit_policy":                                  resourceAliCloudMongodbAuditPolicy(),
			"alicloud_mongodb_instance":                                      resourceAliCloudMongoDBInstance(),
			"alicloud_mongodb_private_srv_network_address":                   resourceAliCloudMongodbPrivateSrvNetworkAddress(),
			"alicloud_mongodb_public_network_address":                        resourceAliCloudMongoDBPublicNetworkAddress(),
			"alicloud_mongodb_replica_set_role":                              resourceAliCloudMongoDBReplicaSetRole(),
			"alicloud_mongodb_serverless_instance":                           resourceAliCloudMongodbServerlessInstance(),
			"alicloud_mongodb_sharding_instance":                             resourceAliCloudMongoDBShardingInstance(),
			"alicloud_mongodb_sharding_network_private_address":              resourceAliCloudMongodbShardingNetworkPrivateAddress(),
			"alicloud_mongodb_sharding_network_public_address":               resourceAliCloudMongodbShardingNetworkPublicAddress(),
			"alicloud_msc_sub_contact":                                       resourceAliCloudMscSubContact(),
			"alicloud_msc_sub_subscription":                                  resourceAliCloudMscSubSubscription(),
			"alicloud_msc_sub_webhook":                                       resourceAliCloudMscSubWebhook(),
			"alicloud_mse_cluster":                                           resourceAliCloudMseCluster(),
			"alicloud_mse_engine_namespace":                                  resourceAliCloudMseEngineNamespace(),
			"alicloud_mse_gateway":                                           resourceAliCloudMseGateway(),
			"alicloud_mse_nacos_config":                                      resourceAliCloudMseNacosConfig(),
			"alicloud_mse_znode":                                             resourceAliCloudMseZnode(),
			"alicloud_nas_access_group":                                      resourceAliCloudNasAccessGroup(),
			"alicloud_nas_access_point":                                      resourceAliCloudNasAccessPoint(),
			"alicloud_nas_access_rule":                                       resourceAliCloudNasAccessRule(),
			"alicloud_nas_auto_snapshot_policy":                              resourceAliCloudNasAutoSnapshotPolicy(),
			"alicloud_nas_data_flow":                                         resourceAliCloudNasDataFlow(),
			"alicloud_nas_file_system":                                       resourceAliCloudNasFileSystem(),
			"alicloud_nas_fileset":                                           resourceAliCloudNasFileset(),
			"alicloud_nas_lifecycle_policy":                                  resourceAliCloudNasLifecyclePolicy(),
			"alicloud_nas_mount_target":                                      resourceAliCloudNasMountTarget(),
			"alicloud_nas_recycle_bin":                                       resourceAliCloudNasRecycleBin(),
			"alicloud_nas_smb_acl_attachment":                                resourceAliCloudNasSmbAclAttachment(),
			"alicloud_nas_snapshot":                                          resourceAliCloudNasSnapshot(),
			"alicloud_nat_gateway":                                           resourceAliCloudNatGateway(),
			"alicloud_network_acl":                                           resourceAliCloudVpcNetworkAcl(),
			"alicloud_network_acl_attachment":                                resourceAliyunNetworkAclAttachment(),
			"alicloud_network_acl_entries":                                   resourceAliyunNetworkAclEntries(),
			"alicloud_network_interface":                                     resourceAliCloudEcsNetworkInterface(),
			"alicloud_network_interface_attachment":                          resourceAliCloudEcsNetworkInterfaceAttachment(),
			"alicloud_nlb_listener":                                          resourceAliCloudNlbListener(),
			"alicloud_nlb_listener_additional_certificate_attachment":        resourceAliCloudNlbListenerAdditionalCertificateAttachment(),
			"alicloud_nlb_load_balancer":                                     resourceAliCloudNlbLoadBalancer(),
			"alicloud_nlb_load_balancer_security_group_attachment":           resourceAliCloudNlbLoadBalancerSecurityGroupAttachment(),
			"alicloud_nlb_load_balancer_zone_shifted_attachment":             resourceAliCloudNlbLoadBalancerZoneShiftedAttachment(),
			"alicloud_nlb_loadbalancer_common_bandwidth_package_attachment":  resourceAliCloudNlbLoadbalancerCommonBandwidthPackageAttachment(),
			"alicloud_nlb_security_policy":                                   resourceAliCloudNlbSecurityPolicy(),
			"alicloud_nlb_server_group":                                      resourceAliCloudNlbServerGroup(),
			"alicloud_nlb_server_group_server_attachment":                    resourceAliCloudNlbServerGroupServerAttachment(),
			"alicloud_ocean_base_instance":                                   resourceAliCloudOceanBaseInstance(),
			"alicloud_ons_group":                                             resourceAliCloudOnsGroup(),
			"alicloud_ons_instance":                                          resourceAliCloudOnsInstance(),
			"alicloud_ons_topic":                                             resourceAliCloudOnsTopic(),
			"alicloud_oos_application":                                       resourceAliCloudOosApplication(),
			"alicloud_oos_application_group":                                 resourceAliCloudOosApplicationGroup(),
			"alicloud_oos_default_patch_baseline":                            resourceAliCloudOosDefaultPatchBaseline(),
			"alicloud_oos_execution":                                         resourceAliCloudOosExecution(),
			"alicloud_oos_parameter":                                         resourceAliCloudOosParameter(),
			"alicloud_oos_patch_baseline":                                    resourceAliCloudOosPatchBaseline(),
			"alicloud_oos_secret_parameter":                                  resourceAliCloudOosSecretParameter(),
			"alicloud_oos_service_setting":                                   resourceAliCloudOosServiceSetting(),
			"alicloud_oos_state_configuration":                               resourceAliCloudOosStateConfiguration(),
			"alicloud_oos_template":                                          resourceAliCloudOosTemplate(),
			"alicloud_open_search_app_group":                                 resourceAliCloudOpenSearchAppGroup(),
			"alicloud_oss_access_point":                                      resourceAliCloudOssAccessPoint(),
			"alicloud_oss_account_public_access_block":                       resourceAliCloudOssAccountPublicAccessBlock(),
			"alicloud_oss_bucket":                                            resourceAliCloudOssBucket(),
			"alicloud_oss_bucket_access_monitor":                             resourceAliCloudOssBucketAccessMonitor(),
			"alicloud_oss_bucket_acl":                                        resourceAliCloudOssBucketAcl(),
			"alicloud_oss_bucket_cname":                                      resourceAliCloudOssBucketCname(),
			"alicloud_oss_bucket_cname_token":                                resourceAliCloudOssBucketCnameToken(),
			"alicloud_oss_bucket_cors":                                       resourceAliCloudOssBucketCors(),
			"alicloud_oss_bucket_data_redundancy_transition":                 resourceAliCloudOssBucketDataRedundancyTransition(),
			"alicloud_oss_bucket_https_config":                               resourceAliCloudOssBucketHttpsConfig(),
			"alicloud_oss_bucket_lifecycle":                                  resourceAliCloudOssBucketLifecycle(),
			"alicloud_oss_bucket_logging":                                    resourceAliCloudOssBucketLogging(),
			"alicloud_oss_bucket_meta_query":                                 resourceAliCloudOssBucketMetaQuery(),
			"alicloud_oss_bucket_object":                                     resourceAliCloudOssBucketObject(),
			"alicloud_oss_bucket_policy":                                     resourceAliCloudOssBucketPolicy(),
			"alicloud_oss_bucket_public_access_block":                        resourceAliCloudOssBucketPublicAccessBlock(),
			"alicloud_oss_bucket_referer":                                    resourceAliCloudOssBucketReferer(),
			"alicloud_oss_bucket_replication":                                resourceAliCloudOssBucketReplication(),
			"alicloud_oss_bucket_request_payment":                            resourceAliCloudOssBucketRequestPayment(),
			"alicloud_oss_bucket_server_side_encryption":                     resourceAliCloudOssBucketServerSideEncryption(),
			"alicloud_oss_bucket_style":                                      resourceAliCloudOssBucketStyle(),
			"alicloud_oss_bucket_transfer_acceleration":                      resourceAliCloudOssBucketTransferAcceleration(),
			"alicloud_oss_bucket_user_defined_log_fields":                    resourceAliCloudOssBucketUserDefinedLogFields(),
			"alicloud_oss_bucket_versioning":                                 resourceAliCloudOssBucketVersioning(),
			"alicloud_oss_bucket_website":                                    resourceAliCloudOssBucketWebsite(),
			"alicloud_oss_bucket_worm":                                       resourceAliCloudOssBucketWorm(),
			"alicloud_ots_index":                                             resourceAliCloudOtsIndex(),
			"alicloud_ots_instance":                                          resourceAliCloudOtsInstance(),
			"alicloud_ots_instance_attachment":                               resourceAliCloudOtsInstanceAttachment(),
			// "alicloud_ots_tunnel":                                            resourceAliCloudOtsTunnel(),
			"alicloud_ots_search_index":                                 resourceAliCloudOtsSearchIndex(),
			"alicloud_ots_table":                                        resourceAliCloudOtsTable(),
			"alicloud_pai_flow_pipeline":                                resourceAliCloudPaiFlowPipeline(),
			"alicloud_pai_service":                                      resourceAliCloudPaiService(),
			"alicloud_pai_workspace_code_source":                        resourceAliCloudPaiWorkspaceCodeSource(),
			"alicloud_pai_workspace_dataset":                            resourceAliCloudPaiWorkspaceDataset(),
			"alicloud_pai_workspace_datasetversion":                     resourceAliCloudPaiWorkspaceDatasetversion(),
			"alicloud_pai_workspace_experiment":                         resourceAliCloudPaiWorkspaceExperiment(),
			"alicloud_pai_workspace_member":                             resourceAliCloudPaiWorkspaceMember(),
			"alicloud_pai_workspace_model":                              resourceAliCloudPaiWorkspaceModel(),
			"alicloud_pai_workspace_model_version":                      resourceAliCloudPaiWorkspaceModelVersion(),
			"alicloud_pai_workspace_run":                                resourceAliCloudPaiWorkspaceRun(),
			"alicloud_pai_workspace_user_config":                        resourceAliCloudPaiWorkspaceUserConfig(),
			"alicloud_pai_workspace_workspace":                          resourceAliCloudPaiWorkspaceWorkspace(),
			"alicloud_polardb_account":                                  resourceAliCloudPolarDBAccount(),
			"alicloud_polardb_account_privilege":                        resourceAliCloudPolarDBAccountPrivilege(),
			"alicloud_polardb_backup_policy":                            resourceAliCloudPolarDBBackupPolicy(),
			"alicloud_polardb_cluster":                                  resourceAliCloudPolarDBCluster(),
			"alicloud_polardb_cluster_endpoint":                         resourceAliCloudPolarDBClusterEndpoint(),
			"alicloud_polardb_database":                                 resourceAliCloudPolarDBDatabase(),
			"alicloud_polardb_endpoint":                                 resourceAliCloudPolarDBEndpoint(),
			"alicloud_polardb_endpoint_address":                         resourceAliCloudPolarDBEndpointAddress(),
			"alicloud_polardb_global_database_network":                  resourceAliCloudPolarDBGlobalDatabaseNetwork(),
			"alicloud_polardb_parameter_group":                          resourceAliCloudPolarDBParameterGroup(),
			"alicloud_polardb_primary_endpoint":                         resourceAliCloudPolarDBPrimaryEndpoint(),
			"alicloud_privatelink_vpc_endpoint":                         resourceAliCloudPrivateLinkVpcEndpoint(),
			"alicloud_privatelink_vpc_endpoint_connection":              resourceAliCloudPrivateLinkVpcEndpointConnection(),
			"alicloud_privatelink_vpc_endpoint_service":                 resourceAliCloudPrivateLinkVpcEndpointService(),
			"alicloud_privatelink_vpc_endpoint_service_resource":        resourceAliCloudPrivateLinkVpcEndpointServiceResource(),
			"alicloud_privatelink_vpc_endpoint_service_user":            resourceAliCloudPrivateLinkVpcEndpointServiceUser(),
			"alicloud_privatelink_vpc_endpoint_zone":                    resourceAliCloudPrivateLinkVpcEndpointZone(),
			"alicloud_pvtz_endpoint":                                    resourceAliCloudPvtzEndpoint(),
			"alicloud_pvtz_rule":                                        resourceAliCloudPvtzRule(),
			"alicloud_pvtz_rule_attachment":                             resourceAliCloudPvtzRuleAttachment(),
			"alicloud_pvtz_user_vpc_authorization":                      resourceAliCloudPvtzUserVpcAuthorization(),
			"alicloud_pvtz_zone":                                        resourceAliCloudPvtzZone(),
			"alicloud_pvtz_zone_attachment":                             resourceAliCloudPvtzZoneAttachment(),
			"alicloud_pvtz_zone_record":                                 resourceAliCloudPvtzZoneRecord(),
			"alicloud_quick_bi_user":                                    resourceAliCloudQuickBiUser(),
			"alicloud_quotas_application_info":                          resourceAliCloudQuotasQuotaApplication(),
			"alicloud_quotas_quota_alarm":                               resourceAliCloudQuotasQuotaAlarm(),
			"alicloud_quotas_quota_application":                         resourceAliCloudQuotasQuotaApplication(),
			"alicloud_quotas_template_applications":                     resourceAliCloudQuotasTemplateApplications(),
			"alicloud_quotas_template_quota":                            resourceAliCloudQuotasTemplateQuota(),
			"alicloud_quotas_template_service":                          resourceAliCloudQuotasTemplateService(),
			"alicloud_ram_access_key":                                   resourceAliCloudRamAccessKey(),
			"alicloud_ram_account_alias":                                resourceAliCloudRamAccountAlias(),
			"alicloud_ram_account_password_policy":                      resourceAliCloudRamAccountPasswordPolicy(),
			"alicloud_ram_alias":                                        resourceAliCloudRamAccountAlias(),
			"alicloud_ram_group":                                        resourceAliCloudRamGroup(),
			"alicloud_ram_group_membership":                             resourceAliCloudRamGroupMembership(),
			"alicloud_ram_group_policy_attachment":                      resourceAliCloudRamGroupPolicyAttachment(),
			"alicloud_ram_login_profile":                                resourceAliCloudRamLoginProfile(),
			"alicloud_ram_password_policy":                              resourceAliCloudRamPasswordPolicy(),
			"alicloud_ram_policy":                                       resourceAliCloudRamPolicy(),
			"alicloud_ram_role":                                         resourceAliCloudRamRole(),
			"alicloud_ram_role_attachment":                              resourceAliCloudRamRoleAttachment(),
			"alicloud_ram_role_policy_attachment":                       resourceAliCloudRamRolePolicyAttachment(),
			"alicloud_ram_saml_provider":                                resourceAliCloudRamSamlProvider(),
			"alicloud_ram_security_preference":                          resourceAliCloudRamSecurityPreference(),
			"alicloud_ram_user":                                         resourceAliCloudRamUser(),
			"alicloud_ram_user_group_attachment":                        resourceAliCloudRamUserGroupAttachment(),
			"alicloud_ram_user_policy_attachment":                       resourceAliCloudRamUserPolicyAttachment(),
			"alicloud_rdc_organization":                                 resourceAliCloudRdcOrganization(),
			"alicloud_rds_account":                                      resourceAliCloudRdsAccount(),
			"alicloud_rds_backup":                                       resourceAliCloudRdsBackup(),
			"alicloud_rds_clone_db_instance":                            resourceAliCloudRdsCloneDbInstance(),
			"alicloud_rds_custom":                                       resourceAliCloudRdsCustom(),
			"alicloud_rds_custom_deployment_set":                        resourceAliCloudRdsCustomDeploymentSet(),
			"alicloud_rds_custom_disk":                                  resourceAliCloudRdsCustomDisk(),
			"alicloud_rds_db_instance_endpoint":                         resourceAliCloudRdsDBInstanceEndpoint(),
			"alicloud_rds_db_instance_endpoint_address":                 resourceAliCloudRdsDBInstanceEndpointAddress(),
			"alicloud_rds_db_node":                                      resourceAliCloudRdsDBNode(),
			"alicloud_rds_db_proxy":                                     resourceAliCloudRdsDBProxy(),
			"alicloud_rds_db_proxy_public":                              resourceAliCloudRdsDBProxyPublic(),
			"alicloud_rds_ddr_instance":                                 resourceAliCloudRdsDdrInstance(),
			"alicloud_rds_instance_cross_backup_policy":                 resourceAliCloudRdsInstanceCrossBackupPolicy(),
			"alicloud_rds_parameter_group":                              resourceAliCloudRdsParameterGroup(),
			"alicloud_rds_service_linked_role":                          resourceAliCloudRdsServiceLinkedRole(),
			"alicloud_rds_upgrade_db_instance":                          resourceAliCloudRdsUpgradeDbInstance(),
			"alicloud_realtime_compute_vvp_instance":                    resourceAliCloudRealtimeComputeVvpInstance(),
			"alicloud_redis_tair_instance":                              resourceAliCloudRedisTairInstance(),
			"alicloud_reserved_instance":                                resourceAliCloudReservedInstance(),
			"alicloud_resource_manager_account":                         resourceAliCloudResourceManagerAccount(),
			"alicloud_resource_manager_auto_grouping_rule":              resourceAliCloudResourceManagerAutoGroupingRule(),
			"alicloud_resource_manager_control_policy":                  resourceAliCloudResourceManagerControlPolicy(),
			"alicloud_resource_manager_control_policy_attachment":       resourceAliCloudResourceManagerControlPolicyAttachment(),
			"alicloud_resource_manager_delegated_administrator":         resourceAliCloudResourceManagerDelegatedAdministrator(),
			"alicloud_resource_manager_folder":                          resourceAliCloudResourceManagerFolder(),
			"alicloud_resource_manager_handshake":                       resourceAliCloudResourceManagerHandshake(),
			"alicloud_resource_manager_policy":                          resourceAliCloudResourceManagerPolicy(),
			"alicloud_resource_manager_policy_attachment":               resourceAliCloudResourceManagerPolicyAttachment(),
			"alicloud_resource_manager_policy_version":                  resourceAliCloudResourceManagerPolicyVersion(),
			"alicloud_resource_manager_resource_directory":              resourceAliCloudResourceManagerResourceDirectory(),
			"alicloud_resource_manager_resource_group":                  resourceAliCloudResourceManagerResourceGroup(),
			"alicloud_resource_manager_resource_share":                  resourceAliCloudResourceManagerResourceShare(),
			"alicloud_resource_manager_role":                            resourceAliCloudResourceManagerRole(),
			"alicloud_resource_manager_saved_query":                     resourceAliCloudResourceManagerSavedQuery(),
			"alicloud_resource_manager_service_linked_role":             resourceAliCloudResourceManagerServiceLinkedRole(),
			"alicloud_resource_manager_shared_resource":                 resourceAliCloudResourceManagerSharedResource(),
			"alicloud_resource_manager_shared_target":                   resourceAliCloudResourceManagerSharedTarget(),
			"alicloud_rocketmq_account":                                 resourceAliCloudRocketmqAccount(),
			"alicloud_rocketmq_acl":                                     resourceAliCloudRocketmqAcl(),
			"alicloud_rocketmq_consumer_group":                          resourceAliCloudRocketmqConsumerGroup(),
			"alicloud_rocketmq_instance":                                resourceAliCloudRocketmqInstance(),
			"alicloud_rocketmq_topic":                                   resourceAliCloudRocketmqTopic(),
			"alicloud_ros_change_set":                                   resourceAliCloudRosChangeSet(),
			"alicloud_ros_stack":                                        resourceAliCloudRosStack(),
			"alicloud_ros_stack_group":                                  resourceAliCloudRosStackGroup(),
			"alicloud_ros_stack_instance":                               resourceAliCloudRosStackInstance(),
			"alicloud_ros_template":                                     resourceAliCloudRosTemplate(),
			"alicloud_ros_template_scratch":                             resourceAliCloudRosTemplateScratch(),
			"alicloud_route_entry":                                      resourceAliyunRouteEntry(),
			"alicloud_route_table":                                      resourceAliCloudVpcRouteTable(),
			"alicloud_route_table_attachment":                           resourceAliCloudVpcRouteTableAttachment(),
			"alicloud_router_interface":                                 resourceAliCloudRouterInterface(),
			"alicloud_router_interface_connection":                      resourceAliCloudRouterInterfaceConnection(),
			"alicloud_sae_application":                                  resourceAliCloudSaeApplication(),
			"alicloud_sae_application_scaling_rule":                     resourceAliCloudSaeApplicationScalingRule(),
			"alicloud_sae_config_map":                                   resourceAliCloudSaeConfigMap(),
			"alicloud_sae_grey_tag_route":                               resourceAliCloudSaeGreyTagRoute(),
			"alicloud_sae_ingress":                                      resourceAliCloudSaeIngress(),
			"alicloud_sae_load_balancer_internet":                       resourceAliCloudSaeLoadBalancerInternet(),
			"alicloud_sae_load_balancer_intranet":                       resourceAliCloudSaeLoadBalancerIntranet(),
			"alicloud_sae_namespace":                                    resourceAliCloudSaeNamespace(),
			"alicloud_sag_acl":                                          resourceAliCloudSagAcl(),
			"alicloud_sag_acl_rule":                                     resourceAliCloudSagAclRule(),
			"alicloud_sag_client_user":                                  resourceAliCloudSagClientUser(),
			"alicloud_sag_dnat_entry":                                   resourceAliCloudSagDnatEntry(),
			"alicloud_sag_qos":                                          resourceAliCloudSagQos(),
			"alicloud_sag_qos_car":                                      resourceAliCloudSagQosCar(),
			"alicloud_sag_qos_policy":                                   resourceAliCloudSagQosPolicy(),
			"alicloud_sag_snat_entry":                                   resourceAliCloudSagSnatEntry(),
			"alicloud_scdn_domain":                                      resourceAliCloudScdnDomain(),
			"alicloud_scdn_domain_config":                               resourceAliCloudScdnDomainConfig(),
			"alicloud_schedulerx_app_group":                             resourceAliCloudSchedulerxAppGroup(),
			"alicloud_schedulerx_job":                                   resourceAliCloudSchedulerxJob(),
			"alicloud_schedulerx_namespace":                             resourceAliCloudSchedulerxNamespace(),
			"alicloud_sddp_config":                                      resourceAliCloudSddpConfig(),
			"alicloud_sddp_data_limit":                                  resourceAliCloudSddpDataLimit(),
			"alicloud_sddp_instance":                                    resourceAliCloudSddpInstance(),
			"alicloud_sddp_rule":                                        resourceAliCloudSddpRule(),
			"alicloud_security_center_group":                            resourceAliCloudSecurityCenterGroup(),
			"alicloud_security_center_service_linked_role":              resourceAliCloudSecurityCenterServiceLinkedRole(),
			"alicloud_security_group":                                   resourceAliCloudEcsSecurityGroup(),
			"alicloud_security_group_rule":                              resourceAliyunSecurityGroupRule(),
			"alicloud_selectdb_cluster":                                 resourceAliCloudSelectDBCluster(),
			"alicloud_selectdb_cluster_binding":                         resourceAliCloudSelectDBClusterBinding(),
			"alicloud_selectdb_instance":                                resourceAliCloudSelectDBInstance(),
			"alicloud_selectdb_public_connection":                       resourceAliCloudSelectDBPublicConnection(),
			"alicloud_service_catalog_portfolio":                        resourceAliCloudServiceCatalogPortfolio(),
			"alicloud_service_catalog_principal_portfolio_association":  resourceAliCloudServiceCatalogPrincipalPortfolioAssociation(),
			"alicloud_service_catalog_product":                          resourceAliCloudServiceCatalogProduct(),
			"alicloud_service_catalog_product_portfolio_association":    resourceAliCloudServiceCatalogProductPortfolioAssociation(),
			"alicloud_service_catalog_product_version":                  resourceAliCloudServiceCatalogProductVersion(),
			"alicloud_service_catalog_provisioned_product":              resourceAliCloudServiceCatalogProvisionedProduct(),
			"alicloud_service_mesh_extension_provider":                  resourceAliCloudServiceMeshExtensionProvider(),
			"alicloud_service_mesh_service_mesh":                        resourceAliCloudServiceMeshServiceMesh(),
			"alicloud_service_mesh_user_permission":                     resourceAliCloudServiceMeshUserPermission(),
			"alicloud_simple_application_server_custom_image":           resourceAliCloudSimpleApplicationServerCustomImage(),
			"alicloud_simple_application_server_firewall_rule":          resourceAliCloudSimpleApplicationServerFirewallRule(),
			"alicloud_simple_application_server_instance":               resourceAliCloudSimpleApplicationServerInstance(),
			"alicloud_simple_application_server_snapshot":               resourceAliCloudSimpleApplicationServerSnapshot(),
			"alicloud_slb":                                              resourceAliCloudSlbLoadBalancer(),
			"alicloud_slb_acl":                                          resourceAliCloudSlbAcl(),
			"alicloud_slb_acl_entry_attachment":                         resourceAliCloudSlbAclEntryAttachment(),
			"alicloud_slb_attachment":                                   resourceAliyunSlbAttachment(),
			"alicloud_slb_backend_server":                               resourceAliyunSlbBackendServer(),
			"alicloud_slb_ca_certificate":                               resourceAliCloudSlbCaCertificate(),
			"alicloud_slb_domain_extension":                             resourceAliCloudSlbDomainExtension(),
			"alicloud_slb_listener":                                     resourceAliCloudSlbListener(),
			"alicloud_slb_load_balancer":                                resourceAliCloudSlbLoadBalancer(),
			"alicloud_slb_master_slave_server_group":                    resourceAliyunSlbMasterSlaveServerGroup(),
			"alicloud_slb_rule":                                         resourceAliyunSlbRule(),
			"alicloud_slb_server_certificate":                           resourceAliCloudSlbServerCertificate(),
			"alicloud_slb_server_group":                                 resourceAliyunSlbServerGroup(),
			"alicloud_slb_server_group_server_attachment":               resourceAliCloudSlbServerGroupServerAttachment(),
			"alicloud_slb_tls_cipher_policy":                            resourceAliCloudSlbTlsCipherPolicy(),
			"alicloud_sls_alert":                                        resourceAliCloudSlsAlert(),
			"alicloud_sls_collection_policy":                            resourceAliCloudSlsCollectionPolicy(),
			"alicloud_sls_etl":                                          resourceAliCloudLogETL(),
			"alicloud_sls_scheduled_sql":                                resourceAliCloudSlsScheduledSQL(),
			"alicloud_sls_consumer_group":                               resourceAliCloudSlsConsumerGroup(),
			"alicloud_smartag_flow_log":                                 resourceAliCloudSmartagFlowLog(),
			"alicloud_sms_short_url":                                    resourceAliCloudSmsShortUrl(),
			"alicloud_snapshot":                                         resourceAliCloudEcsSnapshot(),
			"alicloud_snapshot_policy":                                  resourceAliCloudEcsAutoSnapshotPolicy(),
			"alicloud_snat_entry":                                       resourceAliCloudNATGatewaySnatEntry(),
			"alicloud_ssl_certificates_service_certificate":             resourceAliCloudSslCertificatesServiceCertificate(),
			"alicloud_ssl_vpn_client_cert":                              resourceAliyunSslVpnClientCert(),
			"alicloud_ssl_vpn_server":                                   resourceAliyunSslVpnServer(),
			"alicloud_subnet":                                           resourceAliCloudVpcVswitch(),
			"alicloud_tag_associated_rule":                              resourceAliCloudTagAssociatedRule(),
			"alicloud_tag_meta_tag":                                     resourceAliCloudTagMetaTag(),
			"alicloud_tag_policy":                                       resourceAliCloudTagPolicy(),
			"alicloud_tag_policy_attachment":                            resourceAliCloudTagPolicyAttachment(),
			"alicloud_threat_detection_anti_brute_force_rule":           resourceAliCloudThreatDetectionAntiBruteForceRule(),
			"alicloud_threat_detection_asset_bind":                      resourceAliCloudThreatDetectionAssetBind(),
			"alicloud_threat_detection_asset_selection_config":          resourceAliCloudThreatDetectionAssetSelectionConfig(),
			"alicloud_threat_detection_backup_policy":                   resourceAliCloudThreatDetectionBackupPolicy(),
			"alicloud_threat_detection_baseline_strategy":               resourceAliCloudThreatDetectionBaselineStrategy(),
			"alicloud_threat_detection_client_file_protect":             resourceAliCloudThreatDetectionClientFileProtect(),
			"alicloud_threat_detection_client_user_define_rule":         resourceAliCloudThreatDetectionClientUserDefineRule(),
			"alicloud_threat_detection_file_upload_limit":               resourceAliCloudThreatDetectionFileUploadLimit(),
			"alicloud_threat_detection_honey_pot":                       resourceAliCloudThreatDetectionHoneyPot(),
			"alicloud_threat_detection_honeypot_node":                   resourceAliCloudThreatDetectionHoneypotNode(),
			"alicloud_threat_detection_honeypot_preset":                 resourceAliCloudThreatDetectionHoneypotPreset(),
			"alicloud_threat_detection_honeypot_probe":                  resourceAliCloudThreatDetectionHoneypotProbe(),
			"alicloud_threat_detection_image_event_operation":           resourceAliCloudThreatDetectionImageEventOperation(),
			"alicloud_threat_detection_instance":                        resourceAliCloudThreatDetectionInstance(),
			"alicloud_threat_detection_log_meta":                        resourceAliCloudThreatDetectionLogMeta(),
			"alicloud_threat_detection_malicious_file_whitelist_config": resourceAliCloudThreatDetectionMaliciousFileWhitelistConfig(),
			"alicloud_threat_detection_oss_scan_config":                 resourceAliCloudThreatDetectionOssScanConfig(),
			"alicloud_threat_detection_sas_trail":                       resourceAliCloudThreatDetectionSasTrail(),
			"alicloud_threat_detection_vul_whitelist":                   resourceAliCloudThreatDetectionVulWhitelist(),
			"alicloud_threat_detection_web_lock_config":                 resourceAliCloudThreatDetectionWebLockConfig(),
			"alicloud_tsdb_instance":                                    resourceAliCloudTsdbInstance(),
			"alicloud_video_surveillance_system_group":                  resourceAliCloudVideoSurveillanceSystemGroup(),
			"alicloud_vod_domain":                                       resourceAliCloudVodDomain(),
			"alicloud_vod_editing_project":                              resourceAliCloudVodEditingProject(),
			"alicloud_vpc":                                              resourceAliCloudVpcVpc(),
			"alicloud_vpc_bgp_group":                                    resourceAliCloudVpcBgpGroup(),
			"alicloud_vpc_bgp_network":                                  resourceAliCloudVpcBgpNetwork(),
			"alicloud_vpc_bgp_peer":                                     resourceAliCloudExpressConnectBgpPeer(),
			"alicloud_vpc_dhcp_options_set":                             resourceAliCloudVpcDhcpOptionsSet(),
			"alicloud_vpc_dhcp_options_set_attachment":                  resourceAliCloudVpcDhcpOptionsSetAttachement(),
			"alicloud_vpc_flow_log":                                     resourceAliCloudVpcFlowLog(),
			"alicloud_vpc_gateway_endpoint":                             resourceAliCloudVpcGatewayEndpoint(),
			"alicloud_vpc_gateway_endpoint_route_table_attachment":      resourceAliCloudVpcGatewayEndpointRouteTableAttachment(),
			"alicloud_vpc_gateway_route_table_attachment":               resourceAliCloudVpcGatewayRouteTableAttachment(),
			"alicloud_vpc_ha_vip":                                       resourceAliCloudVpcHaVip(),
			"alicloud_vpc_ipam_ipam":                                    resourceAliCloudVpcIpamIpam(),
			"alicloud_vpc_ipam_ipam_pool":                               resourceAliCloudVpcIpamIpamPool(),
			"alicloud_vpc_ipam_ipam_pool_allocation":                    resourceAliCloudVpcIpamIpamPoolAllocation(),
			"alicloud_vpc_ipam_ipam_pool_cidr":                          resourceAliCloudVpcIpamIpamPoolCidr(),
			"alicloud_vpc_ipam_ipam_resource_discovery":                 resourceAliCloudVpcIpamIpamResourceDiscovery(),
			"alicloud_vpc_ipam_ipam_scope":                              resourceAliCloudVpcIpamIpamScope(),
			"alicloud_vpc_ipam_service":                                 resourceAliCloudVpcIpamService(),
			"alicloud_vpc_ipv4_cidr_block":                              resourceAliCloudVpcIpv4CidrBlock(),
			"alicloud_vpc_ipv4_gateway":                                 resourceAliCloudVpcIpv4Gateway(),
			"alicloud_vpc_ipv6_address":                                 resourceAliCloudVpcIpv6Address(),
			"alicloud_vpc_ipv6_egress_rule":                             resourceAliCloudVpcIpv6EgressRule(),
			"alicloud_vpc_ipv6_gateway":                                 resourceAliCloudVpcIpv6Gateway(),
			"alicloud_vpc_ipv6_internet_bandwidth":                      resourceAliCloudVpcIpv6InternetBandwidth(),
			"alicloud_vpc_nat_ip":                                       resourceAliCloudVpcNatIp(),
			"alicloud_vpc_nat_ip_cidr":                                  resourceAliCloudVpcNatIpCidr(),
			"alicloud_vpc_network_acl_attachment":                       resourceAliCloudVpcNetworkAclAttachment(),
			"alicloud_vpc_peer_connection":                              resourceAliCloudVpcPeerPeerConnection(),
			"alicloud_vpc_peer_connection_accepter":                     resourceAliCloudVpcPeerPeerConnectionAccepter(),
			"alicloud_vpc_prefix_list":                                  resourceAliCloudVpcPrefixList(),
			"alicloud_vpc_public_ip_address_pool":                       resourceAliCloudVpcPublicIpAddressPool(),
			"alicloud_vpc_public_ip_address_pool_cidr_block":            resourceAliCloudVpcPublicIpAddressPoolCidrBlock(),
			"alicloud_vpc_route_entry":                                  resourceAliCloudVpcRouteEntry(),
			"alicloud_vpc_traffic_mirror_filter":                        resourceAliCloudVpcTrafficMirrorFilter(),
			"alicloud_vpc_traffic_mirror_filter_egress_rule":            resourceAliCloudVpcTrafficMirrorFilterEgressRule(),
			"alicloud_vpc_traffic_mirror_filter_ingress_rule":           resourceAliCloudVpcTrafficMirrorFilterIngressRule(),
			"alicloud_vpc_traffic_mirror_session":                       resourceAliCloudVpcTrafficMirrorSession(),
			"alicloud_vpc_vbr_ha":                                       resourceAliCloudVpcVbrHa(),
			"alicloud_vpc_vswitch_cidr_reservation":                     resourceAliCloudVpcVswitchCidrReservation(),
			"alicloud_vpn_connection":                                   resourceAliCloudVPNGatewayVpnConnection(),
			"alicloud_vpn_customer_gateway":                             resourceAliCloudVPNGatewayCustomerGateway(),
			"alicloud_vpn_gateway":                                      resourceAliCloudVPNGatewayVPNGateway(),
			"alicloud_vpn_gateway_vco_route":                            resourceAliCloudVpnGatewayVcoRoute(),
			"alicloud_vpn_gateway_vpn_attachment":                       resourceAliCloudVpnGatewayVpnAttachment(),
			"alicloud_vpn_ipsec_server":                                 resourceAliCloudVpnIpsecServer(),
			"alicloud_vpn_pbr_route_entry":                              resourceAliCloudVpnPbrRouteEntry(),
			"alicloud_vpn_route_entry":                                  resourceAliyunVpnRouteEntry(),
			"alicloud_vswitch":                                          resourceAliCloudVpcVswitch(),
			"alicloud_waf_certificate":                                  resourceAliCloudWafCertificate(),
			"alicloud_waf_domain":                                       resourceAliCloudWafDomain(),
			"alicloud_waf_instance":                                     resourceAliCloudWafInstance(),
			"alicloud_waf_protection_module":                            resourceAliCloudWafProtectionModule(),
			"alicloud_wafv3_defense_template":                           resourceAliCloudWafv3DefenseTemplate(),
			"alicloud_wafv3_domain":                                     resourceAliCloudWafv3Domain(),
			"alicloud_wafv3_instance":                                   resourceAliCloudWafv3Instance(),
			"alicloud_yundun_bastionhost_instance":                      resourceAliCloudBastionhostInstance(),
			"alicloud_yundun_dbaudit_instance":                          resourceAliCloudDbauditInstance(),
		},
	}
	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		return providerConfigure(d, provider)
	}
	return provider
}

var providerConfig map[string]interface{}

func providerConfigure(d *schema.ResourceData, p *schema.Provider) (interface{}, error) {
	log.Println("using terraform version:", p.TerraformVersion)
	var getProviderConfig = func(schemaKey string, profileKey string) string {
		if schemaKey != "" {
			if v, ok := d.GetOk(schemaKey); ok && v != nil && v.(string) != "" {
				return v.(string)
			}
		}
		if v, err := getConfigFromProfile(d, profileKey); err == nil && v != nil {
			return v.(string)
		}
		return ""
	}

	accessKey := getProviderConfig("access_key", "access_key_id")
	secretKey := getProviderConfig("secret_key", "access_key_secret")
	region := getProviderConfig("region", "region_id")
	if region == "" {
		region = DEFAULT_REGION
	}
	securityToken := getProviderConfig("security_token", "sts_token")

	ecsRoleName := getProviderConfig("ecs_role_name", "ram_role_name")
	var profileName string
	var credential credentials.Credential
	if v, ok := d.GetOk("profile"); ok && v != nil {
		profileName = v.(string)
	}

	// TODO: supports all of profile modes after credentials supporting setting timeout
	if (accessKey == "" || secretKey == "") && profileName != "" && fmt.Sprint(providerConfig["mode"]) == "ChainableRamRoleArn" {
		var profileFile string
		if v, ok := d.GetOk("shared_credentials_file"); ok && v.(string) != "" {
			profileFile = absPath(v.(string))
		}
		provider, err := providers.NewCLIProfileCredentialsProviderBuilder().WithProfileName(profileName).WithProfileFile(profileFile).Build()
		if err != nil {
			return nil, fmt.Errorf("failed to create profile credentials provider: %v", err)
		}
		credential = credentials.FromCredentialsProvider("cli_profile", provider)
		creds, err := credential.GetCredential()
		if err != nil {
			return nil, fmt.Errorf("failed to get credential from profile: %v", err)
		}
		accessKey, secretKey, securityToken = *creds.AccessKeyId, *creds.AccessKeySecret, *creds.SecurityToken
	}

	if accessKey == "" || secretKey == "" {
		if v, ok := d.GetOk("credentials_uri"); ok && v.(string) != "" {
			credentialsURIResp, err := getClientByCredentialsURI(v.(string))
			if err != nil {
				return nil, err
			}
			accessKey = credentialsURIResp.AccessKeyId
			secretKey = credentialsURIResp.AccessKeySecret
			securityToken = credentialsURIResp.SecurityToken
		}
	}

	config := &connectivity.Config{
		AccessKey:            strings.TrimSpace(accessKey),
		SecretKey:            strings.TrimSpace(secretKey),
		SecurityToken:        strings.TrimSpace(securityToken),
		EcsRoleName:          strings.TrimSpace(ecsRoleName),
		Region:               connectivity.Region(strings.TrimSpace(region)),
		RegionId:             strings.TrimSpace(region),
		SkipRegionValidation: d.Get("skip_region_validation").(bool),
		ConfigurationSource:  d.Get("configuration_source").(string),
		Protocol:             d.Get("protocol").(string),
		ClientReadTimeout:    d.Get("client_read_timeout").(int),
		ClientConnectTimeout: d.Get("client_connect_timeout").(int),
		SourceIp:             strings.TrimSpace(d.Get("source_ip").(string)),
		SecureTransport:      strings.TrimSpace(d.Get("secure_transport").(string)),
		MaxRetryTimeout:      d.Get("max_retry_timeout").(int),
		TerraformTraceId:     strings.Trim(uuid.New().String(), "-"),
		TerraformVersion:     p.TerraformVersion,
	}
	if credential != nil {
		config.Credential = credential
	}
	log.Println("alicloud provider trace id:", config.TerraformTraceId)
	if accessKey != "" && secretKey != "" && credential == nil {
		credentialConfig := new(credentials.Config).SetType("access_key").
			SetAccessKeyId(accessKey).
			SetAccessKeySecret(secretKey).
			SetTimeout(config.ClientReadTimeout).
			SetConnectTimeout(config.ClientConnectTimeout)
		if v := strings.TrimSpace(securityToken); v != "" {
			credentialConfig.SetType("sts").SetSecurityToken(v)
		}
		if config.ClientConnectTimeout != 0 {
			credentialConfig.SetConnectTimeout(config.ClientConnectTimeout)
		}
		if config.ClientReadTimeout != 0 {
			credentialConfig.SetTimeout(config.ClientReadTimeout)
		}
		credential, err := credentials.NewCredential(credentialConfig)
		if err != nil {
			return nil, err
		}
		config.Credential = credential
	}
	if account, ok := d.GetOk("account_id"); ok && account.(string) != "" {
		config.AccountId = strings.TrimSpace(account.(string))
	}
	if v, ok := d.GetOk("account_type"); ok && v.(string) != "" {
		config.AccountType = v.(string)
	}
	if v, ok := d.GetOk("security_transport"); config.SecureTransport == "" && ok && v.(string) != "" {
		config.SecureTransport = v.(string)
	}

	config.RamRoleArn = getProviderConfig("", "ram_role_arn")
	config.RamRoleSessionName = getProviderConfig("", "ram_session_name")
	expiredSeconds, err := getConfigFromProfile(d, "expired_seconds")
	if err == nil && expiredSeconds != nil {
		config.RamRoleSessionExpiration = (int)(expiredSeconds.(float64))
	}

	assumeRoleList := d.Get("assume_role").(*schema.Set).List()
	if len(assumeRoleList) == 1 {
		assumeRole := assumeRoleList[0].(map[string]interface{})
		if assumeRole["role_arn"].(string) != "" {
			config.RamRoleArn = assumeRole["role_arn"].(string)
		}
		if assumeRole["session_name"].(string) != "" {
			config.RamRoleSessionName = assumeRole["session_name"].(string)
		}
		if config.RamRoleSessionName == "" {
			config.RamRoleSessionName = "terraform"
		}
		config.RamRolePolicy = assumeRole["policy"].(string)
		if assumeRole["session_expiration"].(int) == 0 {
			if v := os.Getenv("ALICLOUD_ASSUME_ROLE_SESSION_EXPIRATION"); v != "" {
				if expiredSeconds, err := strconv.Atoi(v); err == nil {
					config.RamRoleSessionExpiration = expiredSeconds
				}
			}
			if config.RamRoleSessionExpiration == 0 {
				config.RamRoleSessionExpiration = 3600
			}
		} else {
			config.RamRoleSessionExpiration = assumeRole["session_expiration"].(int)
		}
		if v := assumeRole["external_id"].(string); v != "" {
			config.RamRoleExternalId = v
		}

		log.Printf("[INFO] assume_role configuration set: (RamRoleArn: %q, RamRoleSessionName: %q, RamRolePolicy: %q, RamRoleSessionExpiration: %d, RamRoleExternalId: %s)",
			config.RamRoleArn, config.RamRoleSessionName, config.RamRolePolicy, config.RamRoleSessionExpiration, config.RamRoleExternalId)
	}

	if v, ok := d.GetOk("assume_role_with_oidc"); ok && len(v.([]interface{})) == 1 {
		config.AssumeRoleWithOidc, err = getAssumeRoleWithOIDCConfig(v.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] assume_role_with_oidc configuration set: (RoleArn: %q, SessionName: %q, SessionExpiration: %d, OIDCProviderArn: %s)",
			config.AssumeRoleWithOidc.RoleARN, config.AssumeRoleWithOidc.RoleSessionName, config.AssumeRoleWithOidc.DurationSeconds, config.AssumeRoleWithOidc.OIDCProviderArn)
	}

	endpointsSet := d.Get("endpoints").(*schema.Set)
	var endpointInit sync.Map
	config.Endpoints = &endpointInit

	for _, endpointsSetI := range endpointsSet.List() {
		endpoints := endpointsSetI.(map[string]interface{})
		for key, val := range endpoints {
			// Compatible with the deprecated endpoint setting
			if val == nil || val.(string) == "" {
				if v, ok := deprecatedEndpointMap[key]; ok {
					val = endpoints[v]
				}
			}
			endpointInit.Store(key, val.(string))
		}
		config.EcsEndpoint = strings.TrimSpace(endpoints["ecs"].(string))
		config.RdsEndpoint = strings.TrimSpace(endpoints["rds"].(string))
		config.SlbEndpoint = strings.TrimSpace(endpoints["slb"].(string))
		config.VpcEndpoint = strings.TrimSpace(endpoints["vpc"].(string))
		config.EssEndpoint = strings.TrimSpace(endpoints["ess"].(string))
		config.OssEndpoint = strings.TrimSpace(endpoints["oss"].(string))
		config.OnsEndpoint = strings.TrimSpace(endpoints["ons"].(string))
		config.AlikafkaEndpoint = strings.TrimSpace(endpoints["alikafka"].(string))
		config.DnsEndpoint = strings.TrimSpace(endpoints["dns"].(string))
		config.RamEndpoint = strings.TrimSpace(endpoints["ram"].(string))
		config.CsEndpoint = strings.TrimSpace(endpoints["cs"].(string))
		config.CrEndpoint = strings.TrimSpace(endpoints["cr"].(string))
		config.CdnEndpoint = strings.TrimSpace(endpoints["cdn"].(string))
		config.KmsEndpoint = strings.TrimSpace(endpoints["kms"].(string))
		config.OtsEndpoint = strings.TrimSpace(endpoints["ots"].(string))
		config.CmsEndpoint = strings.TrimSpace(endpoints["cms"].(string))
		config.PvtzEndpoint = strings.TrimSpace(endpoints["pvtz"].(string))
		config.StsEndpoint = strings.TrimSpace(endpoints["sts"].(string))
		config.LogEndpoint = strings.TrimSpace(endpoints["log"].(string))
		config.DrdsEndpoint = strings.TrimSpace(endpoints["drds"].(string))
		config.DdsEndpoint = strings.TrimSpace(endpoints["dds"].(string))
		config.GpdbEnpoint = strings.TrimSpace(endpoints["gpdb"].(string))
		config.KVStoreEndpoint = strings.TrimSpace(endpoints["kvstore"].(string))
		config.PolarDBEndpoint = strings.TrimSpace(endpoints["polardb"].(string))
		config.FcEndpoint = strings.TrimSpace(endpoints["fc"].(string))
		config.ApigatewayEndpoint = strings.TrimSpace(endpoints["apigateway"].(string))
		config.DatahubEndpoint = strings.TrimSpace(endpoints["datahub"].(string))
		config.MnsEndpoint = strings.TrimSpace(endpoints["mns"].(string))
		config.LocationEndpoint = strings.TrimSpace(endpoints["location"].(string))
		config.ElasticsearchEndpoint = strings.TrimSpace(endpoints["elasticsearch"].(string))
		config.NasEndpoint = strings.TrimSpace(endpoints["nas"].(string))
		config.ActiontrailEndpoint = strings.TrimSpace(endpoints["actiontrail"].(string))
		config.BssOpenApiEndpoint = strings.TrimSpace(endpoints["bssopenapi"].(string))
		config.DdoscooEndpoint = strings.TrimSpace(endpoints["ddoscoo"].(string))
		config.DdosbgpEndpoint = strings.TrimSpace(endpoints["ddosbgp"].(string))
		config.EmrEndpoint = strings.TrimSpace(endpoints["emr"].(string))
		config.CasEndpoint = strings.TrimSpace(endpoints["cas"].(string))
		config.MarketEndpoint = strings.TrimSpace(endpoints["market"].(string))
		config.AdbEndpoint = strings.TrimSpace(endpoints["adb"].(string))
		config.CbnEndpoint = strings.TrimSpace(endpoints["cbn"].(string))
		config.MaxComputeEndpoint = strings.TrimSpace(endpoints["maxcompute"].(string))
		config.DmsEnterpriseEndpoint = strings.TrimSpace(endpoints["dms_enterprise"].(string))
		config.WafOpenapiEndpoint = strings.TrimSpace(endpoints["waf_openapi"].(string))
		config.ResourcemanagerEndpoint = strings.TrimSpace(endpoints["resourcemanager"].(string))
		config.EciEndpoint = strings.TrimSpace(endpoints["eci"].(string))
		config.OosEndpoint = strings.TrimSpace(endpoints["oos"].(string))
		config.DcdnEndpoint = strings.TrimSpace(endpoints["dcdn"].(string))
		config.Endpoints.Store("mse", strings.TrimSpace(endpoints["mse"].(string)))
		config.ConfigEndpoint = strings.TrimSpace(endpoints["config"].(string))
		config.RKvstoreEndpoint = strings.TrimSpace(endpoints["r_kvstore"].(string))
		config.FnfEndpoint = strings.TrimSpace(endpoints["fnf"].(string))
		config.RosEndpoint = strings.TrimSpace(endpoints["ros"].(string))
		config.PrivatelinkEndpoint = strings.TrimSpace(endpoints["privatelink"].(string))
		config.ResourcesharingEndpoint = strings.TrimSpace(endpoints["ressharing"].(string))
		config.GaEndpoint = strings.TrimSpace(endpoints["ga"].(string))
		config.HitsdbEndpoint = strings.TrimSpace(endpoints["hitsdb"].(string))
		config.BrainIndustrialEndpoint = strings.TrimSpace(endpoints["brain_industrial"].(string))
		config.EipanycastEndpoint = strings.TrimSpace(endpoints["eipanycast"].(string))
		config.ImsEndpoint = strings.TrimSpace(endpoints["ims"].(string))
		config.QuotasEndpoint = strings.TrimSpace(endpoints["quotas"].(string))
		config.SgwEndpoint = strings.TrimSpace(endpoints["sgw"].(string))
		config.DmEndpoint = strings.TrimSpace(endpoints["dm"].(string))
		config.EventbridgeEndpoint = strings.TrimSpace(endpoints["eventbridge"].(string))
		config.OnsproxyEndpoint = strings.TrimSpace(endpoints["onsproxy"].(string))
		config.CdsEndpoint = strings.TrimSpace(endpoints["cds"].(string))
		config.HbrEndpoint = strings.TrimSpace(endpoints["hbr"].(string))
		config.ArmsEndpoint = strings.TrimSpace(endpoints["arms"].(string))
		config.ServerlessEndpoint = strings.TrimSpace(endpoints["serverless"].(string))
		config.AlbEndpoint = strings.TrimSpace(endpoints["alb"].(string))
		config.RedisaEndpoint = strings.TrimSpace(endpoints["redisa"].(string))
		config.GwsecdEndpoint = strings.TrimSpace(endpoints["gwsecd"].(string))
		config.CloudphoneEndpoint = strings.TrimSpace(endpoints["cloudphone"].(string))
		config.ScdnEndpoint = strings.TrimSpace(endpoints["scdn"].(string))
		config.DataworkspublicEndpoint = strings.TrimSpace(endpoints["dataworkspublic"].(string))
		config.HcsSgwEndpoint = strings.TrimSpace(endpoints["hcs_sgw"].(string))
		config.CddcEndpoint = strings.TrimSpace(endpoints["cddc"].(string))
		config.MscopensubscriptionEndpoint = strings.TrimSpace(endpoints["mscopensubscription"].(string))
		config.SddpEndpoint = strings.TrimSpace(endpoints["sddp"].(string))
		config.BastionhostEndpoint = strings.TrimSpace(endpoints["bastionhost"].(string))
		config.SasEndpoint = strings.TrimSpace(endpoints["sas"].(string))
		config.AlidfsEndpoint = strings.TrimSpace(endpoints["alidfs"].(string))
		config.EhpcEndpoint = strings.TrimSpace(endpoints["ehpc"].(string))
		config.EnsEndpoint = strings.TrimSpace(endpoints["ens"].(string))
		config.IotEndpoint = strings.TrimSpace(endpoints["iot"].(string))
		config.ImmEndpoint = strings.TrimSpace(endpoints["imm"].(string))
		config.ClickhouseEndpoint = strings.TrimSpace(endpoints["clickhouse"].(string))
		config.SelectDBEndpoint = strings.TrimSpace(endpoints["selectdb"].(string))
		config.DtsEndpoint = strings.TrimSpace(endpoints["dts"].(string))
		config.DgEndpoint = strings.TrimSpace(endpoints["dg"].(string))
		config.CloudssoEndpoint = strings.TrimSpace(endpoints["cloudsso"].(string))
		config.WafEndpoint = strings.TrimSpace(endpoints["waf"].(string))
		config.SwasEndpoint = strings.TrimSpace(endpoints["swas"].(string))
		config.VsEndpoint = strings.TrimSpace(endpoints["vs"].(string))
		config.QuickbiEndpoint = strings.TrimSpace(endpoints["quickbi"].(string))
		config.VodEndpoint = strings.TrimSpace(endpoints["vod"].(string))
		config.OpensearchEndpoint = strings.TrimSpace(endpoints["opensearch"].(string))
		config.GdsEndpoint = strings.TrimSpace(endpoints["gds"].(string))
		config.DbfsEndpoint = strings.TrimSpace(endpoints["dbfs"].(string))
		config.DevopsrdcEndpoint = strings.TrimSpace(endpoints["devopsrdc"].(string))
		config.EaisEndpoint = strings.TrimSpace(endpoints["eais"].(string))
		config.CloudauthEndpoint = strings.TrimSpace(endpoints["cloudauth"].(string))
		config.ImpEndpoint = strings.TrimSpace(endpoints["imp"].(string))
		config.MhubEndpoint = strings.TrimSpace(endpoints["mhub"].(string))
		config.ServicemeshEndpoint = strings.TrimSpace(endpoints["servicemesh"].(string))
		config.AcrEndpoint = strings.TrimSpace(endpoints["acr"].(string))
		config.EdsuserEndpoint = strings.TrimSpace(endpoints["edsuser"].(string))
		config.GaplusEndpoint = strings.TrimSpace(endpoints["gaplus"].(string))
		config.DdosbasicEndpoint = strings.TrimSpace(endpoints["ddosbasic"].(string))
		config.SmartagEndpoint = strings.TrimSpace(endpoints["smartag"].(string))
		config.TagEndpoint = strings.TrimSpace(endpoints["tag"].(string))
		config.EdasEndpoint = strings.TrimSpace(endpoints["edas"].(string))
		config.EdasschedulerxEndpoint = strings.TrimSpace(endpoints["edasschedulerx"].(string))
		config.EhsEndpoint = strings.TrimSpace(endpoints["ehs"].(string))
		config.CloudfwEndpoint = strings.TrimSpace(endpoints["cloudfw"].(string))
		config.DysmsEndpoint = strings.TrimSpace(endpoints["dysms"].(string))
		config.CbsEndpoint = strings.TrimSpace(endpoints["cbs"].(string))
		config.NlbEndpoint = strings.TrimSpace(endpoints["nlb"].(string))
		config.VpcpeerEndpoint = strings.TrimSpace(endpoints["vpcpeer"].(string))
		config.EbsEndpoint = strings.TrimSpace(endpoints["ebs"].(string))
		config.DmsenterpriseEndpoint = strings.TrimSpace(endpoints["dmsenterprise"].(string))
		config.BpStudioEndpoint = strings.TrimSpace(endpoints["bpstudio"].(string))
		config.DasEndpoint = strings.TrimSpace(endpoints["das"].(string))
		config.CloudfirewallEndpoint = strings.TrimSpace(endpoints["cloudfirewall"].(string))
		config.SrvcatalogEndpoint = strings.TrimSpace(endpoints["srvcatalog"].(string))
		config.VpcPeerEndpoint = strings.TrimSpace(endpoints["vpcpeer"].(string))
		config.EfloEndpoint = strings.TrimSpace(endpoints["eflo"].(string))
		config.OceanbaseEndpoint = strings.TrimSpace(endpoints["oceanbase"].(string))
		config.BeebotEndpoint = strings.TrimSpace(endpoints["beebot"].(string))
		config.ComputeNestEndpoint = strings.TrimSpace(endpoints["computenest"].(string))
		if endpoint, ok := endpoints["alidns"]; ok {
			config.AlidnsEndpoint = strings.TrimSpace(endpoint.(string))
		} else {
			config.AlidnsEndpoint = strings.TrimSpace(endpoints["dns"].(string))
		}
		config.CassandraEndpoint = strings.TrimSpace(endpoints["cassandra"].(string))
	}

	if otsInstanceName, ok := d.GetOk("ots_instance_name"); ok && otsInstanceName.(string) != "" {
		config.OtsInstanceName = strings.TrimSpace(otsInstanceName.(string))
	}

	if logEndpoint, ok := d.GetOk("log_endpoint"); ok && logEndpoint.(string) != "" {
		config.LogEndpoint = strings.TrimSpace(logEndpoint.(string))
	}
	if mnsEndpoint, ok := d.GetOk("mns_endpoint"); ok && mnsEndpoint.(string) != "" {
		config.MnsEndpoint = strings.TrimSpace(mnsEndpoint.(string))
	}

	if fcEndpoint, ok := d.GetOk("fc"); ok && fcEndpoint.(string) != "" {
		config.FcEndpoint = strings.TrimSpace(fcEndpoint.(string))
	}
	if config.StsEndpoint == "" {
		config.StsEndpoint = connectivity.LoadRegionalEndpoint(config.RegionId, "sts")
	}

	configurationSources := []string{
		fmt.Sprintf("Default/%s", config.TerraformTraceId),
	}

	// configuration source final value should also contain TF_APPEND_USER_AGENT value
	// there is need to deduplication
	config.ConfigurationSource += " " + strings.TrimSpace(os.Getenv("TF_APPEND_USER_AGENT"))
	if config.ConfigurationSource != "" {
		for _, s := range strings.Split(config.ConfigurationSource, " ") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			exist := false
			for _, con := range configurationSources {
				if s == con {
					exist = true
					break
				}
			}
			if !exist {
				configurationSources = append(configurationSources, s)
			}
		}
	}
	config.ConfigurationSource = strings.Join(configurationSources, " ") + getModuleAddr()

	var signVersion sync.Map
	config.SignVersion = &signVersion
	for _, version := range d.Get("sign_version").(*schema.Set).List() {
		for key, val := range version.(map[string]interface{}) {
			signVersion.Store(key, val)
		}
	}

	if err := config.RefreshAuthCredential(); err != nil {
		return nil, err
	}

	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("configuring Terraform Alibaba Cloud Provider: no valid credential sources for Terraform Alibaba Cloud Provider found.\n\n%s",
			"Please see https://registry.terraform.io/providers/aliyun/alicloud/latest/docs#authentication\n"+
				"for more information about providing credentials.")
	}

	client, err := config.Client()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// This is a global MutexKV for use within this plugin.
var alicloudMutexKV = mutexkv.NewMutexKV()

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"access_key": "The access key for API operations. You can retrieve this from the 'Security Management' section of the Alibaba Cloud console.",

		"secret_key": "The secret key for API operations. You can retrieve this from the 'Security Management' section of the Alibaba Cloud console.",

		"ecs_role_name": "The RAM Role Name attached on a ECS instance for API operations. You can retrieve this from the 'Access Control' section of the Alibaba Cloud console.",

		"region": "The region where Alibaba Cloud operations will take place. Examples are cn-beijing, cn-hangzhou, eu-central-1, etc.",

		"security_token": "security token. A security token is only required if you are using Security Token Service.",

		"account_id": "The account ID for some service API operations. You can retrieve this from the 'Security Settings' section of the Alibaba Cloud console.",

		"profile": "The profile for API operations. If not set, the default profile created with `aliyun configure` will be used.",

		"shared_credentials_file": "The path to the shared credentials file. If not set this defaults to ~/.aliyun/config.json",

		"assume_role_role_arn": "The ARN of a RAM role to assume prior to making API calls.",

		"assume_role_session_name": "The session name to use when assuming the role. If omitted, `terraform` is passed to the AssumeRole call as session name.",

		"assume_role_policy": "The permissions applied when assuming a role. You cannot use, this policy to grant further permissions that are in excess to those of the, role that is being assumed.",

		"assume_role_session_expiration": "The time after which the established session for assuming role expires. Valid value range: [900-3600] seconds. Default to 0 (in this case Alicloud use own default value).",

		"skip_region_validation": "Skip static validation of region ID. Used by users of alternative AlibabaCloud-like APIs or users w/ access to regions that are not public (yet).",

		"configuration_source": "Use this to mark a terraform configuration file source.",

		"client_read_timeout":    "The maximum timeout of the client read request.",
		"client_connect_timeout": "The maximum timeout of the client connection server.",
		"source_ip":              "The source ip for the assume role invoking.",
		"secure_transport":       "The security transport for the assume role invoking.",
		"credentials_uri":        "The URI of sidecar credentials service.",
		"max_retry_timeout":      "The maximum retry timeout of the request.",

		"ecs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ECS endpoints.",

		"rds_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom RDS endpoints.",

		"slb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom SLB endpoints.",

		"vpc_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom VPC and VPN endpoints.",

		"ess_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Autoscaling endpoints.",

		"oss_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom OSS endpoints.",

		"ons_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ONS endpoints.",

		"alikafka_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ALIKAFKA endpoints.",

		"dns_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom DNS endpoints.",

		"ram_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom RAM endpoints.",

		"cs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Container Service endpoints.",

		"cr_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Container Registry endpoints.",

		"cdn_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom CDN endpoints.",

		"kms_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom KMS endpoints.",

		"ots_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Table Store endpoints.",

		"cms_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Cloud Monitor endpoints.",

		"pvtz_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Private Zone endpoints.",

		"sts_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom STS endpoints.",

		"log_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Log Service endpoints.",

		"drds_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom DRDS endpoints.",

		"dds_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom MongoDB endpoints.",

		"polardb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom PolarDB endpoints.",

		"gpdb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom GPDB endpoints.",

		"kvstore_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom R-KVStore endpoints.",

		"fc_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Function Computing endpoints.",

		"apigateway_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Api Gateway endpoints.",

		"datahub_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Datahub endpoints.",

		"mns_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom MNS endpoints.",

		"location_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Location Service endpoints.",

		"elasticsearch_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Elasticsearch endpoints.",

		"nas_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom NAS endpoints.",

		"actiontrail_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Actiontrail endpoints.",

		"cas_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom CAS endpoints.",

		"bssopenapi_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom BSSOPENAPI endpoints.",

		"ddoscoo_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom DDOSCOO endpoints.",

		"ddosbgp_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom DDOSBGP endpoints.",

		"emr_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom EMR endpoints.",

		"market_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom Market Place endpoints.",

		"hbase_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom HBase endpoints.",

		"adb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom AnalyticDB endpoints.",

		"cbn_endpoint":        "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cbn endpoints.",
		"maxcompute_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom MaxCompute endpoints.",

		"dms_enterprise_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dms_enterprise endpoints.",

		"waf_openapi_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom waf_openapi endpoints.",

		"resourcemanager_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom resourcemanager endpoints.",

		"alidns_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom alidns endpoints.",

		"cassandra_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cassandra endpoints.",

		"eci_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom eci endpoints.",

		"oos_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom oos endpoints.",

		"dcdn_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dcdn endpoints.",

		"mse_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom mse endpoints.",

		"config_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom config endpoints.",

		"r_kvstore_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom r_kvstore endpoints.",

		"fnf_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom fnf endpoints.",

		"ros_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ros endpoints.",

		"privatelink_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom privatelink endpoints.",

		"resourcesharing_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom resourcesharing endpoints.",

		"ga_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ga endpoints.",

		"hitsdb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom hitsdb endpoints.",

		"brain_industrial_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom brain_industrial endpoints.",

		"eipanycast_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom eipanycast endpoints.",

		"ims_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ims endpoints.",

		"quotas_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom quotas endpoints.",

		"sgw_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom sgw endpoints.",

		"dm_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dm endpoints.",

		"eventbridge_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom eventbridge_share endpoints.",

		"onsproxy_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom onsproxy endpoints.",

		"cds_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cds endpoints.",

		"hbr_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom hbr endpoints.",

		"arms_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom arms endpoints.",

		"serverless_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom serverless endpoints.",

		"alb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom alb endpoints.",

		"redisa_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom redisa endpoints.",

		"gwsecd_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom gwsecd endpoints.",

		"cloudphone_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cloudphone endpoints.",

		"scdn_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom scdn endpoints.",

		"dataworkspublic_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dataworkspublic endpoints.",

		"hcs_sgw_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom hcs_sgw endpoints.",

		"cddc_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cddc endpoints.",

		"mscopensubscription_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom mscopensubscription endpoints.",

		"sddp_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom sddp endpoints.",

		"bastionhost_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom bastionhost endpoints.",

		"sas_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom sas endpoints.",

		"alidfs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom alidfs endpoints.",

		"ehpc_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ehpc endpoints.",

		"ens_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ens endpoints.",

		"iot_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom iot endpoints.",

		"imm_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom imm endpoints.",

		"clickhouse_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom clickhouse endpoints.",

		"selectdb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom selectdb endpoints.",

		"dts_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dts endpoints.",

		"dg_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dg endpoints.",

		"cloudsso_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cloudsso endpoints.",

		"waf_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom waf endpoints.",

		"swas_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom swas endpoints.",

		"vs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom vs endpoints.",

		"quickbi_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom quickbi endpoints.",

		"vod_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom vod endpoints.",

		"opensearch_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom opensearch endpoints.",

		"gds_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom gds endpoints.",

		"dbfs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dbfs endpoints.",

		"devopsrdc_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom devopsrdc endpoints.",

		"eais_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom eais endpoints.",

		"cloudauth_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cloudauth endpoints.",

		"imp_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom imp endpoints.",

		"mhub_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom mhub endpoints.",

		"servicemesh_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom servicemesh endpoints.",

		"acr_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom acr endpoints.",

		"edsuser_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom edsuser endpoints.",

		"gaplus_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom gaplus endpoints.",

		"ddosbasic_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ddosbasic endpoints.",

		"smartag_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom smartag endpoints.",

		"tag_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom tag endpoints.",

		"edas_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom edas endpoints.",

		"edasschedulerx_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom edasschedulerx endpoints.",

		"ehs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ehs endpoints.",

		"cloudfw_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cloudfw endpoints.",

		"dysmsapi_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dysmsapi endpoints.",

		"cbs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cbs endpoints.",

		"nlb_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom nlb endpoints.",

		"vpcpeer_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom vpcpeer endpoints.",

		"ebs_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom ebs endpoints.",

		"dmsenterprise_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom dmsenterprise endpoints.",

		"bpstudio_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom bpstudio endpoints.",

		"das_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom das endpoints.",

		"cloudfirewall_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom cloudfirewall endpoints.",

		"srvcatalog_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom srvcatalog endpoints.",

		"eflo_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom eflo endpoints.",

		"eflo_controller_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom efloctrl endpoints.",

		"eflo_cnp": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom eflocnp endpoints.",

		"oceanbase_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom oceanbase endpoints.",

		"beebot_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom beebot endpoints.",

		"computenest_endpoint": "Use this to override the default endpoint URL constructed from the `region`. It's typically used to connect to custom computenest endpoints.",
	}
}

func assumeRoleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"role_arn": {
					Type:        schema.TypeString,
					Required:    true,
					Description: descriptions["assume_role_role_arn"],
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_ASSUME_ROLE_ARN", "ALIBABA_CLOUD_ROLE_ARN"}, nil),
				},
				"session_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: descriptions["assume_role_session_name"],
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ALICLOUD_ASSUME_ROLE_SESSION_NAME", "ALIBABA_CLOUD_ROLE_SESSION_NAME"}, nil),
				},
				"policy": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: descriptions["assume_role_policy"],
				},
				"session_expiration": {
					Type:         schema.TypeInt,
					Optional:     true,
					Description:  descriptions["assume_role_session_expiration"],
					ValidateFunc: IntBetween(900, 43200),
				},
				"external_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: descriptions["external_id"],
				},
			},
		},
	}
}

func assumeRoleWithOidcSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"oidc_provider_arn": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "ARN of the OIDC IdP.",
					DefaultFunc: schema.EnvDefaultFunc("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", ""),
				},
				"oidc_token_file": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The file path of OIDC token that is issued by the external IdP.",
					DefaultFunc: schema.EnvDefaultFunc("ALIBABA_CLOUD_OIDC_TOKEN_FILE", ""),
					//ExactlyOneOf: []string{"assume_role_with_oidc.0.oidc_token", "assume_role_with_oidc.0.oidc_token_file"},
				},
				"oidc_token": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(4, 20000),
					DefaultFunc:  schema.EnvDefaultFunc("ALIBABA_CLOUD_OIDC_TOKEN", nil),
					//ExactlyOneOf: []string{"assume_role_with_oidc.0.oidc_token", "assume_role_with_oidc.0.oidc_token_file"},
				},
				"role_arn": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "ARN of a RAM role to assume prior to making API calls.",
					DefaultFunc: schema.EnvDefaultFunc("ALIBABA_CLOUD_ROLE_ARN", ""),
				},
				"role_session_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The custom name of the role session. Set this parameter based on your business requirements. In most cases, this parameter is set to the identity of the user who calls the operation, for example, the username.",
					DefaultFunc: schema.EnvDefaultFunc("ALIBABA_CLOUD_ROLE_SESSION_NAME", ""),
				},
				"policy": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The policy that specifies the permissions of the returned STS token. You can use this parameter to grant the STS token fewer permissions than the permissions granted to the RAM role.",
				},
				"session_expiration": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The validity period of the STS token. Unit: seconds. Default value: 3600. Minimum value: 900. Maximum value: the value of the MaxSessionDuration parameter when creating a ram role.",
				},
			},
		},
	}
}

func signVersionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"oss": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"sls": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func endpointsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"computenest": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["computenest_endpoint"],
				},

				"beebot": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["beebot_endpoint"],
				},
				"chatbot": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["beebot_endpoint"],
				},

				"eflo": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eflo_endpoint"],
				},

				"eflo_controller": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eflo_controller_endpoint"],
				},

				"eflo_cnp": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eflo_cnp_endpoint"],
				},

				"srvcatalog": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["srvcatalog_endpoint"],
				},
				"servicecatalog": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["srvcatalog_endpoint"],
				},
				"cloudfirewall": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cloudfirewall_endpoint"],
				},

				"das": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["das_endpoint"],
				},

				"bpstudio": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["bpstudio_endpoint"],
				},

				"dmsenterprise": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dmsenterprise_endpoint"],
				},

				"ebs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ebs_endpoint"],
				},

				"nlb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["nlb_endpoint"],
				},

				"cbs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cbs_endpoint"],
				},
				"dbs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cbs_endpoint"],
				},

				"vpcpeer": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["vpcpeer_endpoint"],
				},

				"dysms": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dysms_endpoint"],
				},
				"dysmsapi": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dysmsapi_endpoint"],
				},

				"edas": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["edas_endpoint"],
				},

				"edasschedulerx": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["edasschedulerx_endpoint"],
				},
				"schedulerx2": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["edasschedulerx_endpoint"],
				},
				"ehs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ehs_endpoint"],
				},

				"tag": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["tag_endpoint"],
				},

				"ddosbasic": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ddosbasic_endpoint"],
				},

				"antiddos_public": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ddosbasic_endpoint"],
				},
				"smartag": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["smartag_endpoint"],
				},

				"oceanbase": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["oceanbase_endpoint"],
				},
				"oceanbasepro": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["oceanbase_endpoint"],
				},

				"gaplus": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["gaplus_endpoint"],
				},

				"cloudfw": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cloudfw_endpoint"],
				},

				"edsuser": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["edsuser_endpoint"],
				},
				"eds_user": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["edsuser_endpoint"],
				},

				"acr": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["acr_endpoint"],
				},

				"imp": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["imp_endpoint"],
				},
				"eais": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eais_endpoint"],
				},
				"cloudauth": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cloudauth_endpoint"],
				},

				"mhub": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["mhub_endpoint"],
				},
				"servicemesh": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["servicemesh_endpoint"],
				},
				"quickbi": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["quickbi_endpoint"],
				},
				"quickbi_public": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["quickbi_endpoint"],
				},
				"vod": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["vod_endpoint"],
				},
				"opensearch": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["opensearch_endpoint"],
				},
				"gds": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["gds_endpoint"],
				},
				"gdb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["gds_endpoint"],
				},
				"dbfs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dbfs_endpoint"],
				},
				"devopsrdc": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["devopsrdc_endpoint"],
				},
				"dg": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dg_endpoint"],
				},
				"waf": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["waf_endpoint"],
				},
				"vs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["vs_endpoint"],
				},
				"dts": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dts_endpoint"],
				},
				"cloudsso": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cloudsso_endpoint"],
				},

				"iot": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["iot_endpoint"],
				},
				"swas": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["swas_endpoint"],
				},
				"swas_open": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["swas_endpoint"],
				},

				"imm": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["imm_endpoint"],
				},
				"clickhouse": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["clickhouse_endpoint"],
				},
				"selectdb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["selectdb_endpoint"],
				},

				"alidfs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["alidfs_endpoint"],
				},
				"dfs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["alidfs_endpoint"],
				},

				"ens": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ens_endpoint"],
				},

				"bastionhost": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["bastionhost_endpoint"],
				},
				"cddc": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cddc_endpoint"],
				},
				"sddp": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["sddp_endpoint"],
				},

				"mscopensubscription": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["mscopensubscription_endpoint"],
				},

				"sas": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["sas_endpoint"],
				},

				"ehpc": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ehpc_endpoint"],
				},

				"dataworkspublic": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dataworkspublic_endpoint"],
				},
				"dataworks_public": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dataworkspublic_endpoint"],
				},
				"hcs_sgw": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["hcs_sgw_endpoint"],
				},

				"cloudphone": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cloudphone_endpoint"],
				},

				"alb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["alb_endpoint"],
				},
				"redisa": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["redisa_endpoint"],
				},
				"gwsecd": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["gwsecd_endpoint"],
				},
				"ecd": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["gwsecd_endpoint"],
				},
				"scdn": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["scdn_endpoint"],
				},

				"arms": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["arms_endpoint"],
				},
				"serverless": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["serverless_endpoint"],
				},
				"sae": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["serverless_endpoint"],
				},
				"hbr": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["hbr_endpoint"],
				},

				"amqp": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["onsproxy_endpoint"],
				},

				"onsproxy": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["onsproxy_endpoint"],
				},
				"cds": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cds_endpoint"],
				},

				"dm": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dm_endpoint"],
				},

				"eventbridge": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eventbridge_endpoint"],
				},

				"sgw": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["sgw_endpoint"],
				},

				"quotas": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["quotas_endpoint"],
				},

				"ims": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ims_endpoint"],
				},

				"brain_industrial": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["brain_industrial_endpoint"],
				},

				"ressharing": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["resourcesharing_endpoint"],
				},
				"resourcesharing": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"ga": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ga_endpoint"],
				},

				"hitsdb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["hitsdb_endpoint"],
				},

				"privatelink": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["privatelink_endpoint"],
				},

				"eipanycast": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eipanycast_endpoint"],
				},

				"fnf": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["fnf_endpoint"],
				},

				"ros": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ros_endpoint"],
				},

				"r_kvstore": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["r_kvstore_endpoint"],
				},

				"config": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["config_endpoint"],
				},

				"dcdn": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dcdn_endpoint"],
				},

				"mse": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["mse_endpoint"],
				},

				"oos": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["oos_endpoint"],
				},

				"eci": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["eci_endpoint"],
				},

				"alidns": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["alidns_endpoint"],
				},

				"resourcemanager": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["resourcemanager_endpoint"],
				},

				"waf_openapi": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["waf_openapi_endpoint"],
				},

				"dms_enterprise": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dms_enterprise_endpoint"],
				},

				"cassandra": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cassandra_endpoint"],
				},

				"cbn": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cbn_endpoint"],
				},

				"ecs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ecs_endpoint"],
				},
				"rds": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["rds_endpoint"],
				},
				"slb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["slb_endpoint"],
				},
				"vpc": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["vpc_endpoint"],
				},
				"ess": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ess_endpoint"],
				},
				"oss": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["oss_endpoint"],
				},
				"ons": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ons_endpoint"],
				},
				"alikafka": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["alikafka_endpoint"],
				},
				"dns": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dns_endpoint"],
				},
				"ram": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ram_endpoint"],
				},
				"cs": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cs_endpoint"],
				},
				"cr": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cr_endpoint"],
				},
				"cdn": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cdn_endpoint"],
				},

				"kms": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["kms_endpoint"],
				},

				"ots": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ots_endpoint"],
				},

				"cms": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cms_endpoint"],
				},

				"pvtz": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["pvtz_endpoint"],
				},

				"sts": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["sts_endpoint"],
				},
				// log service is sls service
				"log": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["log_endpoint"],
				},
				"drds": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["drds_endpoint"],
				},
				"polardbx": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["drds_endpoint"],
				},
				"dds": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["dds_endpoint"],
				},
				"polardb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["polardb_endpoint"],
				},
				"gpdb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["gpdb_endpoint"],
				},
				"kvstore": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["kvstore_endpoint"],
				},
				"fc": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["fc_endpoint"],
				},
				"fc_open": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["fc_endpoint"],
				},
				"apigateway": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["apigateway_endpoint"],
				},
				"cloudapi": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["apigateway_endpoint"],
				},
				"apig": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"datahub": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["datahub_endpoint"],
				},
				"devops_rdc": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"mns": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["mns_endpoint"],
				},
				"mns_open": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["mns_endpoint"],
				},
				"rocketmq": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"location": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["location_endpoint"],
				},
				"elasticsearch": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["elasticsearch_endpoint"],
				},
				"nas": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["nas_endpoint"],
				},
				"actiontrail": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["actiontrail_endpoint"],
				},
				"cas": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["cas_endpoint"],
				},
				"bssopenapi": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["bssopenapi_endpoint"],
				},
				"ddoscoo": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ddoscoo_endpoint"],
				},
				"ddosbgp": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["ddosbgp_endpoint"],
				},
				"emr": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["emr_endpoint"],
				},
				"market": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["market_endpoint"],
				},
				"adb": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["adb_endpoint"],
				},
				"maxcompute": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: descriptions["maxcompute_endpoint"],
				},
				"aiworkspace": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"vpcipam": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"gwlb": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"esa": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
			},
		},
		Set: endpointsToHash,
	}
}

func endpointsToHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["ecs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["rds"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["slb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["vpc"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ess"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["oss"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ons"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["alikafka"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dns"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ram"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cdn"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["kms"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ots"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cms"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["pvtz"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["sts"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["log"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["drds"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dds"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["gpdb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["kvstore"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["polardb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["fc"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["apigateway"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["datahub"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["mns"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["location"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["elasticsearch"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["nas"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["actiontrail"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cas"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["bssopenapi"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ddoscoo"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ddosbgp"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["emr"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["market"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["adb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cbn"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["maxcompute"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dms_enterprise"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["waf_openapi"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["resourcemanager"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["alidns"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cassandra"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["eci"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["oos"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dcdn"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["mse"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["config"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["r_kvstore"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["fnf"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ros"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["privatelink"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ressharing"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ga"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["hitsdb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["brain_industrial"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["eipanycast"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ims"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["quotas"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["sgw"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dm"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["eventbridge"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["onsproxy"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cds"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["hbr"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["arms"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["serverless"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["alb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["redisa"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["gwsecd"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cloudphone"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["scdn"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dataworkspublic"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["hcs_sgw"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cddc"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["mscopensubscription"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["sddp"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["bastionhost"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["sas"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["alidfs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ehpc"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ens"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["iot"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["imm"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["clickhouse"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["selectdb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dts"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dg"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cloudsso"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["waf"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["swas"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["vs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["quickbi"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["vod"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["opensearch"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["gds"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dbfs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["devopsrdc"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["eais"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cloudauth"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["imp"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["mhub"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["servicemesh"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["acr"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["edsuser"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["gaplus"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ddosbasic"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["smartag"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["tag"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["edas"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["edasschedulerx"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ehs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cloudfw"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dysms"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cbs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["nlb"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["vpcpeer"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ebs"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dmsenterprise"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["bpstudio"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["das"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["cloudfirewall"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["srvcatalog"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["vpcpeer"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["eflo"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["oceanbase"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["beebot"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["computenest"].(string)))
	return hashcode.String(buf.String())
}

// deprecatedEndpointMap is used to map old service name to new service name,
// and its value equals to the gateway code of the API after converting it to lowercase and using underscores
// key: new endpoint key
// value: deprecated endpoint key
var deprecatedEndpointMap = map[string]string{
	"resourcesharing":  "ressharing",
	"ga":               "gaplus",
	"dms_enterprise":   "dmsenterprise",
	"sgw":              "hcs_sgw",
	"amqp":             "onsproxy",
	"cassandra":        "cds",
	"cloudfw":          "cloudfirewall",
	"sae":              "serverless",
	"r_kvstore":        "redisa",
	"ecd":              "gwsecd",
	"dataworks_public": "dataworkspublic",
	"dfs":              "alidfs",
	"swas_open":        "swas",
	"quickbi_public":   "quickbi",
	"gdb":              "gds",
	"cr":               "acr",
	"eds_user":         "edsuser",
	"antiddos_public":  "ddosbasic",
	"schedulerx2":      "edasschedulerx",
	"ehpc":             "ehs",
	"dysmsapi":         "dysms",
	"dbs":              "cbs",
	"mns_open":         "mns",
	"servicecatalog":   "srvcatalog",
	"oceanbasepro":     "oceanbase",
	"chatbot":          "beebot",
	"cloudapi":         "apigateway",
}

func getConfigFromProfile(d *schema.ResourceData, ProfileKey string) (interface{}, error) {

	if providerConfig == nil {
		if v, ok := d.GetOk("profile"); !ok && v.(string) == "" {
			return nil, nil
		}
		current := d.Get("profile").(string)
		// Set Credentials filename, expanding home directory
		profilePath, err := homedir.Expand(d.Get("shared_credentials_file").(string))
		if err != nil {
			return nil, WrapError(err)
		}
		if profilePath == "" {
			profilePath = fmt.Sprintf("%s/.aliyun/config.json", os.Getenv("HOME"))
			if runtime.GOOS == "windows" {
				profilePath = fmt.Sprintf("%s/.aliyun/config.json", os.Getenv("USERPROFILE"))
			}
		}
		providerConfig = make(map[string]interface{})
		_, err = os.Stat(profilePath)
		if !os.IsNotExist(err) {
			data, err := ioutil.ReadFile(profilePath)
			if err != nil {
				return nil, WrapError(err)
			}
			config := map[string]interface{}{}
			err = json.Unmarshal(data, &config)
			if err != nil {
				return nil, WrapError(err)
			}
			for _, v := range config["profiles"].([]interface{}) {
				if current == v.(map[string]interface{})["name"] {
					providerConfig = v.(map[string]interface{})
				}
			}
		}
	}

	mode := ""
	if v, ok := providerConfig["mode"]; ok {
		mode = v.(string)
	} else {
		return v, nil
	}
	if ProfileKey == "region_id" {
		return providerConfig["region_id"], nil
	}
	if mode == "ChainableRamRoleArn" {
		return nil, nil
	}
	switch ProfileKey {
	case "access_key_id", "access_key_secret":
		if mode == "EcsRamRole" {
			return "", nil
		}
	case "ram_role_name":
		if mode != "EcsRamRole" {
			return "", nil
		}
	case "sts_token":
		if mode != "StsToken" {
			return "", nil
		}
	case "ram_role_arn", "ram_session_name":
		if mode != "RamRoleArn" {
			return "", nil
		}
	case "expired_seconds":
		if mode != "RamRoleArn" {
			return float64(0), nil
		}
	}

	return providerConfig[ProfileKey], nil
}

func getAssumeRoleWithOIDCConfig(tfMap map[string]interface{}) (*connectivity.AssumeRoleWithOidc, error) {
	if tfMap == nil {
		return nil, nil
	}

	assumeRole := connectivity.AssumeRoleWithOidc{}

	if v, ok := tfMap["session_expiration"].(int); ok && v != 0 {
		assumeRole.DurationSeconds = v
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
		assumeRole.Policy = v
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		assumeRole.RoleARN = v
	}

	if v, ok := tfMap["role_session_name"].(string); ok && v != "" {
		assumeRole.RoleSessionName = v
	}
	if assumeRole.RoleSessionName == "" {
		assumeRole.RoleSessionName = "terraform"
	}

	if v, ok := tfMap["oidc_provider_arn"].(string); ok && v != "" {
		assumeRole.OIDCProviderArn = v
	}

	missingOidcToken := true
	if v, ok := tfMap["oidc_token"]; ok && v.(string) != "" {
		assumeRole.OIDCToken = v.(string)
		missingOidcToken = false
	}

	if v, ok := tfMap["oidc_token_file"].(string); ok && v != "" {
		assumeRole.OIDCTokenFile = v
		if assumeRole.OIDCToken == "" {
			token, err := os.ReadFile(v)
			if err != nil {
				return nil, fmt.Errorf("reading oidc_token_file failed. Error: %s", err)
			}
			assumeRole.OIDCToken = string(token)
		}
		missingOidcToken = false
	}
	if missingOidcToken {
		return nil, fmt.Errorf("\"assume_role_with_oidc.0.oidc_token\": one of `assume_role_with_oidc.0.oidc_token,assume_role_with_oidc.0.oidc_token_file` must be specified")
	}

	if assumeRole.OIDCToken == "" {
		return nil, fmt.Errorf("\"assume_role_with_oidc.0.oidc_token\" or \"assume_role_with_oidc.0.oidc_token_file\" content can not be empty")
	}

	return &assumeRole, nil
}

type CredentialsURIResponse struct {
	Code            string
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	Expiration      string
}

func getClientByCredentialsURI(credentialsURI string) (*CredentialsURIResponse, error) {
	res, err := http.Get(credentialsURI)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("get Credentials from %s failed, status code %d", credentialsURI, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	var response CredentialsURIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal credentials failed, the body %s", string(body))
	}

	if response.Code != "Success" {
		return nil, fmt.Errorf("fetching sts token from %s got an error and its Code is not Success", credentialsURI)
	}

	return &response, nil
}

func getModuleAddr() string {
	moduleMeta := make(map[string]interface{})
	str, err := os.ReadFile(".terraform/modules/modules.json")
	if err != nil {
		return ""
	}
	err = json.Unmarshal(str, &moduleMeta)
	if err != nil || len(moduleMeta) < 1 || moduleMeta["Modules"] == nil {
		return ""
	}
	var result string
	for _, m := range moduleMeta["Modules"].([]interface{}) {
		module := m.(map[string]interface{})
		moduleSource := fmt.Sprint(module["Source"])
		moduleVersion := fmt.Sprint(module["Version"])
		if strings.HasPrefix(moduleSource, "registry.terraform.io/") {
			parts := strings.Split(moduleSource, "/")
			if len(parts) == 4 {
				result += " " + "terraform-" + parts[3] + "-" + parts[2] + "/" + moduleVersion
			}
		}
	}
	return result
}

func absPath(filePath string) string {
	if v, err := homedir.Expand(filePath); err != nil {
		log.Printf("[WARN] failed to expand profile file path: %v", err)
	} else {
		filePath = v
	}

	if v, err := filepath.Abs(filePath); err != nil {
		log.Printf("[WARN] failed to get absolute path of profile file: %v", err)
	} else {
		filePath = v
	}

	return filePath
}
