---
name: code-organization
description: Enforce backend-v2 file-size, naming, CRUD separation, readability, and source layout rules. Use for every handwritten source, test, migration, configuration, skill, or documentation change in backend-v2.
---

# Code Organization

## Baslamadan once

1. Degisecek dosyalari ve her dosyanin tek sorumlulugunu yaz.
2. Mevcut satir sayilarini kontrol et.
3. Dosya 160 satira yaklasacaksa degisiklikten once bolme noktasi belirle.
4. CRUD veya use-case'in hangi katmanlara dokundugunu ayir.

## 200 satir kontrati

- Elle yazilan her dosyayi en fazla 200 fiziksel satirda tut.
- 160 satiri tasarim hedefi, 180 satiri zorunlu bolme uyarisi kabul et.
- Bosluklari silerek, coklu statement'i tek satira sikistirarak siniri asma.
- Test, SQL migration, Markdown, YAML ve script dosyalarini da sinira dahil et.
- Repo-authored dosya UTF-8 text'tir; symlink veya binary artifact commit etme.
- Generated istisnasi header ile exact `contracts/generated.yaml` provenance ister.
- Yalniz verified generated output, lockfile ve vendor dependency istisnadir.
- Yeni bir istisna ekleme; gerekiyorsa once ADR ve owner onayi iste.

Her degisiklik sonunda calistir:

```powershell
go run ./tools/fileguard
```

## CRUD dosyalama

Her resource ve katmanda operasyonlari ayir:

```text
application/
  create.go
  get.go
  list.go
  update.go
  delete.go
adapters/http/
  create.go
  get.go
  list.go
  update.go
  delete.go
  routes.go
adapters/postgres/
  create.go
  get.go
  list.go
  update.go
  delete.go
```

- Read'i `get.go` ve `list.go` olarak ayir.
- Patch semantigi update'den farkliysa `patch.go` kullan.
- Bulk, restore, archive, import ve export icin ayri use-case dosyasi ac.
- Testi ayni operasyon adi ile yaz: `create_test.go`, `list_test.go`.
- `routes.go` yalniz route baglasin; validation veya business logic tasimasin.

## Yasak dumping-ground dosyalari

Yeni su dosya adlarini kullanma:

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

Niyet belirten ad kullan: `compile_filter.go`, `hash_password.go`,
`claim_due_jobs.go`, `publish_form.go`.

## Tek sorumluluk

- Bir dosyada bir primary use-case veya bir cohesive type ailesi tut.
- HTTP DTO, domain entity ve database row modelini ayni struct yapma.
- Constructor/wiring'i `module.go` veya composition paketinde tut.
- Interface'i consumer'a yakin port paketinde tanimla.
- Abstraction'i varsayimla degil, en az iki gercek kullanim ile cikar.
- Generic helper yerine domain dilini kullanan fonksiyon yaz.

## Bolme sinyalleri

Asagidakilerden biri varsa dosyayi bol:

- Birden fazla CRUD fiili.
- Birden fazla hata/transaction policy'si.
- Uc veya daha fazla bagimsiz private helper grubu.
- Birden fazla external adapter.
- Test kurmak icin ilgisiz bagimliliklar gerekiyor.
- Dosya adi icerigi tam anlatamiyor.

## Tamamlama checklist'i

- Her dosya 200 satir veya altinda.
- Operasyonlar ayri dosyalarda.
- Dosya adlari niyeti acikliyor.
- Testler operasyonla eslesiyor.
- Yeni generic dumping-ground yok.
- Fileguard green.
