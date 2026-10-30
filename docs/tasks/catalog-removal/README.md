# Harpy'den ForgeCRUD Catalog Sisteminin Çıkarılması

## Karar ve amaç

Harpy içindeki Catalog business modülü, Harpy'nin genel backend platform kabiliyeti değildir. Bu modül; ForgeCRUD ürününün HR, Procurement, Sales CRM, Accounting, Inventory, Quality, Production, Maintenance ve Project gibi ürün modüllerini imzalı paketler halinde tanımlama, bağımlılıklarını çözme, tenant'a kurma/güncelleme, component owner'lara uygulatma ve menüye yansıtma sistemidir.

Bu paket dağıtım/kurulum sistemi ForgeCRUD ürün katmanına aittir. Harpy'nin bağımsız backend modül mimarisinde yer almamalıdır. Amaç, ForgeCRUD'a özgü Catalog sistemini Harpy'den tüm dikey ve yatay bağımlılıklarıyla çıkarmaktır.

**Önemli sınır:** HR, CRM, Inventory vb. ForgeCRUD ürün modülleri Harpy'ye taşınmayacak. Catalog kaldırılırken bunların Harpy içinde yeni business module olarak oluşturulması, paket sistemi yerine başka bir generic module loader kurulması veya Catalog'un farklı adla yeniden yazılması bu işin kapsamında değildir.

## Mevcut durumun gözlem anı

Envanter GitHub main ağacından 30 Ekim 2026 tarihinde çıkarıldı. Mevcut ağaçta:

- core/internal/modules/catalog/ altında 646 tracked file,
- Catalog modülünün kendi migration ağacında 54 SQL migration dosyası,
- API composition ve Worker composition içinde ayrı Catalog runtime wiring,
- IAM ve Analytics tarafında Catalog component-owner/projection entegrasyonları,
- contracts/, cmd/setup/, tools/localrelease/, attestation, deploy evidence ve runtime config'e yayılmış signed package release yüzeyi bulunuyor.

contracts/catalog/firstparty-parity.v1.json içindeki ForgeCRUD parity verisi dokuz ürün paketini listeliyor: hr, procurement, sales-crm, accounting, inventory, quality, production, maintenance ve project. Aynı kayıt form, workflow, menu, portal, analytics ve demo seed adetlerini de taşıyor. Bu veri Harpy'nin çekirdek modül envanteri değildir.

Bu sayılar başlangıç envanteridir; uygulama fazı başlamadan önce güncel main üzerinde tekrar sayılmalı ve farklar raporlanmalıdır.

## Okuma sırası

1. [Mevcut kod envanteri ve kaldırma haritası](01-mevcut-durum-ve-kaldirma-haritasi.md)
2. [Hedef sınırlar ve veri kararı](02-hedef-sinirlar-ve-veri-karari.md)
3. [Agent iş emirleri ve faz sırası](03-agent-is-emirleri.md)
4. [Kabul kriterleri ve recovery](04-kabul-ve-recovery.md)
5. [Korunacaklar / kaldırılacaklar matrisi](05-koruma-ve-kaldirma-matrisi.md)

## Çalışma ilkeleri

- Bu dosya seti bir uygulama planıdır; bu planı oluşturmak Catalog kodunu silmek veya DB nesnelerini düşürmek için izin değildir.
- Önce salt okunur envanter, call graph ve gerçek consumer analizi yapılır. Tahminle dosya silinmez.
- Catalog'a ait olmayan her catalog adı ayrı bir kavram olarak sınıflandırılır. Genel arama sonucunu toplu silme listesi olarak kullanma.
- IAM ve Analytics gibi gerçek owner modüllerinden Catalog bağımlılıkları temizlenmeden Catalog modülünün kaynağı kaldırılmaz.
- Shared contract aggregate'leri, permission katalogları, OpenAPI, storage owner ve migration manifestleri yalnız ilgili modüller kapandıktan sonra tek bir entegrasyon sahibi tarafından güncellenir.
- Mevcut tenant/platform verisi hakkında karar verilmeden destructive SQL, DROP veya veri temizliği yapılmaz.
- Local named volume, kullanıcının mevcut DB'si veya production verisi otomatik sıfırlanmaz.
- Migration baseline/Goose çalışması aynı dosya alanlarını etkiliyorsa iki iş paralel yürütülmez; ortak migration yüzeyinin sahibi ve sırası açıkça belirlenir.
- Tamamlanma yalnızca build'in geçmesi değildir: API, worker, migration, authorization, owner, release/setup ve fresh-install akışlarında Catalog kalıntısı olmamalıdır.

## Nihai sonuç

Harpy; IAM, Tenancy, Billing, Work, Documents, Notifications, Analytics, Audit, Usage, Platform Authority, Platform Operations, Support ve System Resource Layout gibi kendi backend kabiliyetleriyle çalışmaya devam eder. ForgeCRUD paket registry, module install/upgrade lifecycle, package dependency resolution, Catalog menu, demo package handoff ve bu özelliklere özel component projection/owner bağları Harpy'den çıkarılır.

Genel amaçlı async catalog, PostgreSQL pg_catalog, REST Query field catalog, IAM permission/resource catalog, Billing commercial catalog, System Resource Layout catalog ve migration/storage catalog altyapıları korunur. Kesin koruma listesi 05 numaralı matriste tanımlıdır.
