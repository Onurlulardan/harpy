---
name: api-contract
description: Design and change backend-v2 HTTP, streaming, WebSocket, webhook, and published REST contracts with stable DTOs, auth, tenant, permission, idempotency, errors, OpenAPI, and V1 compatibility. Use for every route or wire-format change.
---

# API Contract

## Baslamadan once

1. Consumer ve use-case owner'ini belirle.
2. V1 route/wire fixture'i varsa oku.
3. Auth, tenant, permission ve idempotency kararini yaz.
4. Normal JSON mi, stream/binary/multipart mi oldugunu ayir.
5. OpenAPI contract'i uygulamadan once tanimla.

## Katmanlama

```text
route -> decoder -> application command/query -> presenter
```

- Route dosyasi yalniz method/path/middleware baglasin.
- Decoder transport validation yapsin; business karari vermesin.
- Application HTTP veya Gin tipi almasin.
- Presenter domain entity'yi stable response DTO'ya cevirsin.
- Create/get/list/update/delete handler'larini ayri dosyala.

## Zorunlu endpoint kararlari

- Method ve path.
- Actor tipi ve tenant kaynagi.
- Permission resource/action.
- Request, response ve error DTO'lari.
- Idempotency semantigi.
- Transaction ve side effect siniri.
- Pagination/query contract'i.
- Rate ve query budget'i.
- Audit action ve correlation ID.

## Error kontrati

Stable machine code kullan:

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

- Her issue tam olarak RFC 6901 `path` ve stable rule `code` tasir; `expected`,
  `pointer`, `rule` veya baska alternate member ekleme.
- Internal SQL/provider mesajini client'a verme.
- Validation, conflict, forbidden ve dependency failure'i ayir.
- Ayni hata kodunu farkli semantik icin yeniden kullanma.
- Localization'i machine code'dan ayir.

## Security

- Tenant header'ini token/actor context ile eslestir; kaynak kabul etme.
- Permission'i yalniz edge'de degil application use-case'inde de uygula.
- Write endpoint'inde DB-backed idempotency gereksinimini degerlendir.
- CSRF, cookie ve CORS davranisini auth tasarimina gore belirt.
- Field-level read/write permission'ini response ve query'ye uygula.
- Public endpoint'e explicit allowlist, rate limit ve query budget koy.

## REST Query

List, filter, search, sort, include veya aggregate icin yeni ad-hoc parser yazma.
`rest-query` skill'ini ve `docs/rest-query-contract.md` dosyasini kullan.

## V1 compatibility

- `backend-v2` adi otomatik `/api/v2` path'i gerektirmez.
- Mevcut consumer'i ayni anda rewrite etme.
- Eski syntax'i edge adapter'da canonical command/query'ye cevir.
- Breaking degisikligi ADR, migration ve contract test olmadan yapma.
- SSE, WS, binary ve multipart response'u JSON envelope middleware'ine sokma.

## Verification

- OpenAPI semantic diff.
- Decoder ve presenter unit testleri.
- Auth, tenant, permission ve idempotency integration testleri.
- Success ve her stable error icin contract test.
- V1 differential fixture, gerekiyorsa frontend-v3/MCP E2E.
- Streaming/binary icin protokole ozel test.
