---
name: analytics
description: Build and change backend-v2 saved queries, dashboards, publications, ACLs, public or capability access, REST Query execution, result caching, analytics jobs, and V1 cutover. Use for every analytics-owned table, API, event, permission, cache, or migration change.
---

# Analytics

Saved query, dashboard, publication, approval, cache, budget, retention ve cutover
semantigi icin ADR-0020 ve ADR-0031; purge/reference icin ADR-0026 authoritative'dir;
uygulamadan once tamamini oku. Query icin `rest-query`, endpoint icin
`api-contract`, SQL/tenant icin `data-tenancy`, async is icin `events-jobs`
skill'ini de oku.

## Sahiplik siniri

- `analytics` saved query, dashboard, ACL/preference, publication ve execution
  audit'in tek owner'i olan core module olsun; ayri deployable acma.
- REST Query tek parse/authorize/compile/execute capability'sidir. Analytics
  ikinci SQL builder, raw SQL veya V1 config evaluator tasimasin.
- Resource owner'lari immutable catalog/query portu sunsun; analytics baska
  modulun tablosuna SQL/JOIN veya adapter importu yazmasin.
- Draft'i mutable, published query/dashboard version ve publication release'ini
  immutable kur; dependency/version/policy/budget pinlerini sessizce ilerletme.
- Public route platform pointer'i ile tenant'i server-side cozsun; header, slug,
  token claim veya dashboard config'i tenant authority olmasin.

## Use-case dosyalama

Her davranisi application, HTTP ve PostgreSQL adapter katmaninda ayri dosyala:

```text
create_query.go, get_query.go, list_queries.go, update_query.go, delete_query.go
publish_query.go, execute_query.go, revalidate_query.go
create_dashboard.go, get_dashboard.go, list_dashboards.go
update_dashboard.go, delete_dashboard.go, publish_dashboard.go
grant_dashboard_acl.go, revoke_dashboard_acl.go, set_default_dashboard.go
create_publication.go, approve_publication.go, activate_publication.go
get_publication.go, list_publications.go, revoke_publication.go
execute_publication.go, issue_capability.go, revoke_capability.go
```

- Get ve list'i tek `read.go` dosyasinda birlestirme.
- Draft update, publish, approve, activate ve revoke'u ayri command tut.
- Her davranisa ayni operasyon adli test yaz; `routes.go` yalniz route baglasin.
- Dosya/CRUD/isim sinirinda `skills/code-organization/SKILL.md`yi uygula.

## Execution ve authorization sirasi

1. Authenticated istekte actor/tenant'i; public istekte opaque platform route'u
   dogrula. Client tenant secimini fail-closed reddet.
2. Base IAM permission, ACL/audience, step-up/approval ve capability lifecycle'ini
   application katmaninda kontrol et.
3. Exact immutable version, typed params, output schema, dependency catalog ve
   mandatory predicate'i cozumle.
4. Canonical REST Query authorization ve budget'ini her execution'da yeniden uygula.
5. Cache hit oncesi generation, auth context, predicate ve data revision'i
   yeniden dogrula; Redis/event'i correctness authority yapma.
6. Audit'i classification-aware digest ile yaz; raw param, result veya token tutma.

ACL base IAM'i genisletemez. PUBLIC synthetic principal yalniz ADR allowlist'iyle
calisir. Capability token opaque/hashli olur; URL, log veya response body'de tekrar
tasima. Unauthorized widget'i partial dashboard response'una sizdirme.

## Publish, event ve job akisi

- Publish tum resource, form, output ve layout dependency'lerini fail-closed
  validate et; revalidation yeni immutable version/release uretsin.
- Activation/revoke platform generation ve tenant release pointer'ini ADR'deki
  CAS/transaction sinirinda degistirsin.
- Event payload'ina config, token, param veya result koyma.
- Revalidation, cache warm ve repair'i ADR-0011 DB job'i yap; RabbitMQ state
  kaynagi olmasin.
- Revision-bound cache'i ancak tum dependency owner'lari authoritative
  `DataRevision` portu sundugunda ac; aksi halde final policy `none` olsun.

## Nihai mimari ve migration kapisi

- V1 JSON builder, slug+tenant routing, JWT embed secret veya ikinci result-cache
  authority'sini V2 core'a tasima.
- Unsupported V1 query'yi silent downgrade etme; typed replacement'i onaylat veya
  retire et. Gecici evaluator/query engine kurma.
- Migration'i ADR-0013 `(tenant,analytics)` single-writer fence'iyle yap; exact
  catalog/form reference, parameter/output contract ve result checksum'u persist et.
- V1 `is_public` flag'ini otomatik PUBLIC grant yapma. Approval ve differential
  gate tamamlanmadan publication switch etme.
- Rollback down migration, production dual-write veya second engine degil;
  route/image ve reverse reconciliation'dir.

## Verification

- [ ] Owner table/migration/API siniri ve cross-module SQL yasak testi var.
- [ ] Her CRUD, execute, publish, approval ve revoke islemi ayri dosya/testte.
- [ ] `code-organization` tamamlama kapisi green.
- [ ] QuerySpec conversion, unsupported syntax ve output golden testleri geciyor.
- [ ] Field/row permission, ACL, approval ve cross-tenant negative testleri var.
- [ ] Dashboard batch, reference, preference ve partial-failure testleri geciyor.
- [ ] Public route, capability expiry/revoke ve token-redaction testleri geciyor.
- [ ] Cache auth/revision/revoke race, Redis loss ve event delay testleri geciyor.
- [ ] Fresh migration, eski snapshot upgrade, V1 differential, fence ve rollback
  tatbikati tamam.
