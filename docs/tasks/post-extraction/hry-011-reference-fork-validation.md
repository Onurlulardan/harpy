# HRY-011 — Minimal reference fork ile fork modelini kanıtla

**Öncelik:** P2  
**Durum:** Opsiyonel; HRY-009 sonrasında değerlendir.

## Amaç
Dokümantasyondaki fork modelinin pratikte çalıştığını göstermek. Bu bir demo ürünü geliştirme işi değil.

## Yapılacaklar
- Ayrı bir test fork veya örnek repository oluştur.
- Harpy generic altyapısına dokunmadan küçük, proje-specific bir domain modülü ekle.
- Modül için migration, permission, HTTP contract, basic use-case ve gerekiyorsa worker job wiring'ini tamamla.
- Harpy upstream'den küçük bir değişikliği örnek fork'a taşı ve testleri tekrar çalıştır.

## Kabul kriterleri
- Örnek proje, Harpy'nin mevcut core modüllerini değiştirerek değil, belirlenen extension/ownership noktalarını kullanarak çalışıyor.
- Yeni project module clean install ve tenant scope testlerini geçiyor.
- Harpy upstream update örneği proje-specific davranışı bozmadan alınabiliyor.
