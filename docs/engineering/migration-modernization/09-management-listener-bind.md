# Task 09 — Management listener bind adresini ve erişim sınırını doğrula

## Bulgular

İlk incelemede cmd/setup/harpy_env.go management listener'ı 0.0.0.0:9090 olarak ayarlıyor; public API adresi ise 127.0.0.1. core/internal/platform/management/dispatch.go içinde /livez, /readyz ve /version handler'larının authentication olmadan servis edildiği görüldü. Deployment modeline göre management portu beklenmedik network arayüzlerine açılabilir.

Önce deploy topolojisini, reverse proxy/health-check gereksinimlerini ve firewall varsayımlarını doğrula. Listener'ı körlemesine localhost'a bağlamak container health-check'lerini bozabilir.

## Yapılacaklar

1. Listener'ın hangi proses/network namespace'te çalıştığını ve container port publishing/proxy kurallarını incele.
2. Endpoint'leri public, internal, unauthenticated ve sensitive metadata içerme açısından sınıflandır.
3. Local single-machine setup için loopback default ve container deployment için açık override gereksinimini değerlendir.
4. Bind address config/env üzerinden ayarlanacaksa güvenli default ve migration docs'u belirle.
5. Resolved bind address, endpoint erişimi ve deployment health-check davranışı için test ekle.
6. Bunu migration bundle refactor'ından ayrı küçük değişiklik olarak yürüt.

## Kabul kriterleri

- Erişim gereksinimleri deployment modeline göre belgeli.
- Management endpoint güvenlik varsayımları açık.
- Güvenli varsayılan seçilmiş veya mevcut ayar kanıtla gerekçelendirilmiş.
- Health checks ve container deployment bozulmamış.
