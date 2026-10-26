---
name: billing
description: Build and change backend-v2 billing plans, immutable offerings and prices, subscriptions, discounts, invoices, provider webhooks, entitlements, metered usage, limits, and V1 cutover. Use for every billing-owned table, API, event, job, provider, permission, or migration change.
---

# Billing

Para, period, state, entitlement, usage, provider, retention ve cutover
semantigi icin ADR-0019, ADR-0022 ve ADR-0028 authoritative'dir;
uygulamadan once tamamini oku. Endpoint icin `api-contract`, SQL ve tenant icin
`data-tenancy`, provider/job/outbox icin `events-jobs` skill'ini de oku.
Installation/license base offering'i etkileniyorsa ADR-0079, ADR-0082,
ADR-0118..0120, ADR-0122 ve restore/head ayrimi icin ADR-0124..0127'yi oku.

## Sahiplik siniri

- `billing` bir core business module'dur; ayri deployable veya ikinci storage acma.
- Plan katalogu, offering/price version, subscription, discount, invoice/credit,
  provider normalization, entitlement ve usage tablolarinin tek owner'i olsun.
- Diger moduller yalniz `billing/api` command/query kontratlarini kullansin;
  billing tablosuna SQL, JOIN veya adapter importu yazmasin.
- Money, state, entitlement ve usage'i typed PostgreSQL kaydinda tut; JSONB,
  Redis, webhook veya provider'i paralel authority yapma.
- Offering, price, invoice, credit ve applied policy history'sini immutable
  version/snapshot olarak modelle; in-place edit veya silent reprice yapma.

## Use-case dosyalama

Her davranisi application, HTTP ve PostgreSQL adapter katmaninda ayri dosyala:

```text
create_offering.go, get_offering.go, list_offerings.go, retire_offering.go
create_price_version.go, get_price_version.go, list_price_versions.go
create_subscription.go, get_subscription.go, list_subscriptions.go
change_subscription.go, cancel_subscription.go
create_discount.go, redeem_discount.go, revoke_discount.go
get_invoice.go, list_invoices.go, create_adjustment.go
resolve_entitlements.go, reserve_usage.go, release_usage.go, adjust_gauge.go
receive_webhook.go, reconcile_provider.go
```

- Get ve list'i tek `read.go` dosyasinda birlestirme.
- Provider ingress, normalized processing ve reconciliation'i ayri use-case tut.
- Her davranisa ayni operasyon adli test yaz; `routes.go` yalniz route baglasin.
- Dosya/CRUD/isim sinirinda `skills/code-organization/SKILL.md`yi uygula.

## Command ve authorization sirasi

1. Actor, tenant ve permission'i application katmaninda fail-closed dogrula.
2. Subscription/offering version'i, currency/period ve entitlement definition'i
   ADR kontratina gore cozumle; eksigi unlimited veya zero-price sayma.
3. Idempotency receipt'i ve gereken row/unique lock'lari ayni scope'ta al.
4. Hard usage reserve veya gauge delta'sini business mutation ile ayni
   `TenantScope` transaction'inda yaz.
5. State, decision audit, outbox ve gerekiyorsa job'u ayni commit'e kat.
6. Commit basarili olmadan success veya entitlement grant dondurme.

HTTP, MCP, workflow, import ve worker mutation'lari ayni billing application
guard'ini kullanir. Handler status yazmaz; provider network cagrisi, event publish
ve secret erisimi acik DB transaction'i icinde calismaz.

## Provider ve async akisi

- Webhook tenant'ini header veya payload metadata'sindan degil verified provider
  object mapping'inden cozumle.
- Signature/replay/body limitini adapter registry kontratiyla dogrula; durable
  receipt ve processor job'u olmadan accepted response verme.
- Provider command ve reconcile akisini ADR-0011 job/idempotency kurallariyla kur.
- Secret, kart verisi, raw invoice veya usage satirini event/log payload'ina koyma.
- Manual provider'i tam desteklenen final provider kabul et; compatibility dali
  veya ileride kaldirilacak gecici model olarak kurma.

## Nihai mimari ve migration kapisi

- V1 JSON limit, Redis counter, transaction-disi billing event veya dual-write
  adapter'ini V2 core'a tasima.
- Gecici money modeli, mutable plan, fake invoice veya ikinci entitlement engine
  kurma; yeni semantik ancak Accepted ADR ile degisir.
- Migration'i ADR-0013 `(tenant,billing)` single-writer fence'iyle yap; inventory,
  immutable version mapping, usage flush/seed ve evidence checksum'u persist et.
- Cutover gate sifir fark kanitlamadan writer switch yapma. Rollback down migration
  veya V1 runtime adapter'i degil, route/image ve reverse reconciliation'dir.

## Verification

- [ ] Owner table/migration/API siniri ve cross-module SQL yasak testi var.
- [ ] Her CRUD, entitlement, usage, webhook ve provider islemi ayri dosya/testte.
- [ ] `code-organization` tamamlama kapisi green.
- [ ] Money/period/version ve immutable history property testleri geciyor.
- [ ] Concurrent usage, redemption, transition ve idempotency race testleri geciyor.
- [ ] Commit/rollback, reserve/release, retry ve fail-closed limit testleri geciyor.
- [ ] Permission, cross-tenant, webhook mapping/replay ve secret redaction testleri var.
- [ ] Fresh migration, eski snapshot upgrade, V1 differential, fence ve rollback
  tatbikati tamam.
