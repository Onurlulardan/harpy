# HRY-004 — Açık kaynak lisansını belirle ve repository'ye ekle

**Öncelik:** P0  
**Durum:** Kapatıldı. Lisans MIT olarak seçildi. Harpy açık kaynak bir fork tabanıdır; MIT ticari kullanıma, değiştirmeye ve özel türev dağıtmaya izin verir. Kök `go.mod` require bloğunda GPL veya AGPL modül görülmedi. Doğrudan bağımlılıklar Apache-2.0, MIT ve BSD ailesindedir. Bu bir hukuki denetim raporu değildir.

## Bulgu
Harpy README'si şu anda repository'de lisans dosyası yayımlanmadığını açıkça belirtiyor. Repository'nin public olması, tek başına yeniden kullanım, değiştirme veya dağıtım izni sağlamıyor.

## Yapılacaklar
- Harpy'nin kullanım amacına göre lisans seçeneklerini değerlendir: izin verilen ticari kullanım, patent hükümleri, türev çalışmalar ve dağıtım yükümlülükleri.
- Üçüncü taraf dependency ve taşınmış kodların lisans uyumunu kontrol et.
- Seçilen lisans metnini `LICENSE` dosyasına ekle.
- README'deki lisans bölümünü gerçek karara göre güncelle.

## Kabul kriterleri
- Repository'de açık ve tutarlı bir `LICENSE` bulunuyor.
- README lisansla çelişmiyor.
- Seçilen lisansın Harpy'nin fork edilip yeni projelerde kullanılma amacıyla uyumu değerlendirilmiş.

**Not:** Bu task teknik bir dosya eklemekten ibaret değil; lisans seçimi bilinçli bir proje/iş kararı olmalı.

## İlgili dosya
[README — Proje kapsamı ve lisans](https://github.com/Onurlulardan/harpy#proje-kapsami-ve-lisans)
