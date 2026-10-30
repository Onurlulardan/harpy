# Modül — Notifications

## 1. Owner alanı

Notifications canonical inbox message, notification preference, template, delivery attempt, provider config ve realtime ticket projection alanlarının owner'ıdır.

Kaynak kök: core/internal/modules/notifications/

## 2. Message domain

Message:
- RecipientID ve TemplateVersionID
- Producer ve NotificationType
- Category, TemplateKind ve Locale
- Subject/Body
- DataDigest, SubjectDigest, BodyDigest
- Revision ve InboxRevision
- CreatedAt, ReadAt, DismissedAt, PurgedAt

Message New constructor ID, recipient/template identity, category, template kind, producer/type/locale, content size ve initial revision doğrular. MarkRead ve Dismiss expected revision ister. Tekrar aynı state'e geçiş idempotent kabul edilir. PurgeContent subject/body içeriğini boşaltır ve PurgedAt kaydeder.

## 3. Preference ve delivery policy

Preference recipient/category/channel bazında tutulur. Default preference kanala ve kategoriye göre değişir:
- In-app inbox normalde etkin gelir.
- Security, workflow action ve billing e-posta bildirimleri default etkin olabilir.
- Product ve bazı system/document external channel tercihleri default kapalıdır.
- Telegram yalnız explicit ve verified target varsa kullanılabilir.

Mandatory notification policy kullanıcının kapalı tercihine rağmen gerekli in-app veya security/legal bildirimlerini etkin tutabilir.

QuietHours timezone adı, start minute ve end minute içerir. IANA timezone ile validate edilir. Saat aralığı gece yarısını geçebilir; notification NotBefore bir sonraki uygun local time'a materialize edilir.

## 4. Delivery state machine

Delivery state:
- pending
- suppressed
- sending
- accepted
- delivered
- bounced
- failed
- cancelled

Delivery kaydı channel, provider config, attempt count, maximum attempts, lease owner/generation/until, NotBefore, outcome code, outcome unknown, mandatory flag ve revision taşır.

NewDelivery maximum attempt sayısını 8 olarak doğrular. Claim:
- Yalnız pending ve due delivery'yi alır.
- Outcome unknown durumunu tekrar gönderim için güvenli varsaymaz.
- Attempt ve lease generation artırır.
- Sending state ve lease owner/until yazar.

Lease süresi bir dakika olarak materialize edilir; stale worker aynı lease generation ile delivery sonucunu yazamaz.

## 5. RecordIntent akışı

Producer modülü notification send işlemini doğrudan provider'a yapmaz. RecordIntent:
1. Producer/type authorization doğrular.
2. Recipient'i IAM principal resolver üzerinden çözer.
3. Recipient tenant ve authorization version eşleşmesini doğrular.
4. Tenant scope store'larını bind eder.
5. Published template ve locale fallback'ini çözer.
6. Template'e göre effective channel ve provider config seçer.
7. Effect key ve canonical request digest üretir.
8. Duplicate intent'i lock/receipt ile kontrol eder.
9. Canonical message ve delivery state'lerini persist eder.

Böylece consumer yalnız notification intent kaydeder; dış teslimatı worker yapar.

## 6. Template ve provider registry

Template published version üzerinden seçilir; key olarak notification type, locale ve template version kullanılır. Template body HTML-system veya plain-text türlerinden biri olabilir. Content ve render data digest'leri audit/idempotency amaçlı kullanılır.

Provider registry somut sender'ları provider kind ile eşler. Kaynakta SMTP ve Telegram provider sender tipleri vardır. SMTP/Telegram secret configuration encrypted/opened secret portları üzerinden yönetilir.

## 7. Delivery worker

DeliveryWorker dependencies:
- Delivery runtime store
- Recipient/target resolver
- Secret opener
- Provider registry
- Clock
- ID source

Worker delivery'yi acquire eder; suppressed state varsa provider send yapmaz. Sending state için provider outcome classify edilir. Accepted/delivered/bounced/permanent failure/transient/unknown outcome ayrı state ve receipt semantiği taşır.

SMTP send sonucu belirsizse blind retry ile duplicate mail gönderilmez; provider outcome evidence esas alınır.

## 8. Realtime ve inbox

In-app inbox DB'deki message state'in canonical görünümüdür. Realtime ticket ve projection/WebSocket sinyali yalnız client'a değişiklik olduğunu bildirir; mesajın read/dismiss state'i değildir.

Read, dismiss ve preference update revision-aware use-case'lerdir. Inbox list/read model ve realtime projection aynı authorization scope'a bağlı kalır.

## 9. Kaynak haritası

- Domain: core/internal/modules/notifications/domain/
- Application: core/internal/modules/notifications/application/
- Provider registry: core/internal/modules/notifications/adapters/providerregistry/
- SMTP/Telegram: core/internal/modules/notifications/adapters/smtp/, telegram/
- Secret crypto: core/internal/modules/notifications/adapters/secretcrypto/
- PostgreSQL/HTTP: core/internal/modules/notifications/adapters/
- Composition: core/internal/composition/api/build_notifications_http.go, worker/build_notifications_*.go
- Contracts: contracts/authority/notifications.json, contracts/events/, contracts/http/
- Skill: skills/notifications/SKILL.md
