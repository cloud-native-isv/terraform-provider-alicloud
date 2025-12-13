package alicloud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Layer methods for FCService

// EncodeLayerVersionId encodes layer name and version into a resource ID string
// Format: layerName:version
func EncodeLayerVersionId(layerName, version string) string {
	return fmt.Sprintf("%s:%s", layerName, version)
}

// DecodeLayerVersionId decodes layer version ID string to layer name and version
func DecodeLayerVersionId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid layer version ID format, expected layerName:version, got %s", id)
	}
	return parts[0], parts[1], nil
}

// DescribeFCLayerVersion retrieves layer version information by combined ID
func (s *FCService) DescribeFCLayerVersion(id string) (*aliyunFCAPI.Layer, error) {
	layerName, version, err := DecodeLayerVersionId(id)
	if err != nil {
		return nil, err
	}
	return s.DescribeFCLayerVersionByNameAndVersion(layerName, version)
}

// DescribeFCLayerVersionByNameAndVersion retrieves layer version information by layer name and version
func (s *FCService) DescribeFCLayerVersionByNameAndVersion(layerName, version string) (*aliyunFCAPI.Layer, error) {
	if layerName == "" {
		return nil, fmt.Errorf("layer name cannot be empty")
	}
	if version == "" {
		return nil, fmt.Errorf("version cannot be empty")
	}

	// Convert version string to int32
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	return s.GetAPI().GetLayer(layerName, int32(versionInt))
}

// ListFCLayerVersions lists all layer versions for a layer with optional filters
func (s *FCService) ListFCLayerVersions(layerName string, limit *int32, startVersion *string) ([]*aliyunFCAPI.Layer, error) {
	var limitInt int32
	var startVersionStr string

	if limit != nil {
		limitInt = *limit
	}
	if startVersion != nil {
		startVersionStr = *startVersion
	}

	layerVersions, _, err := s.GetAPI().ListLayerVersions(layerName, startVersionStr, limitInt)
	return layerVersions, err
}

// ListFCLayers lists all layers with optional filters
func (s *FCService) ListFCLayers(prefix *string, limit *int32, nextToken *string, official *bool) ([]*aliyunFCAPI.Layer, error) {
	var prefixStr, nextTokenStr string
	var limitInt int32
	var publicBool, officialBool bool

	if prefix != nil {
		prefixStr = *prefix
	}
	if nextToken != nil {
		nextTokenStr = *nextToken
	}
	if limit != nil {
		limitInt = *limit
	}
	if official != nil {
		officialBool = *official
	}

	layers, _, err := s.GetAPI().ListLayers(prefixStr, nextTokenStr, limitInt, publicBool, officialBool)
	return layers, err
}

// CreateFCLayerVersion creates a new layer version and returns the created layer information
func (s *FCService) CreateFCLayerVersion(layerName string, layer *aliyunFCAPI.Layer) (*aliyunFCAPI.Layer, error) {
	if layerName == "" {
		return nil, fmt.Errorf("layer name cannot be empty")
	}
	if layer == nil {
		return nil, fmt.Errorf("layer cannot be nil")
	}
	return s.GetAPI().CreateLayer(layerName, layer)
}

// DeleteFCLayerVersion deletes an FC layer version
func (s *FCService) DeleteFCLayerVersion(layerName, version string) error {
	if layerName == "" {
		return fmt.Errorf("layer name cannot be empty")
	}
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Convert version string to int32
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		return fmt.Errorf("invalid version format: %s", version)
	}

	return s.GetAPI().DeleteLayer(layerName, int32(versionInt))
}

// LayerVersionStateRefreshFunc returns a StateRefreshFunc to wait for layer version status changes
func (s *FCService) LayerVersionStateRefreshFunc(layerName, version string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCLayerVersionByNameAndVersion(layerName, version)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentStatus := "Active" // Layers don't have explicit status, assume Active if retrievable

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// WaitForFCLayerVersionCreating waits for layer version creation to complete
func (s *FCService) WaitForFCLayerVersionCreating(layerName, version string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.LayerVersionStateRefreshFunc(layerName, version, []string{"Failed"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, EncodeLayerVersionId(layerName, version))
}

// WaitForFCLayerVersionDeleting waits for layer version deletion to complete
func (s *FCService) WaitForFCLayerVersionDeleting(layerName, version string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			obj, err := s.DescribeFCLayerVersionByNameAndVersion(layerName, version)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return obj, "Deleting", nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, EncodeLayerVersionId(layerName, version))
}
