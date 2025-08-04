TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)  
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=alicloud
RELEASE_ALPHA_VERSION=$(VERSION)-alpha$(shell date +'%Y%m%d')
RELEASE_ALPHA_NAME=terraform-provider-alicloud_v$(RELEASE_ALPHA_VERSION)

# 定义颜色和前缀
BLUE := \033[1;34m
GREEN := \033[1;32m
YELLOW := \033[1;33m
RED := \033[1;31m
NC := \033[0m# No Color
PREFIX_INFO := [$(BLUE)INFO$(NC)]
PREFIX_SUCCESS := [$(GREEN)SUCCESS$(NC)]
PREFIX_WARNING := [$(YELLOW)WARNING$(NC)]
PREFIX_ERROR := [$(RED)ERROR$(NC)]
PREFIX_BUILD := [$(BLUE)BUILD$(NC)]
PREFIX_FMT := [$(YELLOW)FORMAT$(NC)]

# 静默执行标志
Q := @

# 默认目标: 自动检测操作系统并构建
.PHONY: default auto-build help
default: auto-build

# 自动检测操作系统类型并构建对应平台的二进制文件
auto-build: fmt
	$(Q)echo "$(PREFIX_INFO) 检测操作系统类型..."
	$(Q)OS=$$(uname -s); \
	case "$$OS" in \
		Darwin*) \
			echo "$(PREFIX_INFO) 检测到 macOS，构建 Darwin 版本"; \
			$(MAKE) -s mac; \
			;; \
		Linux*) \
			echo "$(PREFIX_INFO) 检测到 Linux，构建 Linux 版本"; \
			$(MAKE) -s linux; \
			;; \
		CYGWIN*|MINGW*|MSYS*) \
			echo "$(PREFIX_INFO) 检测到 Windows，构建 Windows 版本"; \
			$(MAKE) -s windows; \
			;; \
		*) \
			echo "$(PREFIX_WARNING) 未知操作系统: $$OS，构建所有平台版本"; \
			$(MAKE) -s all; \
			;; \
	esac
	$(Q)echo "$(PREFIX_SUCCESS) 构建完成"

# 测试操作系统检测逻辑（不实际编译）
test-os-detection:
	$(Q)echo "$(PREFIX_INFO) 检测操作系统类型..."
	$(Q)OS=$$(uname -s); \
	echo "$(PREFIX_INFO) 检测到的操作系统: $$OS"; \
	case "$$OS" in \
		Darwin*) \
			echo "$(PREFIX_SUCCESS) 将执行: make mac"; \
			;; \
		Linux*) \
			echo "$(PREFIX_SUCCESS) 将执行: make linux"; \
			;; \
		CYGWIN*|MINGW*|MSYS*) \
			echo "$(PREFIX_SUCCESS) 将执行: make windows"; \
			;; \
		*) \
			echo "$(PREFIX_WARNING) 未知操作系统，将执行: make all"; \
			;; \
	esac

help:
	@echo "Terraform Provider Alibaba Cloud 构建工具"
	@echo ""
	@echo "使用方法:"
	@echo "  make [目标]"
	@echo ""
	@echo "目标:"
	@echo "  help              显示此帮助信息"
	@echo "  default           默认目标，自动检测操作系统并构建对应平台版本"
	@echo "  auto-build        自动检测操作系统并构建对应平台版本"
	@echo "  test-os-detection 测试操作系统检测逻辑（不实际编译）"
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

#-----------------------------------
# 主构建目标
#-----------------------------------
build: fmtcheck 
	$(Q)echo "$(PREFIX_BUILD) 开始构建所有平台版本"
	$(Q)$(MAKE) -s all
	$(Q)echo "$(PREFIX_SUCCESS) 所有平台构建完成"

all: mac windows linux

#-----------------------------------
# 开发构建目标
#-----------------------------------
dev: clean 
	$(Q)echo "$(PREFIX_BUILD) 开始 macOS 开发构建"
	$(Q)$(MAKE) -s mac copy
	$(Q)echo "$(PREFIX_SUCCESS) macOS 开发构建完成"

devlinux: clean fmt 
	$(Q)echo "$(PREFIX_BUILD) 开始 Linux 开发构建"
	$(Q)$(MAKE) -s linux linuxcopy
	$(Q)echo "$(PREFIX_SUCCESS) Linux 开发构建完成"

devwin: clean fmt 
	$(Q)echo "$(PREFIX_BUILD) 开始 Windows 开发构建"
	$(Q)$(MAKE) -s windows windowscopy
	$(Q)echo "$(PREFIX_SUCCESS) Windows 开发构建完成"

copy:
	$(Q)echo "$(PREFIX_INFO) 安装 macOS 二进制文件到 Terraform 目录"
	$(Q)tar -xf bin/terraform-provider-alicloud_darwin-amd64.tgz 2>/dev/null || true
	$(Q)mv bin/terraform-provider-alicloud $(shell dirname `which terraform`) 2>/dev/null || true

linuxcopy:
	$(Q)echo "$(PREFIX_INFO) 安装 Linux 二进制文件到 Terraform 目录"
	$(Q)tar -xf bin/terraform-provider-alicloud_linux-amd64.tgz 2>/dev/null || true
	$(Q)mv bin/terraform-provider-alicloud $(shell dirname `which terraform`) 2>/dev/null || true

windowscopy:
	$(Q)echo "$(PREFIX_INFO) 安装 Windows 二进制文件到 Terraform 目录"
	$(Q)tar -xf bin/terraform-provider-alicloud_windows-amd64.tgz 2>/dev/null || true
	$(Q)mv bin/terraform-provider-alicloud $(shell dirname `which terraform`) 2>/dev/null || true

clean:
	$(Q)echo "$(PREFIX_INFO) 清理 bin 目录"
	$(Q)rm -rf bin/*
	$(Q)echo "$(PREFIX_SUCCESS) 清理完成"

#-----------------------------------
# 平台特定构建
#-----------------------------------
mac:
	$(Q)echo "$(PREFIX_BUILD) 构建 macOS (amd64) 版本..."
	$(Q)if GOOS=darwin GOARCH=amd64 go build -o bin/terraform-provider-alicloud; then \
		echo "$(PREFIX_SUCCESS) macOS 版本构建成功"; \
	else \
		echo "$(PREFIX_ERROR) macOS 版本构建失败"; \
		exit 1; \
	fi

windows:
	$(Q)echo "$(PREFIX_BUILD) 构建 Windows (amd64) 版本..."
	$(Q)if GOOS=windows GOARCH=amd64 go build -o bin/terraform-provider-alicloud.exe; then \
		echo "$(PREFIX_SUCCESS) Windows 版本构建成功"; \
	else \
		echo "$(PREFIX_ERROR) Windows 版本构建失败"; \
		exit 1; \
	fi

linux:
	$(Q)echo "$(PREFIX_BUILD) 构建 Linux (amd64) 版本..."
	$(Q)if GOOS=linux GOARCH=amd64 go build -o bin/terraform-provider-alicloud; then \
		echo "$(PREFIX_SUCCESS) Linux 版本构建成功"; \
	else \
		echo "$(PREFIX_ERROR) Linux 版本构建失败"; \
		exit 1; \
	fi

macarm:
	$(Q)echo "$(PREFIX_BUILD) 构建 macOS ARM64 版本并安装到插件目录..."
	$(Q)if GOOS=darwin GOARCH=arm64 go build -o bin/terraform-provider-alicloud_v1.0.0; then \
		cp bin/terraform-provider-alicloud_v1.0.0 ~/.terraform.d/plugins/registry.terraform.io/aliyun/alicloud/1.0.0/darwin_arm64/ 2>/dev/null || true; \
		mv bin/terraform-provider-alicloud_v1.0.0 ~/.terraform.d/plugins/registry.terraform.io/hashicorp/alicloud/1.0.0/darwin_arm64/ 2>/dev/null || true; \
		echo "$(PREFIX_SUCCESS) macOS ARM64 版本构建并安装成功"; \
	else \
		echo "$(PREFIX_ERROR) macOS ARM64 版本构建失败"; \
		exit 1; \
	fi

alpha:
	$(Q)echo "$(PREFIX_BUILD) 构建 Alpha 版本并上传到 OSS..."
	$(Q)if GOOS=linux GOARCH=amd64 go build -o bin/$(RELEASE_ALPHA_NAME); then \
		echo "$(PREFIX_INFO) 上传到阿里云 OSS..."; \
		if aliyun oss cp bin/$(RELEASE_ALPHA_NAME) oss://iac-service-terraform/terraform/alphaplugins/registry.terraform.io/aliyun/alicloud/$(RELEASE_ALPHA_VERSION)/linux_amd64/$(RELEASE_ALPHA_NAME) --profile terraformer --region cn-hangzhou; then \
			echo "$(PREFIX_SUCCESS) Alpha 版本构建并上传成功"; \
		else \
			echo "$(PREFIX_ERROR) OSS 上传失败"; \
			exit 1; \
		fi; \
		rm -rf bin/$(RELEASE_ALPHA_NAME); \
	else \
		echo "$(PREFIX_ERROR) Alpha 版本构建失败"; \
		exit 1; \
	fi \
	else \
		echo '$(PREFIX_ERROR) Alpha 版本构建失败'; \
		exit 1; \
	fi"

#-----------------------------------
# 代码质量工具
#-----------------------------------
fmt:
	$(Q)echo "$(PREFIX_FMT) 格式化 Go 代码..."
	$(Q)gofmt -w $(GOFMT_FILES)
	$(Q)goimports -w $(GOFMT_FILES)
	$(Q)echo "$(PREFIX_SUCCESS) 代码格式化完成"

fmtcheck:
	$(Q)echo "$(PREFIX_FMT) 检查代码格式..."
	$(Q)if "$(CURDIR)/scripts/gofmtcheck.sh"; then \
		echo "$(PREFIX_SUCCESS) 代码格式检查通过"; \
	else \
		echo "$(PREFIX_ERROR) 代码格式检查失败"; \
		exit 1; \
	fi

vet:
	$(Q)echo "$(PREFIX_INFO) 执行 go vet 静态分析..."
	$(Q)if go vet $$(go list ./... | grep -v scripts | grep -v vendor/); then \
		echo "$(PREFIX_SUCCESS) go vet 检查通过"; \
	else \
		echo "$(PREFIX_ERROR) go vet 发现问题，请检查并修复"; \
		exit 1; \
	fi

errcheck:
	$(Q)echo "$(PREFIX_INFO) 检查错误处理..."
	$(Q)if sh -c "$(CURDIR)/scripts/errcheck.sh"; then \
		echo "$(PREFIX_SUCCESS) 错误处理检查通过"; \
	else \
		echo "$(PREFIX_ERROR) 错误处理检查失败"; \
		exit 1; \
	fi

importscheck:
	$(Q)echo "$(PREFIX_FMT) 检查 imports 格式..."
	$(Q)if "$(CURDIR)/scripts/goimportscheck.sh"; then \
		echo "$(PREFIX_SUCCESS) imports 格式检查通过"; \
	else \
		echo "$(PREFIX_ERROR) imports 格式检查失败"; \
		exit 1; \
	fi

#-----------------------------------
# 测试
#-----------------------------------
test: fmtcheck
	$(Q)echo "$(PREFIX_INFO) 运行单元测试..."
	$(Q)bash -c "set -o pipefail; if go test $(TEST) -timeout=30s -parallel=4 2>&1 | sed 's/^/$(PREFIX_INFO) test: /'; then \
		echo '$(PREFIX_SUCCESS) 单元测试通过'; \
	else \
		echo '$(PREFIX_ERROR) 单元测试失败'; \
		exit 1; \
	fi"

testacc: fmtcheck
	$(Q)echo "$(PREFIX_INFO) 运行验收测试..."
	$(Q)bash -c "set -o pipefail; if TF_ACC=1 go test $(TEST) -v \$(TESTARGS) -timeout 120m 2>&1 | sed 's/^/$(PREFIX_INFO) testacc: /'; then \
		echo '$(PREFIX_SUCCESS) 验收测试通过'; \
	else \
		echo '$(PREFIX_ERROR) 验收测试失败'; \
		exit 1; \
	fi"

test-compile:
	$(Q)if [ "$(TEST)" = "./..." ]; then \
		echo "$(PREFIX_ERROR) 请设置 TEST 为具体的包，例如:"; \
		echo "$(PREFIX_INFO)   make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	$(Q)echo "$(PREFIX_INFO) 编译测试二进制文件..."
	$(Q)bash -c "set -o pipefail; if go test -c $(TEST) \$(TESTARGS) 2>&1 | sed 's/^/$(PREFIX_INFO) test-compile: /'; then \
		echo '$(PREFIX_SUCCESS) 测试编译完成'; \
	else \
		echo '$(PREFIX_ERROR) 测试编译失败'; \
		exit 1; \
	fi"

#-----------------------------------
# 文档
#-----------------------------------
website:
	$(Q)echo "$(PREFIX_INFO) 构建网站文档..."
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	$(Q)echo "$(PREFIX_INFO) 获取网站仓库..."
	$(Q)git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	$(Q)$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)
	$(Q)echo "$(PREFIX_SUCCESS) 网站文档构建完成"

website-test:
	$(Q)echo "$(PREFIX_INFO) 测试网站文档..."
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	$(Q)echo "$(PREFIX_INFO) 获取网站仓库..."
	$(Q)git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	$(Q)ln -sf ../../../../ext/providers/alicloud/website/docs $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/docs/providers/alicloud
	$(Q)ln -sf ../../../ext/providers/alicloud/website/alicloud.erb $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/layouts/alicloud.erb
	$(Q)$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)
	$(Q)echo "$(PREFIX_SUCCESS) 网站文档测试完成"

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile website website-test all mac windows linux dev devlinux devwin copy linuxcopy windowscopy clean alpha macarm help default auto-build test-os-detection
