# ============================================================
# start-local.ps1 - Start all local services with LATEST code
#   - 4 gRPC services (driversvc/ordersvc/dispatchsvc/locationsvc)
#   - api/driver (reads api/driver/etc/driver.yaml -> local gRPC)
#   - frontend is expected to run separately on localhost:5175
#
# Usage:
#   powershell -ExecutionPolicy Bypass -File scripts/start-local.ps1
#   powershell -ExecutionPolicy Bypass -File scripts/stop-local.ps1
# ============================================================
param(
    [string]$Root = "D:\gocode\src\XiaoLong-Ridy"
)
$ErrorActionPreference = "Continue"
$logs = Join-Path $Root "logs"
$bin  = Join-Path $Root "bin"
New-Item -ItemType Directory -Force -Path $logs | Out-Null
New-Item -ItemType Directory -Force -Path $bin  | Out-Null

function Build-Service($name, $pkg) {
    $exe = Join-Path $bin "$name.exe"
    Write-Host "[build] $name ..." -NoNewline
    Push-Location $Root
    $out = & go build -o $exe ".\$pkg\" 2>&1
    Pop-Location
    if ($LASTEXITCODE -ne 0) { Write-Host " FAIL"; Write-Host $out; return $null }
    Write-Host " OK -> $exe"
    return $exe
}

function Start-Svc($name, $exe, $yaml, $workdir) {
    $outLog = Join-Path $logs "$name.out.log"
    $errLog = Join-Path $logs "$name.err.log"
    $p = Start-Process -FilePath $exe -ArgumentList "-f", $yaml -WorkingDirectory $workdir `
        -RedirectStandardOutput $outLog -RedirectStandardError $errLog -WindowStyle Hidden -PassThru
    Write-Host "[start] $name pid=$($p.Id) out=$outLog err=$errLog"
    return $p
}

function Wait-Port($port, $name) {
    $ok = $false
    for ($i = 0; $i -lt 30; $i++) {
        $r = Test-NetConnection -ComputerName 127.0.0.1 -Port $port -WarningAction SilentlyContinue
        if ($r.TcpTestSucceeded) { $ok = $true; break }
        Start-Sleep -Milliseconds 1000
    }
    Write-Host ("  {0,-12} {1,6} -> {2}" -f $name, $port, $(if ($ok) { "UP" } else { "DOWN" }))
}

# ---------- 1. build & start 4 gRPC services ----------
$grpc = @(
    @{ Name="driversvc";   Pkg="rpc/driversvc";   Yaml="rpc/driversvc/etc/driversvc.yaml";   Port=50055; WD=$Root },
    @{ Name="ordersvc";    Pkg="rpc/ordersvc";    Yaml="rpc/ordersvc/etc/ordersvc.yaml";     Port=50051; WD=$Root },
    @{ Name="dispatchsvc"; Pkg="rpc/dispatchsvc"; Yaml="rpc/dispatchsvc/etc/dispatchsvc.yaml"; Port=50056; WD=$Root },
    @{ Name="locationsvc"; Pkg="rpc/locationsvc"; Yaml="rpc/locationsvc/etc/locationsvc.yaml"; Port=9001;  WD=$Root }
)
foreach ($s in $grpc) {
    $exe = Build-Service $s.Name $s.Pkg
    if ($exe) { Start-Svc $s.Name $exe (Join-Path $Root $s.Yaml) $s.WD | Out-Null }
}

# ---------- 2. build & start api/driver ----------
$apiExe = Build-Service "api-driver" "api/driver"
if ($apiExe) {
    $apiWd = Join-Path $Root "api\driver"
    Start-Svc "api-driver" $apiExe (Join-Path $apiWd "etc\driver.yaml") $apiWd | Out-Null
}

# ---------- 3. wait for ports ----------
Write-Host ""
Write-Host "Waiting for ports..."
$ports = @{ 50055="driversvc"; 50051="ordersvc"; 50056="dispatchsvc"; 9001="locationsvc"; 18082="api/driver" }
foreach ($kv in $ports.GetEnumerator()) { Wait-Port $kv.Key $kv.Value }
Write-Host ""
Write-Host "Done. Frontend localhost:5175 should now work."
Write-Host "Logs: $logs"
Write-Host "Stop: scripts/stop-local.ps1"
