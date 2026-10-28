# Task 10 — Geçiş dokümantasyonu ve regression gate'leri

## Amaç

Migration refactor'ının determinism, release compatibility, offline upgrade ve yeniden üretilebilir artifact garantilerini koruması.

## Yapılacaklar

1. Gerçek lifecycle'ı belgeleyin: source discovery → selection/supersession → manifest canonicalization → attestation → bundle packaging → runtime authorization → Goose execution → history verification. Koddaki gerçek akış farklıysa diyagramı düzeltin.
2. Test matrisi ekleyin: fresh install; adopted migration; platform/tenant ayrımı; manifest digest/signature mismatch; unsupported/out-of-order migration; superseded historical row; desteklenen eski release'ten upgrade; offline install/restore; stale schema contract ile API fail-closed.
3. Generated manifest, schema contract, currency registry ve diğer release material için canonical generation komutlarını belgeleyin; CI'ya generated-output check ekleyin.
4. Determinism testi ekleyin: aynı input iki kez generate edilince byte-for-byte aynı output/digest.
5. Binary/startup benchmark'larını ilk aşamada flaky hard CI threshold yapmayın. Stabil baseline sonrasında toleranslı regression budget belirleyin.
6. Eski freeze/release docs ile güncel implementasyon farklarını listeleyin; tarihsel dokümanı sessizce silmeyin.
7. Her refactor için rollback yolu, eski formatı okuma süresi ve kaldırma kriteri yazın.

## Kabul kriterleri

- Upgrade/offline uyumluluk matrisi mevcut.
- Generator deterministik ve CI'da denetleniyor.
- Security/attestation/history kontrolleri regression testleriyle korunuyor.
- Benchmark baseline ve maliyet bütçesi ölçüme dayanıyor.
- Operasyon ve developer docs kodla tutarlı.
