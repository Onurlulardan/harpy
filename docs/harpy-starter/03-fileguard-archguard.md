# fileguard ve archguard

İki araç modül kökünden çalışır:

```text
go run ./tools/fileguard
go run ./tools/archguard
```

Satır ve dosya adı gerekçesi [code-organization](../../skills/code-organization/SKILL.md) skill'indedir. Bağımlılık gerekçesi [modular-architecture](../../skills/modular-architecture/SKILL.md) skill'indedir.

CI, `push` ve `pull_request` üzerinde üç iş çalıştırır: `go test ./...`, `go run ./tools/fileguard`, `go run ./tools/archguard`. Birim işi Postgres, RabbitMQ ve RustFS istemez.

## fileguard ne ölçer

Elle yazılan her dosyada en fazla 200 fiziksel satır vardır. Go, test, SQL, Markdown, YAML ve script aynı bütçededir. Boşluk silerek veya ifadeyi tek satıra sıkıştırarak satır düşürülmez.

Ayrıca şunlar kırmızıdır:

- UTF-8 olmayan dosya
- symlink
- yasak Go adları: `crud.go`, `handler.go`, `service.go`, `repository.go`, `utils.go`, `helpers.go`, `common.go`, `misc.go`

Lockfile sayılmaz: `go.sum`, `yarn.lock`, `package-lock.json`, `pnpm-lock.yaml` ve `*.lock`.

## Generated provenance

İlk beş satırda `code generated` ve `do not edit` birlikte geçiyorsa dosya generated sayılır. Kayıt `contracts/generated.yaml` içindedir. Bu dosya JSON'dur. Her kayıtta `generator`, `input_sha256` ve `output_sha256` vardır. Dosyanın SHA-256 çıktısı kayıtla aynı olmak zorundadır.

Header var, kayıt yoksa kırmızıdır. Kayıt var, dosya doğrulanmazsa kırmızıdır. Yeni bir istisna bu sayfadan eklenmez.

## archguard ne ölçer

`go.mod` içindeki module yolunu okur. Yol değişince `go run ./tools/archguard` yeniden çalışır.

Import yolu archguard'ın reddettiği bir ağaca girerse komut kırmızıdır.

Kırmızı import yönleri:

- Platform, modül import ederse: `platform cannot import modules`.
- `core/cmd/api`, `worker`, `migrator` ve `operator`, kendi composition girişi dışında bir paket import ederse. Standart kütüphane serbesttir. `core/cmd/api` için giriş `.../internal/composition/api` paketidir.
- `cmd/setup`, bir modülün `application` veya `adapter` paketini import ederse.
- Başka modülün `api` dışındaki paketi import edilirse: `cross-module import must target the other module api package`.
- Aynı modülde application adapter import ederse, ya da domain adapter veya application import ederse.
- Domain `net/http`, `database/sql`, `pgx` veya bulut SDK import ederse.
- MinIO SDK veya RustFS SDK, her dosyada.
- `aws-sdk-go-v2`, blob kodu dışında. Blob kodu, yolunda `blob`, `s3` veya `artifact` geçen dosyadır.

`work` application'ı IAM domain'ini import ederse archguard cross-module kuralını basar. Hedef `core/internal/modules/iam/api` olur.
