---
name: notifications
description: Build backend-v2 durable in-app notifications, preferences, templates, realtime hints, email and Telegram delivery, provider secrets, receipts, retries, suppression, retention, and V1 notification migration. Use for every notification intent, inbox, WebSocket, or provider-delivery change.
---

# Notifications

Message/delivery state'leri, route, permission, preference, retention, retry ve
cutover degerleri icin ADR-0018 ve ADR-0030 authoritative'dir;
uygulamadan once tamamini oku. Endpoint icin `api-contract`, SQL/tenant icin
`data-tenancy`, worker/outbox icin `events-jobs`, recipient/RBAC icin
`iam-system-resources` skill'ini de oku.

## Sahiplik siniri

- `notifications` bir core business module'dur; ayri notification service acma.
- Message, delivery, attempt, preference, template ve provider config/secret
  tablo/migration'larinin tek owner'i olsun.
- Provider adapter'larini `core-worker` calistirsin; provider SDK tipini
  domain/application kontratina tasima.
- Business modulu notification tablosuna yazmasin. Commit sonrasi intent icin
  versioned outbox uretsin veya gerekli senkron sinirda public command'i cagirsin.
- RabbitMQ, WebSocket connection'i ve provider receipt'i canonical inbox state'i
  yapma; persisted message ve delivery state'ini authority kabul et.
- IAM public resolver'i kullan; request'teki user, email veya tenant degerine
  dogrudan guvenme.

## Use-case dosyalama

Her operasyonu application, HTTP adapter ve PostgreSQL adapter'da ayri dosyala:

```text
create_intent.go, get_message.go, list_messages.go
mark_read.go, mark_all_read.go, dismiss_message.go
get_preferences.go, update_preferences.go
create_template.go, publish_template.go
create_provider_config.go, update_provider_config.go, rotate_provider_secret.go
claim_delivery.go, send_delivery.go, record_receipt.go, retry_delivery.go
create_realtime_ticket.go, consume_realtime_ticket.go
```

- Desteklenmeyen operasyon icin bos katman veya generic CRUD abstraction'i acma.
- Her davranis dosyasina ayni operasyon adli test yaz.
- `routes.go` yalniz route baglasin; policy, rendering veya delivery karari tasimasin.
- Dosya/CRUD/isim sinirinda `skills/code-organization/SKILL.md`yi uygula.

## Intent ve delivery sirasi

1. Actor/tenant, producer capability, recipient ve notification type'ini dogrula.
2. Stable effect key, published template/config version, locale ve canonical data'yi
   validate et; unknown type, variable veya channel'i fail-closed reddet.
3. Preference ve mandatory policy'yi uygula; message, deliveries, audit, jobs ve
   realtime outbox'i tek tenant Unit of Work'unde commit et.
4. Her external attempt'ten hemen once preference/config durumunu tekrar kontrol et.
5. Provider sonucunu immutable attempt/receipt olarak kaydet; retry kararini ADR'deki
   transient, permanent ve outcome-unknown sinifina gore ver.
6. Realtime'i yalniz canonical GET'i tetikleyen recipient-scoped hint olarak gonder.

Effect key'i ayni effect'in tekrarini etkisizlestirsin; farkli payload ile reuse
conflict olsun. Rendered body, credential, recipient adresi ve provider token'ini
event, job payload, log, trace veya DLQ'ya koyma. List endpoint'inde yalniz REST
Query allowlist'ini kullan.

## Nihai kontrat kapisi

- Fire-and-forget goroutine, in-memory toast authority, Rabbit-only state,
  plaintext credential veya arbitrary public send endpoint'i ekleme.
- Provider timeout'unda kanit olmadan otomatik duplicate effect uretme.
- V1 queue/WebSocket uyumlulugunu core modele kalici dal yapma. Compatibility edge;
  persisted state, owner, telemetry ve ADR'deki otomatik exit gate'iyle sinirli olsun.
- "Simdilik farkli inbox/provider modeli, sonra birlestiririz" yaklasimini reddet.
- Cutover'da V1/V2 consumer veya provider dual-delivery acma.

## Verification

- [ ] Owner table/migration/API siniri ve yasak cross-module SQL testi var.
- [ ] Her inbox, preference, template, provider ve delivery islemi ayri dosya/testte.
- [ ] `code-organization` tamamlama kapisi green.
- [ ] Message/delivery atomicity, effect-key conflict ve duplicate/reorder testleri geciyor.
- [ ] Preference/mandatory/quiet-hour/locale ve opt-out race testleri geciyor.
- [ ] Retry, stale lease, receipt, bounce, DLQ ve outcome-unknown testleri geciyor.
- [ ] Secret redaction/rotation, TLS ve malicious-template testleri geciyor.
- [ ] Own-inbox IDOR, tenant isolation, realtime ticket ve dropped-hint testleri geciyor.
- [ ] V1 mapping, queue drain, cutover fence ve rollback kaniti mevcut.
