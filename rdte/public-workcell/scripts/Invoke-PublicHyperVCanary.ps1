#requires -Version 5.1
<#
.SYNOPSIS
  Opt-in synthetic Hyper-V disk/VM create-observe-remove canary on a
  disposable PUBLIC GitHub Actions Windows runner only. NO guest boot,
  no network/switch, no host software install, no owner workstation effects.
.NOTES
  Not a local installer or authorization gate. Does not prove nested
  virtualization is supported, and never touches preexisting resources.
#>
[CmdletBinding()]
param()
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if ($env:GITHUB_ACTIONS -ne 'true' -or [string]::IsNullOrEmpty($env:RUNNER_TEMP)) {
    throw 'HOST_NOT_DISPOSABLE_GHA: live Hyper-V canary forbidden'
}
$baseTemp = [System.IO.Path]::GetFullPath($env:RUNNER_TEMP)
if (-not (Test-Path -LiteralPath $baseTemp -PathType Container)) {
    throw 'RUNNER_TEMP unavailable'
}
$id = [guid]::NewGuid().ToString('N')
$root = Join-Path $baseTemp ('public-rdte-hv-' + $id)
$vmName = 'public-rdte-' + $id
$ownerMarker = Join-Path $root '.rdte-synthetic-owner'
$baseVhd = Join-Path $root 'parent.vhdx'
$diffVhd = Join-Path $root 'child.vhdx'
$vm = $null
$status = 'NOT_RUN'
$phase = 'preflight'
$errorClass = ''
$cleanupOkay = $false
$guestBooted = $false
$existingNetworkTouched = $false
$createdDirectory = $false
try {
    foreach ($cmdName in @('New-VHD','Get-VHD','New-VM','Get-VM','Get-VMHardDiskDrive','Remove-VM')) {
        if (-not (Get-Command -Name $cmdName -ErrorAction SilentlyContinue)) {
            throw [System.NotSupportedException]::new('Required Hyper-V command unavailable')
        }
    }
    # Owner-generated globally unique name; mkdir MUST fail if it exists.
    $null = New-Item -Path $root -ItemType Directory -ErrorAction Stop
    $createdDirectory = $true
    [System.IO.File]::WriteAllText($ownerMarker, 'PUBLIC_SYNTHETIC_HV_V1')
    $phase = 'parent'
    $null = New-VHD -Path $baseVhd -SizeBytes 1GB -Dynamic -ErrorAction Stop
    $phase = 'child'
    $null = New-VHD -Path $diffVhd -ParentPath $baseVhd -Differencing -ErrorAction Stop
    $phase = 'vm'
    # Intentionally no vSwitch, network adapter, ISO, boot or credentials.
    $vm = New-VM -Name $vmName -MemoryStartupBytes 256MB -Generation 2 -VHDPath $diffVhd -Path $root -ErrorAction Stop
    $phase = 'verify'
    $v = Get-VM -Id $vm.Id -ErrorAction Stop
    if ($v.Id -ne $vm.Id -or $v.State -ne 'Off') { throw 'Unowned or running VM' }
    $drive = @(Get-VMHardDiskDrive -VM $v -ErrorAction Stop)
    if ($drive.Count -ne 1 -or $drive[0].Path -ne $diffVhd) {
        throw 'Synthetic disk not attached'
    }
    $d = Get-VHD -Path $diffVhd -ErrorAction Stop
    if ([System.IO.Path]::GetFullPath($d.ParentPath) -ne [System.IO.Path]::GetFullPath($baseVhd)) {
        throw 'Differencing parent mismatch'
    }
    $status = 'PASS_DISK_AND_VM_CONTROL_ONLY'
} catch {
    # Do not expose host paths, VM identifiers or exception messages.
    $status = 'BLOCKED_HYPERV_CAPABILITY'
    $errorClass = $_.Exception.GetType().Name
} finally {
    $phase = 'cleanup'
    try {
        if ($null -ne $vm) {
            $live = Get-VM -Id $vm.Id -ErrorAction SilentlyContinue
            if ($null -ne $live) {
                if ($live.Id -ne $vm.Id) { throw 'VM identity mismatch before removal' }
                if ($live.State -ne 'Off') { throw 'Unintended running VM; no forced delete' }
                Remove-VM -VM $live -Force -ErrorAction Stop
            }
            if ($null -ne (Get-VM -Id $vm.Id -ErrorAction SilentlyContinue)) {
                throw 'Owned VM remains after destroy'
            }
        }
        if ($createdDirectory) {
            if (-not (Test-Path -LiteralPath $ownerMarker -PathType Leaf)) {
                throw 'Missing ownership marker, preserve root'
            }
            $tag = [System.IO.File]::ReadAllText($ownerMarker)
            if ($tag -ne 'PUBLIC_SYNTHETIC_HV_V1') {
                throw 'Wrong marker, preserve root'
            }
            # Remove only known synthetic files and New-VM-created owned folder.
            # Every path remains under the uniquely owned RUNNER_TEMP directory.
            $null = Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction Stop
            if (Test-Path -LiteralPath $root) {
                throw 'Synthetic workcell disk directory remains'
            }
        }
        $cleanupOkay = $true
    } catch {
        $cleanupOkay = $false
    }
}
$result = [ordered]@{
    schema = 1
    mode = 'PUBLIC_GITHUB_ACTIONS_HYPERV_SYNTHETIC'
    status = $status
    phase = $phase
    cleanup_verified = $cleanupOkay
    guest_boot_tested = $guestBooted
    sovereign_network_tested = $existingNetworkTouched
    no_real_service_deployed = $true
    error_class = $errorClass
}
$result | ConvertTo-Json -Compress
if (-not $cleanupOkay) {
    throw 'OWNED_SYNTHETIC_HYPERV_CLEANUP_UNVERIFIED'
}
