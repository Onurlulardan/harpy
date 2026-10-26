---
name: modular-architecture
description: Protect backend-v2 module ownership, dependency direction, data boundaries, composition, and service extraction decisions. Use for new modules, cross-module behavior, shared code, package moves, or deployable boundary changes.
---

# Modular Architecture

## Once siniri tanimla

Her degisiklik icin cevapla:

1. Business capability'nin sahibi hangi modul?
2. Hangi tablo, migration ve event bu module ait?
3. Disariya acilan command/query API'si nedir?
4. Transaction hangi modul veya application coordinator'da baslar?
5. Gercekten async veya ayri deployable olmasi gereken kisim var mi?

## Modul yapisi

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

Bos katman olusturma. Bir ihtiyac dogdugunda ekle; dependency yonunu koru.

`api/` yalniz gercek cross-module consumer varsa acilir. Export edilen command,
query, request/result DTO ve dar interface'leri tasir; implementation veya
framework tipi tasimaz. `application/` bu kontratlari uygular.

`ports/` module-private adapter ihtiyaclaridir: repository, clock, provider ve
message portlari. Baska modul `ports/` import edemez.

## Izin verilen yon

```text
adapter -> application -> domain
postgres adapter -> application port
composition -> module constructor
module A -> module B `api/` kontrati
```

## Yasak bagimliliklar

- Baska modulun `domain`, `internal`, `postgres` veya migration paketini import etme.
- Baska modulun `ports/` paketini import etme.
- Baska modulun tablosuna SQL, JOIN, foreign-key cascade veya ORM relation yazma.
- Application'i Gin, GORM, RabbitMQ veya provider SDK tipine baglama.
- Domain entity'yi HTTP veya database DTO'su olarak kullanma.
- Business model ve policy'yi `shared`, `common` veya `platform` altina tasima.
- Ayni process icinde cross-module HTTP client yazma.

## Modul ile process'i ayir

Modul kod sahipligi siniridir; deployable runtime siniridir. Yeni servis ancak
asagidakilerden biri kanitlandiginda acilir:

- Farkli guvenlik veya credential siniri.
- Farkli ve olculmus scale/resource profili.
- Bagimsiz availability/failure isolation gereksinimi.
- Dis protokol veya host yetkisi.

Takim sayisi veya klasor buyuklugu tek basina servis gerekcesi degildir.

## Gecici altyapi yasagi

- Nihai olmayan framework, ORM, storage modeli veya transport ile baslama.
- "Once burada tutar sonra tasiriz" seklinde paralel source of truth kurma.
- Cutover adapter'ini core application/domain paketlerine sizdirma.
- Zorunlu migration boundary'sine owner, persisted state, giris/cikis kosulu,
  rollback ve otomatik completion gate'i tanimla.
- Completion gate gecmeden compatibility kodunu silme; gate gectikten sonra
  kullanilmayan boundary'yi architecture testiyle yasakla.

## Cross-module use-case

- Senkron gereksinimde owner modulun public query/command API'sini cagir.
- Tek transaction gerekiyorsa application coordinator ve ortak Unit of Work kullan.
- Commit sonrasi entegrasyon icin versioned outbox event'i yaz.
- Event'i ayni process icindeki normal fonksiyon cagrisinin yerine kullanma.
- Read model gerekiyorsa owner event'inden rebuild edilebilir projection kur.

## Platform paketi

Yalniz teknik primitive tasir:

```text
config, clock, id, postgres, transaction, telemetry,
http transport, cache adapter, message transport, blob adapter
```

Platform hicbir business modulu import edemez. Composition root tum modulleri
gorebilen tek yerdir.

## Yeni shared abstraction kontrolu

1. En az iki gercek consumer goster.
2. Semantigin business mi teknik mi oldugunu belirle.
3. Owner business ise owner modul API'sinde tut.
4. Teknik ise minimum interface'i platforma koy.
5. Tum consumer'lari generic modele zorlayan abstraction'i reddet.

## Tamamlama checklist'i

- Tek owner ve public API belli.
- Tablo/migration/event sahipligi belgeli.
- Dependency yonu test edilebilir.
- Cross-module SQL veya adapter importu yok.
- Transaction ve async siniri acik.
- Servis ayirma gerekcesi olculmus veya kaldirilmis.
