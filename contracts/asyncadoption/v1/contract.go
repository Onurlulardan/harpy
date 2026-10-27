package asyncadoptionv1

const (
	ItemContractVersion          = "async-adoption-item.v1"
	InventoryContractVersion     = "async-adoption-inventory.v1"
	DispositionContractVersion   = "async-adoption-disposition.v1"
	RestartFenceContractVersion  = "async-adoption-restart-fence.v1"
	ZeroInventoryContractVersion = "async-adoption-zero-inventory.v1"
	AcceptanceContractVersion    = "async-adoption-acceptance.v1"

	MaximumDocumentBytes  = 262_144
	MaximumNameBytes      = 128
	MaximumIdentityBytes  = 512
	MaximumReferenceBytes = 2_048
	MaximumReasonBytes    = 2_048
	MaximumTicketBytes    = 256
	MaximumInventoryItems = 100_000
)

type InventoryClass string

const (
	ClassQueue   InventoryClass = "queue"
	ClassTimer   InventoryClass = "timer"
	ClassFuture  InventoryClass = "future"
	ClassRunning InventoryClass = "running"
)

type SourceSurface string

const (
	SurfaceQueue         SourceSurface = "queue"
	SurfaceScheduler     SourceSurface = "scheduler"
	SurfaceDelayedPlugin SourceSurface = "delayed_plugin"
	SurfaceTTLQueue      SourceSurface = "ttl_queue"
	SurfaceFutureState   SourceSurface = "future_state"
	SurfaceRunningState  SourceSurface = "running_state"
)

type EffectClass string

const (
	EffectDatabaseTransactional EffectClass = "database_transactional"
	EffectExternal              EffectClass = "external"
)

type Disposition string

const (
	DispositionAdopted   Disposition = "adopted"
	DispositionCancelled Disposition = "cancelled"
)

type FenceKind string

const (
	FenceEnqueue   FenceKind = "enqueue"
	FenceScheduler FenceKind = "scheduler"
	FenceConsumer  FenceKind = "consumer"
)

const (
	V1CronScheduler          = "workflow.cron.scheduler"
	V1ScheduleQueue          = "schedule.events"
	V1DelayQueue             = "workflow.delay"
	V1WebhookRetryQueue      = "workflow.webhook.retry"
	V1FormEventQueue         = "form.events"
	V1ImportFutureState      = "forms.import.future_state"
	V1QueueContract          = "v1.queue-message.v1"
	V1SchedulerContract      = "v1.workflow-minute-scheduler.v1"
	V1DelayContract          = "v1.workflow-delayed-plugin.v1"
	V1WebhookTTLContract     = "v1.workflow-webhook-ttl-not-delay.v1"
	V1ImportContract         = "v1.forms-import-job.v1"
	V1IdempotencyContract    = "v1.redis-idempotency-cache.v1"
	V1IdempotencyDisposition = "invalidate_at_cutover"
)
