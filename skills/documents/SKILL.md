---
name: documents
description: Build backend-v2 document folders, immutable versions, upload and malware-scan flows, encrypted blob storage, links, downloads, archives, trash, retention, legal hold, purge, quota, and V1 document migration. Use for every document metadata or binary lifecycle change.
---

# Documents

State machine, route, permission, encryption, limit, retention ve cutover
degerleri icin ADR-0017, ADR-0027 ve ADR-0028; purge/reference icin ADR-0026;
environment recovery inventory/copy/locator-remap icin ADR-0123 ve handoff icin ADR-0124
authoritative'dir;
uygulamadan once tamamini oku. Endpoint icin `api-contract`, SQL ve tenant
degisikligi icin `data-tenancy`, job/outbox icin `events-jobs` skill'ini de oku.

## Sahiplik siniri

- `documents` bir core business module'dur; ayri deployable acma.
- Folder, item, version, upload ve link tablo/migration'larinin tek owner'i olsun.
- Byte'lari yalniz platform blob portundan private, versioning-enabled object
  store'a yaz; DB'yi veya gecici local disk'i ikinci byte kaynagi yapma.
- Forms ve diger moduller tabloya SQL yazmasin; Documents public command/query
  API'sini cagirsin ve kendi subject permission'ini once uygulasin.
- Provider SDK tipini domain/application kontratina sizdirma.
- Object network islemi ile tenant DB transaction'ini ayni acik transaction'da
  tutma; aradaki sonucu persisted state ve durable cleanup ile yonet.
- Recovery executor tabloya direct SQL yazmasin; inventory, DR hold ve old->new
  VersionId locator map'i yalniz Documents-owned restricted routine'den gecsin.

## Use-case dosyalama

Her operasyonu application, HTTP adapter ve PostgreSQL adapter'da ayri dosyala:

```text
create_upload.go, put_upload_content.go, cancel_upload.go
create_folder.go, get_folder.go, list_folders.go, update_folder.go
move_folder.go, delete_folder.go
get_document.go, list_documents.go, update_document.go, delete_document.go
move_document.go, copy_document.go, restore_document.go, purge_document.go
create_link.go, delete_link.go, create_archive.go, get_archive.go
download_version.go, download_current.go
```

- Desteklenmeyen operasyon icin bos katman veya generic CRUD abstraction'i acma.
- Her davranis dosyasina ayni operasyon adli test yaz.
- `routes.go` yalniz route baglasin; authorization veya lifecycle karari tasimasin.
- Dosya/CRUD/isim sinirinda `skills/code-organization/SKILL.md`yi uygula.

## Lifecycle uygulama sirasi

1. Actor ve tenant'i dogrula; route permission'i ve subject capability'yi ayir.
2. Input'u normalize et; name, size, MIME, magic signature ve quota'yi fail-closed
   validate et.
3. Stable ID ve idempotency anahtariyla persisted attempt/state olustur.
4. Stream, hash, scan, encryption ve object write'i ADR sirasi ile calistir.
5. Yalniz dogrulanmis object sonucunu version/item/audit/outbox ile tek tenant
   transaction'inda `ready` yap.
6. Failure, cancel, expire ve purge'u idempotent job/checkpoint ile sonlandir.

Yalniz `ready` version indirilebilir veya subject'e baglanabilir. Client filename,
path, MIME veya provider key'i object kimligi yapma. Download'da current permission'i
yeniden kontrol et; plaintext/public URL veya actual object key dondurme.

## Nihai kontrat kapisi

- Local spool, public bucket, presigned plaintext URL, DB byte storage, sync ZIP,
  best-effort orphan cleanup veya paralel metadata authority ekleme.
- V1 uyumlulugunu core lifecycle'a dal olarak ekleme. Compatibility edge; persisted
  state, owner, telemetry ve ADR'deki otomatik exit gate'iyle sinirli olsun.
- "Simdilik farkli key/scan/storage, sonra tasiriz" yaklasimini reddet.
- Migration sonucu writer switch'ten once inventory, checksum ve tenant fence ile
  kanitlansin; V1 ve V2 ayni anda cleanup authority olmasin.

## Verification

- [ ] Owner table/migration/API siniri ve yasak cross-module SQL testi var.
- [ ] Her CRUD, upload, archive, download ve lifecycle islemi ayri dosya/testte.
- [ ] `code-organization` tamamlama kapisi green.
- [ ] Duplicate/restart, partial stream, crash ve cleanup retry testleri geciyor.
- [ ] Hash/MIME/scan, encryption tamper/rotation ve private bucket testleri geciyor.
- [ ] Cross-tenant ID/key/link, permission revoke ve legal-hold testleri geciyor.
- [ ] Trash/restore/purge/quota/archive state-machine limitleri dogrulaniyor.
- [ ] V1 inventory, wire fixture, cutover fence ve rollback kaniti mevcut.
- [ ] Recovery-set snapshot/hold/ciphertext/Merkle/locator-remap testleri mevcut.
