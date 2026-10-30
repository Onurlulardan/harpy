# Modül — IAM

## 1. Owner alanı

IAM, Harpy'nin canonical identity ve access owner modülüdür. User/credential, organization, role, permission binding, session, authentication, actor assertion ve permission evaluation bu modülde bulunur.

Kaynak kök: core/internal/modules/iam/

## 2. Domain modeli

### User ve profile

User record şu alanları taşır:
- IdentityID ve Status
- TokenVersion, AuthorizationVersion ve Revision
- IdentifierID, display/normalized email
- RoleIDs ve OrganizationIDs
- EmailVerified ve AccessModes
- ProfileDocument, CreatedAt/UpdatedAt/TombstonedAt

Access mode şu anda operator_console ve portal değerleriyle sınırlıdır. TokenVersion ve AuthorizationVersion farklı amaçlarla kullanılır: oturum/token geçerliliği ile yetki görünümünün sürümünü ayırır.

ProfileDocument user, organization veya role resource türüne bağlıdır. SchemaID, JSON values ve SHA-256 digest içerir; maksimum profile boyutu 1 MiB'tır.

### Role ve organization

Role:
- ScopeKind/ScopeID ile scope'u,
- OrganizationID,
- Name/Description,
- Status,
- Protected/Default işaretleri,
- Revision ve profile
alanlarını taşır.

Organization command'ları parent ID ve owner set'ini açıkça kabul eder. Self-parent ilişkisi ve geçersiz owner/membership kombinasyonları domain validation tarafından reddedilir.

### RBAC

RBAC modelinde assignment hedefi USER, ROLE veya ORGANIZATION olabilir. Assignment item; ordinal, action ID, permission ID ve binding ID içerir. EffectivePermission sonucu permission'ı hangi resource/action için ve hangi subject kaynağından elde ettiğini taşır.

Protected admin policy minimum active assignment, role permission binding ve authorization epoch gibi invariant'ları içerir. Mutasyonlarda expected revision ve mutation evidence istenir.

## 3. Authentication application

Authentication application aşağıdaki konuları ayrı alt paketlere ayırır:
- Password credential verification ve password policy
- Login issue/complete/consume
- Initial admin password claim/finalize
- Reauthentication ve step-up
- Browser refresh ve refresh credential rotation
- Session issue/revoke
- TOTP ve WebAuthn credential/evidence
- Device evidence ve credential purpose binding

Authenticator constructor password crypto, refresh generator/keyring, access issuer, security crypto, ID source ve clock bağımlılıklarını zorunlu tutar. Eksik dependency ile usable authenticator oluşturulmaz.

## 4. Access token

Access token signer/verifier Ed25519 kullanır:
- Signer profile issuer/audience ve key ID ister.
- Private key'in seed'den tutarlı türediği kontrol edilir.
- Token header'da type, algorithm ve key ID vardır.
- Claims issuer/audience, issued-at ve expiry doğrulamasından geçer.
- Verifier key ID üzerinden public key çözer ve signature doğrular.
- Close çağrısı private key byte'larını temizleyip signer'ı kapatır.

Kaynak: application/accesstoken/signer.go, verifier.go ve claims/encoding dosyaları.

## 5. Refresh ve reauthentication

Refresh credential opaque formatta tutulur. Secret HMAC/keyring üzerinden doğrulanır; rotation bir successor credential üretir. Önceki credential kullanımının tekrar tespiti ve client access mode binding güvenlik akışının parçasıdır.

Reauthentication amaçları password_v1, webauthn_v1 veya totp_v1 olabilir. Credential scope, doğrulamanın session/refresh/reauthentication amacıyla aynı olduğunu bağlar. Step-up evidence zaman, purpose digest ve authentication method içerir.

## 6. Authorization ve tenant sınırı

IAM authorization yalnız role name karşılaştırmasından ibaret değildir. User, role ve organization scope; permission/action/resource binding ve authorization version birlikte değerlendirilir.

- Cross-module consumer IAM internal application/domain import etmez; IAM api kontratını kullanır.
- Tenant kimliği verified actor context ile eşleştirilir.
- Protected operation için expected revision/evidence ve gerektiğinde step-up gerekir.
- Permission cache güvenilir source of truth değildir; invalidation/version kontrolleri uygulanır.
- Field ve row visibility data access anında kontrol edilir.

## 7. HTTP surface ve integrations

IAM HTTP yüzeyi API composition üzerinden bağlanır. Source tree içinde auth, lifecycle, invitation, permission request, portal user, profile, RBAC, service principal ve system resource route alanları bulunur.

Adapter grupları:
- postgres
- cryptography
- gcpkms
- passwordblocklist
- smtp
- custodymanifest
- systemtransferartifact

Key material, password hashing, email delivery ve KMS bağımlılıkları application domain'inden adapter sınırına çıkarılır.

## 8. Test alanları

IAM içinde unit, fixture, negative, security ve integration testleri bulunur. Özellikle:
- Token signature/issuer/audience/expiry
- Refresh rotation ve reuse
- Cookie/CSRF ve HTTP peer validation
- Permission/step-up ve authorization version
- Organization owner invariants
- Access mode isolation
- Cross-tenant identity ve actor binding
kontrolleri test edilmelidir.

## 9. Kaynak haritası

- Identity domain: core/internal/modules/iam/domain/identity/
- RBAC domain: core/internal/modules/iam/domain/rbac/
- Authentication app: core/internal/modules/iam/application/authentication/
- Access token: core/internal/modules/iam/application/accesstoken/
- Refresh token: core/internal/modules/iam/application/refreshtoken/
- HTTP/security wiring: core/internal/composition/api/
- IAM contracts: contracts/authority/iam_*.json, contracts/http/
- Normative skill: skills/iam-system-resources/SKILL.md
