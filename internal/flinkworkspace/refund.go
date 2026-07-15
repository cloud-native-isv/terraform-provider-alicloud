package flinkworkspace

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	RefundProductCodeDomestic      = "sc"
	RefundProductTypeDomestic      = "sc_flinkserverless_public_cn"
	RefundProductTypeInternational = "sc_flinkserverless_public_intl"
)

func RefundProductType(international bool) string {
	if international {
		return RefundProductTypeInternational
	}
	return RefundProductTypeDomestic
}

func RefundClientToken(regionID, instanceID string) string {
	digest := sha256.Sum256([]byte(regionID + "\x00" + instanceID))
	return "TF-RefundInstance-" + hex.EncodeToString(digest[:])[:32]
}

func BuildRefundRequest(instanceID string, international bool, clientToken string) map[string]interface{} {
	return map[string]interface{}{
		"InstanceId":         instanceID,
		"ClientToken":        clientToken,
		"ImmediatelyRelease": "1",
		"ProductCode":        RefundProductCodeDomestic,
		"ProductType":        RefundProductType(international),
	}
}
