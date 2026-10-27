---
name: cutover-compatibility
description: Design or implement backend-v2 V1 compatibility, shadow reads, logical adoption, writer fencing, cohort rollout, reverse reconciliation, contract retirement, or mixed-domain cutover. Use for every migration edge and production authority switch.
---

# Cutover Compatibility

## Zorunlu okuma

1. `docs/v1-compatibility.md`
2. `docs/adr/0013-v1-cutover.md`
3. `docs/adr/0022-cutover-authority-and-adoption.md`
4. Etkilenen domain ADR ve skill'i
5. SQL/role varsa `skills/data-tenancy/SKILL.md`
6. Job/event varsa `skills/events-jobs/SKILL.md`
7. Phase 14 tenant/owner/license/recovery adoption'i varsa ADR-0116..0127

## Baslamadan once

- V1 route, storage, event, job, client ve operator consumer'larini envanterle.
- Scope/domain owner'ini, cutover group'unu ve dependency set'ini yaz.
- Source/target object manifest'i, transform hash'i ve acceptance checksum'ini dondur.
- Forward switch, soak, reverse reconcile ve rollback closure kosullarini yaz.
- Intentional compatibility farkini merkezi dokumana kaydet.

## Authority kurali

- Scope/domain basina bir writer vardir; route flag'i tek basina authority degildir.
- Business, adopter, decoder ve migrator ayri credential/lease siniridir.
- Fence row'u statement boyunca kilitlenir; missing state fail-closed olur.
- Target apply yalniz typed procedure ve persisted operation manifest'iyle olur.
- Job, event, token ve direct command current generation'i effect oncesi dogrular.
- Cross-domain command tum dependency row'larini canonical sirada kilitler.

## Shadow ve adoption

- Initial copy exported snapshot'a; delta built-in logical decoding'e baglanir.
- Target mutations, change receipts ve checkpoint tek transaction'da commit olur.
- Comparator redacted normalization ve ayni RR snapshot eligibility'si kullanir.
- Blob/job/future timer gibi SQL disi state owner-specific checkpoint ister.
- WAL, lag, mismatch veya ownership preflight asimi writer'i V1'de tutar.

## Migration edge kurali

Compatibility edge final domain API'sinden once kurulmaz. Edge canonical V2
domain modeline ikinci storage/transport yolu eklemez; persisted hit counter,
owner, expiry/exit gate ve automatic CI removal tasir. Yeni tenant V1 schema veya
V1 writer ile olusmaz.

## Tamamlama kapisi

- Direct-DML, missing-row, lock race ve stale-generation negative testleri var.
- Snapshot/delta crash-resume, checksum ve failover-slot rehearsal'i var.
- Future job/event/blob inventory'si sifirlanmadan switch testi gecmiyor.
- Frontend/MCP consumer capability ve intentional-difference sonucu kayitli.
- Reverse reconcile, rollback SLO ve contract-retirement testi var.
