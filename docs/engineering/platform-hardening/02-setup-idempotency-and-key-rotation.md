# Task 02 — Setup idempotency ve secret/key rotation garantileri

## Neden

Önceki incelemede setup'ın birden fazla runtime identity/key material türünü yeniden üretme ihtimali ile docs'ta belirtilen “normal ikinci setup secrets'ı korur; rotation için `--rotate-secrets` gerekir” sözleşmesi arasında olası tutarsızlık görüldü. Bu henüz bütün anahtarların her koşulda değiştiğinin kanıtı değildir. Fakat doğrulanması gereken yüksek öncelikli operasyonel bir konudur.

## İncelenecek alanlar

- `cmd/setup/runtime_material.go`
- `cmd/setup/runtime_identity.go`
- `cmd/setup/runtime_services.go`
- `cmd/setup/runtime_authority.go`
- `cmd/setup/release_material.go`
- `--rotate-secrets` parsing, dispatch ve setup orchestration.
- `docs/harpy-starter/00-nedir.md` ve ilgili setup/restore docs.

## Yapılacaklar

1. Her secret/key için tablo hazırla: tür, üretildiği fonksiyon, kalıcı storage yolu, tekrar setup'ta okuma/üretme davranışı, rotation koşulu ve tüketiciler.
2. Test ortamında ilk setup'tan sonra kalıcı material'ın secret içeriğini göstermeyen fingerprint'lerini kaydet. Normal ikinci setup sonrası hangilerinin aynı kalması gerektiğini sözleşmeye bağla.
3. Normal setup'ı tekrar çalıştır ve fingerprint'leri karşılaştır. Test ortamı gerçek bir production key store'u veya kullanıcı verisi içermemeli.
4. `--rotate-secrets` için expected-change matrix oluştur. Yalnız açıkça rotate edilmesi gereken material değişmeli; ilişkili trust roots, keyrings, token/cursor ve artifact etkileri açıklanmalı.
5. Yarım kalmış setup sonrası retry, backup/restore ve eski material ile oluşturulmuş artifact'ların okunması senaryolarını test et.
6. Mevcut kod gerçekten her çalıştırmada yeni key üretiyorsa idempotent persistence'ı düzelt; sadece dokümantasyonu gerçeğe uydurmakla yetinme, güvenli ürün davranışını değerlendir.
7. Secret veya private key değerlerini test output'una/loglara yazma; yalnız non-reversible fingerprint kullan.

## Kabul kriterleri

- Normal setup idempotency testleri tüm kritik key/secret kategorilerini kapsar.
- Rotation açıkça istenmedikçe kalıcı identity/trust material beklenmedik şekilde değişmez.
- Rotation davranışı, etkileri ve recovery adımları docs'ta yazılı.
- Eski artifact/token uyumluluğu ve rollback etkileri test edilir.
