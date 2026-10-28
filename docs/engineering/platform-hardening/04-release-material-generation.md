# Task 04 — Release material üretimi ve generated artifact tutarlılığı

## Amaç

Setup/release sırasında üretilen material'ın deterministik olup olmadığını, tekrar üretildiğinde mevcut installation state'ini beklenmedik şekilde değiştirip değiştirmediğini ve generated artifact hash'lerinin canonical kaynaktan türetilip türetilmediğini netleştirmek.

## İlgili alanlar

- `cmd/setup/release_material.go`
- `cmd/setup/runtime_material.go`
- `cmd/setup/runtime_identity.go`
- `cmd/setup/runtime_authority.go`
- `tools/currencyregistry/main.go`
- `contracts/generated.yaml` ve ilgili generated source dosyaları.

## Yapılacaklar

1. “Build-time deterministic artifact” ile “runtime per-install unique secret/key” kavramlarını ayır. Her generated output'un hangi sınıfta olduğunu belgele.
2. Her release artifact için canonical input, generation command, output path, hash/signature ve verification consumer'ını tabloya koy.
3. Random key material üreten işlemler için determinism talep etme; bunun yerine persistence, explicit rotation ve backup/restore sözleşmesini test et.
4. Source-derived generated files için temiz checkout'ta generator çalıştır ve diff'i incele. Manifest hash'lerini elle güncelleme; canonical generator/check komutunu kullan.
5. Generated outputs'ın stale olup olmadığını CI'da kontrol eden bir gate oluştur. CI'da gizli production secrets veya key material üretilmesini gerektiren test yazma.
6. Build artifact ve per-install runtime state ayrımını release packaging docs'ta açıkça belirt.

## Kabul kriterleri

- Her artifact'ın source-of-truth ve regeneration komutu belli.
- Deterministic generated source ile random per-install key material ayrılmış.
- Fileguard/generation check temiz checkout'ta geçiyor.
- Runtime key persistence ve rotation testleri Task 02 ile tutarlı.
