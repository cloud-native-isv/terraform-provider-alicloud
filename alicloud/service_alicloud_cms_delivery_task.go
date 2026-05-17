package alicloud

func (s *CmsService) ListCmsDeliveryTasks() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsDeliveryTasks")
}

func (s *CmsService) GetCmsDeliveryTask(taskID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsDeliveryTask")
}
