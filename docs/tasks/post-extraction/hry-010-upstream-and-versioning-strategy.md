# HRY-010 — Harpy upstream güncelleme ve sürümleme modelini belirle

**Öncelik:** P2  
**Durum:** Planlanacak.

## Amaç
Harpy'den fork edilen projelerin ileride Harpy'deki güvenlik ve altyapı iyileştirmelerini kontrollü biçimde alabilmesi.

## Yapılacaklar
- Harpy için tag/release ve changelog stratejisi belirle.
- Fork'larda Harpy upstream değişiklikleri ile proje-specific değişikliklerin ayrı commit/branch alanlarında tutulmasını öneren workflow tanımla.
- Upstream merge/rebase sırasında conflict çözüm ve test sırasını belgele.
- HTTP/event/runtime/storage contract değişikliklerinin compatibility politikasını tanımla.
- Migration değişikliklerinin fork'a taşınma ve sıralanma kurallarını açıkla.
- Güvenlik düzeltmeleri için ayrı duyuru/patch release akışı belirle.

## Kabul kriterleri
- Harpy'nin her kullanılabilir sürümü tag ve değişiklik notlarıyla tanımlı.
- Bir fork, proje-specific modülleri koruyarak upstream güncellemesi alabiliyor.
- Contract ve migration değişikliklerinin breaking/non-breaking sınıflandırması var.
- Güncelleme sonrası çalıştırılacak minimum CI ve E2E kontrol listesi belgelenmiş.
