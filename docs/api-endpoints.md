# Harpy API endpoint kataloğu

Yayındaki işlem sayısı **240**. Bunun 234’ü Core API, 6’sı ayrı management portu. Sayı method ile path çiftidir: `GET` ve `HEAD` iki işlemdir. `HEAD` aynı adresin gövdesiz kontrolüdür.

Kaynak, `contracts/http/core-api-v1.openapi.yaml` ve `contracts/http/management-v1.openapi.yaml` dosyalarıdır.

## Management (6)

Süreç ayakta mı, bağımlılıklar hazır mı ve hangi sürüm çalışıyor, bunu söyler. İş verisi taşımaz.

- `GET /livez` — Süreç olay döngüsü yaşıyor mu, buna bakar.
- `HEAD /livez` — Aynı canlılık kontrolünü gövdesiz yapar.
- `GET /readyz` — API’nin istek almaya hazır olup olmadığını söyler.
- `HEAD /readyz` — Aynı hazır olma kontrolünü gövdesiz yapar.
- `GET /version` — Çalışan sürümü döner.
- `HEAD /version` — Aynı sürüm bilgisini gövdesiz döner.

## Uygulama (4)

İstemcinin açılışta ihtiyaç duyduğu genel bilgiyi verir.

- `GET /.well-known/jwks.json` — Oturum token’ını doğrulamak için kullanılan açık anahtarları döner.
- `HEAD /.well-known/jwks.json` — Aynı anahtar setinin değişip değişmediğini gövdesiz kontrol eder.
- `GET /api/app/config` — Kurulumun on-prem mi SaaS mı olduğunu ve varsayılan tenant bilgisini döner.
- `HEAD /api/app/config` — Aynı kurulum bilgisini gövdesiz kontrol eder.

## Giriş (15)

Kim olduğunu kanıtlaman, oturum açman ve tarayıcı oturumunu yenilemen içindir.

- `GET /api/auth/csrf` — Tarayıcı istekleri için CSRF jetonu verir.
- `POST /api/auth/csrf` — CSRF jetonunu yeniler.
- `POST /api/auth/login` — E-posta ve parola ile operatör oturumu açar.
- `POST /api/auth/logout` — Tarayıcı oturumunu kapatır.
- `POST /api/auth/refresh` — Süresi dolan tarayıcı oturumunu yeniler.
- `GET /api/auth/session` — Şu anki oturumun kim olduğunu döner.
- `POST /api/auth/initial-password` — İlk kurulum parolasını kalıcı parolaya çevirir.
- `POST /api/auth/initial-admin/setup/complete` — İlk yöneticinin kurulumunu bitirir.
- `POST /api/auth/password/reset` — Parola sıfırlama isteği başlatır.
- `POST /api/auth/password/reset/complete` — Parola sıfırlamayı yeni parolayla bitirir.
- `POST /api/auth/identifiers/verify` — E-posta gibi bir kimlik adresinin doğrulama kodunu kabul eder.
- `POST /api/auth/reauthentication/start` — Hassas işlem öncesi yeniden doğrulamayı başlatır.
- `POST /api/auth/reauthentication/complete` — Yeniden doğrulamayı bitirir.
- `POST /api/auth/post-restore/challenge` — Geri yükleme sonrası sahip hesabı için doğrulama sorusu üretir.
- `POST /api/auth/post-restore/reactivate-owner` — Geri yükleme sonrası sahip hesabını yeniden açar.

## Hesap güvenliği (17)

Giriş yaptıktan sonra parolanı, ikinci faktörünü ve açık oturumlarını yönetmen içindir.

- `POST /api/iam/password/change` — Oturum açıkken parolayı değiştirir.
- `POST /api/iam/identifiers` — Hesaba yeni bir e-posta veya kimlik adresi ekler.
- `POST /api/iam/identifiers/{id}/verification` — Bu adres için doğrulama gönderir.
- `POST /api/iam/identifiers/{id}/retire` — Adresi kullanımdan çıkarır.
- `POST /api/iam/identifiers/{id}/tombstone` — Adresi geri dönülmez şekilde kapatır.
- `GET /api/iam/authenticators` — TOTP, WebAuthn ve kurtarma kodu kayıtlarını listeler.
- `POST /api/iam/authenticators/totp/enrollments` — TOTP kurulumunu başlatır.
- `POST /api/iam/authenticators/totp/enrollments/{authenticator_id}/complete` — TOTP kurulumunu kodla bitirir.
- `POST /api/iam/authenticators/webauthn/enrollments` — Güvenlik anahtarı kaydını başlatır.
- `POST /api/iam/authenticators/webauthn/enrollments/{authenticator_id}/complete` — Güvenlik anahtarı kaydını bitirir.
- `GET /api/iam/authenticators/recovery-codes` — Kurtarma kodlarının durumunu döner.
- `POST /api/iam/authenticators/recovery-codes` — Yeni kurtarma kodları üretir.
- `POST /api/iam/authenticators/{id}/suspend` — Bir doğrulayıcıyı geçici durdurur.
- `POST /api/iam/authenticators/{id}/retire` — Bir doğrulayıcıyı kalıcı kapatır.
- `GET /api/iam/sessions` — Hesabın açık oturumlarını listeler.
- `DELETE /api/iam/sessions/{id}` — Tek bir oturumu kapatır.
- `DELETE /api/iam/sessions` — Hesabın bütün oturumlarını kapatır.

## Profil (7)

Giriş yapan kişinin kendi kartı ve kullanıcı, rol, organizasyon kayıtlarına eklenen özel alanlar içindir.

- `GET /api/profile` — Giriş yapan kişinin profilini döner.
- `PUT /api/profile` — Giriş yapan kişinin profilini günceller.
- `GET /api/iam/profile-schemas/{resource}/active` — Kullanıcı, rol veya organizasyon için yürürlükteki özel alan şemasını döner.
- `GET /api/iam/profile-schemas/{resource}/{schema_id}` — Belirli bir şema sürümünü döner.
- `POST /api/iam/profile-schemas/{resource}/publish` — Yeni şema sürümünü yürürlüğe alır.
- `POST /api/iam/profile-schemas/{resource}/fields/{field_id}/retire` — Bir özel alanı kullanımdan çıkarır.
- `POST /api/iam/profile-schemas/{resource}/backfills/{plan_id}/recover` — Yarım kalmış şema doldurma işini sürdürür.

## Kullanıcı, rol ve organizasyon (27)

Sistemdeki kişi, rol ve organizasyon kayıtlarını tutmak içindir. Üçü de aynı kayıt düzenini kullanır.

- `GET /api/users` — Kullanıcıları listeler.
- `POST /api/users` — Kullanıcı oluşturur.
- `POST /api/users/query` — Kullanıcıları filtreli arar.
- `DELETE /api/users` — Seçilen kullanıcıları toplu siler.
- `GET /api/users/{id}` — Tek kullanıcıyı döner.
- `PUT /api/users/{id}` — Kullanıcı kaydını günceller.
- `PATCH /api/users/{id}/state` — Kullanıcının yaşam durumunu değiştirir.
- `DELETE /api/users/{id}` — Tek kullanıcıyı siler.
- `GET /api/users/{id}/permissions` — Kullanıcıya fiilen uygulanmış yetkileri döner.
- `GET /api/roles` — Rolleri listeler.
- `POST /api/roles` — Rol oluşturur.
- `POST /api/roles/query` — Rolleri filtreli arar.
- `DELETE /api/roles` — Seçilen rolleri toplu siler.
- `GET /api/roles/{id}` — Tek rolü döner.
- `PUT /api/roles/{id}` — Rol kaydını günceller.
- `PATCH /api/roles/{id}/state` — Rolün yaşam durumunu değiştirir.
- `DELETE /api/roles/{id}` — Tek rolü siler.
- `GET /api/roles/{id}/permissions` — Role fiilen uygulanmış yetkileri döner.
- `GET /api/organizations` — Organizasyonları listeler.
- `POST /api/organizations` — Organizasyon oluşturur.
- `POST /api/organizations/query` — Organizasyonları filtreli arar.
- `DELETE /api/organizations` — Seçilen organizasyonları toplu siler.
- `GET /api/organizations/{id}` — Tek organizasyonu döner.
- `PUT /api/organizations/{id}` — Organizasyon kaydını günceller.
- `PATCH /api/organizations/{id}/state` — Organizasyonun yaşam durumunu değiştirir.
- `DELETE /api/organizations/{id}` — Tek organizasyonu siler.
- `GET /api/organizations/{id}/permissions` — Organizasyona fiilen uygulanmış yetkileri döner.

## Yetki kataloğu (17)

Hangi kaynağa hangi işlemin izinli olduğunu tanımlamak ve bir karar sormak içindir.

- `GET /api/permissions` — Tanımlı izinleri listeler.
- `POST /api/permissions` — Yeni izin oluşturur.
- `GET /api/permissions/{id}` — Tek izni döner.
- `PUT /api/permissions/{id}` — İzni günceller.
- `DELETE /api/permissions/{id}` — İzni siler.
- `POST /api/permissions/check` — Tek bir yetki kararını sorar.
- `POST /api/permissions/batch-check` — Birden fazla yetki kararını tek istekte sorar.
- `GET /api/permissions/resources` — Yetki kaynaklarını listeler.
- `POST /api/permissions/resources` — Yetki kaynağı oluşturur.
- `GET /api/permissions/resources/{id}` — Tek kaynağı döner.
- `PUT /api/permissions/resources/{id}` — Kaynağı günceller.
- `DELETE /api/permissions/resources/{id}` — Kaynağı siler.
- `GET /api/permissions/actions` — Yetki işlemlerini listeler.
- `POST /api/permissions/actions` — Yetki işlemi oluşturur.
- `GET /api/permissions/actions/{id}` — Tek işlemi döner.
- `PUT /api/permissions/actions/{id}` — İşlemi günceller.
- `DELETE /api/permissions/actions/{id}` — İşlemi siler.

## Yetki talepleri (10)

Birinin eksik izni istemesi ve yöneticinin bu isteği karara bağlaması içindir.

- `GET /api/permission-requests` — Yetki taleplerini listeler.
- `GET /api/permission-requests/mine` — Giriş yapan kişinin kendi taleplerini listeler.
- `POST /api/permission-requests` — Yeni yetki talebi açar.
- `GET /api/permission-requests/{id}` — Tek talebi döner.
- `POST /api/permission-requests/{id}/review` — Talebi onaylar veya reddeder.
- `POST /api/permission-requests/{id}/cancel` — Talebi geri çeker.
- `POST /api/permission-requests/{id}/expire` — Süresi dolmuş talebi kapatır.
- `GET /api/permission-requests/catalog/resources` — Talep edilebilecek kaynakları listeler.
- `GET /api/permission-requests/catalog/actions` — Talep edilebilecek işlemleri listeler.
- `GET /api/permission-requests/catalog/missing-resources` — Katalogda eksik görünen kaynakları listeler.

## Davet (6)

Birini tenant’a e-posta davetiyle almak içindir.

- `GET /api/invitations` — Davetleri listeler.
- `POST /api/invitations` — Davet oluşturur.
- `GET /api/invitations/{id}` — Tek daveti döner.
- `DELETE /api/invitations/{id}` — Daveti geri alır.
- `GET /api/invitations/verify` — Davet bağlantısının geçerli olup olmadığına bakar.
- `POST /api/invitations/accept` — Daveti kabul edip hesabı bağlar.

## Kayıt aktarımı (9)

Kullanıcı, rol ve organizasyon kayıtlarını dosyayla dışarı ve içeri almak içindir.

- `POST /api/iam/system-resources/{resource}/record-exports` — Dışa aktarma işi başlatır.
- `GET /api/iam/system-resources/{resource}/record-exports/{export_id}` — Dışa aktarma işinin durumunu döner.
- `GET /api/iam/system-resources/{resource}/record-exports/{export_id}/content` — Hazır dışa aktarma dosyasını indirir.
- `DELETE /api/iam/system-resources/{resource}/record-exports/{export_id}` — Dışa aktarma işini iptal eder.
- `POST /api/iam/system-resources/{resource}/record-import-sources` — İçe aktarılacak dosyayı yükler.
- `POST /api/iam/system-resources/{resource}/record-imports` — İçe aktarma işi başlatır.
- `GET /api/iam/system-resources/{resource}/record-imports/{import_id}` — İçe aktarma işinin durumunu döner.
- `GET /api/iam/system-resources/{resource}/record-imports/{import_id}/report` — İçe aktarma sonucunun raporunu döner.
- `DELETE /api/iam/system-resources/{resource}/record-imports/{import_id}` — İçe aktarma işini iptal eder.

## Ekran düzeni (7)

Kullanıcı, rol ve organizasyon ekranında hangi alanın nerede duracağını tutmak içindir. Form uygulaması değildir.

- `GET /api/system-resources/{resource}/presentation` — Ekranda kullanılacak yayımlanmış düzeni döner.
- `GET /api/system-resources/{resource}/layout-revisions` — Düzen sürümlerini listeler.
- `GET /api/system-resources/{resource}/layout-revisions/{revision_id}` — Tek düzen sürümünü döner.
- `PUT /api/system-resources/{resource}/layout-revisions/{revision_id}` — Düzen sürümünü yazar.
- `PATCH /api/system-resources/{resource}/layout-revisions/{revision_id}` — Düzen sürümünü kısmen günceller.
- `DELETE /api/system-resources/{resource}/layout-revisions/{revision_id}` — Düzen sürümünü siler.
- `PUT /api/system-resources/{resource}/layout-revisions/{revision_id}/publication` — Bu sürümü ekranda yürürlüğe alır.

## Kiracı (3)

SaaS’ta yeni müşteri tenant’ı açmak içindir.

- `POST /api/platform/tenants` — Yeni müşteri tenant’ı açma işini başlatır.
- `GET /api/platform/tenant-provisioning-operations/{operation_id}` — Tenant açma işinin durumunu döner.
- `POST /api/platform/tenants/{tenant_id}/admin-setup/reissue` — O tenant’ın ilk yönetici kurulumunu yeniden üretir.

## Dosya (30)

Klasör, dosya, yükleme ve arşiv içindir.

- `GET /api/folders` — Klasörleri listeler.
- `POST /api/folders` — Klasör oluşturur.
- `GET /api/folders/{folder_id}` — Tek klasörü döner.
- `PUT /api/folders/{folder_id}` — Klasörü günceller.
- `DELETE /api/folders/{folder_id}` — Klasörü siler.
- `GET /api/folders/{folder_id}/contents` — Klasörün içindeki öğeleri listeler.
- `GET /api/folders/{folder_id}/download` — Klasörü arşiv olarak indirir.
- `POST /api/folders/{folder_id}/move` — Klasörü başka bir klasörün altına taşır.
- `POST /api/folders/{folder_id}/restore` — Silinmiş klasörü geri alır.
- `GET /api/documents` — Dosyaları listeler.
- `POST /api/documents/batch` — Birden fazla dosyaya aynı işlemi uygular.
- `GET /api/documents/{document_id}` — Tek dosyayı döner.
- `PUT /api/documents/{document_id}` — Dosya kaydını günceller.
- `DELETE /api/documents/{document_id}` — Dosyayı siler.
- `POST /api/documents/{document_id}/move` — Dosyayı başka klasöre taşır.
- `POST /api/documents/{document_id}/copy` — Dosyanın kopyasını çıkarır.
- `POST /api/documents/{document_id}/restore` — Silinmiş dosyayı geri alır.
- `POST /api/documents/{document_id}/purge` — Dosyayı kalıcı silme kuyruğuna alır.
- `POST /api/documents/{document_id}/purge/cancel` — Kalıcı silmeyi geri çeker.
- `GET /api/documents/{document_id}/download` — Dosyanın güncel sürümünü indirir.
- `GET /api/documents/{document_id}/versions` — Dosya sürümlerini listeler.
- `GET /api/documents/{document_id}/versions/latest` — En son sürümün kaydını döner.
- `GET /api/documents/{document_id}/versions/{version_id}/content` — Seçilen sürümün içeriğini indirir.
- `POST /api/document-uploads` — Yeni yükleme oturumu açar.
- `GET /api/document-uploads/{upload_id}` — Yüklemenin durumunu döner.
- `PUT /api/document-uploads/{upload_id}/content` — Yükleme içeriğini yazar.
- `POST /api/document-uploads/{upload_id}/cancel` — Yüklemeyi iptal eder.
- `POST /api/document-archives` — Seçilen dosyalardan arşiv üretme işi başlatır.
- `GET /api/document-archives/{archive_id}` — Arşiv işinin durumunu döner.
- `GET /api/document-archives/{archive_id}/content` — Hazır arşivi indirir.

## Bildirim (10)

Kullanıcının gelen kutusu, tercihleri ve canlı bildirim bağlantısı içindir.

- `GET /api/notifications` — Gelen kutusunu listeler.
- `GET /api/notifications/{notification_id}` — Tek bildirimi döner.
- `PUT /api/notifications/{notification_id}/read` — Bildirimi okundu işaretler.
- `PUT /api/notifications/mark-all-as-read` — Gelen kutusunun tamamını okundu işaretler.
- `DELETE /api/notifications/{notification_id}` — Bildirimi gelen kutusundan kaldırır.
- `GET /api/notification-preferences` — Bildirim tercihlerini döner.
- `PUT /api/notification-preferences` — Bildirim tercihlerini günceller.
- `POST /api/notification-intents` — Gönderilecek bir bildirimi kuyruğa alır.
- `POST /api/notification-realtime-tickets` — Canlı bildirim bağlantısı için kısa ömürlü bilet verir.
- `GET /ws/notifications` — Canlı bildirim akışına WebSocket ile bağlanır.

## Kiracı faturalaması (9)

Tenant’ın kendi planını, indirimini ve kullanımını görmesi içindir.

- `GET /api/billing/v2/offerings` — Satın alınabilecek planları listeler.
- `GET /api/billing/v2/promotions/active` — Yürürlükteki kampanyayı döner.
- `POST /api/billing/v2/discounts/validate` — İndirim kodunun geçerli olup olmadığına bakar.
- `POST /api/billing/v2/subscriptions/current/discount-redemptions` — İndirim kodunu mevcut aboneliğe uygular.
- `GET /api/billing/v2/subscriptions/current` — Tenant’ın güncel aboneliğini döner.
- `POST /api/billing/v2/subscriptions/current/changes` — Plan değişikliğini ileri bir tarihe yazar.
- `POST /api/billing/v2/subscriptions/current/cancellations` — Abonelik iptalini planlar.
- `POST /api/billing/v2/subscriptions/current/cancellations/resume` — Planlanmış iptali geri alır.
- `GET /api/billing/v2/usage` — Tenant’ın ölçülen kullanımını döner.

## Platform faturalaması (25)

Plan, kampanya, indirim ve abonelik kataloğunu işleten tarafın yönetimidir. Ödeme sağlayıcısının haberi de buraya girer.

- `GET /api/platform/billing/offerings` — Planları listeler.
- `POST /api/platform/billing/offerings` — Plan oluşturur.
- `POST /api/platform/billing/offerings/query` — Planları filtreli arar.
- `GET /api/platform/billing/offerings/{offering_id}` — Tek planı döner.
- `POST /api/platform/billing/offerings/{offering_id}/versions` — Plana yeni fiyat sürümü ekler.
- `POST /api/platform/billing/offerings/{offering_id}/retire` — Planı satıştan kaldırır.
- `GET /api/platform/billing/campaigns` — Kampanyaları listeler.
- `POST /api/platform/billing/campaigns` — Kampanya oluşturur.
- `POST /api/platform/billing/campaigns/query` — Kampanyaları filtreli arar.
- `GET /api/platform/billing/campaigns/{campaign_id}` — Tek kampanyayı döner.
- `POST /api/platform/billing/campaigns/{campaign_id}/versions` — Kampanyaya yeni sürüm ekler.
- `POST /api/platform/billing/campaigns/{campaign_id}/retire` — Kampanyayı sonlandırır.
- `GET /api/platform/billing/discounts` — İndirimleri listeler.
- `POST /api/platform/billing/discounts` — İndirim oluşturur.
- `POST /api/platform/billing/discounts/query` — İndirimleri filtreli arar.
- `GET /api/platform/billing/discounts/{discount_id}` — Tek indirimi döner.
- `POST /api/platform/billing/discounts/{discount_id}/versions` — İndirime yeni sürüm ekler.
- `POST /api/platform/billing/discounts/{discount_id}/retire` — İndirimi sonlandırır.
- `GET /api/platform/billing/subscriptions` — Abonelikleri listeler.
- `POST /api/platform/billing/subscriptions/query` — Abonelikleri filtreli arar.
- `POST /api/platform/billing/subscriptions/{subscription_id}/transitions` — Aboneliğin durumunu değiştirir.
- `GET /api/platform/billing/events` — Faturalama denetim kayıtlarını listeler.
- `POST /api/platform/billing/events/query` — Faturalama denetim kayıtlarını filtreli arar.
- `POST /api/platform/billing/revenue/query` — Gelir özetini sorar.
- `POST /webhooks/billing/{provider}` — Ödeme sağlayıcısından gelen imzalı bildirimi kabul eder.

## Platform operasyonu (38)

Sağlık, iz, alarm, veritabanı durumu ve yedekten geri dönüş içindir.

- `GET /api/platform/operations/health` — Bağımlılıkların sağlık özetini döner.
- `GET /api/platform/operations/traces` — İz kayıtlarını listeler.
- `POST /api/platform/operations/traces/search` — İz kayıtlarında arama yapar.
- `GET /api/platform/operations/traces/stats` — İz istatistiklerini döner.
- `GET /api/platform/operations/traces/{correlation_id}` — Tek iz kaydını döner.
- `GET /api/platform/operations/traces/{correlation_id}/timeline` — İzin zaman çizelgesini döner.
- `GET /api/platform/operations/traces/spans/{span_id}` — Tek span kaydını döner.
- `GET /api/platform/operations/alert-rules` — Alarm kurallarını listeler.
- `POST /api/platform/operations/alert-rules` — Alarm kuralı oluşturur.
- `GET /api/platform/operations/alert-rules/{alert_rule_id}` — Tek alarm kuralını döner.
- `PUT /api/platform/operations/alert-rules/{alert_rule_id}` — Alarm kuralını günceller.
- `DELETE /api/platform/operations/alert-rules/{alert_rule_id}` — Alarm kuralını siler.
- `POST /api/platform/operations/alert-rules/{alert_rule_id}/test` — Alarm kuralını deneme ateşler.
- `GET /api/platform/operations/alert-events` — Alarm olaylarını listeler.
- `GET /api/platform/operations/alert-events/{alert_event_id}` — Tek alarm olayını döner.
- `POST /api/platform/operations/permission-cache/flush` — Yetki karar önbelleğini boşaltır.
- `GET /api/platform/database/stats` — Veritabanı istatistiklerini döner.
- `GET /api/platform/database/schema` — Veritabanı şema özetini döner.
- `GET /api/platform/database/diagnostics/recovery-readiness` — Veritabanının geri yüklemeye hazır olup olmadığına bakar.
- `GET /api/platform/recovery-protection` — Kurtarma korumasının açık olup olmadığını döner.
- `POST /api/platform/recovery-protection/enable` — Kurtarma korumasını açar.
- `POST /api/platform/recovery-protection/rotate` — Kurtarma korumasının anahtarını döndürür.
- `POST /api/platform/recovery-protection/disable` — Kurtarma korumasını kapatır.
- `GET /api/platform/recovery-protection/anchors` — Kurtarma çapalarını listeler.
- `POST /api/platform/recovery-protection/anchors/registration-options` — Yeni çapa kaydı için seçenekleri üretir.
- `POST /api/platform/recovery-protection/anchors` — Yeni kurtarma çapası kaydeder.
- `DELETE /api/platform/recovery-protection/anchors/{anchor_id}` — Kurtarma çapasını kaldırır.
- `GET /api/platform/recovery-sets` — Kurtarma setlerini listeler.
- `POST /api/platform/recovery-sets` — Yeni kurtarma seti oluşturur.
- `GET /api/platform/recovery-sets/{recovery_set_id}` — Tek kurtarma setini döner.
- `DELETE /api/platform/recovery-sets/{recovery_set_id}` — Kurtarma setini siler.
- `POST /api/platform/recovery-sets/{recovery_set_id}/manifest-download` — Kurtarma setinin manifestosunu indirme işi başlatır.
- `POST /api/platform/recovery-sets/{recovery_set_id}/restore` — Bu setten geri yükleme ister.
- `GET /api/platform/recovery-operations/{operation_id}` — Kurtarma kiti işinin durumunu döner.
- `POST /api/platform/recovery-operations/{operation_id}/kit-download` — Kurtarma kitini indirme hakkı ister.
- `POST /api/platform/recovery-operations/{operation_id}/kit-ack` — Kurtarma kitinin alındığını onaylar.
- `GET /api/platform/restore-operations/{operation_id}` — Geri yükleme işinin durumunu döner.
- `POST /api/platform/restore-operations/{operation_id}/plan-download` — Geri yükleme planını indirir.
