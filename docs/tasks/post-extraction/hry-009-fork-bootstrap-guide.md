# HRY-009 — Yeni proje için fork/bootstrap rehberi hazırla

**Öncelik:** P2  
**Durum:** Planlanacak.

## Amaç
Harpy'yi fork eden geliştiricinin ilk günden itibaren hangi alanları değiştireceğini, hangi alanlara dokunmayacağını ve yeni business modülünü nasıl ekleyeceğini bilmesi.

## Yapılacaklar
- Go module path ve import path değişikliği stratejisini belirle.
- Projeye özel modüllerin `core/internal/modules/<project-module>` altında nasıl konumlanacağını açıkla.
- Yeni modülün domain/application/adapter/API sınırlarını ve cross-module iletişim kurallarını göster.
- Modülün migration owner, storage owner, HTTP contract, authority/permission, event/job ve runtime composition kayıtlarını nasıl ekleyeceğini belge.
- Projeye özel env/config, secret, deployment profile ve local development ayarlarını açıkla.
- Harpy generic altyapısında kalması gereken kodla fork'a ait business kodunun sınırını tarif et.
- Yeni proje başlatma rehberini temiz bir fork üzerinde uygulayıp doğrula.

## Kabul kriterleri
- Geliştirici temiz fork üzerinden Go module path'i ve proje kimliğini ayarlayabiliyor.
- Bir örnek project-owned module ekleyip API/DB/tenant/worker akışına bağlayabiliyor.
- Yeni modül Harpy'nin generic modüllerinin private paketlerine doğrudan bağlanmıyor; belirlenen API/ownership kurallarına uyuyor.
- README veya docs içinde sıfırdan çalışan bir fork/bootstrap walkthrough bulunuyor.

**Mimari karar:** Bu rehber dinamik plugin loader veya runtime marketplace gerektirmiyor. Proje modülleri fork içinde açıkça tanımlanıp composition root üzerinden bağlanabilir.
