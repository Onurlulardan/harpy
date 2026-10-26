---
name: testing-quality
description: Define and enforce backend-v2 unit, integration, contract, architecture, migration, tenant isolation, no-code fixture, differential, performance, and rollout verification. Use for every change and for CI or quality-gate work.
---

# Testing Quality

## Test secim sirasi

1. Degisen invariant'i yaz.
2. En dar deterministic testi ekle.
3. Gercek adapter davranisi icin integration testi ekle.
4. Public wire veya event varsa contract testi ekle.
5. Cross-module/tenant/migration riski varsa ilgili gate'i calistir.
6. Sonucu ve calistirilmayan kontrolleri raporla.

## Katmanlar

### Unit

- Domain policy, parser, normalizer, compiler ve pure mapping.
- Table-driven operator, field type, node type ve error matrix.
- Clock, ID ve random kaynagini explicit dependency yap.

### Integration

- Gercek ephemeral PostgreSQL ve tenant schema kullan.
- Transaction, constraint, lock, query plan ve migration davranisini test et.
- Redis/RabbitMQ/MinIO yalniz ilgili adapter degistiginde gercek container kullan.
- ORM/mock SQL'i database dogrulamasinin yerine koyma.

### Contract

- OpenAPI request/response/error semantic diff.
- Event ve job envelope version fixture'i.
- SSE, WS, multipart ve binary protokol testi.
- frontend-v3 ve MCP tarafindan kullanilan wire shape fixture'i.

### Architecture

- `code-organization` fileguard kapisi.
- Yasak dumping-ground dosya adlari.
- Cross-module internal/adapter importu.
- Owner disi tablo/SQL erisimi.
- Source dosyasi ile operation test eslesmesi.

Create/get/list/update/delete gibi davranis dosyalari ayni adli test tasir.
`routes.go`, `row.go`, DTO ve `module.go` gibi destek dosyalari davranis
iceriyorsa kendi testini tasir; salt wiring/data ise ilgili integration veya
module testinde kapsandigi kanitlanir. Her destek dosyasina anlamsiz 1:1 test yazma.

## Zorunlu risk testleri

Tenant/data:

- Tenant A kaydi Tenant B tarafindan okunamaz veya degistirilemez.
- Header/token mismatch fail-closed olur.
- Fresh migration ve eski snapshot upgrade edilir.
- Backfill resume ve rollback/forward-fix denenir.

Async:

- Commit rollback'te outbox/job yoktur.
- Duplicate, reorder, retry, crash ve stale lease state kaybetmez.

No-code:

- Her field/operator/node capability contract suite'ten gecer.
- Unknown config fail-closed olur.
- Published eski version ve in-flight instance fixture'i calisir.
- Query ve workflow budget'lari zorlanir.

Security:

- Auth, RBAC, field permission, CSRF ve idempotency negative testleri.
- SQL identifier/operator fuzz ve expression sandbox testleri.
- Secret/password/token response ve log redaction testleri.

## V1 differential

Ayni fixture'i V1 ve V2'ye ver. Timestamp, request ID gibi oynak alanlari
normalize et. Wire response ile birlikte DB mutation, audit, outbox ve visible
record setini karsilastir. Bilincli farki ADR/fixture metadata'sinda acikla.

## Performance

Kritik akislarda sunlari baseline ile karsilastir:

- p50/p95/p99 latency.
- SQL query sayisi ve connection kullanimi.
- Exact count ve query plan maliyeti.
- Worker throughput, lag ve retry orani.
- Allocation ve payload boyutu.

## Tamamlama raporu

```text
Commands run:
Passed:
Failed:
Skipped and why:
Compatibility result:
Residual risk:
Rollback evidence:
```

Her degisiklikte en az targeted test ve
`go run ./tools/fileguard` ve `go run ./tools/archguard` calistir. Testi gecirmek icin production
guard'ini gevsetme veya fixture'i aciklama olmadan yeniden yazma.
