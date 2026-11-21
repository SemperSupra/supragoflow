#requires -Version 5.1
<#
.SYNOPSIS
  Read-only Windows / Hyper-V capability census for a synthetic public RDTE.
.DESCRIPTION
  This script NEVER creates a VM, modifies a switch, installs a tool, reads
  credentials, opens a network connection, or writes an artifact.
  It deliberately emits no host name, username, IP, serial, or absolute path.
#>
[CmdletBinding()]
param(
    [ValidateSet('Observe','Plan')]
    [string]$Stage = 'Observe'
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
function Test-CommandPresent([string]$CommandName) {
    try { return [bool](Get-Command -Name $CommandName -ErrorAction SilentlyContinue) }
    catch { return $false }
}
function Test-HyperVQuery([string]$CommandName) {
    if (-not (Test-CommandPresent $CommandName)) { return 'NOT_AVAILABLE' }
    try {
        # Enumerate only in memory; do not print VM or switch names or IDs.
        if ($CommandName -eq 'Get-VM') {
            $null = @(Get-VM -ErrorAction Stop).Count
        }
        elseif ($CommandName -eq 'Get-VMSwitch') {
            $null = @(Get-VMSwitch -ErrorAction Stop).Count
        }
        return 'READABLE'
    }
    catch [System.UnauthorizedAccessException] {
        return 'DENIED'
    }
    catch {
        return 'DENIED_OR_UNAVAILABLE'
    }
}
$hypervModule = [bool](Get-Module -ListAvailable -Name Hyper-V -ErrorAction SilentlyContinue)
$serviceState = 'NOT_AVAILABLE'
try {
    $service = Get-Service -Name 'vmms' -ErrorAction SilentlyContinue
    if ($null -ne $service) { $serviceState = [string]$service.Status }
} catch { $serviceState = 'UNKNOWN' }
$vmQuery = Test-HyperVQuery 'Get-VM'
$switchQuery = Test-HyperVQuery 'Get-VMSwitch'
$tools = [ordered]@{}
foreach ($toolName in @('ssh','oci','kaggle','colab','gcloud','tailscale','cloudflared','go','git')) {
    $tools[$toolName] = Test-CommandPresent $toolName
}
$canLaunch = ($hypervModule -and $serviceState -eq 'Running' -and $vmQuery -eq 'READABLE' -and $switchQuery -eq 'READABLE')
$nextGate = if ($canLaunch) {
    'PLAN_REVIEW_AND_OWNER_EFFECT_ADMISSION_REQUIRED'
} else {
    'HYPERV_MANAGEMENT_AUTHORIZATION_OR_FEATURE_BLOCKED'
}
$result = [ordered]@{
    schema = 1
    mode = 'READ_ONLY_WINDOWS_PS51'
    stage = $Stage
    no_host_changes = $true
    no_credentials_read = $true
    powershell_major = $PSVersionTable.PSVersion.Major
    hyperv_module_present = $hypervModule
    hyperv_service_state = $serviceState
    vm_inventory_status = $vmQuery
    switch_inventory_status = $switchQuery
    command_presence = $tools
    next_gate = $nextGate
    live_vm_creation_tested = $false
    sovereign_network_tested = $false
}
$result | ConvertTo-Json -Depth 5 -Compress
