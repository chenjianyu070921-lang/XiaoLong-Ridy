# 重新生成 rpc/adminsvc 的 protobuf 与 go-zero 包装代码，并清理 goctl 生成的旧 stub 层。
# 使用方式：在项目根目录执行 `powershell -ExecutionPolicy Bypass -File scripts/admin-test/regenerate_adminsvc_proto.ps1`。
# 注意：真实实现只允许保留在 rpc/adminsvc/internal/logic/adminservice，脚本末尾会删除 internal/logic/*.go 旧壳文件。

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Set-Location $root

goctl rpc protoc rpc/adminsvc/admin.proto `
  --go_out=rpc/adminsvc `
  --go-grpc_out=rpc/adminsvc `
  --zrpc_out=rpc/adminsvc `
  --style=go_zero

$logicDir = Resolve-Path "rpc/adminsvc/internal/logic"
Get-ChildItem -LiteralPath $logicDir -File -Filter "*.go" | ForEach-Object {
  Remove-Item -LiteralPath $_.FullName
}

$legacyServers = @(
  "rpc/adminsvc/internal/server/adminServiceServer.go",
  "rpc/adminsvc/internal/server/adminserviceserver.go"
)
foreach ($path in $legacyServers) {
  if (Test-Path -LiteralPath $path) {
    Remove-Item -LiteralPath $path
  }
}

gofmt -w rpc/adminsvc/adminservice/adminService.go rpc/adminsvc/client/adminservice/adminservice.go rpc/adminsvc/internal/server/adminservice/adminserviceserver.go
