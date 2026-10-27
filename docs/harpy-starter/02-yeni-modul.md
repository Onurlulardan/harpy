# Yeni modül

Harita budur. Kural [modular-architecture](../../skills/modular-architecture/SKILL.md) skill'indedir.

## Dizin

```text
modules/<name>/
  api/
  domain/
  application/
  ports/
  adapters/http/
  adapters/postgres/
  adapters/events/
  migrations/
  module.go
```

Harpy'de bu ağaç `core/internal/modules/<name>/` altındadır. Boş katman açılmaz. Katman, ihtiyaç doğunca eklenir. `api/` yalnız gerçek bir cross-module consumer varsa açılır. `ports/` modüle özeldir. Başka modül `ports/` import etmez.

`work` bu şekildedir: `api`, `domain`, `application`, `ports`, `adapters`, `migrations`.

## Sahip sınırı

Yeni işe başlamadan skill'deki beş soru cevaplanır:

1. Business capability'nin sahibi hangi modül?
2. Hangi tablo, migration ve event bu modüle ait?
3. Dışarıya açılan command/query API'si nedir?
4. Transaction hangi modül veya application coordinator'da başlar?
5. Gerçekten async veya ayrı deployable olması gereken kısım var mı?

Örnek: pano, görev, yorum ve takvim `work` sahibindedir. Tabloları ve migration'ı `core/internal/modules/work/migrations/` altındadır. Dışarıya açılan kontrat `core/internal/modules/work/api` paketidir. Başka modül `work` tablosuna SQL yazmaz. Bunu skill yasaklar.

## Migration modülde durur

Şema SQL'i o modülün `migrations/` dizininde durur. Composition migrator bu dosyaları yükler.

## Composition'a tek bağ

Her süreç modülü kendi composition paketinden bir kez kurar. `cmd` dosyası modülü import etmez.

`work` örnekleri:

- HTTP: `core/internal/composition/api/build_work_http.go`
- Kuyruk: `core/internal/composition/worker/build_work_jobs.go`
- Migration yükü: `core/internal/composition/migrator`

`core/cmd/api/main.go` yalnız `core/internal/composition/api` import eder. Aynı process içinde modüller HTTP ile konuşmaz.
