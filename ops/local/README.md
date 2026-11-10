# Harpy local stack

Bu stack production composition yolunu yerel bagimliliklarla calistirir. API,
worker ve migrator ayni current release image'ini kullanir; API/worker startup
sirasinda migration calistirmaz. Pinned foundation ve current image ayrimi
Compose icindeki migration detayi olarak kalir, `.env.local` ayari degildir.

## Gereksinimler

- Docker Desktop veya Docker Engine + Compose v2
- Go Task (`task` komutu)
- Ilk build icin internet erisimi
- ClamAV imza yuklemesi nedeniyle Docker'a en az 4 GB RAM

Komutlari repository kokunden calistir.

## Tek komutla baslatma

```powershell
task local:up
```

Bu komut sirayla local secret/sertifika uretir, release image'larini build eder,
servis bazli least-authority secret volume'larini project eder, altyapiyi
baslatir, fresh cluster bootstrap'unu yapar, canonical migration'i uygular,
runtime servislerini acar.

Normal `local:build` ve `local:up`, digest-pinned build girdilerini mevcut
Docker/BuildKit cache'inden kullanir ve registry refresh'i zorlamaz. Ilk build
veya bilincli cache refresh'i internet ister; exact pinned girdileri yeniden
cekme yolu `task local:build:refresh` komutudur.

Ilk ClamAV baslangici imza indirme nedeniyle birkac dakika surebilir. Her
`local:infra` ve `local:up`, korunan named volume'daki imzalari once gunceller,
ardindan `clamd` container'ini volume silmeden yeniden yaratir. Boylece daemon
bellekte eski imza generation'ini tutmaz; sonraki calistirmalar volume
korundugu icin daha hizlidir. Updater yalniz bu lifecycle adiminda calisir;
sonraki runtime service baslatmalari updater'i dependency olarak yeniden kosmaz.

## Adim bazli komutlar

```powershell
task local:init       # ignored local secret ve sertifikalar
task local:config     # birlesik Compose modelini dogrula
task local:build      # dahili foundation ve current release image'lari
task local:build:refresh # digest-pinned build girdilerini explicit yenile
task local:prepare    # init + build + secret/state volume projection
task local:project    # yalniz volume projection adimi
task local:infra      # PostgreSQL, RabbitMQ, S3, KMS, SMTP, ClamAV
task local:initialize # foundation ve goose lifecycle girisi
task local:services   # API, worker ve backup executor
task local:verify     # API ve worker readiness kontrolu
task local:check      # config, unit test ve mimari guard'lar
task local:status     # backend container sagligi
```

`task dev`, `task local:up` icin; `task stop`, `task local:down` icin kisa
yoldur.

## Local adresler

| Yuzey | Adres |
| --- | --- |
| Frontend v3 | `http://localhost:5173` |
| Core API | `http://localhost:8000` |
| API management | `http://localhost:9001` |
| Worker management | `http://localhost:9002` |
| PostgreSQL | `localhost:55432` |
| RabbitMQ AMQP | `localhost:5673` |
| RabbitMQ UI | `http://localhost:15673` |
| S3 owner proxy | `http://localhost:19000` |
| Mailpit SMTP | `localhost:1026` |
| Mailpit UI | `http://localhost:8026` |

Portlar `.env.local` icindeki `LOCAL_*_PORT` degerleriyle degistirilebilir.
Frontend tarayicida ayni origin `/api` ve `/ws` yollarini kullanir; Vite bunlari
`LOCAL_API_PORT` uzerindeki Backend V2 API'sine proxy eder. Kalici ikinci bir
frontend API URL ayari yoktur.

Forms Engine ve Workflow Engine ayri container veya microservice degildir;
`core-worker` icindeki modullerdir. Isleri RabbitMQ ve PostgreSQL uzerinden
ayni modular-monolith authority sinirlariyla yuruturler.

## Log, durdurma ve sifirlama

```powershell
task local:logs          # API, worker ve backup executor
task local:logs:api
task local:logs:worker
task local:logs:infra
task local:down          # containerlari siler, named volume'lari korur
task local:destroy       # onay ister; tum local volume ve veriyi siler
task local:cleanup       # cache, kullanilmayan image ve volume'lari siler
task local:cleanup:all   # local:cleanup ile ayni
task local:cleanup:disk  # prune + Windows docker_data.vhdx compact (yonetici)
```

Normal kapatma icin `local:down` kullan. Sonraki `local:up` ayni initialize
girisini tekrar calistirir. Eski veya dolu bir veritabani fail closed durur ve
volume otomatik silinmez. `local:destroy` PostgreSQL, RabbitMQ,
S3, Mailpit, ClamAV, signed release evidence, spool ve projected secret
volume'larini kalici olarak siler; ardindan exact manifest/hash ile dogrulanan
host localconfig setini de resetler.

Host'taki `state/secrets` dizinini yalniz root one-shot projector gorur. API,
worker, initialize ve local sidecar'lar non-root calisir ve her biri yalniz
kendi explicit allowlist secret volume'ini mount eder.

## Sorun giderme

- Bir servis hazir degilse once `task local:status`, sonra ilgili log komutunu
  calistir.
- ClamAV uzun sure `starting` ise Docker bellek limitini en az 4 GB yap ve
  `task local:infra` komutunu yeniden calistir.
- Port doluysa `.env.local` icindeki ilgili local portu degistir; container ici
  portlari degistirme.
- Config/certificate expired veya bozuksa `task local:init` fail-closed olur.
  Once `task local:destroy`, sonra `task local:up` ile volume ve tool-owned host
  setini birlikte temiz bir generation'a al.
- Migration tamamlanmadan kesilen signed attempt suresi dolarsa kanit parca
  parca yenilenmez. `task local:destroy`, ardindan `task local:up` ile exact
  fresh attempt baslat.
- Tam temiz fresh-cluster denemesi icin once `task local:destroy`, sonra
  `task local:up` calistir.

`.env.local` ve `ops/local/state/` yalniz makineye ozeldir ve Git tarafindan
ignore edilir. `state/secrets` altindaki icerigi commit etme.
