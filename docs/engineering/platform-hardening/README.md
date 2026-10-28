# Harpy — Platform Hardening Görev Planı

Bu plan, migration-modernization görevlerinden ayrı olarak önceki kod incelemesinde tespit edilen CI, setup güvenilirliği, management endpoint erişimi ve generated-artifact sorunlarını kapsar. Branch: `docs/platform-hardening-tasks`.

## Genel çalışma kuralları

- Başlangıç bulguları `5044757b522d95c4e0afc9df3d214d6145b15016` commit'indeki incelemeye dayanıyor. Her görevde güncel `main` üzerinde kodu ve CI run'ını yeniden doğrula.
- Bir testin veya guard'ın amacını anlamadan silme, skip etme ya da gevşetme.
- Güvenlik/operasyon riski olarak işaretlenen bulguları kanıtlanmış exploit veya kesin üretim arızası diye sunma; ilgili akışla doğrula.
- Her görev ayrı küçük PR/commit olabilir. Her değişiklikte test komutlarını ve sonuçlarını yaz.
- Secret, private key, token veya gerçek credential değerlerini loglama.
- Bu görevler migration bundle refactor'ından bağımsız tutulmalı.

## Görevler

1. [CI testlerini ve fileguard hatalarını düzelt](01-ci-test-and-fileguard.md)
2. [Setup idempotency ve secret/key rotation garantilerini doğrula](02-setup-idempotency-and-key-rotation.md)
3. [Management listener'ın network exposure sınırını doğrula](03-management-listener-network-exposure.md)
4. [Release material ve generated artifact üretimini deterministik hale getir](04-release-material-generation.md)
5. [Platform hardening regression gate'leri ekle](05-regression-tests-and-operations-docs.md)

## Öncelik önerisi

- **P0:** CI'da doğrulanmış test ve fileguard hataları; temiz checkout'ta test/build güvenilirliği.
- **P1:** Setup'ın ikinci kez çalıştırılması ve key rotation semantiği; operasyonel veri/kimlik kaybı ihtimali varsa düzeltme önceliğini yükselt.
- **P1:** Management portunun gerçekten hangi arayüzlerden erişilebilir olduğunun kanıtlanması; deployment'a göre güvenli bind default'u.
- **P2:** Generation komutlarının determinism'i, docs ve regression gates.

Bu sıra geçici önceliklendirmedir. Güncel CI ve deployment modeli farklı sonuç verirse görev sahibi gerekçeli olarak değiştirebilir.
