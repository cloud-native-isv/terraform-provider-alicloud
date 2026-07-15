package flinkworkspace

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

func BuildRefundRequest(instanceID string, international bool, clientToken string) map[string]interface{} {
	return map[string]interface{}{
		"InstanceId":         instanceID,
		"ClientToken":        clientToken,
		"ImmediatelyRelease": "1",
		"ProductCode":        RefundProductCodeDomestic,
		"ProductType":        RefundProductType(international),
	}
}
