# HRY-001 — Integration-tag testlerini derlenebilir ve CI tarafından görünür hale getir

**Öncelik:** P0  
**Durum:** Derleme kapanışı yapıldı. `catalogPackageKeyring` atamaları kaldırıldı. Silinmiş supersession zincirine bağlı integration testleri çıkarıldı. `go test -tags=integration -count=0 -run '^$' ./...` geçiyor ve CI `integration compile` job'u bunu kapıya aldı. Built entrypoint testi derleniyor. PostgreSQL 18 üzerinde startup/readiness/shutdown koşusu Docker istediği için bu kapanışta çalıştırılmadı.

## Bulgu
`core/internal/platform/migrator/built_entrypoint_gate_integration_test.go` içinde API/Worker/Migrator test profillerine `catalogPackageKeyring` alanı atanıyor. Ancak aynı paketteki `builtEntrypointProfile`, `built_entrypoint_process_integration_test.go` dosyasında bu alanı tanımlamıyor. Bu nedenle integration build tag'iyle ilgili test paketi derlenirken hata oluşması bekleniyor. Normal `go test ./...` bu dosyaları build tag nedeniyle kapsamıyor.

## Yapılacaklar
- Catalog kaldırma işleminden kalmış `catalogPackageKeyring` ve `evidence.catalog` referanslarını incele; artık kullanılmıyorlarsa temizle.
- `integration` build tag'i altındaki tüm test paketlerinin en azından derlendiğini doğrula.
- Built API, Worker ve Migrator entrypoint testlerinin gerçek PostgreSQL ortamında startup, readiness ve graceful shutdown kontrollerini yapmasını sağla.
- CI'a integration-tag compile kontrolü ekle. Gerçek servis bağımlılığı isteyen seçili integration testlerini ayrı CI job'ında çalıştır.

## Kabul kriterleri
- `go test -tags=integration -run '^$' ./...` başarılı.
- İlgili built-entrypoint integration testi PostgreSQL 18 üzerinde geçiyor.
- Worker ve Migrator test config'lerinde artık Catalog package keyring bağımlılığı bulunmuyor.
- CI, normal testlerin kapsamadığı integration-tag derleme hatalarını yakalayabiliyor.

## İlgili dosyalar
- [built_entrypoint_gate_integration_test.go](https://github.com/Onurlulardan/harpy/blob/main/core/internal/platform/migrator/built_entrypoint_gate_integration_test.go)
- [built_entrypoint_process_integration_test.go](https://github.com/Onurlulardan/harpy/blob/main/core/internal/platform/migrator/built_entrypoint_process_integration_test.go)
- [CI workflow](https://github.com/Onurlulardan/harpy/blob/main/.github/workflows/ci.yml)
