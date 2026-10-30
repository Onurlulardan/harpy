# Modül — Billing

## 1. Owner alanı

Billing, money arithmetic, offering/price catalogue, campaign/discount, subscription, invoice, provider integration, entitlement ve metered usage alanlarının owner'ıdır.

Kaynak kök: core/internal/modules/billing/

## 2. Domain alt paketleri

| Paket | Sorumluluk |
| --- | --- |
| money | Typed currency/minor-unit arithmetic |
| period | Canonical billing period |
| offering | Offering/plan definition ve lifecycle |
| subscription | Subscription state ve transition |
| invoice | Invoice line, total, pricing snapshot |
| discount | Discount policy ve application |
| campaign | Campaign identity ve publication |
| entitlement | Feature/resource definition ve effective value |
| provider | Payment provider state/contract |
| usage | Metered usage, reservation, commit ve release |

## 3. Money model

Money immutable minor-unit değeridir. Alanları:
- minor units (int64)
- ISO currency code
- Currency registry version
- Currency exponent

Constructor registry version'ın current supported version ile aynı olmasını ister; currency registry'de bulunmalı ve minor unit kullanımına izin vermelidir. Aritmetik overflow checked integer ve big.Int tabanlı doğrulamalarla sınırlanır. Maximum absolute minor unit 9,000,000,000,000,000 değeridir.

UI decimal string'ini iş mantığında float64 ile toplamak yerine typed Money kullanılır. Registry kaynağı ve generated currency map birbirine bağlıdır.

## 4. Period ve subscription

Period canonical UTC instants [start,end) aralığını temsil eder. Start, end'den önce olmalıdır; Contains başlangıcı dahil, bitişi hariç tutar.

Subscription state değerleri:
- pending
- trialing
- active
- past_due
- paused
- cancelled
- expired

Cancelled ve expired terminal kabul edilir. Trialing, active ve past_due grant-bearing state olarak tanımlanır. Geçiş kuralları domain/subscription içinde tutulur; HTTP handler state ataması yapmaz.

## 5. Offering, campaigns ve discounts

Offering katmanı plan/ürün tekliflerini ve bunların version/revision doğrulamasını taşır. Campaign/discount domain'i:
- Campaign identity ve version
- Discount rule
- Fixed/percentage policy
- Duration policy
- Redemption/application constraint
- Publication/retirement behavior
alanlarını yönetir.

Fiyatlandırma kararının sonradan denetlenebilmesi için invoice line ve subscription kararlarında kaynak offering/price snapshot'ı korunur. Mutable katalog değerine geçmiş invoice'ı bağlamak doğru değildir.

## 6. Entitlement resolution

Entitlement registry her key için typed Definition tanımlar. Effective value şu deterministik katmanlarda çözülür:
1. Base offering
2. Subscription
3. Explicit override

Her katmandaki key definition içinde bulunmalı ve value definition'ın beklediği tipe uymalıdır. Bilinmeyen key veya uyumsuz value resolution hatasıdır. Bu sıra öncelik davranışını açık ve test edilebilir yapar.

## 7. Metered resource usage

Usage alt sistemi reservation/commit/decrease/release/keep gibi operasyonlar sağlar. Reservation isteği:
- Entitlement key
- Period ID
- Source operation ID
- Amount
- Expiry ve idempotency bilgisi
taşır.

Reserve, mevcut limit, committed ve reserved amount, gauge ve entitlement definition'ı birlikte değerlendirir. Aynı request tekrarında önceki reservation sonucu replay edilebilir. Soft limit ve hard admission sonucu ayrı tutulur.

Documents ve Analytics gibi consumer modülleri usage API üzerinden kaynak tüketimini kaydeder; kendi başına billing tablosuna yazmaz.

## 8. Provider entegrasyonu

Provider adapter dış servis wire protocol'ünü domain modelinden ayırır. Composition provider client ve webhook doğrulama bağımlılıklarını kurar. Provider sonucu canonical state'e normalize edilir, receipt/replay kontrolü yapılır ve subscription/invoice state application katmanından güncellenir.

## 9. HTTP ve platform API

Billing route grupları:
- Public tenant subscription/license ve metered usage
- Platform offering/catalog/campaign/discount/reporting
- Invoice ve provider/webhook
- Retention/legal hold ve audit read surface'leri

Exact paths ve schema için contracts/http/billing* dosyaları, route registration için composition/api/billing_routes.go ve ilgili billing HTTP builder'ları incelenir.

## 10. Kaynak haritası

- Domain: core/internal/modules/billing/domain/
- Application: core/internal/modules/billing/application/
- Postgres/provider/crypto adapters: core/internal/modules/billing/adapters/
- Metered resource API: core/internal/modules/billing/api/
- License runtime composition: core/internal/composition/api/
- Worker: core/internal/composition/worker/build_billing_*.go
- Contracts: contracts/authority/billing_*.json, contracts/http/billing*
- Skill: skills/billing/SKILL.md
