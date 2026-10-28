# Task 01 — CI test ve fileguard hatalarını düzelt

## Başlangıç kanıtı

Önceki incelemede [GitHub Actions run 40593236189](https://github.com/Onurlulardan/harpy/actions/runs/40593236189), commit `5044757b522d95c4e0afc9df3d214d6145b15016` için test job ve `fileguard` başarısız, `archguard` başarılı görünüyordu. Bu eski bir snapshot'tır; önce güncel main'deki CI durumunu ve logları kontrol et.

## A. Recovery fixture eksikliği

- İlgili test: `ops/local/cmd/recovery-state/recovery_state_test.go`
- Raporlanan sorun: `contracts/recoverykit/v1/testdata/browser-vector.json` bekleniyor fakat dosya bulunamıyor.
- Yapılacaklar:
  1. Fixture'ın üretim kaynağını, schema'sını ve testteki beklenen semantiği bul.
  2. Dosya generator ile üretiliyorsa generator'ı çalıştır; elle fixture oluşturmak gerekiyorsa gerçek contract'a uygun, deterministic bir fixture oluştur.
  3. Testin sadece dosya varlığını değil doğru recovery/browser vector davranışını da kontrol ettiğini doğrula.
- Kabul: Temiz checkout'ta test geçer; fixture kaynağı ve yeniden üretme komutu belgelenir.

## B. Read-only temporary directory cleanup

- İlgili test: `ops/local/cmd/secret-projector/target_project_readonly_test.go`
- Raporlanan sorun: Test temp directory'yi `0500` moduna getiriyor; test cleanup child dosyaları silemiyor.
- Yapılacaklar:
  1. Testin gerçek read-only invariant'ını ve filesystem davranışını doğrula.
  2. Cleanup'ta yalnız testin sahip olduğu temp path'in izinlerini geri açıp cleanup yap; production code'un readonly güvenlik davranışını gevşetme.
  3. Cleanup başarısızlığının asıl assertion'ı maskelemediğini doğrula.
- Kabul: Test ve cleanup başarılı; read-only davranış test edilmeye devam eder.

## C. Windows-specific Go executable path

- İlgili test: `cmd/setup/release_material_test.go`
- Raporlanan sorun: Test `$GOROOT/bin/go.exe` çağırıyor; Linux runner'da bu dosya yok. Aynı package içinde `goTool()` helper'ı olduğu önceki incelemede görülmüştü.
- Yapılacaklar: Güncel kodu doğrula; mevcut portable helper'ı kullan veya tek canonical helper oluştur. OS-specific executable suffix'ini hardcode etme.
- Kabul: Linux CI'da ve desteklenen platformlarda test çalışır.

## D. Currency registry fileguard mismatch

- İlgili alanlar: `tools/currencyregistry/main.go`, `currency_registry_gen.go`, `contracts/generated.yaml`.
- Raporlanan sorun: Generated Go dosyasının SHA256 değeri manifestteki recorded output hash ile uyuşmuyordu.
- Yapılacaklar:
  1. Canonical input → generator → output → manifest hash akışını takip et.
  2. Generator'ı çalıştır, oluşan diff'i incele ve kaynak verinin doğru olduğunu doğrula.
  3. Manifesti doğru generation yoluyla güncelle; yalnızca hash değerini elle değiştirerek uyuşmazlığı örtme.
  4. Temiz checkout'ta aynı inputların aynı output'u verdiğini doğrula.
- Kabul: Generator deterministik; fileguard temiz checkout'ta geçer; komut belgelenir.

## Genel kabul kriterleri

- Her alt sorun ayrı test/log kanıtıyla kapatılır.
- Test skip edilmez, assertion silinmez ve fileguard gevşetilmez.
- Son raporda başlangıçtaki CI durumu ile düzeltme sonrası run linki yer alır.
