-- +goose Up
-- The cutover-owned routine is idempotent and rejects identity or count drift.
SELECT platform.contract_retirement_provision_v1('rq.alias.af9bc8b0b54c8c03defc4201d6a87c20f157f24f16ec0d2e908f1f7396cf74cc', 'billing_v1_http', 'billing', 5::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.93e5400786a02340bd04c387af3aa518a7fb31973528262804b09ce4d7bc287f', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.8a197db1f851b65f7beecce6e4bb5151b697827582001a456df2a3f97340f5c1', 'billing_v1_http', 'billing', 5::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.15f0d1a3d6dedbf0ee474c79c3e0646e3758b9d9a4b488fba0184ba6509f8eac', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.d7e6a7ede9af7b310fa8b86e3db31e48f2bb145de0c5af44ba7681dc1a81154f', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.2526b296f4cbd48b6fe8968b90a3a943e8eba2ee392f81c2465bea77e7960a81', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.d35c61f4158f271124d5928b0717576725490fcee158a36b8e77dfb1c6793836', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.3aefe76d337363ab07313ade68d7b45c9ff4b47c9c7f3cacd5560e4592c52313', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.c553539cbb76bef1ca61ac33deec5b44b81c336f184912f2f28e348b6aa1ea5f', 'billing_v1_http', 'billing', 2::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.c217f19b93c4d379b1d6d0fa9471df089a4362c42f94ab4cf5afa63f929d655e', 'billing_v1_http', 'billing', 5::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.6ba785a4a37eec2eb65e6fd5c14a1155a428fc15d598a254304cee626ff63c79', 'billing_v1_http', 'billing', 4::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.7cb20259cf288ea30b4e29fb86084f9be0ff2db44b3ef2a61def7bd8808f8ed5', 'billing_v1_http', 'billing', 2::bigint);
SELECT platform.contract_retirement_provision_v1('rq.alias.a67418ca56b2f99da5e9ce97ece2505b3020f083233bf5fc24a57a661f83a42e', 'billing_v1_http', 'billing', 4::bigint);
