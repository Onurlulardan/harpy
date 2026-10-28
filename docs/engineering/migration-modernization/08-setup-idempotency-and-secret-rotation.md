# Task 08 — Setup tekrar çalıştırıldığında secrets ve key material korunuyor mu?

## Risk — henüz tam doğrulanmamış

İlk incelemede cmd/setup/runtime_material.go, runtime_identity.go, runtime_services.go, runtime_authority.go ve release_material.go yollarında her setup çalıştırmasında yeni secret/key material üretilebildiği görüldü. Buna karşılık docs/harpy-starter/00-nedir.md ikinci setup çalıştırmasının secret'ları koruduğunu ve rotation için --rotate-secrets gerektiğini söylüyor.

Bu bir operasyonel risk hipotezidir; tüm anahtarların her koşulda değiştiği veya bütün etkilerin oluştuğu kesinleşmiş değildir. Gerçek persistence akışını test et.

## İncelenecekler

- cmd/setup/runtime_material.go
- cmd/setup/runtime_identity.go
- cmd/setup/runtime_services.go
- cmd/setup/runtime_authority.go
- cmd/setup/release_material.go
- --rotate-secrets parsing/dispatch ve setup orchestration.
- docs/harpy-starter/00-nedir.md ve setup docs.
- Mevcut setup testlerinin hangi secret/key türlerini kapsadığı.

## Yapılacaklar

1. Her secret/key türü için üretim yeri, saklandığı yer, tekrar setup'ta okuma davranışı, rotation koşulu ve tüketicileri tabloya dök.
2. Testte ilk setup sonrası kalıcı material fingerprint'lerini al; ikinci normal setup'ta aynı kalması beklenenler aynı kalmalı. Secret değerlerini loglama.
3. --rotate-secrets için beklenen değişiklik matrisini tanımla. Sadece rotate edilmesi gereken material değişmeli; token/keyring/trust bağımlılıkları planlanmalı.
4. Normal setup idempotency ve başarısız/yarım setup sonrası retry senaryosunu test et.
5. Eski FCART artifact'ları, IAM token'ları, cursor'lar ve release signing/trust material davranışını test et; etkiyi kanıtla.
6. Docs ve implementation uyuşmuyorsa sözleşmeyi netleştirip ikisini uyumlu hale getir.

## Kabul kriterleri

- İkinci normal setup persistence sözleşmesini test ediyor.
- Rotation etkileri açıkça belgelenmiş ve test edilmiş.
- Secret değerleri loglanmıyor.
- Restore/retry ve eski artifact/token senaryoları kapsanmış.
