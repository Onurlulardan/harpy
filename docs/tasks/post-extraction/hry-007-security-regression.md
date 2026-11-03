# HRY-007 — Extraction sonrası security regression denetimi yap

**Öncelik:** P1  
**Durum:** Planlanacak.

## Amaç
Genel bir security suite'i sıfırdan yazmak değil; Catalog removal ve baseline sadeleştirmesi sonrasında korunması gereken güvenlik sınırlarını hedefli olarak doğrulamak.

## Yapılacaklar
- IAM authentication, RBAC, permission request ve retired permission davranışlarını kontrol et.
- Tenant identity'nin request header/query param üzerinden güvenilir kabul edilmediğini doğrula.
- `TenantScope`, DB role, schema `search_path` ve connection/session cleanup sınırlarını kontrol et.
- Cross-tenant ID substitution, yetkisiz belge indirme ve stale permission/cache senaryolarını hedefli test et.
- API, Worker ve Migrator runtime config'lerinde artık gerekmeyen Catalog keyring/secret alanı kalmadığını kontrol et.
- Dependency vulnerability ve secret scanning için CI'a uygun, düşük bakım maliyetli kontroller seç.

## Kabul kriterleri
- Extraction'dan etkilenen IAM/Analytics/tenant sınırlarında negatif testler geçiyor.
- Yetkisiz veya tenant dışı erişim fail-closed davranıyor.
- Secret, credential ve document content log/trace alanlarına sızmıyor.
- Dependency ve secret scan sonuçları raporlanmış; kritik bulgular çözülmüş veya gerekçeli olarak kabul edilmiş.
- Mevcut güvenlik testleri tekrar yazılmıyor; yalnız eksik regresyonlar ekleniyor.

## İlgili dokümanlar
- [Güvenlik modeli](https://github.com/Onurlulardan/harpy/blob/main/docs/architecture/08-guvenlik-modeli.md)
- [Veri ve tenant izolasyonu](https://github.com/Onurlulardan/harpy/blob/main/docs/architecture/06-veri-ve-tenant-izolasyonu.md)
