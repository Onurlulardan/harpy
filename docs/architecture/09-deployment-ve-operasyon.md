# 09 — Deployment ve operasyon

## 1. Deployment profilleri

Repository'de iki standart compose profili vardır:
- deploy/single-tenant/compose.yml
- deploy/multi-tenant/compose.yml

Her iki profil ortak third-party stack'i deploy/third-party.yml üzerinden include eder. Profil dosyası TENANCY_MODE değerini seçer ve gerekli secret dosyalarını compose secret olarak tanımlar.

| Profil | Mod | Kullanım |
| --- | --- | --- |
| single-tenant | single | Tek müşteri/on-premise ortamı |
| multi-tenant | multi | SaaS tenant yönetimi |

Runtime deployment config ayrıca DEPLOYMENT_MODE=onpremise|saas değerini kullanır. Compose tenant mode ve uygulama deployment mode birbiriyle uyumlu olmalıdır.

## 2. Üçüncü taraf servisler

deploy/third-party.yml ortak servis tanımlarının sahibidir. Image tag yanında digest pin'i bulunur.

| Servis | Rol | Veri |
| --- | --- | --- |
| PostgreSQL 18 + pgvector | Transactional DB ve vector extension | postgres-data |
| RabbitMQ 4 management | Event transport ve broker yönetimi | rabbitmq-data |
| RustFS | Yerel S3-compatible object store | rustfs-data |

Servisler restart: unless-stopped ile tanımlanır. Healthcheck'ler pg_isready, rabbitmq-diagnostics ping ve RustFS /health endpoint'ini kontrol eder. Port yayınları loopback'e bind edilir; yerel geliştirme servisleri varsayılan olarak public network'e açılmaz.

RustFS S3 API portu 9000, console portu 9001'dir. Tek data volume /data üzerinde tutulur. Access/secret key dosya tabanlı secret olarak container'a aktarılır.

## 3. Setup ve secret dosyaları

cmd/setup profili ve local config'i hazırlar. Profilin .env dosyası ile secrets/ dizini deployment ayarlarını ve secret materyalini taşır. Secret dosyaları source control'e alınmaz; .gitignore bu local path'leri dışarıda bırakır.

Secret rotation gerekiyorsa setup aracının açık rotate seçeneği vardır. Normal setup tekrarında önceki runtime secret materyalinin yanlışlıkla üzerine yazılmaması hedeflenir.

## 4. Local stack topolojisi

Local stack ops/local altında compose parçalarına bölünmüştür:
- compose.infra.yml: PostgreSQL, RabbitMQ, S3/MinIO, KMS, SMTP ve scanner dependency'leri.
- compose.tools.yml: Local setup ve kontrol utility'leri.
- compose.runtime.yml: API ve worker gibi uygulama process'leri.
- compose.recovery.yml: Recovery/test operasyonları için izole bileşenler.
- compose.inventory.yml: Secret inventory kontrolleri.
- compose.lifecycle.yml: Local lifecycle servisleri.

Bu dosyalar .taskfiles/local.yml içindeki ortak COMPOSE komutuyla bir araya getirilir.

## 5. Local developer prerequisites

Local stack README'sinde belirtilen gereksinimler:
- Docker Desktop veya Docker Engine + Compose v2
- Go Task (task)
- İlk build için internet erişimi
- ClamAV signature yüklemesi için Docker'a en az 4 GB RAM
- Sibling frontend-v3 çalıştırılacaksa Node.js 22+ ve Yarn 1

Backend tek başına geliştirilirken frontend toolchain'i zorunlu değildir.

## 6. Local giriş komutları

| Komut | Sonuç |
| --- | --- |
| task dev | Tam local stack'i başlatır (local:up kısayolu). |
| task stop | Stack'i durdurur ve named volume'ları korur. |
| task local:init | Ignore edilen local secret/sertifikaları hazırlar. |
| task local:config | Birleşik Compose modelini doğrular. |
| task local:build | Pinned build image'larını cache kullanarak build eder. |
| task local:infra | Local infra servislerini başlatır ve probe eder. |
| task local:logs | Backend ve infra loglarını izler. |
| task local:logs:api / task local:logs:worker | Process loglarını ayrı izler. |
| task local:down | Container'ları durdurur; volume'ları korur. |
| task local:destroy | Onay sonrası local DB, broker ve object volume'larını siler. |
| task local:check | Local tool testlerini ve repository guard'larını çalıştırır. |

## 7. Local endpoint referansı

ops/local/README.md'de verilen varsayılan portlar:

| Yüzey | Adres |
| --- | --- |
| Sibling frontend | http://localhost:5173 |
| Core API | http://localhost:8000 |
| API management | http://localhost:9001 |
| Worker management | http://localhost:9002 |
| Schema executor management | http://localhost:9003 |
| PostgreSQL | localhost:55432 |
| RabbitMQ AMQP | localhost:5673 |
| RabbitMQ management | http://localhost:15673 |
| S3 owner proxy | http://localhost:19000 |
| Mailpit SMTP/UI | localhost:1026 / http://localhost:8026 |

Portlar .env.local içindeki LOCAL_*_PORT ayarlarıyla değiştirilebilir.

## 8. Gözlem ve ilk seviye troubleshooting

Arıza incelemesini şu sırada yap:
1. Container health ve restart count.
2. İlgili API/worker management readiness.
3. Process logları ve correlation/operation bilgisi.
4. PostgreSQL, RabbitMQ, S3 owner proxy, KMS, SMTP ve scanner dependency health'i.
5. API request mi, worker job mı, operator process mi başarısız ayrımı.
6. Secret/config/schema contract hatası ile business authorization hatasını ayır.

## 9. Kaynak haritası

- Standart deployment: deploy/single-tenant/, deploy/multi-tenant/
- Third-party services: deploy/third-party.yml
- Local compose: ops/local/compose.*.yml
- Task workflows: .taskfiles/local*.yml, Taskfile.yml
- Local operations guide: ops/local/README.md
- Local generated state: ops/local/state/ (source control dışında tutulur)
