package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

type AdbpgService struct {
	client   *connectivity.AliyunClient
	adbpgAPI *adbpg.AdbpgAPI
}

func NewAdbpgService(client *connectivity.AliyunClient) (*AdbpgService, error) {
	creds := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}
	api, err := adbpg.NewAdbpgAPI(creds)
	if err != nil {
		return nil, WrapError(err)
	}
	return &AdbpgService{client: client, adbpgAPI: api}, nil
}

// Instance methods (C-2 through C-6)

func (s *AdbpgService) DescribeAdbpgInstance(instanceId string) (*adbpg.AdbpgInstanceDetail, error) {
	detail, err := s.adbpgAPI.DescribeInstance(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return detail, nil
}

func (s *AdbpgService) CreateAdbpgInstance(input *adbpg.AdbpgInstanceCreate) (*adbpg.AdbpgInstanceDetail, error) {
	detail, err := s.adbpgAPI.CreateInstance(input)
	if err != nil {
		return nil, WrapError(err)
	}
	return detail, nil
}

func (s *AdbpgService) DeleteAdbpgInstance(instanceId string) error {
	_, err := s.adbpgAPI.DeleteInstance(instanceId)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) ModifyAdbpgInstanceDescription(instanceId, description string) error {
	_, err := s.adbpgAPI.ModifyInstanceDescription(instanceId, description)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) ModifyAdbpgInstanceMaintainTime(instanceId, startTime, endTime string) error {
	_, err := s.adbpgAPI.ModifyInstanceMaintainTime(instanceId, startTime, endTime)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// Instance state methods (C-7 through C-9)

func (s *AdbpgService) AdbpgInstanceStateRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAdbpgInstance(instanceId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if object.DBInstanceStatus == failState {
				return object, object.DBInstanceStatus, WrapError(fmt.Errorf("ADBPG instance %s status is %s", instanceId, object.DBInstanceStatus))
			}
		}
		return object, object.DBInstanceStatus, nil
	}
}

func (s *AdbpgService) WaitForAdbpgInstanceRunning(instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating", "ClassChanging", "NetAddressCreating", "Restarting"},
		Target:     []string{"Running"},
		Refresh:    s.AdbpgInstanceStateRefreshFunc(instanceId, []string{}),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *AdbpgService) WaitForAdbpgInstanceDeleted(instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{},
		Refresh: s.AdbpgInstanceStateRefreshFunc(instanceId, []string{}),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return WrapError(err)
}

// Account methods (C-10 through C-14)

func (s *AdbpgService) ListAdbpgAccounts(instanceId string) ([]adbpg.AdbpgAccount, error) {
	accounts, err := s.adbpgAPI.ListAccounts(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return accounts, nil
}

func (s *AdbpgService) DescribeAdbpgAccount(id string) (*adbpg.AdbpgAccount, error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	instanceId := parts[0]
	accountName := parts[1]

	accounts, err := s.ListAdbpgAccounts(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	for _, account := range accounts {
		if account.AccountName == accountName {
			return &account, nil
		}
	}
	return nil, WrapErrorf(Error(GetNotFoundMessage("AdbpgAccount", id)), NotFoundMsg, ProviderERROR)
}

func (s *AdbpgService) CreateAdbpgAccount(instanceId string, input *adbpg.AdbpgAccountCreate) error {
	_, err := s.adbpgAPI.CreateAccount(instanceId, input)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) DeleteAdbpgAccount(instanceId, accountName string) error {
	_, err := s.adbpgAPI.DeleteAccount(instanceId, accountName)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) ResetAdbpgAccountPassword(instanceId, accountName, newPassword string) error {
	_, err := s.adbpgAPI.ResetAccountPassword(instanceId, accountName, newPassword)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// ModifyAdbpgAccountDescription modifies the description of an account.
// Uses RPC because cws-lib-go wrapper is pending (T046 deferred — submodule read-only).
func (s *AdbpgService) ModifyAdbpgAccountDescription(instanceId, accountName, description string) error {
	action := "ModifyAccountDescription"
	request := map[string]interface{}{
		"DBInstanceId":       instanceId,
		"AccountName":        accountName,
		"AccountDescription": description,
	}
	_, err := s.client.RpcPost("gpdb", "2016-05-03", action, nil, request, true)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// Database methods (C-15 through C-18)

func (s *AdbpgService) CreateAdbpgDatabase(instanceId string, input *adbpg.AdbpgDatabaseCreate) error {
	_, err := s.adbpgAPI.CreateDatabase(instanceId, input)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) DeleteAdbpgDatabase(instanceId, dbName string) error {
	_, err := s.adbpgAPI.DeleteDatabase(instanceId, dbName)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) ListAdbpgDatabases(instanceId string) ([]adbpg.AdbpgDatabase, error) {
	databases, err := s.adbpgAPI.ListDatabases(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return databases, nil
}

func (s *AdbpgService) DescribeAdbpgDatabase(id string) (*adbpg.AdbpgDatabase, error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	instanceId := parts[0]
	dbName := parts[1]

	databases, err := s.ListAdbpgDatabases(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	for _, db := range databases {
		if db.DBName == dbName {
			return &db, nil
		}
	}
	return nil, WrapErrorf(Error(GetNotFoundMessage("AdbpgDatabase", id)), NotFoundMsg, ProviderERROR)
}

// Connection methods (C-19 through C-21)

func (s *AdbpgService) AllocateAdbpgPublicConnection(instanceId, connectionStringPrefix string) error {
	_, err := s.adbpgAPI.AllocatePublicConnection(instanceId, connectionStringPrefix)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) ReleaseAdbpgPublicConnection(instanceId, connectionString string) error {
	_, err := s.adbpgAPI.ReleasePublicConnection(instanceId, connectionString)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) DescribeAdbpgPublicConnection(instanceId string) (*adbpg.AdbpgNetInfo, error) {
	netInfos, err := s.adbpgAPI.DescribeInstanceNetInfo(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	for _, info := range netInfos {
		if strings.EqualFold(info.IPType, "Public") {
			return &info, nil
		}
	}
	return nil, WrapErrorf(Error(GetNotFoundMessage("AdbpgPublicConnection", instanceId)), NotFoundMsg, ProviderERROR)
}

// Security IP methods (C-22)

func (s *AdbpgService) ModifyAdbpgSecurityIps(instanceId, securityIPList, arrayName string) error {
	_, err := s.adbpgAPI.ModifySecurityIps(instanceId, securityIPList, arrayName)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// Backup policy methods (C-23)

func (s *AdbpgService) DescribeAdbpgBackupPolicy(instanceId string) (*adbpg.AdbpgBackupPolicy, error) {
	policy, err := s.adbpgAPI.DescribeBackupPolicy(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return policy, nil
}

// ModifyAdbpgBackupPolicy modifies the backup policy for an instance.
// Uses RPC because cws-lib-go wrapper is pending (T045 deferred — submodule read-only).
func (s *AdbpgService) ModifyAdbpgBackupPolicy(instanceId string, policy *adbpg.AdbpgBackupPolicy) error {
	action := "ModifyBackupPolicy"
	request := map[string]interface{}{
		"DBInstanceId":          instanceId,
		"BackupRetentionPeriod": policy.BackupRetentionPeriod,
		"PreferredBackupPeriod": policy.PreferredBackupPeriod,
		"PreferredBackupTime":   policy.PreferredBackupTime,
		"EnableRecoveryPoint":   policy.EnableRecoveryPoint,
	}
	_, err := s.client.RpcPost("gpdb", "2016-05-03", action, nil, request, true)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// SSL methods (C-24, C-25)

func (s *AdbpgService) DescribeAdbpgSSL(instanceId string) (*adbpg.AdbpgSSLConfig, error) {
	config, err := s.adbpgAPI.DescribeInstanceSSL(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return config, nil
}

func (s *AdbpgService) ModifyAdbpgSSL(instanceId string, sslEnabled bool) error {
	_, err := s.adbpgAPI.ModifyInstanceSSL(instanceId, sslEnabled)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// Discovery methods (C-26, C-27)

func (s *AdbpgService) ListAdbpgAvailableResources(regionId string) ([]adbpg.AdbpgAvailableResource, error) {
	resources, err := s.adbpgAPI.ListAvailableResources(regionId)
	if err != nil {
		return nil, WrapError(err)
	}
	return resources, nil
}

func (s *AdbpgService) ListAdbpgInstances(query *adbpg.AdbpgInstanceQuery) ([]adbpg.AdbpgInstance, error) {
	instances, err := s.adbpgAPI.ListInstances(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return instances, nil
}

// Vector collection methods

func (s *AdbpgService) ListAdbpgVectorCollections(instanceId, namespace string) ([]adbpg.AdbpgVectorCollection, error) {
	collections, err := s.adbpgAPI.ListCollections(instanceId, namespace)
	if err != nil {
		return nil, WrapError(err)
	}
	return collections, nil
}

func (s *AdbpgService) DescribeAdbpgVectorCollection(id string) (*adbpg.AdbpgVectorCollection, error) {
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return nil, WrapError(err)
	}
	instanceId := parts[0]
	namespace := parts[1]
	collection := parts[2]

	collections, err := s.ListAdbpgVectorCollections(instanceId, namespace)
	if err != nil {
		return nil, WrapError(err)
	}
	for _, c := range collections {
		if c.CollectionName == collection {
			return &c, nil
		}
	}
	return nil, WrapErrorf(Error(GetNotFoundMessage("AdbpgVectorCollection", id)), NotFoundMsg, ProviderERROR)
}

func (s *AdbpgService) DescribeAdbpgVectorCollectionDetail(instanceId, namespace, namespacePassword, collection string) (*adbpg.AdbpgCollectionDetail, error) {
	detail, err := s.adbpgAPI.DescribeCollection(instanceId, namespace, namespacePassword, collection)
	if err != nil {
		return nil, WrapError(err)
	}
	return detail, nil
}

func (s *AdbpgService) CreateAdbpgVectorCollection(instanceId, collection string, input *adbpg.AdbpgCollectionCreate) error {
	_, err := s.adbpgAPI.CreateCollection(instanceId, collection, input)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) DeleteAdbpgVectorCollection(instanceId, namespace, namespacePassword, collection string) error {
	_, err := s.adbpgAPI.DeleteCollection(instanceId, namespace, namespacePassword, collection)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

// Vector namespace methods

func (s *AdbpgService) ListAdbpgNamespaces(instanceId string) ([]adbpg.AdbpgNamespace, error) {
	namespaces, err := s.adbpgAPI.ListNamespaces(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return namespaces, nil
}

// Data source discovery methods

func (s *AdbpgService) ListAdbpgDataSources(instanceId string) ([]adbpg.AdbpgDataSource, error) {
	dataSources, err := s.adbpgAPI.ListDataSources(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return dataSources, nil
}

func (s *AdbpgService) ListAdbpgStreamingJobs(instanceId string) ([]adbpg.AdbpgStreamingJob, error) {
	jobs, err := s.adbpgAPI.ListStreamingJobs(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return jobs, nil
}

func (s *AdbpgService) ListAdbpgBackups(instanceId string) ([]adbpg.AdbpgBackup, error) {
	backups, err := s.adbpgAPI.ListBackups(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return backups, nil
}

func (s *AdbpgService) ListAdbpgInstancePlans(instanceId string) ([]adbpg.AdbpgInstancePlan, error) {
	plans, err := s.adbpgAPI.ListInstancePlans(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return plans, nil
}

// Governance methods

func (s *AdbpgService) ListAdbpgResourceGroups(instanceId string) ([]adbpg.AdbpgResourceGroup, error) {
	groups, err := s.adbpgAPI.ListResourceGroups(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return groups, nil
}

func (s *AdbpgService) ListAdbpgSupabaseProjects(instanceId string) ([]adbpg.AdbpgSupabaseProject, error) {
	projects, err := s.adbpgAPI.ListSupabaseProjects(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return projects, nil
}

// Tag methods (C-28, C-29)

func (s *AdbpgService) TagAdbpgResources(resourceId string, tags []adbpg.Tag) error {
	_, err := s.adbpgAPI.TagResources(resourceId, "instance", tags)
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *AdbpgService) UntagAdbpgResources(resourceId string, tagKeys []string) error {
	_, err := s.adbpgAPI.UntagResources(resourceId, "instance", tagKeys)
	if err != nil {
		return WrapError(err)
	}
	return nil
}
