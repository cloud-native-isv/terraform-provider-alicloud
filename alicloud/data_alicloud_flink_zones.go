func dataSourceAliCloudFlinkZonesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	// 修改1: 使用NewFlinkService正确初始化FlinkService
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return err
	}

	request := &foasconsole.DescribeSupportedZonesRequest{} // 修改2: 明确指针类型
	response, err := flinkService.DescribeSupportedZones(request)
	if err != nil {
		return err
	}

	zones := make([]map[string]interface{}, 0)
	for _, zone := range response.Body.Zones { // 假设响应结构包含Zones字段
		zones = append(zones, map[string]interface{}{
			"zone_id":   zone.ZoneId,
			"zone_name": zone.ZoneName,
			// 根据实际响应字段补充
		})
	}

	d.SetId(dataResourceIdHash())
	if err := d.Set("zones", zones); err != nil {
		return err
	}

	return nil
}