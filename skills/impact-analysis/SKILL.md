---
name: impact-analysis
description: Trace backend-v2 changes across modules, APIs, tenancy, data, events, jobs, frontend-v3, MCP, compatibility, and rollout before implementation. Use for every cross-module or public-contract change and any change with uncertain ownership.
---

# Impact Analysis

## Prosedur

Kod yazmadan once tek sayfalik etki notu olustur.

### 1. Ownership

- Davranisin owner modulunu belirle.
- Degisecek application API, port ve adapter'lari listele.
- Yeni shared abstraction veya cross-module dependency olup olmadigini kontrol et.

### 2. HTTP ve consumer

- Method, path, auth, permission ve idempotency'yi izle.
- frontend-v3, MCP, webhook ve external consumer'lari bul.
- Envelope, cookie, header, SSE, WS, binary ve pagination etkisini yaz.

### 3. Tenant ve veri

- Platform schema mi tenant schema mi oldugunu belirt.
- Owner table, migration, backfill ve rollback yolunu yaz.
- Cross-tenant negative senaryoyu tanimla.
- Transaction icindeki state, audit, outbox ve job etkisini yaz.

### 4. No-code kontrati

- Etkilenen form field, relation, validation veya query capability'sini bul.
- Etkilenen workflow node, expression, checkpoint veya side effect'i bul.
- Published eski version ve in-flight instance davranisini tanimla.
- Unknown config ve downgrade davranisini belirt.

### 5. Async ve projection

- Event producer/consumer ve event version etkisini listele.
- Retry, duplicate, ordering ve DLQ davranisini yaz.
- Projection rebuild veya repair gereksinimini belirt.

### 6. Compatibility ve rollout

- V1 wire/data davranisi korunuyor mu belirt.
- Expand/backfill/switch/contract adimlarini yaz.
- Writer cutover birimini domain + tenant olarak tanimla.
- Shadow, canary, telemetry ve rollback switch'ini belirt.

## Risk siniflari

High:

- Tenant isolation, auth/RBAC, schema evolution, money, destructive data.
- Published form/workflow semantic degisikligi.
- Public API veya event breaking change.

Medium:

- Yeni projection, cache, query operator, worker veya provider entegrasyonu.

Low:

- Wire ve storage davranisini degistirmeyen lokal refactor.

## Cikis formati

```text
Owner:
Affected contracts:
Affected data/migrations:
Tenant and permission risk:
Events/jobs:
V1 compatibility:
Rollout and rollback:
Required tests:
Open decisions:
```

Planlanan V2 mimarisindeki acik kararlar kapanmadan Phase 1 source koduna
baslama. Yeni bir karar business sonucunu degistiriyorsa accepted ADR veya
acik kullanici karari olmadan implementation yapma.
