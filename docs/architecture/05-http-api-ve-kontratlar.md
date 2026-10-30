# 05 — HTTP, API ve kontratlar

## 1. HTTP request pipeline

Harpy'de HTTP transport, domain kurallarından ayrıdır. Bir endpoint'in olağan akışı:

```text
route registration
  -> edge admission / security middleware
  -> decoder and transport validation
  -> application command/query
  -> owner module + TenantScope
  -> presenter / stable response DTO
  -> HTTP response
```

Route dosyası method/path/middleware bağlar. Decoder wire format'ı doğrular. Application use-case business kararı ve transaction'ı yönetir. Presenter domain entity veya DB row'u doğrudan dışarı vermek yerine stable response DTO üretir.

## 2. Public ve management HTTP

API composition public ve management sunucularını ayrı kurar:
- Public server modül route registry'sini, admission ve security edge'ini kullanır.
- Management server readiness, servis bilgisi ve operasyonel endpoint'leri sunar.
- IAM security runtime public actor assertion gibi dış protokol risklerini ayrıca sınırlar.
- JWKS public endpoint'i `/.well-known/jwks.json` olarak GET/HEAD ile kayıtlıdır.

Public route ile management route aynı authorization ve ağ erişim beklentisine sahip değildir.

## 3. Modül HTTP adapter'ları

API composition'da modül bazlı builder dosyaları bulunur. Örnekler:
- `build_iam_http.go`, `iam_rbac_http.go`, `iam_profile_http.go`
- `build_billing_http.go`, `billing_routes.go`
- `build_catalog_http.go`
- `build_documents_http.go`
- `build_notifications_http.go`
- `build_work_http.go`
- `build_analytics_http.go`
- `build_audit_http.go`
- `build_platformops_http.go`
- `build_support_http.go`
- `build_system_resource_layout_http.go`

Bu dosyalar concrete constructor, shared dependency ve handler'ları kurar. Route path ve wire contract için `contracts/http/` ve modül adapter testlerine bakılır.

## 4. API contract kaynağı

`contracts/http/` HTTP şema ve path tanımlarını tutar. Dosyalar capability/modül bazında ayrılmıştır. OpenAPI aggregate ve runtime contract testleri bu tanımları uygulamaya bağlar.

`contracts/runtime/v1/error-envelope.schema.json` gibi runtime tanımları hata envelope'unun şeklini sabitler. `contracts/runtime/v1/openapi_contract_test.go` ve modül bazlı contract testleri şema/route eşleşmesini kontrol eder.

## 5. Hata envelope'u

API hataları client'a internal SQL/provider mesajı sızdırmamalıdır. Stable machine code, kullanıcıya uygun genel message, issue listesi ve correlation ID taşınır.

Örnek wire biçimi:

```json
{
  "success": false,
  "error": {
    "code": "QUERY_INVALID",
    "message": "Request could not be processed",
    "issues": [
      {"path": "/where/and/0/field", "code": "FIELD_UNKNOWN"}
    ],
    "correlation_id": "..."
  }
}
```

Issue path RFC 6901 JSON Pointer biçimindedir; issue code stable rule identifier'dır. Validation, conflict, forbidden ve dependency failure ayrı semantiklere sahip olmalıdır.

## 6. REST Query — tek dinamik query capability

Dynamic list/read surface'leri için ikinci bir SQL parser veya ad-hoc filter map kurulmaz. Ortak capability `core/internal/capabilities/restquery` altındadır.

Execution pipeline:

1. Transport GET/POST girdisini decode eder.
2. Girdi canonical `QuerySpec` modeline normalize edilir.
3. Public field ID, ilgili resource/version field catalog üzerinden çözülür.
4. Field type ve operator uyumu doğrulanır.
5. Resource, row ve field authorization uygulanır.
6. Zorunlu tenant/owner/visibility predicate'leri eklenir.
7. Predicate sayısı, nesting depth, limit, timeout ve cost budget uygulanır.
8. Compiler parametreli SQL planı üretir.
9. Plan timeout ve row budget içinde çalıştırılır.
10. Canonical response presenter ile üretilir.

Unknown field/operator sessizce atlanmaz. Client raw SQL, raw identifier, arbitrary join veya unsupported operator veremez.

## 7. REST Query alt paketleri

| Paket | Amaç |
| --- | --- |
| `contract/` | Query document, strict member decode ve redaction |
| `api/` | Consumer-facing QuerySpec, field catalog ve execution contracts |
| `authorization/` | Query ve field authorization |
| `resolver/`, `resolved/` | Stable field/relation resolution |
| `planner/` | Query budget, cost ve statement plan |
| `adapters/http/` | GET/POST decode ve presentation |
| `adapters/postgres/` | Parameterized SQL compiler/executor |
| `cursor/` | Stable pagination cursor handling |
| `published/` | Published query execution surface |

Modül consumer'ı capability'nin iç compiler paketini değil, public `restquery/api` kontratını kullanır.

## 8. Pagination, search ve field security

- Stable sort ve ID tie-breaker olmadan pagination deterministik değildir.
- Cursor içeriği actor, tenant, query shape, field catalog ve sort bağlamına bağlanır ve şifreli korunur.
- Field permission sadece response'da değil filter/sort/search kullanımında da kontrol edilir.
- Relation/include yalnız declared relation catalog üzerinden çözülür.
- Search yalnız published search profile alanlarında çalışır.
- Exact count ayrı maliyet kararıdır; bütün list endpoint'lerine varsayılan olarak eklenmez.
- Interactive query statement timeout ve row limit uygular.

## 9. Event/API contract ilişkisi

HTTP request senkron command/query yüzeyidir. Commit sonrası başka process'e taşınması gereken olaylar `contracts/events/` altındaki versioned envelope'larla tanımlanır. API DTO'su, event payload'ı ve database row modeli aynı struct olarak kullanılmaz.

## 10. İnceleme ve değişiklik kaynakları

- Normatif endpoint prosedürü: `skills/api-contract/SKILL.md`
- REST Query prosedürü: `skills/rest-query/SKILL.md`
- HTTP schemas/paths: `contracts/http/`
- Runtime error envelope: `contracts/runtime/v1/error-envelope.schema.json`
- API route composition: `core/internal/composition/api/`
- Modül handler: `core/internal/modules/<module>/adapters/http/` veya modül root HTTP adapter dosyaları
- REST Query implementation: `core/internal/capabilities/restquery/`
