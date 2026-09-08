# 修复 goctl 生成司机端代码时丢失的 __proto 包别名。
#
# 背景：goctl 1.9.2 生成 driversvc 时会把 import 写成
#         "XiaoLong-Ridy/rpc/driversvc/proto"   （无别名）
#       但生成的文件内部仍然用 __proto.XXX 引用类型，导致编译报
#         undefined: __proto
#
# 本脚本在生成后把该 import 补回 __proto 别名；已带别名时不做任何改动（幂等）。

$ErrorActionPreference = 'Stop'

$files = @(
  'rpc/driversvc/driverservice/driver_service.go',
  'rpc/driversvc/internal/server/driver_service_server.go'
)

$fixed = 0
foreach ($rel in $files) {
  $path = Join-Path (Get-Location) $rel
  if (-not (Test-Path $path)) {
    Write-Output "skip (missing): $rel"
    continue
  }

  # 必须显式按 UTF8 读写：否则中文注释会被按系统编码(GBK)误解码成乱码，并破坏换行。
  $content = [System.IO.File]::ReadAllText($path, [System.Text.Encoding]::UTF8)
  if ($content -match '__proto "XiaoLong-Ridy/rpc/driversvc/proto"') {
    Write-Output "already ok : $rel"
    continue
  }

  $updated = $content -replace '(?m)^(\s*)"XiaoLong-Ridy/rpc/driversvc/proto"\s*$', '$1__proto "XiaoLong-Ridy/rpc/driversvc/proto"'
  if ($updated -ne $content) {
    [System.IO.File]::WriteAllText($path, $updated, (New-Object System.Text.UTF8Encoding($false)))
    Write-Output "fixed      : $rel"
    $fixed += 1
  } else {
    Write-Output "no match   : $rel"
  }
}

Write-Output "driversvc proto alias fix done (fixed=$fixed)"
