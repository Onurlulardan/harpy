# 08 — Güvenlik modeli

## 1. Güvenlik katmanları

Harpy security modeli tek bir HTTP middleware'e indirgenmez. Kontroller birbirini tamamlayan katmanlardadır:
- Transport admission ve request validation
- Authentication ve actor resolution
- Tenant membership/deployment authority
- Application-level authorization
- Field/row-level data visibility
- DB role, transaction ve schema isolation
- Secret, token ve blob protection
- Audit trail ve telemetry redaction

Bir endpoint edge'de authenticated görünse bile application use-case kendi permission/owner koşullarını tekrar uygular.

## 2. Actor ve tenant authority

Actor context verified kimlik kanıtından çözülür. Tenant kimliği istemci header'ı, query param veya URL slug'ından tek başına kabul edilmez.

Request pipeline genel olarak:
1. Stable request/correlation ID oluşturulur veya doğrulanır.
2. Authenticator actor'ı çözer.
3. Tenant membership ve route authority kontrol edilir.
4. Application command TenantScope transaction'ını açar.
5. Modül command/query ve owner adapter çağrılır.
6. Commit sonrası canonical response ve telemetry tamamlanır.

Tenant header kullanılıyorsa verified actor ile uyumu doğrulanır; tenant authority kaynağı değildir.

## 3. IAM domain sahipliği

IAM canonical identity ve access modelini yönetir:
- User ve credential lifecycle
- Role ve permission bindings
- Organization/parent/owner ilişkileri
- Session ve refresh credential
- Authentication method ve step-up
- Service principal ve actor bridge
- Permission request ve authorization version

User modelinde TokenVersion ve AuthorizationVersion bulunur. Role modelinde scope, organization, protected/default durumları ve revision tutulur. Organization owner/parent invariant'ları domain command validation'ında korunur.

## 4. Access token ve refresh credential

Access token signer/verifier kaynakta Ed25519 kullanır:
- İmza algoritması Ed25519.
- Token issuer ve audience profil ile eşleşir.
- Header içindeki key ID ile public key resolver üzerinden anahtar bulunur.
- Signature kontrolü yapılır.
- Issued-at ve expiry zamanları, sınırlı clock skew ile doğrulanır.
- Token boyutu ve claim şekli sınırlandırılır.

Refresh token access token gibi taşınan uzun ömürlü JWT değildir. Opaque credential olarak üretilir; secret kısmı hash/HMAC ile doğrulanır, rotation ve reuse detection uygulanır. Browser refresh ve client-specific rotation akışları authentication application içinde ayrılmıştır.

## 5. Step-up ve MFA

Authentication modülü şu reauthentication metotlarını tanımlar:
- `password_v1`
- `webauthn_v1`
- `totp_v1`

Device evidence ve credential scope; hangi amaçla, hangi session/credential için reauthentication yapıldığını sınırlar. RBAC assignment gibi korumalı işlemler step-up evidence (purpose digest, auth time ve doğrulanan method) taşıyabilir.

## 6. RBAC ve data visibility

IAM role/permission binding'leri user, role ve organization hedefleri üzerinden modellenir. Effective permission yalnız bir role label'ı değildir; resource/action ve hangi assignment kaynağından geldiğiyle birlikte hesaplanır.

Application command:
- Global IAM action'ını,
- İlgili domain resource permission'ını,
- Gerekliyse step-up/approval kanıtını,
- Tenant ve row visibility koşullarını
kontrol eder.

Field-level permission query select/filter/sort/search davranışında da uygulanır. UI menu visibility veya projection içeriği tek başına permission grant sayılmaz.

## 7. Tenant isolation

- Tenant schema name server-side tenant ID'den türetilir ve immutable registry kaydıyla eşleştirilir.
- Bir transaction yalnız bir tenant schema'sında çalışır.
- Scope açık değilse owner DB binding yapılamaz.
- Runtime DB role yalnız kendi profiline verilen yetkiyle çalışır.
- Session/search_path temizliği ve pool reuse testleri önemlidir.
- Cross-tenant ID substitution negative testleri zorunludur.

## 8. Secret ve key management

Secret değerleri source code'a yazılmaz. Setup ve runtime config secret dosyası/secret provider üzerinden yükler. Secret değerleri:
- Log/message/error body'ye yazılmaz.
- Trace attribute veya metric label yapılmaz.
- Event/job payload'ına konulmaz.
- Public API response veya generic record DTO içinde gösterilmez.

Key purpose ve keyring ayrımı IAM, Documents, Analytics export, cursor ve platform authority gibi alanlarda uygulanır. Bir key'in başka güvenlik amacı için yeniden kullanılması engellenir. Key material kullanımı mümkün olan yerde kısa ömürlü buffer ve clear/zero davranışıyla sınırlandırılır.

## 9. Documents storage security

Documents modülü private, versioning-enabled S3 object store bekler. S3 adapter initialization sırasında bucket versioning ve private ACL koşullarını doğrular.

Dosya içeriği AES-256-GCM ile chunk'lar halinde şifrelenir. Plaintext ve ciphertext SHA-256 digest'leri receipt'e yazılır. Data Encryption Key (DEK), Key Encryption Key (KEK) altında AES-GCM ile wrap edilir. Object key tenant/document/version kimliklerinden türetilir. Yalnız `ready` version indirilebilir; current permission download anında tekrar kontrol edilir.

ClamAV scanner signature freshness ve protocol validity kontrol eder; scan tamamlanmadan dosya ready olmaz.

## 10. Audit ve redaction

Audit entry owner operation, actor, tenant/org, correlation ID, entity, action, authorization version, reason code ve field change evidence içerir. Field change classification public/internal/confidential/secret değerlerini destekler. Hassas alanlar raw before/after olarak tutulmak yerine digest ile temsil edilebilir.

## 11. Güvenlik testleri

- Header/token tenant mismatch
- Cross-tenant ID substitution
- Missing permission ve field-level access
- Refresh reuse/rotation concurrency
- CSRF, cookie ve CORS davranışı
- Query unknown field/operator ve SQL injection fuzz
- Secret/token/password response/log redaction
- Document MIME/signature/size, ciphertext tamper ve object privacy
- Provider webhook signature/replay ve idempotency

Kaynak ve prosedür: `skills/iam-system-resources/SKILL.md`, `skills/data-tenancy/SKILL.md`, `skills/api-contract/SKILL.md`, `skills/testing-quality/SKILL.md`.
