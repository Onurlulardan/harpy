# Task 03 — Management listener'ın network exposure sınırını doğrula

## Başlangıç bulgusu

Önceki incelemede `cmd/setup/harpy_env.go` management listener için `0.0.0.0:9090`, public API için ise `127.0.0.1` ayarlıyordu. `core/internal/platform/management/dispatch.go` içinde `/livez`, `/readyz`, `/version` endpoint'leri authentication olmadan servis ediliyor görünüyordu. Bu, tek başına internetten erişilebilir olduklarını kanıtlamaz; gerçek erişim container publishing, host firewall, network namespace ve reverse proxy'ye bağlıdır.

## Yapılacaklar

1. Güncel main'de listener bind konfigürasyonunu, server startup'ını, container port publishing'i, reverse proxy ve deployment docs'u incele.
2. Her management endpoint'i için şu alanları belirle: auth var mı, hangi bilgi dönüyor, readiness/liveness için kim erişmeli, dışarıdan erişim gerekli mi?
3. Desteklenen deployment modellerini listele: local single-machine, Docker/container network, reverse proxy/remote health checker.
4. Güvenli varsayılanı belirle. Local setup'ta loopback uygun olabilir; container içinde loopback'e bağlanmak host/peer health-check erişimini bozabilir. Bu yüzden tek bir adresi varsayılan olarak uygulamadan önce gerçek topolojiyi doğrula.
5. Gerekirse açık config/env override ekle ve tüm default/override davranışını test et.
6. Management endpoint'leri sensitive data açığa çıkarıyorsa sadece bind address'e güvenmek yerine uygun auth/network policy düşün.
7. Testlerde resolved address ve gerçek erişim sınırını doğrula. Portu host'a publish etmeden önce/sonra container senaryosunu test et.

## Kabul kriterleri

- Hangi ağ arayüzünden hangi aktörün erişmesi gerektiği açıkça belgelenmiş.
- Güvenli default ve container deployment davranışı test edilmiş.
- Health checks bozulmamış.
- Endpoint'lerin authentication ve veri sızdırma durumu açık.
- Risk doğrulanmadıysa bulgu “kanıtlanmadı” diye kapatılır; varsayımla değişiklik yapılmaz.
