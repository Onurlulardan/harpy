---
name: events-jobs
description: Design backend-v2 transactional outbox, inbox, integration events, database jobs, workers, retries, leases, idempotency, RabbitMQ delivery, DLQ, projections, and external side effects. Use for every asynchronous or cross-process flow.
---

# Events and Jobs

State, lease, retry, retention, idempotency ve DLQ degerlerinin tek kaynagi
ADR-0011; authority generation/adoption icin ADR-0022'dir. Ikisini de oku.

## Once sinifi belirle

- Ayni transaction icinde zorunlu business invariant mi?
- Commit sonrasi internal job mi?
- Cross-process integration event'i mi?
- Rebuild edilebilir projection mi?
- External side effect mi?

Ayni process icindeki normal modul cagrisini RabbitMQ event'ine cevirme.

## Transactional outbox

Business mutation ile ayni transaction'da integration outbox satiri yaz.

Envelope en az sunlari tasir:

```text
event_id, event_type, event_version
tenant_id, aggregate_type, aggregate_id, aggregate_version, event_sequence
occurred_at, actor_id
correlation_id, causation_id, idempotency_key
payload
```

- Payload'i kucuk ve versioned tut.
- Publisher confirm olmadan delivered isaretleme.
- Publish sonrasi crash duplicate uretebilir; at-least-once kabul et.
- Consumer effect'i ile inbox receipt'i ayni transaction'da yaz.
- Unique `(consumer, event_id)` ile duplicate'i etkisizlestir.
- Ayni aggregate stream'ini contiguous `event_sequence` ile sirala;
  `aggregate_version`in contiguous oldugunu varsayma.

## Jobs

DB job source of truth olsun:

```text
kind, tenant_id, payload_ref, run_at, state
attempts, max_attempts, lease_owner, lease_until
dedupe_key, last_error, created_at, completed_at
```

- `FOR UPDATE SKIP LOCKED` ile claim et.
- Lease ve heartbeat ile stale worker'i geri kazan.
- Delay, cron, retry ve long-running import'u job olarak modelle.
- RabbitMQ'yu opsiyonel wake-up/transport yap; schedule state'i yapma.
- Job payload'ina buyuk veya secret data gommek yerine stable reference koy.

## Retry ve hata sinifi

- Transient, permanent, conflict ve cancelled hatayi ayir.
- Exponential backoff + jitter ve hard attempt limiti uygula.
- Permanent validation/auth hatasini retry etme.
- DLQ veya terminal failure'i operator-visible yap.
- Manual retry'nin yeni audit ve causation bilgisi olsun.

## External side effect

- Stable idempotency key'i provider'a tasiyabildigin yerde tasi.
- Network cagrisini database transaction icinde yapma.
- Provider response ve delivery receipt'i persist et.
- Secret'i event veya log payload'ina koyma.
- Compensation'i otomatik rollback ile karistirma; ayri command olarak tasarla.

## Projection

- Projection canonical business source olmasin.
- Event version compatibility'sini acik tut.
- Rebuild, checkpoint, lag metric ve repair command sagla.
- Projection drift'ini ornekleme veya checksum ile tespit et.

## Verification

- Commit rollback'te event/job olusmama testi.
- Publish-confirm failure ve retry testi.
- Duplicate, reorder ve consumer restart testi.
- Stale lease ve worker crash testi.
- Max-attempt/DLQ/operator visibility testi.
- Projection rebuild ve drift repair testi.
