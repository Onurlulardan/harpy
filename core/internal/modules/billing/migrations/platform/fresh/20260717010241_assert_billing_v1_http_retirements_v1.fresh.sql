-- +goose Up
-- Completes the inventory; API Admit checks the exact 26 keys before traffic.
SELECT platform.contract_retirement_provision_v1('rq.alias.220f28df60a85ab64eca5a31e401719b07aadde91c40156944852595acabfe14', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.91bcd921f0c9693bb931dcd2131f9c096863b109de70b6566e83cf8bb18077de', 'billing_v1_http', 'billing', 3::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.5f713d907f7a0c0b63bba718820050bcf3497492a996581e4c8c4023c37ca406', 'billing_v1_http', 'billing', 3::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.f944e8f0f3b8e6ac0b86f82a2bdd157a76e0c998f313a7a7f1317e1e431c7afc', 'billing_v1_http', 'billing', 3::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.5a99293a048a151b8f76c28f20c799c0d853de41ce4098ed4e400e5b8538b85e', 'billing_v1_http', 'billing', 3::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.869d32a5cff513bb1644f9d9fcc31613a438b86578880a37209a1754d7e63d8f', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.707e4d28c5a1995bbb546c94d35bfae6a7878d87d6d4b239e71c84fe536d824e', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.9abdecb48a6a312c12eaa994e92d6b52c15bae3e50dd75ba882b3daa14e3108b', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.0ede27b5e0b01e0f3571bb5f37eb1476d49d0217a18fc3367cfd4c09ea9932dc', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.92478dcfe8e191ddabd6e13543a2a19aeb1c1d271a7e04e103b319d30d8a6c35', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.aeeadbc48be1fe92202d2d5ad7b902c9358b60d5cc8eae032319cc186d5187e7', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.7bb8ed7d09d7b09f3aa25a3f4ba0aeed24403f772a2bde3556d7aa55dd5d2cbf', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.0822cbd7bcc64a6393d6888a1706bcceb344bd8aaaf57ef1d7b8dc98854d3af1', 'billing_v1_http', 'billing', 4::bigint);
