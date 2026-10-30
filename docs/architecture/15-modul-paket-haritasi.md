# 15 — Modül paket haritası

Bu sayfa core/internal/modules altındaki modüllerin source tree'de görülen application ve adapter alt paketlerini listeler. Dizin adları, sorumluluk alanını bulmak için başlangıç noktasıdır; kesin davranış ilgili package source ve testlerinden doğrulanır.

## 1. IAM

Application package groups:
- Identity/access: identity, identifier, authorization, rbac, rbacchange, permissionrequest, permissioncache, profile, profilebackfill, serviceprincipal, portaluser, systemresource.
- Authentication: authentication, accesstoken, refreshtoken, csrf, password, passwordenrollment, initialadminsetup, installationcredential, sessionview, securityepoch.
- Lifecycle/integration: lifecycle, lifecycledelivery, invitation, invitationdelivery, notificationaccess, actorbridgeissue, platformiam, restquery, maintenance, recoveryanchor, recoveryevidence, systemtransfer.

Adapters: cryptography, custodymanifest, gcpkms, passwordblocklist, postgres, smtp, systemtransferartifact.

## 2. Tenancy

Application:
- customerprovisioning ve customerlifecycle
- customerworker
- deploymentpolicy
- onboarding
- installationdelivery
- primarytarget ve recoverytarget
- tenantsettings

Adapter: postgres.

## 3. Billing

Application:
- Commercial catalogue: offering, catalogquery, campaign, discount, discountcatalog
- Subscription and entitlement: subscription, entitlement, resourceentitlement
- Invoicing/provider: invoice, provider
- Platform license: licenseauthority, licenseoffering, licenseclock, platformlicense
- Metering and reporting: usage, platformreports, retention, legalhold, restquery, httpapi

Adapters: cryptography, licenseauthoritysource, postgres, provider, resourceusage.

## 4. Work

Application:
- Board structure: boards, columns, memberships
- Task data: tasks, labels, checklists, comments
- Calendar: calendar, recurrence
- Runtime support: jobs, projections, restquery

Adapters: http, postgres.

## 5. Documents

Application:
- Content lifecycle: folder, item, link, copy, deletion, purge
- File transfer: upload, uploadreconcile, download
- Archive/recovery: archive, archivedownload, recovery
- Additional surfaces: knowledgefile, legalhold, foundation

Adapters: adoptionartifact, archiveartifact, blobcrypto, cipher, clamav, http, keyring, keywrap, postgres, s3blob.

## 6. Notifications

Application code root-level use-case files plus retention and channel/consumer runtime packages bulunur. Adapter grupları:
- http, postgres
- providerregistry
- secretcrypto
- smtp, telegram

Domain package message, delivery, preference, template, provider config ve quiet-hours kurallarını içerir.

## 7. Analytics

Application:
- Query: query, queryexecution
- Dashboard: dashboard, dashboardexport
- Publication/authority: publication, authority
- Support: entitlement, httptransport, retention, compatibility

Adapters: artifact, cache, http, postgres.

## 8. Audit ve Usage

Audit:
- Domain: entry/action/classification/field-change model
- Application: append ve owner operation binding
- Adapters: http, postgres

Usage:
- Domain: normalized observation ve feature mapping
- Application: record, read models/analytics
- Adapters: http, postgres

## 9. Platform Authority ve Platform Ops

Platform Authority application grubu recoveryevidence; adapter'lar airgap, devjournal, gcpkms, localkey, postgres ve providerregistry'dir.

Platform Ops application grupları:
- Diagnostics: alertevent, alertrule, systemhealth, trace, tracespan
- Recovery: recoverycontrol, recoveryexecution, recoverykit, recoverymanifest, recoveryorchestration, recoveryprotection, recoveryset, restoredlicense, restoreexecution, restoreplan

Adapters: backuparchive, pgtools, postgres, recoveryplansigner, recoverys3.

## 10. Support ve System Resource Layout

Support application:
- contact
- platformbug
- reporterbug
- tenantbug

Adapters: http, postgres.

System Resource Layout:
- Public API package ve application entrypoint'leri modül root altında bulunur.
- PostgreSQL adapter persistence owner'dır.
- Presentation compile/publish akışı platform fieldcatalog ile çalışır.

## 11. Dizin haritası nasıl kullanılmalı?

1. Feature'ın owner modülünü belirle.
2. İlgili application alt paketinde mevcut use-case'i bul.
3. Domain modelinin hangi invariant'ı enforce ettiğini oku.
4. Ports ve adapter dosyalarını takip et.
5. HTTP endpoint'i ise composition route builder ve contracts/http dosyalarına bak.
6. Background task ise worker composition ve ilgili job/event contract'ını incele.
7. Testleri aynı package path altında source ile birlikte oku.

Bu harita generated dosya, tüm CRUD dosyaları veya her test dosyası için tek tek dosya listesi değildir; source tree'deki işlevsel package gruplarını gösterir.
