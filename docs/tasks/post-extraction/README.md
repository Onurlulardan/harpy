# Harpy — Extraction Sonrası Olgunlaştırma Task'ları

**Durum:** Planlama backlog'u  
**Amaç:** Harpy'yi gelecekteki projelerin fork edip kendi business katmanlarıyla genişletebileceği, güvenilir ve sürdürülebilir bir backend temeli olarak hazırlamak.

## Kapsam

Harpy, ForgeCRUD Backend V2'den ForgeCRUD'a özgü business/product katmanları ayıklanarak oluşturulmuş bağımsız bir backend temelidir. Bağımsız SaaS ürünü veya müşteri edinmesi gereken ticari platform olarak ele alınmaz. Yeni projeler Harpy fork edilerek geliştirilecektir.

Bu task seti yeni özellik eklemekten ziyade extraction sonrası doğruluğu kapatmayı, operasyonel güvenilirliği kanıtlamayı ve fork/upstream çalışma modelini tanımlamayı hedefler.

## Task listesi

### P0 — Öncelikli kapanışlar
1. [HRY-001 — Integration-tag test derlenebilirliği](hry-001-integration-tag-test-compilation.md)
2. [HRY-002 — Catalog removal ve migration baseline kapanışı](hry-002-catalog-removal-and-migration-baseline-closure.md)
3. [HRY-003 — CI kalite kapıları](hry-003-ci-quality-gates.md)
4. [HRY-004 — Açık kaynak lisansı](hry-004-open-source-license.md)

### P1 — Operasyonel güvenilirlik
5. [HRY-005 — Fresh-install E2E](hry-005-fresh-install-e2e.md)
6. [HRY-006 — Backup/restore recovery provası](hry-006-backup-restore-recovery.md)
7. [HRY-007 — Security regression](hry-007-security-regression.md)
8. [HRY-008 — Observability ve alarm kabul kriterleri](hry-008-observability-and-alerting.md)

### P2 — Fork edilebilirlik ve upstream sürdürülebilirliği
9. [HRY-009 — Fork/bootstrap rehberi](hry-009-fork-bootstrap-guide.md)
10. [HRY-010 — Upstream güncelleme ve sürümleme modeli](hry-010-upstream-and-versioning-strategy.md)
11. [HRY-011 — Reference fork doğrulaması](hry-011-reference-fork-validation.md)

## Kapsam dışı kararlar
- ForgeCRUD Backend V2 üzerinde daha önce yapıldığı belirtilen kapasite/load testleri bu backlog'da yeniden planlanmaz. Orijinal raporlar ve test koşulları gerektiğinde referans olarak bağlanır.
- ForgeCRUD HR, CRM, Inventory, Production, Forms, Workflow veya AI business özellikleri Harpy'ye geri eklenmez.
- Catalog benzeri dinamik package/module installer veya marketplace geliştirilmez.
- Modular monolith gerekçesiz biçimde microservice'lere bölünmez.
- Yalnızca SDK oluşturmak için `internal` mimarisi geniş çaplı public API'ye dönüştürülmez.

## Önerilen yürütme sırası
| Sıra | Task | Öncelik |
| ---: | --- | --- |
| 1 | HRY-001 — Integration-tag test derlenebilirliği | P0 |
| 2 | HRY-002 — Catalog removal ve migration baseline kapanışı | P0 |
| 3 | HRY-003 — CI kalite kapıları | P0 |
| 4 | HRY-004 — Açık kaynak lisansı | P0 |
| 5 | HRY-005 — Fresh-install E2E | P1 |
| 6 | HRY-006 — Backup/restore recovery provası | P1 |
| 7 | HRY-007 — Security regression | P1 |
| 8 | HRY-008 — Observability ve alarm kabul kriterleri | P1 |
| 9 | HRY-009 — Fork/bootstrap rehberi | P2 |
| 10 | HRY-010 — Upstream güncelleme ve sürümleme modeli | P2 |
| 11 | HRY-011 — Reference fork doğrulaması | P2 |

## Genel tamamlanma ölçütü
Harpy'nin extraction sonrası bağımsızlığı testlerle kanıtlanmış; temiz kurulum, recovery ve güvenlik sınırları doğrulanmış; CI ve operasyonel teşhis kabiliyeti yeterli; lisanslama, fork/bootstrap ve upstream güncelleme modeli açıkça belgelenmiş olmalıdır.

Bu doküman seti bir planlama backlog'udur. Task'ların listelenmesi, uygulandıkları veya kabul kriterlerinin karşılandığı anlamına gelmez.
