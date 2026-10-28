# Task 07 — CI test ve generated-artifact hatalarını düzelt

## Kanıtlanan CI durumu

İncelemede workflow run 40593236189, commit 5044757b522d95c4e0afc9df3d214d6145b15016 için test job ve fileguard başarısız, archguard başarılı görünüyordu. Agent çalışmaya başlamadan güncel branch'te run ve hataları tekrar doğrulamalı.

## A. Eksik recovery fixture

- Dosya/test: ops/local/cmd/recovery-state/recovery_state_test.go
- Hata: contracts/recoverykit/v1/testdata/browser-vector.json bekleniyor fakat bulunamıyor.
- Yapılacak: Fixture'ın canonical kaynağını ve formatını bul. Gerçek sözleşmeye uygun fixture üret veya testin yanlış varsayımını düzelt; rastgele test verisi uydurma.
- Kabul: Test geçiyor ve fixture kaynağı belgeli.

## B. Read-only temp directory cleanup

- Dosya/test: ops/local/cmd/secret-projector/target_project_readonly_test.go
- Hata: Test temp directory'yi 0500 yapıyor; cleanup child dosyaları silemiyor.
- Yapılacak: Üretim davranışını değiştirmeden cleanup sırasında izinleri güvenli biçimde geri al. Read-only invariant test edilmeye devam etsin.
- Kabul: Test geçiyor; assertion gevşetilmemiş.

## C. Windows Go executable varsayımı

- Dosya/test: cmd/setup/release_material_test.go
- Hata: Test $GOROOT/bin/go.exe çalıştırıyor; Linux runner'da yok. Aynı package içinde goTool() helper'ı olduğu raporlandı.
- Yapılacak: Portable helper kullan veya tek canonical helper oluştur.
- Kabul: Linux CI ve desteklenen diğer platformlarda test çalışıyor.

## D. Currency registry generated hash uyuşmazlığı

- Hata: fileguard, currency_registry_gen.go SHA256 değerinin contracts/generated.yaml içindeki recorded output hash ile uyuşmadığını bildiriyor.
- İlgili alan: tools/currencyregistry/main.go.
- Yapılacak: Canonical input → generator → output → manifest hash akışını belirle. Önce çıktıyı generator ile yeniden üretip diff'i incele; sonra manifesti doğru üretim yolundan güncelle. Yalnız YAML hash'ini elle değiştirerek tutarsızlığı gizleme.
- Kabul: Temiz checkout'ta generation deterministik; CI geçiyor; komut docs'ta.

## Genel kabul kriterleri

- Hatalar ayrı küçük commit/PR'larda düzeltilebilir.
- Hata öncesi/sonrası log ve komutlar kaydedilir.
- Test skip etme, assertion silme veya guard gevşetme yoluyla CI yeşile çevrilmez.
