TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=alicloud
RELEASE_ALPHA_VERSION=$(VERSION)-alpha$(shell date +'%Y%m%d')
RELEASE_ALPHA_NAME=terraform-provider-alicloud_v$(RELEASE_ALPHA_VERSION)

# 默认目标: 显示帮助信息
.PHONY: help
help:
	@echo "Terraform Provider Alibaba Cloud 构建工具"
	@echo ""
	@echo "使用方法:"
	@echo "  make [目标]"
	@echo ""
	@echo "目标:"
	@echo "  help              显示此帮助信息"
	@echo ""
	@echo "=== 开发构建目标 ==="
	@echo "  build             执行代码格式检查并构建所有平台的二进制文件"
	@echo "  dev               构建Mac版本并安装到本地Terraform路径"
	@echo "  devlinux          构建Linux版本并安装到本地Terraform路径"
	@echo "  devwin            构建Windows版本并安装到本地Terraform路径"
	@echo "  clean             清理bin目录"
	@echo ""
	@echo "=== 平台特定构建目标 ==="
	@echo "  all               构建所有平台的二进制文件 (mac, windows, linux)"
	@echo "  mac               仅构建Mac版本 (amd64)"
	@echo "  windows           仅构建Windows版本 (amd64)"
	@echo "  linux             仅构建Linux版本 (amd64)"
	@echo "  macarm            构建Mac ARM版本并安装到本地插件目录"
	@echo "  alpha             构建Alpha版本并上传到OSS"
	@echo ""
	@echo "=== 代码质量目标 ==="
	@echo "  fmt               格式化代码"
	@echo "  fmtcheck          检查代码格式"
	@echo "  vet               执行go vet静态分析"
	@echo "  errcheck          检查错误处理"
	@echo "  importscheck      检查imports格式"
	@echo ""
	@echo "=== 测试目标 ==="
	@echo "  test              运行单元测试"
	@echo "  testacc           运行验收测试"
	@echo "  test-compile      编译测试二进制文件但不运行测试"
	@echo ""
	@echo "=== 文档目标 ==="
	@echo "  website           构建网站文档"
	@echo "  website-test      测试网站文档"
	@echo ""

# 默认目标改为help
default: help

#-----------------------------------
# 主构建目标
#-----------------------------------
build: fmtcheck all

all: mac windows linux

#-----------------------------------
# 开发构建目标
#-----------------------------------
dev: clean mac copy

devlinux: clean fmt linux linuxcopy

devwin: clean fmt windows windowscopy

copy:
	tar -xvf bin/terraform-provider-alicloud_darwin-amd64.tgz && mv bin/terraform-provider-alicloud $(shell dirname `which terraform`)

linuxcopy:
	tar -xvf bin/terraform-provider-alicloud_linux-amd64.tgz && mv bin/terraform-provider-alicloud $(shell dirname `which terraform`)

windowscopy:
	tar -xvf bin/terraform-provider-alicloud_windows-amd64.tgz && mv bin/terraform-provider-alicloud $(shell dirname `which terraform`)

clean:
	rm -rf bin/*

#-----------------------------------
# 平台特定构建
#-----------------------------------
mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/terraform-provider-alicloud
	tar czvf bin/terraform-provider-alicloud_darwin-amd64.tgz bin/terraform-provider-alicloud
	rm -rf bin/terraform-provider-alicloud

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/terraform-provider-alicloud.exe
	tar czvf bin/terraform-provider-alicloud_windows-amd64.tgz bin/terraform-provider-alicloud.exe
	rm -rf bin/terraform-provider-alicloud.exe

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/terraform-provider-alicloud
	tar czvf bin/terraform-provider-alicloud_linux-amd64.tgz bin/terraform-provider-alicloud
	rm -rf bin/terraform-provider-alicloud

macarm:
	GOOS=darwin GOARCH=arm64 go build -o bin/terraform-provider-alicloud_v1.0.0
	cp bin/terraform-provider-alicloud_v1.0.0 ~/.terraform.d/plugins/registry.terraform.io/aliyun/alicloud/1.0.0/darwin_arm64/
	mv bin/terraform-provider-alicloud_v1.0.0 ~/.terraform.d/plugins/registry.terraform.io/hashicorp/alicloud/1.0.0/darwin_arm64/

alpha:
	GOOS=linux GOARCH=amd64 go build -o bin/$(RELEASE_ALPHA_NAME)
	aliyun oss cp bin/$(RELEASE_ALPHA_NAME) oss://iac-service-terraform/terraform/alphaplugins/registry.terraform.io/aliyun/alicloud/$(RELEASE_ALPHA_VERSION)/linux_amd64/$(RELEASE_ALPHA_NAME)  --profile terraformer --region cn-hangzhou
	#aliyun oss cp bin/$(RELEASE_ALPHA_NAME) oss://iac-service-terraform/terraform/alphaplugins/registry.terraform.io/hashicorp/alicloud/$(RELEASE_ALPHA_VERSION)/linux_amd64/$(RELEASE_ALPHA_NAME)  --profile terraformer --region cn-hangzhou
	rm -rf bin/$(RELEASE_ALPHA_NAME)

#-----------------------------------
# 代码质量工具
#-----------------------------------
fmt:
	gofmt -w $(GOFMT_FILES)
	goimports -w $(GOFMT_FILES)

fmtcheck:
	"$(CURDIR)/scripts/gofmtcheck.sh"

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v scripts | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

importscheck:
	"$(CURDIR)/scripts/goimportscheck.sh"

#-----------------------------------
# 测试
#-----------------------------------
test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

#-----------------------------------
# 文档
#-----------------------------------
website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), getting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), getting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	ln -sf ../../../../ext/providers/alicloud/website/docs $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/docs/providers/alicloud
	ln -sf ../../../ext/providers/alicloud/website/alicloud.erb $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/layouts/alicloud.erb
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile website website-test all mac windows linux dev devlinux devwin copy linuxcopy windowscopy clean alpha macarm help
