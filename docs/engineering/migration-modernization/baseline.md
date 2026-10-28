# Migration maliyet taban çizgisi

Ölçüm makinesi: `linux amd64`, `go1.26.5`, `Intel(R) Xeon(R) Processor`, `GOMAXPROCS` bench satırında `-4`. Kaynak taban: `b520489baed2579bc004de74ebf2fd59268a4a2b`. Before binary'leri bu commit'ten, after binary'leri bu değişiklikten, aynı bayraklarla üretildi. Saatler UTC.

## Arşiv sayımı

```text
go run ./internal/migrationbundle/cmd/census .
```

2026-10-28T10:45Z civarı çıktı: 109 manifest, 115990137 byte. Boyut min 55050, medyan 1234144, p95 1650025, max 1655160. Entry satırı 184036, benzersiz 2735. Phase: foundation 105, expand 2615, contract 15. Generation `(1,1)` 105 entry, 97 farklı çift. Varyant SHA aynı 2714, farklı 21, path aynı 0. En büyük iki manifest çakışması 2608/2610. Benzersiz entry JSON 1732445 byte. `/migrations/` SQL 2000 dosya / 7711677 byte. Repo SQL 3270 dosya / 8287412 byte. `*.v1_adopted.sql` yok.

Medyan ve p95, sıralı boyutlarda doğrusal aradeğerlemedir.

## Registry yükü

```text
go test -bench=BenchmarkLoadReleaseRegistry -benchmem -benchtime=1x -count=3 -run '^$' ./core/internal/platform/migration/
```

2026-10-28T10:46:13Z–10:47:24Z. Üç tek koşunun medyanı: 18412161195 ns/op, 8465312328 B/op, 169000537 allocs/op. Bu yol üretimde duruyor; hız iddiası yok. Süreç başına bir kez yüklenir, önbellek eklenmedi.

## Startup

```text
go test -bench=BenchmarkAPISchemaContractSQL -benchmem -benchtime=1s -count=5 -run '^$' ./core/internal/composition/migrator/
go test -bench=BenchmarkLoadBoundStartup -benchmem -benchtime=1s -count=5 -run '^$' ./core/internal/platform/schemacontract/
```

Aynı pencerede beş koşunun medyanı:

| Yol | ns/op | B/op | allocs/op |
| --- | --- | --- | --- |
| SQL `FreshMigrationDocuments` + `LoadStartup` | 131280984 | 417118549 | 704556 |
| `LoadBoundStartup` SHA-256 bağı | 573686 | 827064 | 382 |

Bound yol manifesti yeniden doğrulamaz; baytların özeti schema contract ile aynı olmak zorundadır. Yapısal doğrulama `TestEmbeddedFreshManifestVerifies` ve `checkfresh` içindedir. Zaman eşiği CI'ya konmadı.

## Binary

```text
go build -trimpath -ldflags="-s -w" -o "$OUT/core-api" ./core/cmd/api
```

Aynı komut worker, migrator, operator için. Before `OUT=/tmp/harpy-bins` (10:33Z). After `OUT=/tmp/harpy-bins-after`.

| Binary | Before | After |
| --- | --- | --- |
| core-api | 64745737 | 57401609 |
| core-worker | 57098505 | 49733897 |
| core-migrator | 159383817 | 159387913 |
| core-operator | 150782217 | 150782217 |

Before API/worker içinde örnek migration SHA'sı 0 kez, after 1 kez. Migrator/operator path'i 108 kez tutuyor; historical embed duruyor. `-tags harpy_measure_no_release_embed` migrator'ı 43372809 byte yaptı (after migrator'dan 116015104 byte küçük). Bu etiket üretim yolu değildir. API deneyi after API ile aynı boyda: linker, çağrılmayan arşivi zaten düşürüyor.

## Sınırlar

Tek makine, soğuk/sıcak cache ayrımı yapılmadı. Registry bench bellek tahsisi tutulan registry'yi de sayar. Census, entry JSON'unu `encoding/json` ile sayar. Docker stack açılmadı.
