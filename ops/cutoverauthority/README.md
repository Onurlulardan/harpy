# External cutover authority workflow

Bu executable core runtime, core config ve release image'inin parcasi degildir.
Yalniz isolated operator job'unda calisir. Cluster root credential'i sadece
authority catalog ownership adoption'inda; leased `harpy_cluster_admin`
credential'i finalizer lease issuance/authorization'inda kullanilir.
Tam rollout completion receipt'i de ayni isolated operator boundary'sinde,
protected cluster-admin routine'i uzerinden kaydedilir.

## Build ve invocation

```powershell
go build -trimpath -o cutover-authority.exe ./ops/cutoverauthority
./cutover-authority.exe -config C:\secure\cutover-authority.json
```

Tek CLI girdisi absolute config path'idir. DSN, password, artifact veya operation
argv/environment ile verilmez. Config kontrati
`cutover-authority-config.schema.json` dosyasidir. Config yalniz DSN secret-file
path'i tasir; raw secret tasimaz. `HARPY_*` key'i bu executable'a ait degildir.
Unix'te DSN ve signing-key secret file mode'u exact `0400` veya `0600` olmalidir.
Windows'ta directory/file ACL izolasyonu operator precondition'idir.
DSN secret'i yalniz explicit `postgres://`/`postgresql://` URL formunda kabul
edilir; user, tek TCP host ve database URL'de bulunur, finalizer password'u da
userinfo icinde explicit olur. Service/passfile ve identity'yi query ile
degistiren key'ler reddedilir. `PGHOST`, `PGPASSWORD`, `PGPASSFILE`, `PGSERVICE`
dahil pgx'in okudugu herhangi bir ambient `PG*` config degeri varsa tool
fail-closed durur; implicit `.pgpass` credential'i kullanilmaz.

## Adoption sirasi

1. External provisioner `harpy_cutover_authority_owner` rolunu signed
   contract'taki exact NOLOGIN profile ile yaratir.
2. Normal migrator paired migration stream'ini uygular; protected catalog bu
   anda `harpy_owner` tarafindan owned olur.
3. Operator `action=adopt_authority` config'ini calistirir. Tool release,
   bootstrap authority ve adoption operation signature'larini DB'ye baglanmadan
   once dogrular.
4. Tool external-root session/system/database/PG18.4 preflight'ini yapar,
   migrator'u signed policy ile quiesce eder ve active session'lari sifirlar.
5. Tek SERIALIZABLE transaction; advisory/catalog lock'larini alir, exact pre
   definition+ACL'leri dogrular, protected object ownership'ini transfer eder,
   dependency setinin degismedigini kanitlar ve migrator'u restore eder.
6. Receipt byte'lari ve Ed25519 signature locked post-state'ten transaction
   icinde uretilir. Canonical receipt protected inventory'ye yazilir; success
   yalniz commit tamamlaninca uretilir.
7. Commit sonrasi receipt ve detached signature atomic no-replace publish olur.

## Finalizer issuance ve authorization sirasi

1. External credential provisioner `harpy_cluster_admin` ve
   `harpy_cutover_finalizer` rollerini exact finite lease profile'ina getirir
   ve birbirinden ayri DSN secret file'lari uretir. Tool plaintext password ile
   `ALTER ROLE` calistirmaz.
2. `action=issue_finalizer`; signed adoption/seal/cleanup/authority artifact'lari,
   canonical terminal evidence, authority signing key ve iki DSN path'i alir.
   Finalizer DSN ile gercek login preflight'i transaction'dan once yapilir.
3. Cluster-admin SERIALIZABLE transaction'inda protected
   `record_credential_receipts_v1` verified canonical seal+cleanup byte'larini
   ve digest'lerini exact-CAS ile kaydeder. Ardindan protected observer bu
   receipt'lere bagli live state digest, procedure OID/signature, database
   identity, eligibility, tenant authority ve finalizer expiry'yi dondurur.
   Tool exact 120-second lease'i Ed25519 ile imzalar; commit sonrasi atomic
   no-replace publish eder.
4. `action=authorize_finalization` uretilen finalizer lease/signature ile ayni
   verified prerequisites'i alir. Root/finalizer DSN bu profile kabul edilmez.
5. Exact cluster-admin session preflight'inden sonra protected
   `cutover_authority.authorize_finalization_v1` 16-parametre ABI'si tek
   SERIALIZABLE transaction'da cagrilir.
6. Core migrator ayni verified ABI16 envelope ile protected finalize'i yapar ve
   `verify_finalization_receipt_v1` DB inventory receipt'ini dogrular. Bundan
   sonra `harpy_cutover_finalizer` login lease'ini kapatmak external
   credential provisioner boundary'sidir; bu tool role password/login mutate
   etmez ve ayri bir receipt file publication ABI'si uretmez.

## Tam rollout completion kaydi

1. `action=record_full_rollout` profili signed authority zincirini tarihsel
   issuance aninda dogrular ve completion `environment_id` degerini bu zincire
   baglar.
2. `completion_evidence`, `full-rollout-completion-evidence.v1` canonical JSON
   kontratidir. Steps, cohort soak metrikleri, sifir V1 yuzeyi, 30 gun sessizlik
   ve rollback-window closure bilgisi core completion evaluator'dan gecer.
3. `evidence_artifacts` listedeki her non-zero acceptance/snapshot/contract
   digest'ini gercek immutable artifact byte'larinin SHA-256 degeriyle kapatir;
   eksik veya ilgisiz artifact fail-closed reddedilir.
4. Exact cluster-admin PG18.4/database/finite-lease preflight'inden sonra kayit
   tek SERIALIZABLE transaction'da protected
   `record_full_rollout_completion_v1` routine'iyle yazilir. stdout yalniz
   receipt ID ve receipt digest'i iceren redacted result kontratidir.

## Crash, replay ve recovery

- Quiescence oncesi crash DB state'i degistirmez.
- Quiescence commit'i sonrasi crash migrator'u gecici NOLOGIN birakabilir. Ayni
  signed operation expiry penceresinde ayni config yeniden calistirilir; tool
  exact temporary profile'dan devam eder ve adoption commit'inde restore eder.
- Adoption transaction'i rollback olursa catchable failure'da tool pre-profile'i
  compensating restore eder ve exact state'i yeniden dogrular. Connection loss
  nedeniyle restore kanitlanamazsa stable `authority adoption recovery required`
  sonucu doner; operator recovery gerekliligi normal hataya gizlenmez.
- DB commit'i sonrasi, file publish oncesi crash'te replay protected inventory'de
  saklanan exact canonical byte'lari okur, live catalog assert'ini calistirir ve
  Ed25519 deterministic signature'i yeniden uretir. Existing ayni file kabul,
  farkli file hard failure'dir.
- Finalizer issuance commit'i sonrasi publish yarim kalirsa ayni verified
  receipt byte'lari protected exact-CAS'ten gecer, ayni lease/signature byte'lari
  yeniden uretilir ve eksik artifact no-replace publish edilir. Ayni operation/run
  icin farkli seal veya cleanup byte'lari protected routine tarafindan reddedilir.
- Finalization authorization ayni artifact byte'lariyla idempotenttir. Ayni
  operation/run/lease key'inde farkli document veya digest protected routine
  tarafindan tamper olarak reddedilir.

Her dort action icin stdout yalniz
`external-cutover-authority-result.v1` JSON'idir. DSN, password, signing seed,
canonical credential artifact'i veya PostgreSQL error text'i stdout/stderr'e
yazilmaz.
