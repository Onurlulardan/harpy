// Package eventdelivery runs the durable RabbitMQ subscriptions declared by
// the event consumer catalog.
//
// A registered handler owns its complete business transaction boundary. On
// every delivery it must open a fresh TenantScope, bind the platform consumer
// repository, and commit the owner mutation with the inbox receipt and stream
// watermark in that same transaction. It must not reuse authorization from a
// previous delivery. Database or dependency failures return an error or
// HandleTransient so the adapter rejects exactly one delivery with requeue.
// HandleCommitted and HandleDuplicate are acknowledged only after their durable
// transaction has committed. HandlePermanent is rejected
// without requeue and reaches the durable dead-letter ingest path.
//
// The platform never fabricates a business handler. Every owning module must
// supply its registration together with the owner transaction and consumer
// contract; incomplete ownership is rejected by registry construction.
package eventdelivery
