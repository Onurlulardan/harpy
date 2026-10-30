# 07 — Event, job ve worker mimarisi

## 1. İki farklı asenkron model

Harpy'de iki kavram ayrı tutulur:
1. **DB job:** Bir işin ne zaman ve hangi state ile çalışacağını belirleyen canonical kayıt.
2. **Integration event:** Bir process'in başka process veya consumer'lara commit sonrası duyurduğu olay.

RabbitMQ transport ve wake-up katmanıdır. İşin schedule/state authority'si veya tek doğruluk kaynağı değildir.

## 2. Transactional outbox

Business state değişikliğiyle entegrasyon event'i aynı DB transaction'ında outbox'a yazılır. Böylece business commit tamamlanmadan event dışarı çıkmaz; commit başarılıysa publisher daha sonra event'i teslim etmeye çalışabilir.

Envelope alanları:
- `event_id`, `event_type`, `event_version`
- `tenant_id`
- `aggregate_type`, `aggregate_id`, `aggregate_version`, `event_sequence`
- `occurred_at`, `actor_id`
- `correlation_id`, `causation_id`, `idempotency_key`
- Versioned payload

Payload küçük ve secret içermeyecek şekilde tutulur. Event schema ve producer/consumer policy `contracts/events/` ve async catalog tanımlarında bulunur.

## 3. Outbox publisher

Worker composition:
1. Async release/producer/consumer catalog'larını yükler.
2. Canlı DB catalog'ının beklenen producer kayıtlarıyla uyumunu doğrular.
3. RabbitMQ bağlantısını ve gerekli exchange/consumer yüzeyini probe eder.
4. Outbox publisher store ve runner'ı kurar.
5. Worker admission gate'iyle runner'ın yaşam döngüsünü yönetir.

Publisher, RabbitMQ publisher confirmation almadan outbox kaydını delivered olarak işaretlemez. Publish başarısızlığında attempt evidence ve retry delay kaydedilir. Publish sonrası process crash olursa duplicate delivery mümkün olduğundan consumer tarafı idempotent tasarlanır.

## 4. Consumer ve inbox

Consumer event envelope'u doğrular, version/producer/consumer policy'yi uygular ve kendi owner state değişikliğini idempotent işler.

Aynı event birden fazla kez gelebilir. Consumer effect'i ve inbox receipt aynı transaction'da tutulur; `(consumer, event_id)` unique kuralı duplicate effect'i engeller. Aggregate ordering için `event_sequence` kullanılır; aggregate version'ın kesintisiz olacağı varsayılmaz.

## 5. DB job modeli

Job state DB'de canonical tutulur. Genel job kaydı şu tür alanları kapsar:
- `kind`, `tenant_id`, `payload_ref`, `run_at`, `state`
- `attempts`, `max_attempts`
- `lease_owner`, `lease_until`
- `dedupe_key`, `last_error`
- `created_at`, `completed_at`

Claim işlemi row lock ve `FOR UPDATE SKIP LOCKED` benzeri concurrency kontrolüyle yapılır. Lease ve heartbeat stale worker'ın bıraktığı işi geri alınabilir hale getirir.

## 6. Retry ve failure sınıfları

Hatalar tek bir retry kategorisine indirgenmez:
- **Transient:** Geçici dependency/network hatası; backoff+jitter ile tekrar denenebilir.
- **Permanent:** Geçersiz payload, yetki veya desteklenmeyen işlem; otomatik retry edilmez.
- **Conflict:** State/version değişmiş; business-specific reconciliation gerekir.
- **Outcome unknown:** Provider sonucu kesin bilinmiyor; kanıt olmadan duplicate external effect yaratılmaz.
- **Cancelled:** Operasyon açıkça iptal edilmiş.

Attempt limit dolduğunda iş terminal state'e veya dead-letter/operator-visible akışa gider. Manual retry yeni audit/causation bilgisi üretir.

## 7. Worker workload grupları

`core/internal/composition/worker/` modül işlerini bir araya getirir. Kaynakta görülen ana gruplar:
- Billing provider, subscription ve retention processor'ları
- Catalog lifecycle ve customer provisioning processor'ları
- Documents archive/upload/deletion ve ilgili job'lar
- IAM invitation/delivery, maintenance ve system transfer processor'ları
- Notifications delivery ve retention
- Work reminder/recurrence/projection job'ları
- Analytics dashboard export ve query-related job'ları
- Event gap, dead-letter ve async retention workload'ları

Her processor kendi domain owner's application/API'sine bağlanır. Worker process bütün business logic'i kendi içine kopyalamaz.

## 8. Async supervisor

`asyncSupervisor`:
- Workload listesi ve failure channel oluşturur.
- Start aşamasında workload'ları sırayla başlatır.
- Başlatma hatasında daha önce açılmış workload'ları Force ile kapatır.
- Failures channel'larını takip eder.
- BeginDrain ile yeni iş alımını durdurur.
- Wait ile aktif workload'ların tamamlanmasını bekler.
- Force ile zorunlu kapatma uygular.

Bu mekanizma worker shutdown'ının tek bir processor'a bağlı kalmadan yönetilmesini sağlar.

## 9. Observability

Async operasyonlarda backlog, attempt, retry delay, lease expiry, terminal failure ve consumer lag gözlenebilir olmalıdır. Harpy'nin telemetry paketi async backlog/retention observer ve instruments içerir. Event/job payload'ına credential veya hassas belge içeriği yazılmaz.

## 10. Kaynak kod haritası

- Async base: `core/internal/platform/async/`
- Job execution: `core/internal/platform/async/jobexecution/`
- Outbox delivery: `core/internal/platform/async/outboxdelivery/`
- Event delivery: `core/internal/platform/async/eventdelivery/`
- Idempotency: `core/internal/platform/async/idempotency/`
- PostgreSQL/Rabbit adapters: `core/internal/platform/async/adapters/`
- Worker composition: `core/internal/composition/worker/`
- Async contracts: `contracts/events/`, `contracts/asyncmanifest/`

Normatif prosedür: `skills/events-jobs/SKILL.md`.
