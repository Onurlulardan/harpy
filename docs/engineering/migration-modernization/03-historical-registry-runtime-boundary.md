# Task 03 — Historical release manifest arşivini runtime'dan ayırmayı araştır

## Problem

core/internal/platform/migration/load_release_registry.go içindeki embed, inceleme commit'inde toplam 115,990,137 byte olan 109 full manifesti binary'ye alıyor. LoadReleaseRegistry() her çağrıda tüm dosyaları okuyor, artifact doğrulaması yapıyor ve entry/document contract bilgilerini registry'ye yüklüyor.

Tüm geçmiş manifestlerin gereksiz olduğunu varsayma: eski migration history satırları, offline upgrade, rollback ve attestation doğrulaması için gerekli olabilir.

## İlgili alanlar

- core/internal/platform/migration/load_release_registry.go
- core/internal/platform/migration/release_documents.go
- core/internal/platform/migration/catalog.go
- core/internal/platform/migration/history_supersession.go
- core/internal/platform/schemacontract/load_embedded.go
- core/internal/composition/operator/ ve core/internal/composition/migrator/ çağrı yerleri.
- Release packaging, restore ve offline setup dokümanları.

## Yapılacaklar

1. Her LoadReleaseRegistry() çağrısı hangi invariant'ı doğruluyor, hangi veriyi kullanıyor ve hangi lifecycle'da çalışıyor belirle.
2. Veriyi sınıflandır: current release manifesti; historical full manifests; eski path/digest eşleşmeleri; history supersession için minimum bilgi; document-contract authorization için minimum bilgi.
3. Şu seçenekleri ölçülebilir biçimde karşılaştır: full registry runtime'da; build/package aşamasında doğrulanan kompakt compatibility index; full archive immutable release artifact olarak ve gerektiğinde yüklenir.
4. Kompakt index'in neyi kanıtlayıp neyi kanıtlayamadığını tanımla. Digest listesi SQL içeriği veya signed attestation doğrulamasının yerine geçmesin.
5. Prototype gerekirse mevcut authorization modelini korusun. Eski manifest gerektiğinde nasıl bulunacak (offline package/local archive/artifact store) açık olsun.
6. Fresh install, desteklenen eski release'ten upgrade, tenant migration, restore, rollback ve history supersession senaryolarını test et.
7. Eski formatı dual-read/compare ile doğrulamadan manifest silme; cutover ve rollback planı oluştur.

## Karar (bu geçiş)

Üretim embed'i yerinde bırakıldı. `Authorize` kanonik bayt eşitliği ister. `ops/production/Dockerfile` son imaja yalnız `/out` binary'lerini koyar; release dizini yoktur. Geçmiş manifestleri embed'den çıkarmak, imaj paketlenmeden offline upgrade'i fail-closed kırardı.

Kanıt: `LoadReleaseManifestFS` aynı dosyaları diskten okur. `TestFilesystemReleaseMatchesEmbed` digest listesini, kanonik baytları, document contract kümesini ve en az bir eski digest için `Authorize` sonucunu embed ile karşılaştırır. Bilinmeyen digest reddedilir.

`-ldflags="-s -w" -trimpath` ile API ve worker, `LoadReleaseRegistry` çağırmadığı için 109 arşivi linker'dan düşürür. Migrator ve operator çağırdığı için tutar. Ölçüm `baseline.md` içindedir. `harpy_measure_no_release_embed` yalnız ölçüm etiketidir; üretim build'i kullanmaz.

## Kabul kriterleri

- Historical manifestlerin her call site'taki gerçek amacı belgelenmiş.
- Authorization, digest, signature/attestation ve history semantiği testlerle korunmuş.
- Binary size ve registry load süresi önce/sonra ölçülmüş.
- Offline deployment ve upgrade davranışı belgelenmiş.
- Gerekli tarihçe korunmadan hiçbir manifest silinmemiş.
