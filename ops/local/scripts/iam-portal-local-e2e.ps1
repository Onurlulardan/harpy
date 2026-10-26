<#
.SYNOPSIS
Runs the local IAM portal-user happy path without printing credentials or tokens.
.DESCRIPTION
Prerequisites: the API, IAM lifecycle worker, Forms runtime, PostgreSQL, RabbitMQ,
and Mailpit are healthy; the Forms subject exists; the operator is initialized.
Protected route contract:
  POST /api/auth/portal-users                         iam.users:create
  GET  /api/iam/portal-setups/{binding_id}            iam.users:read
  POST /api/iam/portal-setups/{binding_id}/token      iam.users:update
  POST /api/iam/portal-setups/{binding_id}/decision   iam.users:update
Create, token, and decision require Idempotency-Key. Decision also requires the
exact If-Match revision. Protected calls carry Bearer plus exact X-Tenant-ID.
Both login calls carry the exact allowed Origin. Setup completion is public and
accepts the one-time credential only in its JSON body.
#>
[CmdletBinding()]
param(
    [ValidatePattern('^https?://')][string]$ApiBase = 'http://localhost:8000',
    [ValidatePattern('^https?://')][string]$MailpitBase = 'http://localhost:8026',
    [ValidatePattern('^https?://')][string]$Origin = 'http://localhost:5173',
    [Parameter(Mandatory)][ValidatePattern('^[a-z0-9]+(?:[-_][a-z0-9]+)*$')]
    [string]$TenantSlug,
    [Parameter(Mandatory)][ValidateNotNullOrEmpty()][string]$OperatorEmail,
    [Parameter(Mandatory)][ValidatePattern('^[a-z0-9]+(?:[-_][a-z0-9]+)*$')]
    [string]$SubjectType,
    [Parameter(Mandatory)][guid]$SubjectID,
    [Parameter(Mandatory)][ValidatePattern('^[a-z0-9]+(?:[-_][a-z0-9]+)*$')]
    [string]$ProfileFormSlug,
    [Parameter(Mandatory)][ValidateNotNullOrEmpty()][string]$PortalIdentifier
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$ApiBase = $ApiBase.TrimEnd('/')
$MailpitBase = $MailpitBase.TrimEnd('/')
$runID = [guid]::NewGuid().ToString('N')
function Get-PlainText {
    param([Parameter(Mandatory)][Security.SecureString]$Value)
    $pointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($Value)
    try {
        return [Runtime.InteropServices.Marshal]::PtrToStringBSTR($pointer)
    }
    finally {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($pointer)
    }
}
function Invoke-JsonPost {
    param(
        [Parameter(Mandatory)][string]$Uri, [Parameter(Mandatory)][hashtable]$Headers,
        [Parameter(Mandatory)][hashtable]$Document
    )
    $wire = $Document | ConvertTo-Json -Compress
    $Document.Clear()
    try {
        return Invoke-RestMethod -Method Post -Uri $Uri -Headers $Headers `
            -ContentType 'application/json' -Body $wire
    }
    finally {
        $wire = $null
    }
}
function Get-PortalSetupState {
    param([string]$BindingID, [hashtable]$Headers)
    return Invoke-RestMethod -Method Get -Headers $Headers `
        -Uri "$ApiBase/api/iam/portal-setups/$BindingID"
}
function Get-PortalSetupCredential {
    param([string]$Query, [string]$ExpectedTenant, [hashtable]$Excluded)
    $pattern = "https?://[^\s<`"]+/portal/setup/" +
        "(?<token>[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43})" +
        "\?tenant=(?<tenant>[A-Za-z0-9._~-]+)"
    for ($attempt = 0; $attempt -lt 60; $attempt++) {
        $search = Invoke-RestMethod -Method Get -Uri "$MailpitBase/api/v1/search?query=$Query"
        foreach ($summary in @($search.messages)) {
            if ($Excluded.ContainsKey([string]$summary.ID)) { continue }
            $messageID = [uri]::EscapeDataString([string]$summary.ID)
            $message = Invoke-RestMethod -Method Get -Uri "$MailpitBase/api/v1/message/$messageID"
            $matched = [regex]::Match([string]$message.Text, $pattern)
            if ($matched.Success -and
                [uri]::UnescapeDataString($matched.Groups['tenant'].Value) -eq
                    $ExpectedTenant) {
                $credential = $matched.Groups['token'].Value
                $message = $null
                $matched = $null
                return $credential
            }
            $message = $null
        }
        Start-Sleep -Seconds 2
    }
    throw 'Mailpit did not receive the portal setup message within 120 seconds.'
}
$operatorPassword = Read-Host 'Operator password' -AsSecureString
$portalPassword = $null
$operatorToken = $null
$portalToken = $null
$setupCredential = $null
try {
    Write-Host '1/8 Operator session is being established.'
    $plain = Get-PlainText $operatorPassword
    try {
        $operatorGrant = Invoke-JsonPost "$ApiBase/api/auth/login" `
            @{ Origin = $Origin } `
            @{ tenant_slug = $TenantSlug; email = $OperatorEmail; password = $plain }
    }
    finally {
        $plain = $null
    }
    $operatorToken = [string]$operatorGrant.token
    $operatorGrant = $null
    if ([string]::IsNullOrWhiteSpace($operatorToken)) {
        throw 'Operator login returned no access token.'
    }
    $operatorAuth = @{ Authorization = "Bearer $operatorToken" }
    $operatorSession = Invoke-RestMethod -Method Get -Uri "$ApiBase/api/auth/session" `
        -Headers $operatorAuth
    if ($operatorSession.access_mode -ne 'operator_console' -or
        $operatorSession.tenant_slug -ne $TenantSlug) {
        throw 'Operator session tenant or access mode does not match.'
    }
    $tenantID = [string]$operatorSession.tenant_id
    $operatorHeaders = @{ Authorization = "Bearer $operatorToken"; 'X-Tenant-ID' = $tenantID }
    Write-Host '2/8 Portal user is being created.'
    $createHeaders = $operatorHeaders.Clone()
    $createHeaders['Idempotency-Key'] = "portal-e2e-create-$runID"
    $created = Invoke-JsonPost "$ApiBase/api/auth/portal-users" $createHeaders `
        @{ subject_type = $SubjectType; subject_id = $SubjectID.ToString()
            portal_profile_form_slug = $ProfileFormSlug }
    $bindingID = [string]$created.data.portal_binding_id
    if ([string]::IsNullOrWhiteSpace($bindingID) -or
        -not [bool]$created.data.portal_setup_required) {
        throw 'Created portal user has no required setup binding.'
    }
    Write-Host "3/8 Setup readback is valid for binding $bindingID."
    $setup = Get-PortalSetupState $bindingID $operatorHeaders
    if ($setup.setup_state -ne 'required') {
        throw "Expected required setup state; received $($setup.setup_state)."
    }
    Write-Host '4/8 One-time setup delivery is being requested.'
    $mailQuery = [uri]::EscapeDataString("to:`"$($created.data.email)`" subject:`"Complete your portal setup`"")
    $existingMail = Invoke-RestMethod -Method Get -Uri "$MailpitBase/api/v1/search?query=$mailQuery"
    $excludedMail = @{}
    foreach ($mail in @($existingMail.messages)) { $excludedMail[[string]$mail.ID] = $true }
    $tokenHeaders = $operatorHeaders.Clone()
    $tokenHeaders['Idempotency-Key'] = "portal-e2e-token-$runID"
    $null = Invoke-JsonPost `
        "$ApiBase/api/iam/portal-setups/$bindingID/token" $tokenHeaders @{}
    $setupCredential = Get-PortalSetupCredential $mailQuery $TenantSlug $excludedMail
    Write-Host '5/8 Portal setup is being completed.'
    $portalPassword = Read-Host 'New portal password' -AsSecureString
    $plain = Get-PlainText $portalPassword
    try {
        $null = Invoke-JsonPost "$ApiBase/api/auth/portal/setup/complete" @{} `
            @{ tenant_slug = $TenantSlug; token = $setupCredential; password = $plain }
    }
    finally {
        $plain = $null
        $setupCredential = $null
    }
    Write-Host '6/8 Approval state is being reconciled.'
    $setup = Get-PortalSetupState $bindingID $operatorHeaders
    if ($setup.approval_state -eq 'pending') {
        $decisionHeaders = $operatorHeaders.Clone()
        $decisionHeaders['Idempotency-Key'] = "portal-e2e-decision-$runID"
        $decisionHeaders['If-Match'] = "`"revision-$($setup.revision)`""
        $null = Invoke-JsonPost "$ApiBase/api/iam/portal-setups/$bindingID/decision" `
            $decisionHeaders @{ decision = 'approve' }
        $setup = Get-PortalSetupState $bindingID $operatorHeaders
    }
    if ($setup.setup_state -ne 'completed' -or
        $setup.approval_state -notin @('approved', 'not_required')) {
        throw "Portal setup is not usable: setup=$($setup.setup_state), " +
            "approval=$($setup.approval_state)."
    }
    Write-Host '7/8 Portal login is being verified.'
    $plain = Get-PlainText $portalPassword
    try {
        $portalGrant = Invoke-JsonPost "$ApiBase/api/auth/portal/login" @{ Origin = $Origin } `
            @{ tenant_slug = $TenantSlug; identifier = $PortalIdentifier; password = $plain }
    }
    finally {
        $plain = $null
    }
    $portalToken = [string]$portalGrant.token
    $portalGrant = $null
    $portalSession = Invoke-RestMethod -Method Get -Uri "$ApiBase/api/auth/session" `
        -Headers @{ Authorization = "Bearer $portalToken"; 'X-Tenant-ID' = $tenantID }
    if ($portalSession.access_mode -ne 'portal' -or
        [string]$portalSession.portal_binding_id -ne $bindingID) {
        throw 'Portal session access mode or binding does not match.'
    }
    Write-Host '8/8 PASS: create, delivery, completion, approval, and portal login passed.'
}
finally {
    $operatorToken = $null
    $portalToken = $null
    $setupCredential = $null
    if ($null -ne $operatorPassword) { $operatorPassword.Dispose() }
    if ($null -ne $portalPassword) { $portalPassword.Dispose() }
}
