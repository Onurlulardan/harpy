# 04 — fileguard ve archguard

İki araç Harpy'dedir: `tools/fileguard` ve `tools/archguard`. forms, workflow, AI ve MCP allowlist'leri yoktur. O modüller Harpy'de yoktur. IAM, tenancy, worker ve diğer modüllerin import yönü araçta durur.

Amaç, ajanın bu şeklin dışına çıkamamasıdır. Şekil değişince araç aynı turda değişir. Listede olmayan bir yasağı araç kendisi eklemez.

Araçlar, Harpy ağacıyla birlikte yeşil olmak zorundadır.

## fileguard

`go run ./tools/fileguard`

Elle yazılan her dosya en fazla 200 fiziksel satırdır. Go, test, SQL, Markdown, YAML ve script aynı bütçededir. Satır, boşluk silinerek veya ifadeyi tek satıra sıkıştırarak düşürülmez.

Yasak Go dosya adları:

```text
crud.go
handler.go
service.go
repository.go
utils.go
helpers.go
common.go
misc.go
```

UTF-8 text dışındaki authored dosya ve symlink geçmez. Lockfile sayılmaz. Generated dosya, Harpy'de kayıtlı provenance yoksa geçmez.

## archguard

`go run ./tools/archguard`

Zorladığı yön, 02'deki ağaçtır:

- `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator` ve `core/cmd/operator` yalnız kendi composition girişini import eder.
- `cmd/setup` modülün application veya adapter paketine girmez.
- Platform, modül import etmez.
- Domain `net/http`, `database/sql`, `pgx` ve bulut SDK import etmez.
- Aynı modülde yön adapter → application → domain şeklindedir.
- Bir modül, başka modülün yalnız `api` paketini import eder.
- `forms`, `workflow`, `ai` ve `mcp` paketleri import edilemez. `services/ai` ve `services/mcp` yoktur.
- Modül yolu `go.mod` içindeki `module` satırından okunur.
- Blob kodu `aws-sdk-go-v2` kullanır. MinIO SDK ve RustFS SDK import edilmez.

## Test

Geçerli küçük bir ağaç yeşildir. Yasak import, forms/workflow/ai/mcp importu veya 200 satırı aşan dosya kırmızıdır. Araç paketlerinin testleri Harpy modülünde çalışır. `go run ./tools/fileguard` ve `go run ./tools/archguard` bu ağaçta çıkar kodu 0 ile biter.

## Kontrol

- [x] 200 satır ve yasak adlar bu metindedir.
- [x] archguard 02'deki mimariyi zorlar. `forms`, `workflow`, `ai`, `mcp`, `services/ai` ve `services/mcp` importunu reddeder. Bu paketlerin allowlist'i yoktur.
