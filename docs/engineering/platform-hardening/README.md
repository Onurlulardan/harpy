# Harpy — Platform Hardening Görev Planı

Bu plan, migration-modernization görevlerinden ayrı olarak önceki kod incelemesinde tespit edilen CI, setup güvenilirliği, management endpoint erişimi ve generated-artifact sorunlarını kapsar. Branch: `docs/platform-hardening-tasks`.

## Genel çalışma kuralları

- Başlangıç bulguları `5044757b522d95c4e0afc9df3d214d6145b15016` commit'indeki incelemeye dayanıyor. Her görevde güncel `main` üzerinde kodu ve CI run'ını yeniden doğrula.
- Bir testin veya guard'ın amacını anlamadan silme, skip etme ya da gevşetme.
- Güvenlik/operasyon riski olarak işaretlenen bulguları kanıtlanmış exploit veya kesin üretim arızası diye sunma; ilgili akışla doğrula.
- Her görev ayrı küçük PR/commit olabilir. Her değişiklikte test komutlarını ve sonuçlarını yaz.
- Secret, private key, token veya gerçek credential değerlerini loglama.
- Bu görevler migration bundle refactor'ından bağımsız tutulmalı.

## Görevler

1. [CI testlerini ve fileguard hatalarını düzelt](01-ci-test-and-fileguard.md)
2. [Setup idempotency ve secret/key rotation garantilerini doğrula](02-setup-idempotency-and-key-rotation.md)
3. [Management listener'ın network exposure sınırını doğrula](03-management-listener-network-exposure.md)
4. [Release material ve generated artifact üretimini deterministik hale getir](04-release-material-generation.md)
5. [Platform hardening regression gate'leri ekle](05-regression-tests-and-operations-docs.md)

## Öncelik önerisi

- **P0:** CI'da doğrulanmış test ve fileguard hataları; temiz checkout'ta test/build güvenilirliği.
- **P1:** Setup'ın ikinci kez çalıştırılması ve key rotation semantiği; operasyonel veri/kimlik kaybı ihtimali varsa düzeltme önceliğini yükselt.
- **P1:** Management portunun gerçekten hangi arayüzlerden erişilebilir olduğunun kanıtlanması; deployment'a göre güvenli bind default'u.
- **P2:** Generation komutlarının determinism'i, docs ve regression gates.

Bu sıra geçici önceliklendirmedir. Güncel CI ve deployment modeli farklı sonuç verirse görev sahibi gerekçeli olarak değiştirebilir.

## Kapanış durumu

Kapanış komutları hedefli paket testleri, `fileguard` ve `archguard`dır. CI birim işi ayrıca `go test ./...` çalıştırır ve Go 1.26.5 kurar. Bu kapanış Postgres, RabbitMQ, RustFS, local stack ve login çalıştırmadı. Yerel komutlar `GOTOOLCHAIN=local` kullanır. Bu makinede o sürüm Go 1.26.1 olduğu için management testleri sabit `go1.26.5` kontratında kırmızı kalır. Gitignored `deploy/**/secrets` içindeki binary key dosyaları yerel fileguard'ı UTF-8 kontrolünde kırmızı yapar. CI checkout kurulum secrets ağacını içermez.

### 01 — CI ve fileguard

Kod değişmedi. Recovery fixture, salt-okunur temp cleanup ve `goTool()` sözleşmeyle uyuyordu.

```text
go test ./ops/local/cmd/recovery-state/
go test ./ops/local/cmd/secret-projector/
```

### 02 — Restore ve rotate

Yalnız `cmd/setup/secret_restore_test.go` eklendi. Persistence yazıcıları değişmedi. Secrets ağacını geri kopyalayıp setup'ı yeniden çalıştırmak SHA-256 fingerprint'lerini korur. İkinci setup sonrası `runtime/access-token-keyring.json` parse edilir. `--rotate-secrets` o fingerprint'i değiştirir. Eski cursor keyring, IAM keyring, FCART KEK keyring ve release trust root geçersiz olur. Normal ikinci setup bunları değiştirmez.

`--rotate-secrets`, `authority-manifest.json` ve `cutover-operation.json` için kurulu issue clock'u korur. Yeni trust root yine değişir. Korunan `issued_at` yeni release clock'tan erkense `not_before` bir saat geriye gider. Cutover receipt `adopted_at` orijinal işlem penceresinde kalır; pencere 12 dakikalık cutover aralığını da kapsar. Sonraki rotate aynı geçerlilik pencerelerini yerinde bırakır.

```text
go test ./cmd/setup/
```

### 03 — Management bind

`cmd/setup/management_listen_test.go` aynı makineden `127.0.0.1` dinleyicisine bağlanır. `unauthenticated_probe_test.go`, `GET /livez`, `/readyz` ve `/version` için auth header olmadan 200 döndüğünü ve gövdede parola, token, DSN veya key material olmadığını kanıtlar. Compose: süreç `0.0.0.0:9000`, health `http://127.0.0.1:9000/readyz`, host publish `127.0.0.1`. Şema default'u `0.0.0.0:9090` bırakıldı.

Kalan risk: elle yazılmış env `HARPY_MANAGEMENT_LISTEN_ADDR` atarsa süreç `0.0.0.0:9090` adresine bağlanır. Local setup her zaman `127.0.0.1:9090` yazar. Sonraki setup, `--management-listen` yoksa bu adresi korur.

```text
go test ./cmd/setup/
go test ./core/internal/platform/management/
```

### 04 — Currency drift

İkinci `go run ./tools/currencyregistry` hash'ler uyuşunca bayt-aynıdır ve yazmaz. Yazıcı yalnız iki currency kaydını yerine koyar. Hash elle düzenlenmez. CI birim işi, `checkfresh` ve `packrelease -check` sonrasında drift kapısını çalıştırır. Üretim secret'ı gerekmez.

Kaynak türevi ve deterministik: currency Go dosyası, taze currency SQL seed, taze migration manifest, release index. Kurulum başına ve deterministik olmayan: runtime keyring, compose parolaları, release trust root.

```text
go run ./tools/currencyregistry
git diff --exit-code -- contracts/generated.yaml
go test ./tools/currencyregistry/
```

### 05 — Regression kapısı

Drift kapısı `.github/workflows/ci.yml` birim işindedir. Restore, rotate pencereleri ve management erişimi `docs/harpy-starter/04-secrets-and-management.md` içindedir. Drift komutu `docs/harpy-starter/03-fileguard-archguard.md` içindedir. `go test`, `fileguard` ve `archguard` işleri durur.

Kalan riskler:

- Elle yazılmış env `HARPY_MANAGEMENT_LISTEN_ADDR` atarsa bind `0.0.0.0:9090` olur.
- `/livez`, `/readyz` ve `/version` auth'suzdur. Sınır, local loopback, konteyner süreci `0.0.0.0:9000` ve host publish `127.0.0.1` yerleşimidir.
- Local stack ve login bu planda doğrulanmadı.

```text
go test ./cmd/setup/
go test ./core/internal/platform/management/
go test ./tools/currencyregistry/
go test ./ops/local/cmd/recovery-state/
go test ./ops/local/cmd/secret-projector/
go run ./tools/fileguard
go run ./tools/archguard
```
