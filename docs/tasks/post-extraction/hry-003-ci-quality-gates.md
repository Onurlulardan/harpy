# HRY-003 — CI kalite kapılarını gerçek kabul kriterleriyle eşitle

**Öncelik:** P0  
**Durum:** PR kapısı güncellendi. `go vet ./...` unit job'unda, integration-tag derlemesi ayrı job'da, PostgreSQL 18 baseline cleanup job'u yerinde. Tam integration koşusu ve `docker build -f ops/production/Dockerfile .` image smoke'u her PR'a bağlanmadı; nasıl çalıştırılacakları mimari test belgesinde yazılı.

## Bulgu
Mevcut CI; `go test ./...`, FileGuard, ArchGuard ve PostgreSQL 18 baseline cleanup kontrolü çalıştırıyor. Buna karşılık migration-baseline kabul belgesi `go vet ./...` ve daha geniş integration doğrulamalarını da bekliyor. `integration` tag derlemesinin normal CI dışında kalması HRY-001'deki problemi görünmez bırakmış.

## Yapılacaklar
- `go vet ./...` kontrolünü CI'a ekle.
- HRY-001 kapsamındaki integration-tag compile kontrolünü ekle.
- PostgreSQL 18 fresh platform/tenant baseline ve cleanup kontrollerinin CI'da çalıştığını koru.
- Docker image build ve deployment smoke testleri için ayrı, tekrarlanabilir bir job tasarla; ilk aşamada bütün pahalı E2E testlerini her PR'a yükleme.
- CI ve mimari dokümanlarının job listesini birbirleriyle tutarlı hale getir.

## Kabul kriterleri
- CI, normal Go testleri yanında vet, architecture/file guard ve integration compile kontrolünü kapsıyor.
- PostgreSQL 18 baseline cleanup testi otomatik çalışmaya devam ediyor.
- PR kontrol adları ve kapsamları README/architecture dokümanlarıyla uyumlu.
- Bir test sınıfı CI dışında bırakılıyorsa gerekçesi ve nasıl çalıştırılacağı belgelenmiş.

## İlgili dosyalar
- [CI workflow](https://github.com/Onurlulardan/harpy/blob/main/.github/workflows/ci.yml)
- [Test ve mühendislik standartları](https://github.com/Onurlulardan/harpy/blob/main/docs/architecture/10-test-ve-muhendislik-standartlari.md)
- [Migration baseline kabul kriterleri](https://github.com/Onurlulardan/harpy/blob/main/docs/engineering/migration-baseline-rework/04-acceptance-and-recovery.md)
