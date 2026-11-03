# HRY-006 — Backup/restore için gerçek uçtan uca recovery provası yap

**Öncelik:** P1  
**Durum:** Planlanacak.

## Amaç
Backup alınabildiğini değil, başka bir temiz ortamda geri yüklenip kullanılabildiğini kanıtlamak.

## Yapılacaklar
- PostgreSQL platform ve tenant verilerini kapsayan backup akışını doğrula.
- Documents için kullanılan object storage verilerini, metadata'yı ve gerekli encryption/KMS key referanslarını birlikte ele al.
- Secret ve key materyalinin backup dışında nasıl korunduğunu; restore sırasında nasıl sağlandığını belgele.
- Backup artifact ve restore adımlarının bütünlük/digest kontrollerini doğrula.
- İzole yeni ortamda restore yap; IAM login, tenant erişimi, belge indirme/decryption, Work kayıtları ve worker devamlılığını kontrol et.
- Restore sonrası permission, tenant isolation ve audit davranışını doğrula.

## Kabul kriterleri
- Restore işlemi orijinal ortamdan bağımsız temiz bir hedefte başarılı.
- Restore edilen belgeler okunabiliyor ve digest/integrity kontrolleri geçiyor.
- Platform ve tenant verileri tutarlı.
- Key/secret eksikliği veya yanlış konfigürasyon sessiz veri kaybına yol açmıyor; işlem açık hata veriyor.
- Restore sırası, gerekli secret'lar, operatör adımları ve bilinen kısıtlar belgelenmiş.

## İlgili doküman
[Deployment ve operasyon](https://github.com/Onurlulardan/harpy/blob/main/docs/architecture/09-deployment-ve-operasyon.md)
