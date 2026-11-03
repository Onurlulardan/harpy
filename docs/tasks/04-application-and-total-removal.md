# 04 — APPLICATION ve eski sistemin tamamen kaldırılması

## Amaç

Dört aşamalı akışın son adımı olarak API, worker ve diğer uygulama servislerini yalnız Foundation ve Goose başarılı olduktan sonra çalıştır. Aynı zamanda repository'nin tamamından eski migration framework'üne ait her executable kalıntıyı kaldır.

Bu görev bir cleanup veya compatibility işi değildir. Eski sistemin tümüyle kaldırılması ve yalnız dört aşamalı mimarinin bırakılmasıdır.

## APPLICATION başlangıç kontratı

- API, worker ve diğer uygulama servisleri yalnız INITIALIZE → DATABASE FOUNDATION → GOOSE başarılı olduktan sonra başlatılır.
- API/worker startup'ı database/schema oluşturmaz, extension kurmaz, role/grant değiştirmez, baseline uygulamaz veya Goose çalıştırmaz.
- Runtime role'leri DDL veya database administration yetkisi taşımaz.
- Runtime readiness normal process bağımlılıklarını (DB connection, runtime role, gerekli altyapı servisleri) doğrular. Migration manifest, schema generation, binary schema contract veya custom schema version receipt kontrolü yapmaz.
- Schema hazır olma garantisi initialize akışının Goose adımından gelir; uygulama içinde ikinci kez uygulanmaz.
- Uygulama normal işlevleri, IAM/tenant/business authorization ve mevcut domain davranışlarını korur.
- Uygulama başlangıç hatası migration state değiştirmez ve migration'ı otomatik başlatmaz.

## Tenant provisioning

Yeni tenant oluşturma, yeni bir migration alt sistemi üretmeden dört aşamalı schema lifecycle'ı kullanır:

1. Tenant kimliği ve canonical schema adı mevcut tenancy domain kurallarıyla doğrulanır.
2. Foundation sorumluluğu yalnız boş PostgreSQL schema namespace'i, owner ve gerekli privilege'ları hazırlar.
3. Aynı Goose runner, tenant migration source'unu tenant schema'sına uygular.
4. Goose başarılı olduktan sonra mevcut tenancy application flow tenant'ı aktif eder.

Ayrı tenant migrator process/container, tenant migration executor, migration worker, özel job claim/lease, migration receipt, migration run tablosu veya completion stored procedure tutulmaz. Genel tenant provisioning işinin başka bir business lifecycle ihtiyacı varsa yalnız o ihtiyaca ait minimal orchestration korunabilir; schema değişikliği ve migration state Goose dışına taşınamaz.

## Repository genelinde kaldırılacaklar

Aşağıdaki alanların hepsini tara. İsim farklı olsa bile aynı davranışı yapan kodu kaldır.

### Go source ve package'lar

- Custom migration domain model, port, adapter, store, runner, planner, selector, validator, classifier, finalizer ve orchestrator.
- Migration bundle/corpus collector ve custom migration discovery.
- Release manifest parser, archive/index, release registry, attestation ve digest binding.
- Migration run/attempt/receipt/checkpoint/requirement lifecycle.
- Tenant migration executor ve buna özel claim, process, heartbeat, retry, reconcile, completion ve failure kaydı.
- Schema contract/binary schema contract binding ve migration version/generation readiness.
- Goose native API'sini custom migration framework gibi saran bütün policy ve state machine kodu.

### Contracts ve generated outputs

- Migration-specific Go contracts, JSON Schema, generated models ve fixture'lar.
- Migration manifest ve release manifest schema'ları.
- Custom migration run, receipt, checkpoint ve schema-contract payload'ları.
- Artık tüketilmeyen generated OpenAPI/runtime aggregate referansları.
- Silinen contract'ları üreten generator ve validation kodları.

Silinen contract'ın yalnız source dosyasını kaldırmak yeterli değildir. Consumer, schema, fixture, generator, generated output, API response ve deployment referansı da kaldırılır.

### PostgreSQL nesneleri

Eski sistemin oluşturduğu migration framework'e özel:

- Table ve index'ler
- Sequence ve type'lar
- Function/procedure/trigger'lar
- Grant ve ownership kayıtları
- Custom version/history/catalog/checkpoint/receipt/release state tabloları

Yeni Goose flow gerçek PostgreSQL üzerinde çalışıp doğrulandıktan sonra bunları Goose forward cleanup migration'larıyla kaldır. Business domain tablolarına, tenant verilerine veya business state'e dokunma. DROP kapsamı nesne bazında doğrulanmadan geniş wildcard cleanup yazma.

### CLI, build ve deployment

- Eski migrator binary/command ve deployable tanımları.
- Migration manifest input/output flag'leri.
- Dockerfile içindeki manifest kopyalama, chmod/ownership ve build artifact adımları.
- Setup/localrelease içindeki migration artifact üretme veya doğrulama mantığı.
- Taskfile içinde duplicate bootstrap/migrate/verify/catalog/release görevleri.
- CI workflow içindeki migration manifest/checkpoint/release doğrulama işleri.
- Compose service, healthcheck, volume, port ve dependency referansları.
- Environment variable, config field, secret, file path ve docs referansları.

Goose çalıştırmak için gereken en küçük bağlantı ve source path ayarı dışında migration'a özgü config bırakma.

### Docs, skills ve repository talimatları

Mevcut mimari ile çelişen bütün talimatları güncelle:

- AGENTS.md
- skills/INDEX.md
- skills/data-tenancy/SKILL.md
- skills/platform-runtime/SKILL.md
- İlgili testing/impact skill'leri
- docs/architecture altındaki platform, tenant, runtime, deployment ve developer guide metinleri
- README, local operations guide ve migration/installation açıklamaları

Özellikle custom migrator deployable, schema contract, release generation, manifest, custom checkpoint ve tenant migration executor'ı öneren eski hükümler silinmelidir. Yeni dokümanlar yalnız INITIALIZE, DATABASE FOUNDATION, GOOSE ve APPLICATION aşamalarını tarif etmelidir.

Bu task seti dışındaki eski migration task/archive klasörleri tutulmaz. docs/tasks/ altında bu yeni dört aşamalı task seti dışında başka bir migration planı veya eski task klasörü bırakma.

## Kaynak taraması ve sıfır kalıntı denetimi

Repository root'undan hem case-sensitive hem case-insensitive arama yap. Aşağıdaki terimleri ve eş anlamlılarını tüm source/config/contracts/SQL/generated/tests/build/docs alanlarında ara:

- migration manifest / fresh migration manifest
- migration digest / migration hash
- release manifest / release registry / release archive
- migration generation / database generation / runtime generation
- selected migration / migration catalog / migration bundle
- migration run / migration attempt / migration receipt
- migration checkpoint / migration requirement
- history alias / supersession / history classification
- schema contract / binary schema contract / schema generation
- tenant migration executor / tenant migrator
- custom Goose store / migration version store
- migration readiness / migration finalization

Beklenen sonuç: aktif Go code, SQL routine, contract, runtime config, build/deployment wiring veya DB schema içinde eski framework'e ait hiçbir sonuç kalmamalıdır.

Bu task dosyalarının kendisi eski isimleri kaldırılacaklar listesinde anabilir. Bu, executable code veya aktif mimari kalıntısı sayılmaz. Fakat eski davranışı açıklayan başka bir aktif architecture/skill dokümanı bırakılmaz.

## Veri güvenliği ve cutover

- Önce yeni Goose mekanizmasının clean install ve mevcut DB üzerinde çalıştığını doğrula.
- Eski migration framework'ü devreden çıkarmadan önce tüm uygulama consumer'larını kaldır.
- Eski metadata tablolarının yalnız migration framework'üne ait olduğunu kanıtla.
- Metadata temizliğini Goose forward migration'ları ile uygula.
- Platform business data ve her tenant'ın business data'sını öncesi/sonrası doğrula.
- Tenant ID, schema name, row identity, foreign key, unique constraint ve kritik aggregate sayıları korunmalıdır.
- Mevcut database drop/recreate, tenant schema reset veya business data reseed yapılamaz.
- Eski framework koduna geri dönme fallback'i oluşturma. Hatalar yeni Goose migration veya Foundation düzeltmesiyle çözülür.

## Zorunlu uçtan uca kabul testleri

### Temiz kurulum
1. Boş PostgreSQL ile INITIALIZE çalışır.
2. DATABASE FOUNDATION gerekli altyapıyı kurar.
3. GOOSE platform ve tenant başlangıç schema'larını oluşturur.
4. APPLICATION başlar ve readiness başarılıdır.

### Repeat/idempotency
1. Aynı kurulumda INITIALIZE yeniden çağrılır.
2. Foundation nesneleri silinmez/değiştirilmez.
3. Goose yeniden uygulanan DDL çalıştırmaz.
4. Hiçbir business data değişmez veya çoğalmaz.

### Mevcut kurulum
1. Önceden dolu platform DB ve en az iki tenant schema ile upgrade yapılır.
2. Goose native history güvenli şekilde adoption edilir.
3. Eksik migration'lar forward uygulanır.
4. Eski migration metadata nesneleri Goose ile kaldırılır.
5. Business data kimlikleri ve sayıları korunur.
6. API/worker yeni akışla çalışır.

### Tenant provisioning
1. Yeni tenant schema'sı Foundation ile oluşturulur.
2. Goose aynı native runner ile tenant migration'larını uygular.
3. Migration başarısız olursa tenant aktif edilmez.
4. Retry custom migration state gerektirmez.
5. Başarılı tenant diğer tenant schema'sına erişemez.

### Güvenlik ve mimari
1. API/worker database role'leri DDL yapamaz.
2. Goose başarısızsa Application başlamaz.
3. Goose dışı migration history/version store yoktur.
4. Ayrı migrator veya tenant migration executor deployable yoktur.
5. Her schema DDL değişikliğinin tek kaynağı Goose migration dosyalarıdır.

## Çalıştırılacak kontroller

Repository'nin mevcut test/guard komutlarını kullan ve sonuçları raporla:

- go test ./...
- go run ./tools/fileguard
- go run ./tools/archguard
- Mevcut contract/generated output doğrulamaları
- Gerçek PostgreSQL fresh install integration testleri
- Mevcut database ve tenant upgrade integration testleri
- Repeat initialize testi
- Tenant A/B isolation testi
- task dev uçtan uca doğrulaması

Docker veya gerçek PostgreSQL erişimi yoksa integration testini başarılı varsayma. Atlanan kontrolü açıkça blocker olarak raporla.

## Nihai teslim raporu

Aşağıdaki başlıklarla somut kanıt ver:

- Dört aşamalı son akış ve entrypoint'ler
- Silinen eski package/contract/SQL/config/build/deployable alanları
- Kaldırılan PostgreSQL metadata nesneleri
- Goose migration source'ları ve Goose native history doğrulaması
- Clean install sonucu
- Mevcut DB/tenant upgrade sonucu ve veri koruma kanıtı
- Repeat initialize sonucu
- API/worker startup ve runtime privilege sonuçları
- Çalıştırılan komutlar, başarılı/başarısız/atlanmış testler
- Repository genelindeki eski kavram taramasının sonuçları

Yalnızca build'in başarılı olması, testlerin bir kısmının geçmesi veya task dosyalarının tamamlanması işi bitirmez. Dört aşama dışında aktif mekanizma veya eski migration framework'ünden tek bir executable kalıntı varsa görev başarısızdır.
