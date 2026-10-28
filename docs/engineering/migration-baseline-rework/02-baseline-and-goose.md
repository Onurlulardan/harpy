# 02 — Hedef mimari: güncel baseline + Goose

## Çalışma modeli

- **İlk kurulum:** Güncel schema snapshot'ını uygulayan baseline(lar) bir kez uygulanır ve Goose version history'ye kaydolur.
- **Her sonraki geliştirme:** Yeni timestamp'li SQL migration eklenir. Migration dosyası değişikliği bir kez merge edildikten sonra değiştirilmez; hata yeni bir migration ile düzeltilir.
- **Normal deploy:** Goose uygulanmamış versiyonları sıralar ve çalıştırır. Out-of-order migration kapalıdır. Goose'un native version store'u migration history'nin tek kaynağıdır.
- **Yeni tenant:** Tenant schema'sı oluşturulduğunda aynı tenant baseline'ı o schema'ya uygulanır; sonrasında tenant'a ait yeni migration'lar çalıştırılır. Tenant şeması yaşam döngüsü platform migration'ı ile karıştırılmaz.
- **Mevcut legacy local volume:** Bu starter geçişi backwards-compatible değildir. Legacy history tablosu adapte edilmeyecek; test için yeni disposable DB/volume kullanılacak. Kullanıcı volume'ları otomatik silinmeyecek.

## Baseline nasıl çıkarılacak?

Baseline'ı mevcut migration dosyalarından elle birleştirme. Kaynak, son şemanın kendisi olacak.

1. GitHub `main` ile aynı commit'in temiz, geçici PostgreSQL 18 instance'ını oluştur. Kullanıcının `ops/local` named volume'larını kullanma.
2. Mevcut migration akışını bu boş DB üzerinde çalıştır. Platform migration'ları tamamlandıktan sonra setup/operator yoluyla temsilî bir tenant yarat ve tenant migration'larını çalıştır.
3. `platform` kapsamını ve tenant schema'sını ayrı ayrı katalog envanteriyle doğrula. Source DB'nin uygulamanın son geçerli şema ve gerekli semantic seed data durumunda olduğundan emin ol. `task dev` eski volume yüzünden fail ederse eski local volume'u source of truth kabul etme.
4. PostgreSQL'in aynı major/digest'iyle schema-only dump üret. Baseline'da tablolar, views/materialized views, sequences, types/enums/domains, defaults, indexes, constraints/FKs, functions/procedures, triggers, RLS/policies, grants/ownership ve güvenlikle ilgili `search_path` ayarlarının tamamını koru.
5. Yalnızca gerçek ve gerekli statik reference/seed satırlarını dahil et. Environment UUID, tenant UUID, admin credential, signing/recovery key, license artifact, secret ve runtime-generated identity gibi kuruluma özgü verileri dump'a koyma. Bunlar mevcut setup/operator authority akışıyla oluşturulmalı.
6. Cluster-level roles, database/schema ownership, grants, extensions ve default privileges `pg_dump --schema-only` çıktısında nasıl temsil edildiğini ayrıca doğrula. `--no-owner` / `--no-acl` gibi seçenekleri incelemeden kullanma; güvenlik sınırlarını sessizce kaybetme.
7. Kaynak şemayı yeni baseline'dan oluşturulmuş boş hedef DB ile karşılaştır. Catalog query'leriyle nesne sayıları ve isimleri, constraint/index definition'ları, routine signature/body, trigger/RLS, extensions, grants ve tenant izolasyonu karşılaştırılmalı.
8. Aynı işlemi yeni bir temporary tenant schema'da tekrarla. Platform ve tenant başlangıç durumu, uygulama readiness'si ve on-premise installation/license offering akışı doğrulanmadan eski SQL kaynaklarını silme.

## Dosya / sürüm modeli

- `20261028000000_initial_schema.sql` baseline için örnek version'dır; son adlandırmayı repo'nun FileGuard ve migration version kurallarıyla uyumlu seç. Future migration timestamp'leri bundan büyük ve benzersiz olmalı.
- Ayrı platform ve tenant history store gerekiyorsa aynı baseline release'i iki kapsam için birer SQL dosyası üretir; platform DB ve her tenant schema'sı kendi uygulanmış migration durumunu izler.
- Eski 1.982 entry'nin hepsi runtime'a paketlenmez. Yeni install yalnız baseline'ı ve baseline'dan sonra gerçekten eklenen migration'ları bilir.
- Baseline bir kere çıkarılır ve immutable kabul edilir. Şema geliştikçe baseline her sprintte yeniden üretilmez; sonraki fark yeni migration olur.

## Goose entegrasyonu

`go.mod` zaten `github.com/pressly/goose/v3 v3.27.2` içeriyor. Yeni sistemde `goose.NewProvider` ile PostgreSQL dialect ve normal Goose store kullanılsın. Kod, manifestteki 1.982 entry'yi `SelectedMigration` dizisine çevirip Goose'a tekrar doğrulatmasın.

- SQL migration filesystem'i için `embed.FS` + `fs.Sub` veya tek açık migration dizini kullan.
- Native PostgreSQL store ve gerektiğinde `WithTableName` ile history tablosu isimlerini ayarla. Platform ve tenant schema'larında tablo adı çözümlemesini pinned Goose sürümünde test et.
- Goose'un desteklediği locker kullan veya gerçekten gerekli olan küçük bir advisory-lock katmanı bırak; iki ayrı locking mekanizması yaratma.
- `WithAllowOutofOrder(false)` davranışını koru; global registry kullanılmıyorsa Goose'un ilgili seçeneğini tut.
- Custom `DialectCustom` + `WithStore(customStore)` kaldırılmalı. Eğer Goose'un native store'u dinamik tenant schema izolasyonunu gerçekten desteklemiyorsa, sadece o eksikliği çözen küçük bir adapter kabul edilir; owner/scope/generation/selected SHA/supersession framework'ü tekrar kurulmaz.
- Goose history tablosu uygulanmış versiyonu ve durumunu kanonik olarak kaydeder. Bu yeni starter için `platform.platform_migration_history` custom sözleşmesinin yerini alır; `run_id`/generation gibi operasyonel bilgileri migration version table'a zorla ekleme.
- Migration hatasında versiyon, dosya adı, kapsam ve SQLSTATE loglansın; connection string/secret loglanmasın.

## Hash / contract yaklaşımı

Migration başına SHA manifest, binary schema contract ile migration digest zinciri ve historical release archive kaldırılacak. Goose standard version table native bir içerik-checksum defteri değildir; bunun yerine CI'da önceki commit'te bulunan `.sql` dosyalarının değiştirilmediği/silinmediği, yalnızca yeni dosyaların eklendiği denetlenir. Aynı dosyanın değişmesi gerekiyorsa yeni migration yazılır.

Uygulama gerçekten runtime DB compatibility generation gerektiriyorsa onu tek küçük schema version/compatibility kontratıyla ölç; manifestteki her migration'a bağlama. Extension, license, installation ve release signing doğrulaması kendi domain'inde kalır ve migration geçmişinin yerine geçmez.

## Reference

Goose Provider, native dialect stores, custom Store kullanımının ne zaman gerekli olduğu ve lock options için: https://pressly.github.io/goose/documentation/provider/ ve https://pressly.github.io/goose/documentation/custom-store/. Custom Store ancak native store'un karşılayamadığı somut bir ihtiyacı testle gösterince kullanılmalı.
