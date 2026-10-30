# Modül — Documents

## 1. Owner alanı

Documents, tenant'a ait dosya/dizin metadata'sı, upload state, file version, archive, download, link, legal hold ve content retention işlemlerinin owner'ıdır.

Kaynak kök: core/internal/modules/documents/

## 2. Domain modeli

| Entity | Sorumluluk |
| --- | --- |
| Folder | Parent/owner, display name, active/trashed lifecycle |
| Item | Document identity, folder, current version ve trash/purge state |
| Version | Immutable file declaration, object receipt, scan evidence |
| Upload | Receiving/scanning/ready state ve upload attempt |
| Link | Owner tarafından verilen typed document link |
| Archive | Archive artifact ve retention state |
| Legal Hold | Purge üzerinde koruyucu constraint |

Folder rename/move/trash/restore işlemleri expected row revision üzerinden yapılır. Folder kendisini parent seçemez. Item ve version UUIDv7 identity kullanır.

## 3. File policy

Maximum upload boyutu 100 MiB'tır. Desteklenen formatlar:
- PDF, TXT, CSV, JSON
- JPEG, PNG, GIF, WebP
- DOC/DOCX, XLS/XLSX, PPT/PPTX
- ODT, ODS, ODP

Extension, declared MIME ve detected file kind birbiriyle eşleşmelidir. Yalnız extension kontrolü yeterli değildir. Filename ve size validation domain katmanında uygulanır.

## 4. Upload state machine

Upload state:
- created
- receiving
- retryable
- scanning
- ready
- rejected
- conflicted
- cancelled
- expired

Upload kaydı tenant/actor/document/candidate version kimliklerini, expected bytes, content hash, request hash, attempt generation ve cleanup evidence'i taşır.

Typical flow:
1. Upload command normalize edilir; idempotency request digest oluşturulur.
2. Tenant scope ve permission bind edilir.
3. Billing usage reservation alınır.
4. Upload object receipt ve candidate version identity oluşturulur.
5. Stream byte count ve content SHA-256 doğrulanır.
6. Scanner content'i inceler.
7. Ciphertext object private S3 store'a yazılır ve object receipt doğrulanır.
8. Candidate version ready state'e alınır ve document current version güncellenir.
9. Hata/retry halinde cleanup requirement ve receipt tutulur.

Ready olmayan version downloadable değildir.

## 5. Object storage

Object key tenant/document/version identity üzerinden deterministic üretilir:

tenants/{tenant-key}/documents/{document-id}/versions/{version-id}/content

Metadata DB'de, content S3-compatible storage'da tutulur. S3 adapter AWS SDK for Go v2 kullanır. Initialization:
- Bucket versioning enabled olmasını,
- Private ACL davranışını,
- Configured expected owner varsa owner eşleşmesini
doğrular.

Bucket version ID ve content digest object receipt içinde tutulur; key tek başına content identity sayılmaz.

## 6. Encryption model

Blob encryption AES-256-GCM kullanır. İçerik chunk'lar halinde stream edilir:
- 32-byte DEK
- 12-byte GCM nonce
- Per-chunk index ve plaintext size
- Header/metadata ile authenticated additional data
- Plaintext ve ciphertext SHA-256

Tüm dosya tek memory buffer'a alınmaz. Object receipt byte count ve digest'i doğrulamaya yarar.

DEK, KEK altında AES-GCM ile wrap edilir. Wrap context tenant/document/version purpose gibi identity'leri authenticated header içine bağlar. Rewrap işlemi eski KEK ile açılan DEK'i yeni KEK altında tekrar korur; DEK geçici buffer'ı işlem sonunda temizlenir.

## 7. Malware scanning

ClamAV adapter constructor sırasında endpoint'i doğrular ve server version/signature bilgisine probe yapar. Scanner signature freshness sınırı 24 saattir. Upload scan zINSTREAM protokolüyle stream edilir ve bounded frame/response limitleri uygulanır.

Scanner unavailable/stale/protocol error durumları upload'ı ready yapmaz. Scan evidence içinde signature revision ve timestamp saklanır.

## 8. Item retention ve legal hold

Item state:
- active
- trashed
- purge_cooling
- purging
- purged

Domain sabitleri:
- Trash retention: 30 gün
- Purge cooling: 24 saat
- Tombstone keep: 365 gün
- Step-up evidence maximum age: 10 dakika

Purge isteği reason ve yakın zamanda gerçekleşmiş step-up evidence ister. Cooling süresi içinde purge iptal edilebilir. Legal hold bulunan item purge edilemez. Manual ve retention purge ayrı domain path'leridir.

## 9. Archive/download ve worker

Archive, archive download, copy, deletion, purge, upload reconciliation ve related work asynchronous olabilir. Worker processor'ları Documents application portlarını kullanır; HTTP request içinde uzun süreli object operation yürütülmez.

Download öncesi:
- Tenant identity ve permission doğrulanır.
- Item current version ve ready state kontrol edilir.
- Object receipt/version/hash doğrulanır.
- DEK unwrap ve stream decrypt yapılır.
- Client response'a yalnız gerekli metadata/stream aktarılır.

## 10. Kaynak haritası

- Domain: core/internal/modules/documents/domain/
- Application: core/internal/modules/documents/application/
- S3 adapter: core/internal/modules/documents/adapters/s3blob/
- Blob crypto: core/internal/modules/documents/adapters/blobcrypto/
- Key wrap: core/internal/modules/documents/adapters/keywrap/
- ClamAV: core/internal/modules/documents/adapters/clamav/
- HTTP/Postgres adapters: core/internal/modules/documents/adapters/
- Composition: core/internal/composition/api/build_documents_http.go, worker/build_documents_jobs.go
- Contracts: contracts/authority/documents.json, contracts/http/documents*, contracts/events/
- Skill: skills/documents/SKILL.md
