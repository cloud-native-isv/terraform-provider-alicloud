package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// AckService adapts the cws-lib-go ACK API layer (cs-20151215 SDK wrapper)
// for the alicloud_cs_* resources and data sources. It replaces the legacy
// CsService/CsClient and AckServiceV2 implementations that talked to the CS
// RoA endpoint directly.
type AckService struct {
	client *connectivity.AliyunClient
	ackAPI *aliyunAckAPI.AckAPI
}

// NewAckService creates a new AckService backed by the cws-lib-go AckAPI.
func NewAckService(client *connectivity.AliyunClient) (*AckService, error) {
	if client == nil {
		return nil, fmt.Errorf("AliyunClient is required to create AckService")
	}
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}
	ackAPI, err := aliyunAckAPI.NewAckAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go AckAPI: %w", err)
	}
	return &AckService{client: client, ackAPI: ackAPI}, nil
}

// GetAPI returns the underlying AckAPI for direct domain API access.
func (s *AckService) GetAPI() *aliyunAckAPI.AckAPI {
	return s.ackAPI
}

// ackClusterStateDeleted is the synthetic state reported once a cluster no
// longer exists; it is the target state of the delete waiters.
const ackClusterStateDeleted = "deleted"

// ackClusterCreatePendingStates are the cluster states observed while a
// cluster is still being provisioned.
var ackClusterCreatePendingStates = []string{"initial", "provisioning", ""}

// ackClusterCreateFailStates are the terminal cluster creation states.
var ackClusterCreateFailStates = []string{"failed"}

// AckClusterStateRefreshFunc polls DescribeClusterDetail until the cluster
// reaches a target or fail state. A NotFound error keeps the refresh pending
// so creation waits survive the window where the cluster is not yet visible.
func (s *AckService) AckClusterStateRefreshFunc(clusterId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := s.GetAPI().DescribeClusterDetail(clusterId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if cluster.State == failState {
				return cluster, cluster.State, WrapError(fmt.Errorf("cluster %s reached fail state %q", clusterId, cluster.State))
			}
		}
		return cluster, cluster.State, nil
	}
}

// AckClusterDeleteStateRefreshFunc polls DescribeClusterDetail until the
// cluster disappears (NotFound is mapped to the deleted target state).
func (s *AckService) AckClusterDeleteStateRefreshFunc(clusterId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := s.GetAPI().DescribeClusterDetail(clusterId)
		if err != nil {
			if NotFoundError(err) {
				return nil, ackClusterStateDeleted, nil
			}
			return nil, "", WrapError(err)
		}
		return cluster, cluster.State, nil
	}
}

// ackNodePoolStateActive is the target state of node pool create/update waits.
const ackNodePoolStateActive = "active"

// ackNodePoolFailStates are the terminal node pool failure states.
var ackNodePoolFailStates = []string{"failed"}

// AckNodePoolStateRefreshFunc polls DescribeClusterNodePoolDetail for the
// node pool encoded as "clusterId:nodepoolId" until it is active. A NotFound
// error keeps the refresh pending while the pool is still being created.
func (s *AckService) AckNodePoolStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusterId, nodepoolId, err := ackParseTwoPartId(id)
		if err != nil {
			return nil, "", WrapError(err)
		}
		pool, err := s.GetAPI().DescribeClusterNodePoolDetail(clusterId, nodepoolId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if pool.Status == failState {
				return pool, pool.Status, WrapError(fmt.Errorf("node pool %s reached fail state %q", id, pool.Status))
			}
		}
		return pool, pool.Status, nil
	}
}

// AckNodePoolDeleteStateRefreshFunc polls DescribeClusterNodePoolDetail until
// the node pool disappears (NotFound maps to the deleted target state).
func (s *AckService) AckNodePoolDeleteStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusterId, nodepoolId, err := ackParseTwoPartId(id)
		if err != nil {
			return nil, "", WrapError(err)
		}
		pool, err := s.GetAPI().DescribeClusterNodePoolDetail(clusterId, nodepoolId)
		if err != nil {
			if NotFoundError(err) {
				return nil, ackClusterStateDeleted, nil
			}
			return nil, "", WrapError(err)
		}
		return pool, pool.Status, nil
	}
}

// ackAddonStateInstalled is the target state of the addon install wait: the
// DescribeClusterAddonInstance call succeeds once the component is installed.
const ackAddonStateInstalled = "installed"

// AckAddonInstanceStateRefreshFunc polls DescribeClusterAddonInstance for the
// addon encoded as "clusterId:name" until the instance is installed.
func (s *AckService) AckAddonInstanceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusterId, name, err := ackParseTwoPartId(id)
		if err != nil {
			return nil, "", WrapError(err)
		}
		instance, err := s.GetAPI().DescribeClusterAddonInstance(clusterId, name)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		return instance, ackAddonStateInstalled, nil
	}
}

// AckAddonInstanceDeleteStateRefreshFunc polls DescribeClusterAddonInstance
// until the addon disappears (NotFound maps to the deleted target state).
func (s *AckService) AckAddonInstanceDeleteStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusterId, name, err := ackParseTwoPartId(id)
		if err != nil {
			return nil, "", WrapError(err)
		}
		instance, err := s.GetAPI().DescribeClusterAddonInstance(clusterId, name)
		if err != nil {
			if NotFoundError(err) {
				return nil, ackClusterStateDeleted, nil
			}
			return nil, "", WrapError(err)
		}
		return instance, instance.State, nil
	}
}

// ackParseTwoPartId splits a "clusterId:xxx" composite resource id into its
// two non-empty parts, mirroring the fork's DescribeAckNodepool convention.
func ackParseTwoPartId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid Resource Id %s. Expected parts' length 2, got %d", id, len(parts))
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid Resource Id %s. Expected non-empty cluster id and resource name", id)
	}
	return parts[0], parts[1], nil
}

// expandAckTags converts a schema tags map into the API layer's tag slice.
// A typed nil map (unset) yields a nil slice; a non-nil map (even empty)
// always yields an allocated slice.
func expandAckTags(raw interface{}) []aliyunAckAPI.AckTag {
	value, ok := raw.(map[string]interface{})
	if !ok || value == nil {
		return nil
	}
	result := make([]aliyunAckAPI.AckTag, 0, len(value))
	for k, v := range value {
		result = append(result, aliyunAckAPI.AckTag{Key: k, Value: v.(string)})
	}
	return result
}

// flattenAckTags converts the API layer's tag slice into a schema tags map.
// A nil slice round-trips to nil so an unset tags attribute stays unset.
func flattenAckTags(tags []aliyunAckAPI.AckTag) map[string]interface{} {
	if tags == nil {
		return nil
	}
	result := make(map[string]interface{})
	for _, tag := range tags {
		result[tag.Key] = tag.Value
	}
	return result
}

// buildAckStateConf builds a StateChangeConf with the ACK wait profile used
// by the cs kubernetes resources (fixed poll interval, no backoff).
func buildAckStateConf(pending, target []string, timeout, delay, poll time.Duration, f resource.StateRefreshFunc) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:      pending,
		Target:       target,
		Refresh:      f,
		Timeout:      timeout,
		Delay:        delay,
		PollInterval: poll,
		MinTimeout:   poll,
	}
}
