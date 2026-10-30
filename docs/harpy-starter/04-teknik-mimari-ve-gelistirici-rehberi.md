# Harpy — Teknik Dokümantasyon

Bu dizin Harpy'nin mevcut backend kodunu, modül sınırlarını, runtime akışlarını ve geliştirme kurallarını ayrıntılı olarak açıklar. Belge seti repository kaynak kodu esas alınarak hazırlanmıştır.

**Tam dokümantasyon dizini:** [Harpy mimari dokümantasyonu](../architecture/README.md)

## Başlangıç

1. [Sistem genel bakış](../architecture/01-genel-bakis.md)
2. [Mimari ve modül sınırları](../architecture/02-mimari-ve-modul-sinirlari.md)
3. [Runtime ve composition](../architecture/03-runtime-ve-composition.md)
4. [Platform temelleri](../architecture/04-platform-temelleri.md)
5. [HTTP, API ve kontratlar](../architecture/05-http-api-ve-kontratlar.md)
6. [Veri ve tenant izolasyonu](../architecture/06-veri-ve-tenant-izolasyonu.md)
7. [Event, job ve worker mimarisi](../architecture/07-event-job-ve-worker.md)
8. [Güvenlik modeli](../architecture/08-guvenlik-modeli.md)
9. [Deployment ve operasyon](../architecture/09-deployment-ve-operasyon.md)
10. [Test ve mühendislik standartları](../architecture/10-test-ve-muhendislik-standartlari.md)
11. [Geliştirici başlangıç rehberi](../architecture/11-gelistirici-rehberi.md)
12. [Paylaşılan capabilities](../architecture/12-paylasilan-capabilities.md)
13. [Contract kataloğu](../architecture/13-kontrat-katalogu.md)
14. [Mimari kararlar](../architecture/14-mimari-kararlar.md)
15. [Modül paket haritası](../architecture/15-modul-paket-haritasi.md)

## Modül belgeleri

- [IAM](../architecture/modules/iam.md)
- [Tenancy](../architecture/modules/tenancy.md)
- [Billing](../architecture/modules/billing.md)
- [Catalog](../architecture/modules/catalog.md)
- [Work](../architecture/modules/work.md)
- [Documents](../architecture/modules/documents.md)
- [Notifications](../architecture/modules/notifications.md)
- [Analytics](../architecture/modules/analytics.md)
- [Diğer iş ve platform modülleri](../architecture/modules/diger-moduller.md)

## Kaynak kod haritası

| Alan | Kaynak dizin |
| --- | --- |
| Süreç giriş noktaları | core/cmd/, cmd/setup/ |
| Composition root | core/internal/composition/ |
| İş modülleri | core/internal/modules/ |
| Teknik altyapı | core/internal/platform/ |
| Paylaşılan capability'ler | core/internal/capabilities/ |
| Orchestration | core/internal/orchestration/ |
| Wire ve runtime kontratları | contracts/ |
| Yerel/production operasyon dosyaları | ops/, deploy/, .taskfiles/ |
| Guard ve yardımcı araçlar | tools/ |
| Beceri/uygulama standartları | skills/ |

## Belge okuma prensibi

Bu sayfalar kaynak kodunun yerine geçmez. Bir davranışın kesin uygulamasını doğrularken ilgili modülün domain, application, api, ports, adapters ve test dosyalarını birlikte incele. skills/INDEX.md iş türüne göre okunması gereken skill dosyalarını listeler; AGENTS.md repository çalışma protokolünün sahibidir.

Repository'de frontend uygulaması bulunmaz. Yerel geliştirme altyapısı, kardeş frontend-v3 çalışma ağacını ayrıca başlatabilen operasyon görevleri içerir; bu belge seti Harpy backend'ini kapsar.
