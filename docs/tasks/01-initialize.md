# 01 — INITIALIZE

## Amaç

Harpy'nin ilk kurulumunu başlatan tek bir giriş noktası oluştur. INITIALIZE, veritabanı migration motoru değildir; ikinci bir setup framework'ü değildir; Foundation ve Goose'un yerine geçmez. Tek görevi dört aşamalı akışın doğru sırayla çalışmasını sağlamaktır.

Nihai sıra değiştirilemez:

1. INITIALIZE
2. DATABASE FOUNDATION
3. GOOSE
4. APPLICATION

## Zorunlu davranış

- İlk kurulum için tek bir açık initialize entrypoint bulunmalıdır.
- Entry point Foundation tamamlanmadan Goose çağırmamalıdır.
- Goose tamamlanmadan API, worker veya diğer uygulama süreçlerini başlatmamalıdır.
- Her aşama hata döndürdüğünde akış durmalıdır. Hata gizlenip sonraki aşamaya geçilmemelidir.
- Her aşamanın gerçek exit code ve sebebi dışarı aktarılmalıdır.
- Initialize aynı kurulum üzerinde tekrar çağrıldığında veritabanını sıfırlamamalı, var olan nesneleri silmemeli ve veri kaybına yol açmamalıdır.
- İlk kurulum yarıda kesilmişse tekrar çalıştırma, özel checkpoint tablosu okumadan PostgreSQL nesneleri ve Goose native history üzerinden güvenle devam edebilmelidir.
- Initialize tamamlandıktan sonra normal API/worker restart'ı initialize veya migration çalıştırmamalıdır.

## Kuruluma ait state

Initialize durumunu takip etmek için özel bir tablo, generation, run, receipt, manifest, digest, release veya checkpoint kaydı oluşturma.

Tekrar çalıştırma güvenliği şu gerçeklere dayanır:

- PostgreSQL database var mı?
- Gerekli Foundation schema/role/extension/ownership/grant nesneleri var mı?
- İlgili platform/tenant schema'sında Goose native history ne gösteriyor?
- Goose uygulanmamış migration var mı?

Bunların dışında ikinci bir initialization state store bulunamaz.

## İş kapsamı

- Mevcut setup CLI, operator, Taskfile, Compose ve deployment entrypoint'lerini incele.
- Mevcut birden fazla initialize/bootstrap/migrate yolunu tek entrypoint etrafında sadeleştir.
- Aynı Foundation işlemini iki farklı task'ın tekrar etmesini kaldır.
- Migration manifest dosyası hazırlama/doğrulama veya migration release hazırlama adımlarını tamamen sil.
- Initialize'ın yalnızca ilk kurulumda ve gerektiğinde yeni tenant schema'sının hazırlanmasında kullanılmasını sağla.
- Normal uygulama process startup'ını initialize'dan ayır.

## Hata ve tekrar çalıştırma kuralları

- Foundation ortasında hata: Application başlamaz. Tekrar denemede mevcut PostgreSQL nesneleri kontrol edilir, var olanlar korunur ve eksikler güvenli şekilde tamamlanır.
- Goose ortasında hata: Application başlamaz. Goose native history baz alınarak Goose yeniden çağrılır. Custom resume/checkpoint mekanizması yazılmaz.
- Application başlangıç hatası: Initialize veya Goose geriye dönük otomatik olarak çalıştırılmaz. Uygulama hatası kendi lifecycle'ında raporlanır.
- Database bağlantı bilgileri veya bootstrap credential eksikse açık hata verilir; varsayılan/boş credential ile devam edilmez.
- Kısmi kurulumda hiçbir aşama database'i drop/recreate ederek temizlemeye çalışmaz.

## Yasaklar

- initialize_state, initialization_runs benzeri özel durum tabloları.
- setup_generation, database_generation, runtime_generation veya bunlara karşılık gelen ikinci kurulum/version kaydı.
- Bootstrap receipt, migration receipt, installation generation, release veya migration manifest bağımlılığı.
- API/worker startup'ında goose.Up, DDL veya Foundation işlemi.
- Ayrı bir migration daemon, migration API'si veya migration worker'ı.
- Başarısız aşamayı "zaten tamamlanmıştır" gibi gösteren fallback.

## Test ve kabul

Gerçek PostgreSQL kullanarak şu senaryoları doğrula:

1. Boş ortamda INITIALIZE çağrılır; Foundation, Goose ve Application sırası korunur.
2. Aynı INITIALIZE ikinci kez çağrılır; hiçbir data/schema reset işlemi yapılmaz.
3. Foundation ortasında kontrollü hata üretilir; Application başlamaz; tekrar deneme güvenle tamamlar.
4. Goose ortasında kontrollü migration hatası üretilir; Application başlamaz; tekrar deneme yalnız Goose history üzerinden devam eder.
5. Var olan database üzerinde tekrar çağrı; mevcut business tabloları ve satırları korunur.
6. API ve worker process'leri bağımsız restart edildiğinde migration başlatmaz.
7. Setup ve local workflow içinde aynı adımın iki kere çalıştırılmadığı kanıtlanır.

## Tamamlanma şartı

Tek bir INITIALIZE giriş noktası vardır. Bu nokta yalnız dört aşamalı akışı başlatır. Özel initialize/migration state tablosu yoktur. Repeat-safe davranış gerçek DB state ve Goose native history ile sağlanır. API/worker startup migration yapmaz.
