# ============================================================
# stop-local.ps1 - Stop all locally started services
# ============================================================
$names = @("driversvc","ordersvc","dispatchsvc","locationsvc","api-driver")
foreach ($n in $names) {
    Get-Process -Name $n -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    Write-Host "[stop] $n"
}
Write-Host "All local services stopped."
