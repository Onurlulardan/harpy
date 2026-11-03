# HRY-002 — Catalog removal ve migration baseline işlerini kanıtlarıyla resmen kapat

**Öncelik:** P0  
**Durum:** Kod tarafı büyük ölçüde tamamlanmış görünüyor; dokümantasyon kapanışı eksik.

## Bulgu
`docs/tasks/catalog-removal/04-kabul-ve-recovery.md` ve `docs/engineering/migration-baseline-rework/04-acceptance-and-recovery.md` içindeki Definition of Done kutuları hâlâ işaretlenmemiş. Son `main` CI çalışmasında PostgreSQL 18 cleanup kontrolü başarılı olsa da bu tek başına iki task setinin tüm kabul kriterlerinin tamamlandığını kanıtlamıyor.

## Yapılacaklar
- Catalog removal task'ındaki kalan referans taramalarını çalıştır; her sonucu kaldırılacak ForgeCRUD Catalog kalıntısı veya korunması gereken generic catalog kavramı olarak sınıflandır.
- IAM ve Analytics native davranışlarının Catalog olmadan çalıştığını doğrula.
- Platform ve tenant baseline sonrası cleanup, ikinci çalıştırmada no-op davranışı ve beklenmeyen dolu veride fail-closed davranışını kanıtla.
- Fresh platform install ve tenant provisioning sonuçlarını kaydet.
- Veri disposition kararını ve gerçek DB/volume kullanılmadığına ilişkin doğrulama notunu task raporuna ekle.
- Kabul kriterlerini yalnızca karşılık gelen test/CI/rapor kanıtı varsa işaretle.

## Kabul kriterleri
- Catalog removal ve migration baseline için tek bir final doğrulama raporu var.
- Kaldırılan Catalog kodu, owner adapter'ları, contract'lar, runtime registration'ları ve DB nesneleri açıkça listelenmiş.
- Korunan IAM, Analytics, Billing, REST Query, async ve migration generic catalog alanları regression testleriyle korunmuş.
- Fresh install ve tenant provisioning Catalog package sistemi olmadan çalışıyor.
- İki task setinin DoD listeleri kanıtlarla kapatılmış; çözülemeyen madde varsa açık blocker olarak kalmış.

## İlgili dosyalar
- [Catalog removal kabul kriterleri](https://github.com/Onurlulardan/harpy/blob/main/docs/tasks/catalog-removal/04-kabul-ve-recovery.md)
- [Catalog removal koruma/kaldırma matrisi](https://github.com/Onurlulardan/harpy/blob/main/docs/tasks/catalog-removal/05-koruma-ve-kaldirma-matrisi.md)
- [Migration baseline kabul kriterleri](https://github.com/Onurlulardan/harpy/blob/main/docs/engineering/migration-baseline-rework/04-acceptance-and-recovery.md)
- [PostgreSQL 18 Catalog cleanup testi](https://github.com/Onurlulardan/harpy/blob/main/core/internal/platform/baseline/catalog_removal_pg18_integration_test.go)
