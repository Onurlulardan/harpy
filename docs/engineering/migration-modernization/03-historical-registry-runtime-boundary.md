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

## Uygulama

`LoadReleaseRegistry` artık 109 manifesti parse etmez. Binary yalnız `release-index.json` taşır: digest listesi ve document contract birleşimi. Manifest baytları yanındaki arşivden, tek digest istenince okunur.

Kaynak dosyalar repoda durur. `go test` ve yerel çalışma, çalışma dizininden `core/internal/platform/migration/release` dizinini bulur. Distroless imaj `migration-release.zip` dosyasını binary ile aynı dizine koyar (`ops/production/Dockerfile`). Zip, executable yanında varsa o kullanılır.

Açılış yalnız index ile arşiv dosya adlarını karşılaştırır. `Authorize` ve `MigrationArtifact` tek dosyayı okur, SHA-256 dosya adıyla aynı mı bakar, attestation doğrular, canonical baytları karşılaştırır. Bilinmeyen digest ve değiştirilmiş bayt fail-closed olur.

Index drift'i `TestReleaseIndexMatchesManifests` yakalar. Yenilemek için `HARPY_WRITE_RELEASE_INDEX=1 go test ./core/internal/platform/migration/ -run TestReleaseIndexMatchesManifests`. İsim kontrolü `go run ./internal/migrationbundle/cmd/packrelease -check`.

Aynı makinede `go1.26.5` `linux/amd64` ölçümü, `BenchmarkLoadReleaseRegistry` medyanı 13985065 ns/op, 7348521 B/op, 38559 allocs/op. Önceki tam parse ölçümü 18412161195 ns/op, 8465312328 B/op, 169000537 allocs/op. `go build -trimpath -ldflags="-s -w"` migrator 45011209 byte, operator 36389129 byte. Önceki migrator 159387913, operator 150782217 byte.

## Kabul kriterleri

- Historical manifestlerin her call site'taki gerçek amacı belgelenmiş.
- Authorization, digest, signature/attestation ve history semantiği testlerle korunmuş.
- Binary size ve registry load süresi önce/sonra ölçülmüş.
- Offline deployment ve upgrade davranışı belgelenmiş.
- Gerekli tarihçe korunmadan hiçbir manifest silinmemiş.
