# Task 05 — Platform hardening regression testleri ve operasyon docs'u

## Amaç

CI düzeltmelerinin, setup idempotency'nin ve management listener güvenliğinin gelecekte geriye dönük bozulmasını engellemek.

## Yapılacaklar

1. CI için reproducible local commands ve CI workflow referanslarını ekle. Fixture ve generated outputs'ın kaynağı açık olsun.
2. Setup test matrisi ekle:
   - ilk setup;
   - aynı config ile normal ikinci setup;
   - `--rotate-secrets` ile explicit rotation;
   - yarım setup sonrası retry;
   - backup/restore;
   - eski artifact/key material ile uyumluluk.
3. Management listener için local setup ve container deployment testleri ekle; beklenen bind address ve health-check erişimi belgeli olsun.
4. Generated-artifact check için “generate then git diff --exit-code” benzeri doğru canonical komutu belirle. Gerçek generator akışı doğrulanmadan örnek komutu körlemesine kopyalama.
5. CI regression gate'leri güvenlik kontrolünü korusun; sırf pipeline yeşil olsun diye test skip veya guard gevşetme olmasın.
6. Operational docs'ta secret rotation, backup/restore, management port exposure ve recovery fixture generation prosedürlerini yaz.
7. Her görev için test evidence, risk kalanları ve rollback/recovery adımları raporlansın.

## Kabul kriterleri

- Regression testleri yeni davranışları otomatik doğrular.
- Setup/rotation ve management network erişimi docs ile testlerde tutarlı.
- Generated artifact drift CI'da yakalanır.
- Her unresolved risk açıkça listelenir.
