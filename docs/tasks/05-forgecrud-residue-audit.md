# 05 — ForgeCRUD Kalıntı Denetimi ve Harpy Ürün Sınırı

## 1. Amaç ve ürün sınırı

Harpy, ForgeCRUD'ün ürün olarak devamı değildir. Harpy, ForgeCRUD kaynak kodundan ayıklanmış bağımsız ve yeniden kullanılabilir bir temel repository'dir. ForgeCRUD yalnız tarihsel kaynak kod ilişkisidir; ürün kapsamı, modülleri, ekranları, iş akışları, schema tasarımları ve kullanıcı davranışları Harpy gereksinimi değildir.

Bir özellik bugün Harpy repository'sinde bulunuyor diye Harpy'nin gereksinimi sayılmaz. ForgeCRUD ürün özelliği sessizce korunamaz, yeniden adlandırılarak taşınamaz veya Harpy roadmap'ine varsayılan olarak eklenemez. Bu task, 04 — APPLICATION ve SIFIR KALINTI görevinin zorunlu girdisidir.

## 2. Denetim ilkeleri

- ForgeCRUD'e özgü ürün davranışı, ayrıca onaylanmış generic Harpy gereksinimi değilse kaldırılır.
- Marka, ürün adı, URL, görsel, pazarlama metni ve ürüne özel deployment kimliği Harpy'de kalmaz.
- Generic altyapı ancak bağımsız contract, gerçek Harpy consumer'ı, owner'ı ve testi varsa tutulabilir.
- Kodun ForgeCRUD kaynaklı olması tek başına silme nedeni değildir; Harpy'de bulunması da koruma gerekçesi değildir.
- Generic kod ForgeCRUD runtime/config/contract/schema bağımlılığı olmadan çalışmalıdır.
- Ürün davranışını korumak için adapter, compatibility layer, placeholder UI veya mock product module oluşturulmaz.
- Belirsiz öğe UNRESOLVED kaydedilir ve 04 tamamlanmasını engeller.

## 3. Repository genelinde incelenecek yüzeyler

Repository root'undan recursive denetim yapılır; yalnız core/ veya grep yeterli değildir.

### 3.1 Marka ve ürün kimliği
- Go module/package, executable, CLI help, version/build metadata.
- README, docs, architecture records, skills, AGENTS.md ve developer guides.
- Frontend title/metadata, favicon, logo, manifest, static assets ve public path.
- Domain, URL, email, OAuth callback, telemetry, analytics, license/activation ve support link'leri.
- Docker image/OCI label, Compose service/container, volume/network ve deployment manifest'leri.
- Environment variable, config key, secret adı, generated artifact, install/release metadata.
- GitHub Actions, release workflow, badges, package/container registry ve artifact destination.

Arama varyantları: ForgeCRUD, ForgeCrud, forgecrud, forge-crud, forge_crud ve gerçek domain/package/URL varyantları. Her eşleşme context ile incelenir; bu audit sözlüğü false positive kabul edilir.

### 3.2 Ürün modülleri ve davranışları
Aşağıdakiler Harpy feature listesi değil, tarihsel ürün kalıntısı arama sözlüğüdür:
- HR/personnel/employee management.
- CRM, sales, customer/pipeline.
- Inventory, warehouse, stock, purchasing.
- Accounting, finance, invoicing.
- Production/manufacturing, quality, maintenance.
- Project management, task/calendar screens.
- No-code form designer, dynamic form publishing, field designer.
- Workflow designer/engine, business process publishing.
- Dashboard/report designer.
- API builder, portal, menu designer, print designer.
- Product-specific AI assistant/agent, prompt/tool catalogue.
- Marketplace/module registry, pricing/subscription, product feature flags.
- Ürüne özel RBAC/ABAC policy preset, navigation tree ve default seed.

Terim eşleşmesi tek başına silme/koruma kararı değildir; gerçek davranış, owner ve consumer graph incelenir.

### 3.3 Veri ve persistence
- ForgeCRUD domain tabloları, schema, enum/type, function, procedure, trigger, view, RLS policy, index ve seed.
- Ürün migration SQL'i, baseline snapshot, fixture, test database ve sample data.
- Ürüne özel tenant tabloları, custom fields, form metadata, record storage ve workflow state.
- Ürün API repository/store/query implementation'ları ve generated storage contract'ları.
- schema-executor, typed DDL IR ve dynamic schema publishing bağlantıları.

Her SQL object için DDL kaynağı, güncel Go consumer, API/domain caller, seed/test kullanımı ve Harpy schema'sında gerekliliği kanıtlanır.

### 3.4 Frontend, API, build ve operasyon
- Route/page/component/menu/navigation, layout, settings ve product administration ekranları.
- API route, handler, DTO, OpenAPI/protobuf contract, permission scope ve generated client.
- Product-specific event, notification, background process, scheduler, webhook ve external integration.
- Static JSON, demo seed, locale/i18n string, sample config ve e2e fixture.
- Go/npm dependencies, generated code, generators, templates, scaffolding.
- Dockerfile, Compose, Kubernetes/diğer deployment manifestleri, Taskfile, shell scripts.
- CI/release pipeline, package artifact, container image, SBOM, license metadata.
- Config profile, secret projection, environment schema, volume/port/service/healthcheck.
- Test helper, mock, fixture generator, benchmark, docs, samples, skills ve historical files.

Harpy'nin gerçekten ihtiyaç duyduğu generic health, identity, authorization, tenancy veya operations surface'i sırf ForgeCRUD'de de vardı diye silinmez; generic olduğu kanıtlanır.

## 4. Generic foundation olarak korunabilecek alanlar

Bunlar otomatik allowlist değildir; bağımsız Harpy ihtiyacı kanıtlanırsa korunabilir:
- PostgreSQL connection/transaction/query/storage primitives.
- Genel tenancy identity/isolation.
- Genel authentication, identity ve authorization.
- Generic permission evaluation (ürün policy preset'leri hariç).
- Config, secrets, logging, metrics, tracing, health/readiness.
- Generic HTTP/API runtime, middleware, error handling, validation.
- Generic object/file storage, backup/restore, recovery.
- INITIALIZE, DATABASE FOUNDATION, Goose, APPLICATION.
- Orchestrator-managed one-shot tenant provisioning Job integration.
- Generic testing, architecture/file guards ve developer tooling.

Her KEEP kararı somut Harpy consumer'ı, bağımsız owner/package sınırı, ForgeCRUD bağımlılığı olmadığının kanıtı ve gerçek test/contract göstermelidir.

## 5. Disposition modeli

| Karar | Anlamı |
| --- | --- |
| REMOVE_FORGECRUD_PRODUCT | ForgeCRUD ürün kodu/contract/schema/asset/deployment kalıntısı; tüm producer/consumer bağlantılarıyla kaldırılır |
| KEEP_GENERIC_FOUNDATION | Harpy generic altyapısı; consumer, owner, bağımsızlık kanıtı ve test belirtilir |
| REPLACE_WITH_GENERIC_PRIMITIVE | Harpy için gerekli genel primitive ürün davranışına gömülmüş; generic owner/contract/test tanımlanır |
| UNRESOLVED | Kanıt yetersiz veya owner/consumer belirsiz; 04 tamamlanamaz |

Her inventory satırı path_or_object, symbol_or_feature, forgecrud_evidence, harpy_consumers, decision, owner, action_or_replacement, test_evidence ve reason alanlarını taşır. Envanter migration manifest/catalog/runtime contract'a dönüştürülmez.

## 6. schema-executor ve typed DDL IR özel denetimi

schema-executor, typed DDL IR, composition/command çağrıları ve config/secret/contract/SQL/test/deployment bağlantıları ayrıca araştırılır. Bunlar Harpy için dynamic form/config publishing gereksinimi oluşturmaz. Form publishing, JSONB/EAV storage veya record CRUD ürünü varsayılmaz; yerine yeni ürün alt sistemi kurulmaz.

Kaldırılacaksa tüm consumer graph kapatılır. Gerçek generic PostgreSQL primitive varsa yalnız ayrı owner altında, ForgeCRUD product contract'ı olmadan ve somut Harpy consumer/test kanıtıyla korunabilir. Eski executor'ı yeniden adlandırmak veya adapter arkasında yaşatmak yasaktır.

## 7. Denetim yöntemi

1. Repository file tree ve büyük dizinlerin tam inventory'si.
2. Case-sensitive/case-insensitive marka ve ürün token taraması.
3. Go package/import ve symbol/call graph.
4. SQL DDL/object inventory ile Go SQL consumer cross-reference.
5. API route/OpenAPI/protobuf/DTO ve frontend route/component/menu graph.
6. Config/env/secret producer-consumer graph.
7. Docker/Compose/orchestrator/deploy/artifact graph.
8. Dependency/module ve generator/template consumer graph.
9. Test/fixture/seed/sample/docs/skills/CI referans taraması.
10. Silme/dönüştürme sonrasında aynı taramaların tekrarı.

Yalnız grep sonucu yeterli değildir. Her eşleşme gerçek kullanım context'iyle sınıflandırılır.

## 8. Kabul kriterleri

- ForgeCRUD marka kimliği, link, asset ve ürün metadata'sı aktif Harpy kod/build/deploy/artifact yüzeylerinde kalmamıştır.
- ForgeCRUD product module, UI, API, route, seed, SQL object, config ve external integration kalıntıları kaldırılmış veya generic foundation olarak kanıtlanmıştır.
- Harpy generic altyapısı ForgeCRUD package/schema/config/contract'a bağlı değildir.
- schema-executor/typed DDL IR ürün davranışı Harpy gereksinimi olarak korunmamış, consumer graph temizlenmiştir.
- Eski ForgeCRUD ürün tabloları/seed'leri yeni Goose başlangıç schema'sına taşınmamıştır.
- README/docs/architecture/skills Harpy'yi ForgeCRUD ürünü gibi anlatmamaktadır.
- Docker/Compose/orchestrator/CI/release/dependency/artifact zincirinde ForgeCRUD product dependency kalmamıştır.
- Bütün eşleşmeler disposition ve kanıt taşır; UNRESOLVED yoktur.
- Go test, frontend/build (varsa), fileguard, archguard, contract/generator ve PostgreSQL fresh-install kontrolleri gerçek sonuçlarıyla raporlanmıştır.
- Compatibility layer, placeholder product module veya generic olmayan abstraction oluşturulmamıştır.

## 9. 04 numaralı task ile bağ

04 görevi bu inventory'yi migration residue inventory ile birlikte kullanır; REMOVE kayıtlarını siler, KEEP_GENERIC kararlarının kanıtını doğrular, REPLACE kararlarını yalnız açık Harpy gereksinimi varsa uygular. Tek bir UNRESOLVED kalırsa sıfır kalıntı tamamlandı demez. Bu task, 04'ün zorunlu repository audit kapısıdır.
