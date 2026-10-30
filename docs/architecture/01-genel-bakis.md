# 01 — Harpy genel bakış

## 1. Harpy nedir?

Harpy, Go ile geliştirilmiş bir backend starter ve uygulama platformudur. IAM, tenant yönetimi, katalog, billing, iş yönetimi, belge yönetimi, bildirim, analytics ve platform operasyonları gibi yetenekleri tek bir modüler backend kod tabanında toplar.

Harpy'nin temel çalışma profilleri:
- **Single-tenant / on-premise:** Bir müşteri kurulumu için ayrılmış çalışma ortamı.
- **Multi-tenant / SaaS:** Aynı platform runtime'ında birden fazla tenant'ın yönetilmesi.

Harpy bir frontend repository'si değildir. Kaynak ağacında React/Vue/Svelte gibi bir frontend uygulamasının yerleşik yapısı bulunmaz. Yerel çalışma görevleri kardeş `frontend-v3` projesini ayrıca çalıştırabilir; backend API ve worker bu repository'dedir.

## 2. Teknoloji seti

| Katman | Kullanım |
| --- | --- |
| Dil/runtime | Go `1.26.0`, toolchain `1.26.5` |
| Veritabanı | PostgreSQL 18, pgx v5 |
| Mesaj taşıma | RabbitMQ, `amqp091-go` |
| Nesne depolama | S3 API, AWS SDK for Go v2; yerel RustFS |
| Şifreleme | Go standard crypto paketleri, AEAD tabanlı içerik ve anahtar koruma |
| Kimlik | Ed25519 imzalı access token; opaque refresh credential |
| Telemetry | OpenTelemetry trace/metric SDK ve OTLP exporter |
| Runtime paketleme | Docker Compose, Task |
| Şema/kontrat | JSON Schema, OpenAPI ve sürümlü Go contract paketleri |

Dependency listesi için `go.mod`; gerçek runtime ayarları için `core/internal/platform/config`; yerel üçüncü parti servisler için `deploy/third-party.yml` esas alınır.

## 3. Repository haritası

```text
cmd/setup/                       İlk kurulum aracı
core/cmd/                        API, worker ve operator giriş noktaları
core/internal/composition/       Süreç bazında runtime assembly
core/internal/modules/           İş modülleri
core/internal/platform/          Teknik runtime primitive'leri
core/internal/capabilities/      Modüller arası teknik capability'ler
core/internal/orchestration/     Çok adımlı platform koordinasyonu
contracts/                       HTTP, event ve runtime sözleşmeleri
compat/                           Sınırlandırılmış uyumluluk adapter'ları
deploy/                           Tekil/çoklu tenant compose profilleri
ops/                              Local ve operasyonel süreç dosyaları
tools/                            Guard, contract ve geliştirme araçları
skills/                           Repository mühendislik prosedürleri
docs/                             İnsan odaklı teknik dokümantasyon
```

## 4. Process/deployable görünümü

Harpy'nin ana uzun yaşayan süreçleri API ve worker'dır. Operator ise belirli kurulum/yönetim işlerini tamamlayan ayrı bir process girişidir. Bu ayrım, modül sınırlarıyla aynı şey değildir: bir modülün kodu API ve worker process'leri içinde birlikte kullanılabilir.

| Process | Giriş paketi | Ana görev |
| --- | --- | --- |
| API | `core/cmd/api` | Dış HTTP isteklerini kabul eder ve modül route'larını sunar. |
| Worker | `core/cmd/worker` | DB job'larını ve event consumer workload'larını işletir. |
| Operator | `core/cmd/operator` | Kurulum ve özel yönetim operasyonlarını yürütür. |
| Setup CLI | `cmd/setup` | Çalışma profili ve yerel secret/env dosyalarını hazırlar. |

## 5. Başlıca iş modülleri

| Modül | Temel kavramlar |
| --- | --- |
| IAM | User, role, organization, permission, credential, session |
| Tenancy | Tenant, tenant schema, deployment policy, onboarding |
| Billing | Offering, price, subscription, invoice, entitlement, usage |
| Work | Board, column, task, membership, calendar event |
| Documents | Folder, item, version, upload, archive, legal hold |
| Notifications | Message, preference, template, delivery, provider |
| Analytics | Saved query, dashboard, publication, capability |
| Audit | Owner operation audit ve field-change evidence |
| Usage | Feature usage observations ve agregasyonlar |
| Platform Authority | Signed authority proposal, acknowledgement, checkpoint |
| Platform Ops | Diagnostics, recovery set, restore operation, traces/alerts |
| Support | Bug report ve support contact |
| System Resource Layout | Revisioned system form/layout/presentation catalog |

## 6. Harpy'nin mimari yaklaşımı

Harpy modüler monolit olarak tasarlanmıştır:
- Kod sahipliği ve veri sahipliği modül bazında ayrılır.
- Modüller arası senkron erişim owner modülün public `api` paketiyle yapılır.
- Teknik altyapı `platform` paketlerine konur; business policy platforma taşınmaz.
- Composition root, somut modülleri ve adapter'ları process başına bağlar.
- Asenkron entegrasyonlar event ve job kontratlarıyla yönetilir.
- Guard ve contract testleri, bu kuralların CI'da korunmasını sağlar.

Ayrıntı: [Mimari ve modül sınırları](02-mimari-ve-modul-sinirlari.md).

## 7. İlk bakılacak dosyalar

| İhtiyaç | Dosya |
| --- | --- |
| Genel amaç ve ilk komut | `README.md` |
| Değişiklik öncesi protokol | `AGENTS.md` |
| Skill eşlemesi | `skills/INDEX.md` |
| Go module ve dependency | `go.mod` |
| Süreç başlangıcı | `core/cmd/api/main.go`, `worker/main.go`, `operator/main.go` |
| API composition | `core/internal/composition/api/` |
| Worker composition | `core/internal/composition/worker/` |
| Domain modülleri | `core/internal/modules/` |
| Runtime sözleşmeleri | `contracts/runtime/` |
