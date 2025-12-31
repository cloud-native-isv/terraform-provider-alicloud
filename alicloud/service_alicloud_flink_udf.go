package alicloud

import (
	"fmt"

	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// UdfArtifact methods
func (s *FlinkService) CreateUdfArtifact(workspaceId string, namespaceName string, artifact *flinkAPI.UdfArtifact) (*flinkAPI.UdfArtifact, error) {
	return s.GetAPI().CreateUdfArtifact(workspaceId, namespaceName, artifact)
}

func (s *FlinkService) GetUdfArtifact(workspaceId string, namespaceName string, artifactName string) (*flinkAPI.UdfArtifact, error) {
	artifacts, err := s.GetAPI().ListUdfArtifacts(workspaceId, namespaceName)
	if err != nil {
		return nil, err
	}
	for _, artifact := range artifacts {
		if artifact.Name == artifactName {
			return artifact, nil
		}
	}
	return nil, fmt.Errorf("UdfArtifact %s not found", artifactName)
}

func (s *FlinkService) UpdateUdfArtifact(workspaceId string, namespaceName string, artifact *flinkAPI.UdfArtifact) (*flinkAPI.UdfArtifact, error) {
	return s.GetAPI().UpdateUdfArtifact(workspaceId, namespaceName, artifact)
}

func (s *FlinkService) DeleteUdfArtifact(workspaceId string, namespaceName string, artifactName string) error {
	return s.GetAPI().DeleteUdfArtifact(workspaceId, namespaceName, artifactName)
}

func (s *FlinkService) FlinkUdfArtifactStateRefreshFunc(workspaceId string, namespaceName string, artifactName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		artifact, err := s.GetUdfArtifact(workspaceId, namespaceName, artifactName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		return artifact, "Available", nil
	}
}
