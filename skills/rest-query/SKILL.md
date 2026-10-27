---
name: rest-query
description: Build backend-v2 typed REST Query parsing, QuerySpec AST, field catalog resolution, authorization, PostgreSQL compilation, filtering, search, sort, projection, relation, aggregation, pagination, and V1 query compatibility. Use for every dynamic read surface.
---

# REST Query

Authoritative wire ve AST karari icin once
`docs/rest-query-contract.md`, `docs/rest-query-wire.md`,
`docs/rest-query-types.md`, `docs/adr/0008-rest-query-execution.md` ve
`docs/adr/0024-rest-query-wire-and-budgets.md`
dosyalarini oku. Bu skill kontrati tekrarlamaz; uygulama prosedurunu tanimlar.

## Pipeline

```text
transport decode
-> normalize QuerySpec
-> resolve field catalog
-> validate types/operators
-> authorize resource and fields
-> inject mandatory predicates
-> enforce budget
-> compile parameterized plan
-> execute with timeout
-> present canonical response
```

Her adimi ayri dosya ve test yuzeyi yap.

## Paketleme

```text
core/internal/capabilities/restquery/
  api/{query_spec,executor,field_catalog}.go
  contract/{request,filter,sort,pagination,relation,aggregate}.go
  application/{normalize,validate,authorize,budget,compile,execute}.go
  adapters/http/{decode_get,decode_post,present}.go
  adapters/postgres/compiler/{select,filter,relation,aggregate,sort,page,args}.go
```

Dosya adlarini capability'ye gore daha da daralt; `code-organization` kapisini uygula.

## Parser kurallari

- GET ve POST'u ayni QuerySpec'e normalize et.
- Unknown key, field veya operatoru sessizce ignore etme.
- Predicate arity, aggregate ref ve include cardinality'sini wire sozlesmesinden
  birebir decode et; ikinci JSON sekli acma.
- String degeri field type'a gore parse et; compiler string coercion yapmasin.
- Nested boolean depth ve predicate count'u parse sonrasinda olc.
- V1 syntax'i yalniz compatibility adapter'da kabul et.
- V1'de yalniz imzali alias manifest'ini kabul et; gercek unknown her zaman 422'dir.
- `limit=0`, raw SQL, raw identifier veya arbitrary join kabul etme.

## Field catalog

- Public name'i published resource/version icinde stable field ID'ye cozumle.
- Physical expression'i yalniz storage adapter'i uretsin.
- Module consumer yalniz `restquery/api` import etsin; capability internal
  application/compiler paketini import etmesin.
- Select, filter, sort ve search capability'lerini ayri bildir.
- Field-level permission'i her kullanim turu icin kontrol et.
- Relation'i declared relation ID disinda kurma.

## Compiler

- Tum degerleri parameterize et.
- Identifier'i client inputundan quote ederek bile kullanma; catalog'dan uret.
- Tenant, soft-delete, portal owner ve approval visibility predicate'lerini zorunlu ekle.
- Stable sort ve ID tie-breaker uygula.
- Cursor'u ADR-0008 keyring'iyle AEAD-encrypt et; query shape, catalog, actor,
  sort ve tenant claim'lerini authenticated payload'a bagla.
- Interactive statement timeout ve row limit uygula.
- Exact count'u ayri cost karari olarak planla.

## Search ve index

- Search'i yalniz published search profile alanlarinda calistir.
- Her kolonu text'e cast eden global search yazma.
- Type'a uygun search expression ve index beklentisini capability'de tut.
- Exact ve casefold operator semantics'ini birbirine cevirmeden uygula.
- Yeni operator icin SQL, null, array, collation ve index testleri yaz.

## Published endpoint

- Saved QuerySpec'i publish aninda validate ve compile et.
- Plan'i referans form/version ID'leriyle sakla.
- Runtime caller'a yalniz ilan edilmis typed parameterleri ac.
- Form publish event'inde etkilenen plani `STALE` yap ve revalidation iste.
- Runtime caller'in yeni filter/sort/join enjekte etmesine izin verme.

## Verification

- GET/POST -> QuerySpec golden testleri.
- Operator x field-type matrix.
- Unknown/type mismatch icin 422 issue path testleri.
- Field permission ve tenant negative testleri.
- SQL injection/property fuzz testleri.
- Budget, timeout, cursor tamper ve stable order testleri.
- V1 Form/OpenAPI differential adapter testleri.
