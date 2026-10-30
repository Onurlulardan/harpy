# 14 — Mimari kararlar ve trade-off'lar

Bu sayfa, mevcut implementasyon ve repository'nin normatif architecture skill'lerinden gözlemlenen kararları özetler. Bu liste resmi ADR kayıt dizini değildir; kararların kaynak kod üzerindeki görünür etkilerini açıklar.

## 1. Modüler monolit

**Karar:** İş yetenekleri tek Go module içinde, owner sınırları ve public API'lerle ayrılmıştır.

**Kazanç:** Aynı transaction ve process içindeki use-case'ler gereksiz network hop olmadan çalışır; dependency yönü archguard ile korunur.

**Bedel:** Composition root ve contract sahibi dikkatle yönetilmelidir; cross-module erişimde shortcut kullanmak cazip olsa da yasaktır.

## 2. Process ile module ayrımı

**Karar:** API, worker ve operator process girişleri ayrıdır; business modüller process'ler içinde composition root ile kurulur.

**Kazanç:** HTTP request load ile background work ve operator task'ları birbirinden ayrılır.

**Bedel:** Aynı modülün farklı process profile'larında doğru dependency ve credentials ile kurulması gerekir.

## 3. PostgreSQL canonical state

**Karar:** Business state, idempotency receipt ve durable job/event intent PostgreSQL'de tutulur.

**Kazanç:** Transaction ve constraint'ler business invariant'larını enforce eder; broker/provider geçici unavailable olduğunda iş kaydı kaybolmaz.

**Bedel:** DB scope, index, isolation ve query budget tasarımı kritiktir. Pool/session hygiene zorunludur.

## 4. Tenant schema isolation

**Karar:** Tenant data schema-per-tenant şeklinde ayrılır; tenant schema adı server-side canonical TenantID'den türetilir.

**Kazanç:** Tenant namespace ve search_path izolasyonu DB katmanında enforce edilir.

**Bedel:** Her işlem doğrulanmış TenantScope içinde çalışmalıdır; global/platform data için ayrı scope ve owner policy gerekir.

## 5. Composition root

**Karar:** Process entrypoint yalnız kendi composition paketini çağırır.

**Kazanç:** Somut adapter seçimi, lifecycle, shutdown ve readiness tek noktada düzenlenir.

**Bedel:** Composition dosyaları sayıca yüksek olabilir; bu yüzden builder dosyaları use-case mantığı değil wiring sorumluluğu taşımalıdır.

## 6. Contract-first interop

**Karar:** HTTP, event, runtime, authority ve storage sözleşmeleri contracts dizininde versioned tutulur.

**Kazanç:** Contract shape source code ve generated outputs üzerinden doğrulanır; consumer/provider anlaşması explicit olur.

**Bedel:** Contract değişikliği producer, consumer, fixture, OpenAPI ve schema testlerine yansıtılmalıdır.

## 7. Outbox ve DB-backed jobs

**Karar:** Broker tek state authority değildir; DB outbox event publish'e, DB job ise scheduled work execution'a kaynak olur.

**Kazanç:** Publish confirmation, retry, lease, idempotency ve dead-letter state'i durable olarak izlenebilir.

**Bedel:** At-least-once delivery beklenir; consumer ve provider effect'leri idempotent tasarlanmalıdır.

## 8. Object storage abstraction

**Karar:** Business modules S3-compatible object port kullanır; S3 SDK yalnız blob adapter katmanında kalır.

**Kazanç:** RustFS, MinIO ve harici S3 uyumlu servisler aynı API arkasında kullanılabilir.

**Bedel:** Object result ambiguity, version ID, digest, ACL ve bucket ownership explicit doğrulanmalıdır.

## 9. Typed domain values

**Karar:** Money, TenantID, SchemaName, TaskStatus, PublicationState, UploadState gibi kavramlar string/int primitive yerine typed domain model ile taşınır.

**Kazanç:** Invalid state ve unit mix-up derleme/test/validation katmanında daha erken yakalanır.

**Bedel:** Wire DTO ve database representation dönüşümleri explicit adapter gerektirir.

## 10. Revision ve idempotency

**Karar:** Mutating use-case'lerde expected revision, operation identity ve idempotency digest kullanılır.

**Kazanç:** Concurrent update conflict'i açık hale gelir; retry aynı command'ın ikinci side effect üretmesini engeller.

**Bedel:** Client mutation request'inde revision ve idempotency semantics'i doğru taşınmalıdır.

## 11. Security fail-closed

**Karar:** Missing tenant scope, authority, permission, verified actor veya contract evidence success varsayılmaz.

**Kazanç:** Belirsiz durumda erişim reddedilir; security boundary explicit kalır.

**Bedel:** Negatif testler ve doğru error mapping zorunludur; ergonomi ile güvenlik kararı birbirine karıştırılmaz.

## 12. Fileguard/archguard

**Karar:** Kod organizasyonu ve dependency graph kuralları automated guard'larla korunur.

**Kazanç:** Dosya büyümesi, dumping-ground pattern ve illegal import erken yakalanır.

**Bedel:** Değişiklik küçük ve intention-revealing file set'lerine bölünmelidir.

## 13. Kaynakların normatif sırası

- AGENTS.md: çalışma protokolü
- skills/INDEX.md: skill routing
- İlgili skill: implementation prosedürü
- contracts/: wire/runtime ownership contract
- Source code ve tests: mevcut davranışın implementasyonu
- docs/architecture/: insan odaklı açıklama

Açıklayıcı doküman kaynak kodla çelişirse kaynak kod ve normatif skill incelenerek doküman düzeltilmelidir.
