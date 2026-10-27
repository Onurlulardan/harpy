# Kalan işler

Bu dosya yarın devam etmek için nottur. Yeni kural değildir.

## Neden yeniden yüklendi

GitHub Contributors listesinde Onurlulardan ile birlikte `cursoragent` görünüyordu. Yerel geçmişte yazar veya committer adı Cursor Agent olan commit yoktu. `b5626ba` sonrasındaki iki commit `Co-authored-by: Cursor <cursoragent@cursor.com>` taşıyordu:

- `c21cf74d` Add the Harpy Go starter to the monorepo branch.
- `45c4b252` Add the remaining Harpy starter sources.

Dal `b5626ba75a1161c0584b725ec1452699b24fa839` noktasına alındı. Bu, 13 Ekim 2025 tarihli "Update docker-compose.yml" kaydıdır. Yazarı Onur Altuntaş, e-posta `50622451+Onurlulardan@users.noreply.github.com`. Üzerine `C:\repos\harpy\harpy-yedek\harpy` ağacı aynı ad ve e-posta ile commitlendi.

## Güvenlik kopyası

`C:\repos\harpy\harpy-2026-10-27-calisma`

Sıfırlamadan önceki çalışma ağacıdır. `.git` kopyalanmadı. Push edilmez.

## Ölçülen ağaçlar

Güvenlik kopyası ile yedek aynı sayıları verdi. Yedekte fresh SQL işinin olmadığı ölçülmedi; ikisi de içeriyor.

- `*.fresh.sql`: güvenlik kopyası 2091, yedek 2091.
- `HARPY_` geçen fresh SQL: her ikisinde 1122.
- `LoadStartupRegistry` her iki ağaçta `SelectFreshDocuments` çağırıyor. `core/internal/composition/migrator/fresh_startup_test.go` seçilen yolu 2071 bekliyor. Bu test her iki ağaçta da var.
- `contracts/runtime/v1/binary-schema-contract.json` her iki ağaçta var, 412 bayt ve içerik aynı.
- `*.v1_adopted.sql`: her ikisinde 0.

## Yarın

Yüklenen ağaç yedektir. Aşağıdakiler bu ağaç için duruyor.

- `deploy up` için gerçek Postgres DSN (`HARPY_POSTGRES_DSN`), imzalı release artifact, operator license ve `HARPY_DOCUMENTS_KEK_KEYRING_FILE` hâlâ kodda duruyor. Bu oturumda bunlar verilmeden login denenmedi.
- SQL bir veritabanına uygulanmadı. Login denenmedi.
- 1122 fresh SQL dosyası `HARPY_` adı taşıyor. Bu sayı yüklenen ağaçta da var; yalnız güvenlik kopyasına ait değil.
- `*.v1_adopted.sql` yok (yedekte de 0). Cutover testleri ve schema executor testleri `v1_adopted` varyantını anıyor. `tenant_executor*.go` dosyalarında `v1_adopted` metni yok.
- `go test ./...` çalıştırılmadı.
