$ErrorActionPreference = 'Stop'

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object Security.Principal.WindowsPrincipal($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw 'task local:cleanup:disk must run in an elevated PowerShell (Run as administrator).'
}

$vhdx = Join-Path $env:LOCALAPPDATA 'Docker\wsl\disk\docker_data.vhdx'
if (-not (Test-Path -LiteralPath $vhdx)) {
    throw "Docker VHDX not found: $vhdx"
}

$cli = Join-Path ${env:ProgramFiles} 'Docker\Docker\DockerCli.exe'
if (Test-Path -LiteralPath $cli) {
    & $cli -Shutdown
}

wsl --shutdown

$script = Join-Path $env:TEMP 'harpy-compact-docker-vhdx.txt'
@(
    "select vdisk file=`"$vhdx`""
    'attach vdisk readonly'
    'compact vdisk'
    'detach vdisk'
) | Set-Content -LiteralPath $script -Encoding ascii
try {
    diskpart /s $script
}
finally {
    Remove-Item -LiteralPath $script -Force -ErrorAction SilentlyContinue
}
