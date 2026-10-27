---
name: iam-system-resources
description: Build backend-v2 identity, authentication, sessions, RBAC, users, roles, organizations, and their Form Engine system-resource facade without duplicate sources of truth. Use for any IAM or protected system-form behavior.
---

# IAM System Resources

Storage, password, access/refresh token ve directory kontrati icin once
ADR-0009'u; login/profile/step-up icin ADR-0023'u oku.
Instance authority veya environment first-admin onboarding icin ADR-0117 ve
ADR-0118'i; bootstrap delivery scope icin ADR-0122'yi; recovery owner authority
icin ADR-0124'u; post-restore safety icin ADR-0125'i; scoped external authority,
ack trust ve signer lifecycle'i icin ADR-0126/ADR-0127'yi;
SaaS customer first-admin setup icin ADR-0119'u oku.

## Temel karar

IAM bir paralel tablo kopyasi degil, business ownership siniridir. Users, roles
ve organizations icin tek canonical record bulunur. Form Engine bu kaynaklari
form gibi sunar fakat command/query davranisini IAM'e delege eder.

## Alan siniflari

```text
CORE: email, password hash, status, account type, token version
RELATION: roles, organization, owner, parent
CUSTOM: tenant tarafindan eklenen guvenli profil alanlari
```

- CORE ve RELATION manifest'ini backend kodunda tanimla.
- Bu alanlarin silinmesini, type veya security semantics degisimini reddet.
- CUSTOM alana auth, permission veya tenant identity anlami verme.
- Password, token, session secret ve MFA secret'ini dynamic response'a koyma.

## Tek yazma kapisi

Tum girisleri IAM application command'larina bagla:

```text
login/register/invitation/portal/form CRUD/MCP
  -> IAM command/query
  -> one transaction
  -> canonical record + audit + projection intent
```

Form hook, auth handler veya seed tarafinda ayni business kuralini kopyalama.
Custom-field mutation'i system-resource orchestrator disindan acma. Direct table
mutation'i yalniz controlled migration/seed adapter'inda kullan.

## Zorunlu use-case kurallari

- Email normalize et, tenant scope'unda unique uygula.
- Password policy, Argon2id ve finite bcrypt migration'ini tek credential
  service'te tut.
- Default role atamasini ayni transaction'da yap.
- Role silmeden assignment ve protected-role bagimliligini kontrol et.
- Organization owner, parent, cycle ve child/member bagimliligini koru.
- Protected super-admin identity'yi generic CRUD'dan fail-closed koru.
- Role/permission degisiminde token veya authorization version stratejisini uygula.

## Form facade

- Mevcut dynamic users/roles/organizations wire shape'ini presenter ile uret.
- Layout, label ve custom alan metadata'sini Forms tarafinda tut.
- Create/update/delete'i generic record SQL'e dusurme.
- IAM query tam record veya typed summary projection sunabilsin.
- Consumer'in tabloya girmek yerine `GetActor`, `GetUserSummary`,
  `ResolvePermissions` gibi dar query kullanmasini sagla.

## Directory ve projection

Tenant seciminden once email lookup gerekiyorsa platform directory bilincli bir
projection'dir. Ayni PostgreSQL transaction'inda yaz; ayri storage'a cikarsa
outbox + idempotent projector + rebuild/repair yolu kur.

Projection canonical password, token veya permission grant tasimaz.

## Session ve token

- Access token kisa omurlu ve versioned olsun.
- Refresh token opaque, hash'li, rotating ve reuse-detection'li olsun.
- Satellite'lara signing secret verme; verification key stratejisi kullan.
- HS256 veya legacy token verifier'i V2 runtime'a ekleme.
- Header tenant degerini verified token actor'u ile eslestir.
- Logout, password reset ve role revocation davranisini contract test et.
- Restore'da ADR-0125 security epoch, full credential revoke, grant quarantine ve recovery-only anchor'i uygula.

## Verification

- Tum mutation girisleri icin ortak invariant suite.
- Login/refresh/logout/reuse concurrency testleri.
- Cross-tenant identity ve permission negative testleri.
- Protected field, role assignment ve organization cycle testleri.
- Form facade V1 differential testleri.
- Directory failure, repair ve drift detection testi.
