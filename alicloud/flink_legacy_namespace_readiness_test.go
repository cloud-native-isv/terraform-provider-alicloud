package alicloud

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type fakeFlinkLegacyNamespaceReadinessService struct {
	workspaceFn  func(int) (*flink.Workspace, error)
	namespacesFn func(int) ([]flink.Namespace, error)
	queuesFn     func(string, int) ([]flink.DeploymentTarget, error)
	createFn     func(*flink.Namespace) (*flink.Namespace, error)

	workspaceCalls int
	namespaceCalls int
	queueCalls     map[string]int
	createCalls    int
	created        bool
}

func (f *fakeFlinkLegacyNamespaceReadinessService) DescribeFlinkWorkspace(string) (*flink.Workspace, error) {
	f.workspaceCalls++
	return f.workspaceFn(f.workspaceCalls)
}

func (f *fakeFlinkLegacyNamespaceReadinessService) ListNamespaces(string) ([]flink.Namespace, error) {
	f.namespaceCalls++
	return f.namespacesFn(f.namespaceCalls)
}

func (f *fakeFlinkLegacyNamespaceReadinessService) ListFlinkDeploymentTargets(_ string, namespace string) ([]flink.DeploymentTarget, error) {
	if f.queueCalls == nil {
		f.queueCalls = make(map[string]int)
	}
	f.queueCalls[namespace]++
	return f.queuesFn(namespace, f.queueCalls[namespace])
}

func (f *fakeFlinkLegacyNamespaceReadinessService) CreateNamespace(_ string, namespace *flink.Namespace) (*flink.Namespace, error) {
	f.createCalls++
	f.created = true
	return f.createFn(namespace)
}

func readyFlinkLegacyWorkspace() *flink.Workspace {
	return &flink.Workspace{Id: "f-paid", Name: "workspace", Status: "RUNNING", ResourceId: "workspace-resource"}
}

func readyFlinkLegacyNamespace(name string) flink.Namespace {
	return flink.Namespace{
		Name:         name,
		Status:       "SUCCESS",
		ResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
	}
}

func readyFlinkLegacyDefaultQueue(namespace string) flink.DeploymentTarget {
	return flink.DeploymentTarget{
		Name:      "default-queue",
		Namespace: namespace,
		Quota: &flink.ResourceQuota{
			Request: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
			Limit:   &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
		},
	}
}

func TestFlinkNamespacesReadinessWaitsForWorkspaceNamespaceAndDefaultQueueVisibility(t *testing.T) {
	service := &fakeFlinkLegacyNamespaceReadinessService{
		workspaceFn: func(call int) (*flink.Workspace, error) {
			switch call {
			case 1:
				return nil, flink.NewFlinkServiceErrorWithCode("", "", "404", "not visible", "")
			case 2:
				return &flink.Workspace{Id: "f-paid", Status: "CREATING"}, nil
			case 3:
				return &flink.Workspace{Id: "f-paid", Status: "RUNNING"}, nil
			default:
				return readyFlinkLegacyWorkspace(), nil
			}
		},
		namespacesFn: func(call int) ([]flink.Namespace, error) {
			switch call {
			case 1:
				return nil, nil
			case 2:
				return []flink.Namespace{{Name: "workspace-default", Status: "CREATING"}}, nil
			default:
				return []flink.Namespace{readyFlinkLegacyNamespace("workspace-default")}, nil
			}
		},
		queuesFn: func(namespace string, call int) ([]flink.DeploymentTarget, error) {
			switch call {
			case 1:
				return nil, nil
			case 2:
				return []flink.DeploymentTarget{{Name: "default-queue", Namespace: namespace}}, nil
			default:
				return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
			}
		},
		createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
	}

	namespaces, err := waitForFlinkLegacyNamespaceReadiness(
		context.Background(), service, "f-paid", flinkLegacyExactNamespaceSelection(map[string]struct{}{"workspace-default": {}}), 250*time.Millisecond, time.Millisecond,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(namespaces) != 1 || namespaces[0].Name != "workspace-default" || namespaces[0].ResourceSpec == nil {
		t.Fatalf("ready namespaces = %#v, want complete requested default namespace", namespaces)
	}
	if service.workspaceCalls < 7 || service.namespaceCalls < 4 || service.queueCalls["workspace-default"] != 3 {
		t.Fatalf("readiness calls workspace/namespaces/queues=%d/%d/%d, want every delayed layer polled", service.workspaceCalls, service.namespaceCalls, service.queueCalls["workspace-default"])
	}
}

func TestFlinkNamespacesReadinessFailsClosedOnTypedTerminalAndBusinessErrors(t *testing.T) {
	for _, test := range []struct {
		name        string
		workspaceFn func(int) (*flink.Workspace, error)
		want        string
	}{
		{
			name: "workspace terminal state",
			workspaceFn: func(int) (*flink.Workspace, error) {
				return &flink.Workspace{Id: "f-paid", Status: "FAILED"}, nil
			},
			want: "terminal state",
		},
		{
			name: "permission service error",
			workspaceFn: func(int) (*flink.Workspace, error) {
				return nil, flink.NewFlinkServiceErrorWithCode("request", "", "Forbidden", "denied", "")
			},
			want: "Forbidden",
		},
		{
			name: "business service error",
			workspaceFn: func(int) (*flink.Workspace, error) {
				return nil, flink.NewFlinkServiceErrorWithCode("request", "", "InvalidParameter", "invalid business request", "")
			},
			want: "InvalidParameter",
		},
		{
			name: "plain not found text is not classified",
			workspaceFn: func(int) (*flink.Workspace, error) {
				return nil, errors.New("workspace not found while permission scope is unknown")
			},
			want: "not found",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &fakeFlinkLegacyNamespaceReadinessService{
				workspaceFn:  test.workspaceFn,
				namespacesFn: func(int) ([]flink.Namespace, error) { t.Fatal("ListNamespaces must not run"); return nil, nil },
				queuesFn: func(string, int) ([]flink.DeploymentTarget, error) {
					t.Fatal("ListFlinkDeploymentTargets must not run")
					return nil, nil
				},
				createFn: func(*flink.Namespace) (*flink.Namespace, error) {
					t.Fatal("CreateNamespace must not run")
					return nil, nil
				},
			}
			_, err := waitForFlinkLegacyNamespaceReadiness(context.Background(), service, "f-paid", flinkLegacyBootstrapNamespaceSelection(), time.Second, time.Millisecond)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			if service.workspaceCalls != 1 {
				t.Fatalf("terminal observer calls = %d, want 1", service.workspaceCalls)
			}
		})
	}
}

func TestFlinkNamespacesReadinessHonorsCancellationAndTimeout(t *testing.T) {
	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		service := &fakeFlinkLegacyNamespaceReadinessService{
			workspaceFn: func(int) (*flink.Workspace, error) {
				cancel()
				return &flink.Workspace{Id: "f-paid", Status: "CREATING"}, nil
			},
			namespacesFn: func(int) ([]flink.Namespace, error) { return nil, nil },
			queuesFn:     func(string, int) ([]flink.DeploymentTarget, error) { return nil, nil },
			createFn:     func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
		}
		_, err := waitForFlinkLegacyNamespaceReadiness(ctx, service, "f-paid", flinkLegacyBootstrapNamespaceSelection(), time.Second, time.Millisecond)
		if !errors.Is(err, context.Canceled) || service.workspaceCalls != 1 {
			t.Fatalf("cancellation error/calls = %v/%d, want context canceled/1", err, service.workspaceCalls)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		service := &fakeFlinkLegacyNamespaceReadinessService{
			workspaceFn:  func(int) (*flink.Workspace, error) { return &flink.Workspace{Id: "f-paid", Status: "CREATING"}, nil },
			namespacesFn: func(int) ([]flink.Namespace, error) { return nil, nil },
			queuesFn:     func(string, int) ([]flink.DeploymentTarget, error) { return nil, nil },
			createFn:     func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
		}
		_, err := waitForFlinkLegacyNamespaceReadiness(context.Background(), service, "f-paid", flinkLegacyBootstrapNamespaceSelection(), 5*time.Millisecond, time.Millisecond)
		if err == nil || !strings.Contains(err.Error(), "timed out") || service.workspaceCalls < 2 {
			t.Fatalf("timeout error/calls = %v/%d, want bounded polling timeout", err, service.workspaceCalls)
		}
	})
}

func TestFlinkNamespacesDataSourcePublishesOnlyCompleteRequestedNamespace(t *testing.T) {
	service := &fakeFlinkLegacyNamespaceReadinessService{
		workspaceFn: func(int) (*flink.Workspace, error) { return readyFlinkLegacyWorkspace(), nil },
		namespacesFn: func(int) ([]flink.Namespace, error) {
			return []flink.Namespace{readyFlinkLegacyNamespace("workspace-default")}, nil
		},
		queuesFn: func(namespace string, _ int) ([]flink.DeploymentTarget, error) {
			return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
		},
		createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
	}
	data := schema.TestResourceDataRaw(t, dataSourceAliCloudFlinkNamespaces().Schema, map[string]interface{}{
		"workspace_id": "f-paid",
		"names":        []interface{}{"workspace-default"},
	})

	if err := readFlinkNamespacesWithService(context.Background(), data, service, time.Second, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if data.Id() == "" {
		t.Fatal("data source ID is empty")
	}
	got := data.Get("namespaces").([]interface{})
	if len(got) != 1 {
		t.Fatalf("namespaces = %#v, want one complete result", got)
	}
	namespace := got[0].(map[string]interface{})
	if namespace["name"] != "workspace-default" || namespace["cpu"] != 4 || namespace["memory"] != 16 {
		t.Fatalf("namespace output = %#v, want complete requested default", namespace)
	}
}

func TestFlinkNamespaceCreateWaitsAtBothDependentBoundaries(t *testing.T) {
	var service *fakeFlinkLegacyNamespaceReadinessService
	service = &fakeFlinkLegacyNamespaceReadinessService{
		workspaceFn: func(int) (*flink.Workspace, error) { return readyFlinkLegacyWorkspace(), nil },
		namespacesFn: func(call int) ([]flink.Namespace, error) {
			base := []flink.Namespace{readyFlinkLegacyNamespace("workspace-default")}
			if !service.created {
				return base, nil
			}
			if call < 3 {
				return base, nil
			}
			if call == 3 {
				return append(base, flink.Namespace{Name: "analytics", Status: "CREATING"}), nil
			}
			return append(base, readyFlinkLegacyNamespace("analytics")), nil
		},
		queuesFn: func(namespace string, call int) ([]flink.DeploymentTarget, error) {
			if namespace == "analytics" && call == 1 {
				return nil, nil
			}
			return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
		},
		createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) {
			if namespace.Name != "analytics" {
				t.Fatalf("CreateNamespace name = %q, want analytics", namespace.Name)
			}
			return namespace, nil
		},
	}

	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkNamespace().Schema, map[string]interface{}{
		"workspace_id":   "f-paid",
		"namespace_name": "analytics",
		"elastic_resource_spec": []interface{}{map[string]interface{}{
			"cpu": 4, "memory_gb": 16,
		}},
	})
	if err := createFlinkNamespaceWithService(context.Background(), data, service, 250*time.Millisecond, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if data.Id() != "f-paid:analytics" || service.createCalls != 1 || service.queueCalls["analytics"] != 2 {
		t.Fatalf("child ID/create/queue calls = %q/%d/%d, want ready child and one create", data.Id(), service.createCalls, service.queueCalls["analytics"])
	}
	if got := data.Get("status"); got != "SUCCESS" {
		t.Fatalf("child status = %#v, want SUCCESS", got)
	}
}

func TestFlinkNamespaceCreateRecoversTypedPostCreateReadWithoutReplayingWrite(t *testing.T) {
	var service *fakeFlinkLegacyNamespaceReadinessService
	service = &fakeFlinkLegacyNamespaceReadinessService{
		workspaceFn: func(int) (*flink.Workspace, error) { return readyFlinkLegacyWorkspace(), nil },
		namespacesFn: func(int) ([]flink.Namespace, error) {
			if !service.created {
				return []flink.Namespace{readyFlinkLegacyNamespace("workspace-default")}, nil
			}
			return []flink.Namespace{readyFlinkLegacyNamespace("workspace-default"), readyFlinkLegacyNamespace("analytics")}, nil
		},
		queuesFn: func(namespace string, _ int) ([]flink.DeploymentTarget, error) {
			return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
		},
		createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) {
			return nil, flink.NewFlinkPostCreateReadError("f-paid", namespace.Name, flink.NewFlinkServiceErrorWithCode("", "", "404", "delayed", ""))
		},
	}
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkNamespace().Schema, map[string]interface{}{
		"workspace_id": "f-paid", "namespace_name": "analytics",
	})

	if err := createFlinkNamespaceWithService(context.Background(), data, service, time.Second, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if service.createCalls != 1 || data.Id() != "f-paid:analytics" {
		t.Fatalf("post-create recovery create calls/ID = %d/%q, want 1/f-paid:analytics", service.createCalls, data.Id())
	}
}

func TestFlinkNamespaceCreateReadinessFailureStaysOnChild(t *testing.T) {
	t.Run("terminal before create", func(t *testing.T) {
		service := &fakeFlinkLegacyNamespaceReadinessService{
			workspaceFn:  func(int) (*flink.Workspace, error) { return &flink.Workspace{Id: "f-paid", Status: "FAILED"}, nil },
			namespacesFn: func(int) ([]flink.Namespace, error) { return nil, nil },
			queuesFn:     func(string, int) ([]flink.DeploymentTarget, error) { return nil, nil },
			createFn:     func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
		}
		data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkNamespace().Schema, map[string]interface{}{
			"workspace_id": "f-paid", "namespace_name": "analytics",
		})
		err := createFlinkNamespaceWithService(context.Background(), data, service, time.Second, time.Millisecond)
		if err == nil || service.createCalls != 0 || data.Id() != "" {
			t.Fatalf("pre-create terminal error/create/ID = %v/%d/%q, want child-only error before write", err, service.createCalls, data.Id())
		}
	})

	t.Run("terminal after create preserves child identity", func(t *testing.T) {
		var service *fakeFlinkLegacyNamespaceReadinessService
		service = &fakeFlinkLegacyNamespaceReadinessService{
			workspaceFn: func(int) (*flink.Workspace, error) { return readyFlinkLegacyWorkspace(), nil },
			namespacesFn: func(int) ([]flink.Namespace, error) {
				if !service.created {
					return []flink.Namespace{readyFlinkLegacyNamespace("workspace-default")}, nil
				}
				return []flink.Namespace{readyFlinkLegacyNamespace("workspace-default"), {Name: "analytics", Status: "FAILED"}}, nil
			},
			queuesFn: func(namespace string, _ int) ([]flink.DeploymentTarget, error) {
				return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
			},
			createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
		}
		data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkNamespace().Schema, map[string]interface{}{
			"workspace_id": "f-paid", "namespace_name": "analytics",
		})
		err := createFlinkNamespaceWithService(context.Background(), data, service, time.Second, time.Millisecond)
		if err == nil || !strings.Contains(err.Error(), "terminal state") || service.createCalls != 1 || data.Id() != "f-paid:analytics" {
			t.Fatalf("post-create terminal error/create/ID = %v/%d/%q, want child identity retained", err, service.createCalls, data.Id())
		}
	})
}

func TestFlinkNamespaceCreateBootstrapReadinessIgnoresUnrelatedNamespaces(t *testing.T) {
	for _, test := range []struct {
		name   string
		status string
	}{
		{name: "unrelated failed namespace", status: "FAILED"},
		{name: "concurrent sibling creating", status: "CREATING"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var service *fakeFlinkLegacyNamespaceReadinessService
			service = &fakeFlinkLegacyNamespaceReadinessService{
				workspaceFn: func(int) (*flink.Workspace, error) { return readyFlinkLegacyWorkspace(), nil },
				namespacesFn: func(int) ([]flink.Namespace, error) {
					namespaces := []flink.Namespace{
						readyFlinkLegacyNamespace("workspace-default"),
						{Name: "unrelated", Status: test.status},
					}
					if service.created {
						namespaces = append(namespaces, readyFlinkLegacyNamespace("analytics"))
					}
					return namespaces, nil
				},
				queuesFn: func(namespace string, _ int) ([]flink.DeploymentTarget, error) {
					if namespace == "unrelated" {
						return nil, nil
					}
					return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
				},
				createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
			}
			data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkNamespace().Schema, map[string]interface{}{
				"workspace_id": "f-paid", "namespace_name": "analytics",
			})

			if err := createFlinkNamespaceWithService(context.Background(), data, service, 30*time.Millisecond, time.Millisecond); err != nil {
				t.Fatalf("unrelated namespace blocked child create: %v", err)
			}
			if service.createCalls != 1 || service.queueCalls["unrelated"] != 0 {
				t.Fatalf("create/unrelated queue calls = %d/%d, want 1/0", service.createCalls, service.queueCalls["unrelated"])
			}
		})
	}
}

func TestFlinkNamespaceCreateBootstrapReadinessRequiresWorkspaceName(t *testing.T) {
	service := &fakeFlinkLegacyNamespaceReadinessService{
		workspaceFn: func(int) (*flink.Workspace, error) {
			return &flink.Workspace{Id: "f-paid", Status: "RUNNING", ResourceId: "workspace-resource"}, nil
		},
		namespacesFn: func(int) ([]flink.Namespace, error) {
			return []flink.Namespace{readyFlinkLegacyNamespace("guessed-default")}, nil
		},
		queuesFn: func(namespace string, _ int) ([]flink.DeploymentTarget, error) {
			return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
		},
		createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
	}
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkNamespace().Schema, map[string]interface{}{
		"workspace_id": "f-paid", "namespace_name": "analytics",
	})

	err := createFlinkNamespaceWithService(context.Background(), data, service, 30*time.Millisecond, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "workspace name") || service.createCalls != 0 {
		t.Fatalf("missing workspace name error/create calls = %v/%d, want fail closed before write", err, service.createCalls)
	}
}

func TestFlinkNamespacesUnfilteredDataSourcePublishesFullSnapshotAfterBootstrapReady(t *testing.T) {
	service := &fakeFlinkLegacyNamespaceReadinessService{
		workspaceFn: func(int) (*flink.Workspace, error) { return readyFlinkLegacyWorkspace(), nil },
		namespacesFn: func(int) ([]flink.Namespace, error) {
			return []flink.Namespace{
				readyFlinkLegacyNamespace("workspace-default"),
				{Name: "failed-unrelated", Status: "FAILED"},
				{Name: "creating-unrelated", Status: "CREATING"},
			}, nil
		},
		queuesFn: func(namespace string, _ int) ([]flink.DeploymentTarget, error) {
			if namespace != "workspace-default" {
				t.Fatalf("unfiltered readiness queried unrelated namespace %q", namespace)
			}
			return []flink.DeploymentTarget{readyFlinkLegacyDefaultQueue(namespace)}, nil
		},
		createFn: func(namespace *flink.Namespace) (*flink.Namespace, error) { return namespace, nil },
	}
	data := schema.TestResourceDataRaw(t, dataSourceAliCloudFlinkNamespaces().Schema, map[string]interface{}{
		"workspace_id": "f-paid",
	})

	if err := readFlinkNamespacesWithService(context.Background(), data, service, time.Second, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	got := data.Get("namespaces").([]interface{})
	if len(got) != 3 {
		t.Fatalf("unfiltered namespaces = %#v, want full three-item snapshot", got)
	}
}

type flinkLegacyTemporaryNetError struct {
	timeout   bool
	temporary bool
	message   string
}

func (e flinkLegacyTemporaryNetError) Error() string   { return e.message }
func (e flinkLegacyTemporaryNetError) Timeout() bool   { return e.timeout }
func (e flinkLegacyTemporaryNetError) Temporary() bool { return e.temporary }

var _ net.Error = flinkLegacyTemporaryNetError{}

func TestFlinkLegacyNamespaceReadinessRetryClassificationUsesTypedStructureOnly(t *testing.T) {
	permission := flink.NewFlinkServiceErrorWithCode("request", "", "Forbidden", "denied", "")
	typed404 := flink.NewFlinkServiceErrorWithCode("request", "", "404", "not visible", "")
	wrappedTimeout := flink.NewFlinkSDKError("foasconsole", "ListNamespaces", "transport", context.DeadlineExceeded)
	wrappedTemporary := flink.NewFlinkSDKError("foasconsole", "ListNamespaces", "transport", flinkLegacyTemporaryNetError{temporary: true, message: "temporary transport"})
	wrappedTeaStatus := flink.NewFlinkSDKError("foasconsole", "ListNamespaces", "tea", tea.NewSDKError(map[string]interface{}{
		"statusCode": 503, "code": "Unrelated", "message": "terminal-looking text",
	}))
	wrappedTeaCode := flink.NewFlinkSDKError("foasconsole", "ListNamespaces", "tea", tea.NewSDKError(map[string]interface{}{
		"statusCode": 400, "code": "SystemBusy", "message": "permission-looking text",
	}))
	wrappedTeaMisleading := flink.NewFlinkSDKError("foasconsole", "ListNamespaces", "tea", tea.NewSDKError(map[string]interface{}{
		"statusCode": 400, "code": "Forbidden", "message": "Post https://example.invalid ServiceUnavailable",
	}))

	for _, test := range []struct {
		name string
		err  error
		want bool
	}{
		{name: "typed deadline", err: wrappedTimeout, want: true},
		{name: "typed temporary net error", err: wrappedTemporary, want: true},
		{name: "plain Post text", err: errors.New("Post https://example.invalid: opaque transport text"), want: false},
		{name: "typed service 404", err: typed404, want: true},
		{name: "permission service error", err: permission, want: false},
		{name: "wrapped Tea retryable status", err: wrappedTeaStatus, want: true},
		{name: "wrapped Tea retryable code", err: wrappedTeaCode, want: true},
		{name: "wrapped Tea misleading text", err: wrappedTeaMisleading, want: false},
		{name: "plain misleading text", err: errors.New("Throttling ServiceUnavailable code: 503"), want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := flinkLegacyNamespaceReadinessRetryableError(test.err); got != test.want {
				t.Fatalf("retryable(%T: %v) = %v, want %v", test.err, test.err, got, test.want)
			}
		})
	}
}
