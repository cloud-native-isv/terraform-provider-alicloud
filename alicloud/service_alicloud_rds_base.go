package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
)

// RdsService provides RDS related operations
type RdsService struct {
	client *connectivity.AliyunClient
}

//	_______________                      _______________                       _______________
//	|              | ______param______\  |              |  _____request_____\  |              |
//	|   Business   |                     |    Service   |                      |    SDK/API   |
//	|              | __________________  |              |  __________________  |              |
//	|______________| \    (obj, err)     |______________|  \ (status, cont)    |______________|
//	                    |                                    |
//	                    |A. {instance, nil}                  |a. {200, content}
//	                    |B. {nil, error}                     |b. {200, nil}
//	               					  |c. {4xx, nil}
//
// The API return 200 for resource not found.
// When getInstance is empty, then throw InstanceNotfound error.
// That the business layer only need to check error.
var DBInstanceStatusCatcher = Catcher{"OperationDenied.DBInstanceStatus", 60, 5}
