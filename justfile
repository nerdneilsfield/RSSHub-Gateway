projectname := "rsshub-gateway"
version := `git describe --abbrev=0 --tags 2>/dev/null || echo dev`
commit := `git rev-parse --short HEAD 2>/dev/null || echo unknown`
build_time := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-s -w -X main.version=" + version + " -X main.buildTime=" + build_time + " -X main.gitCommit=" + commit

# 列出所有可用的命令
default:
    @just --list

# 构建 Golang 二进制文件
build:
    go build -ldflags '{{ldflags}}' -o {{projectname}} ./

# 安装 Golang 二进制文件
install:
    go install -ldflags '{{ldflags}}'

# 运行应用程序
run:
    go run -ldflags '{{ldflags}}' ./ serve -c config.example.yaml

# 安装构建依赖
bootstrap:
    go mod download

# 运行测试并显示覆盖率
test:
    go test ./...

# 清理环境
clean:
    rm -rf coverage.out dist {{projectname}} {{projectname}}.exe

# 显示测试覆盖率
cover:
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# 格式化 Go 文件
fmt:
    gofmt -w .

# 运行 linter
lint:
    golangci-lint run -c .golangci.yml

# 测试发布
release-test:
    goreleaser release --snapshot --clean --skip-publish

# 运行 pre-commit 钩子（已注释）
# pre-commit:
#     pre-commit run --all-files
