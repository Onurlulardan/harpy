//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertBillingV1RetirementInventory(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	rows, err := database.QueryContext(ctx, `
SELECT contract_key, remaining_scope_count
FROM platform.contract_retirements
WHERE owner_module = 'billing_v1_http'
  AND domain = 'billing'
  AND state = 'active'
ORDER BY contract_key COLLATE "C"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := make(map[string]int64, 26)
	for rows.Next() {
		var key string
		var remaining int64
		if err := rows.Scan(&key, &remaining); err != nil {
			t.Fatal(err)
		}
		if remaining <= 0 {
			t.Fatalf("retirement %s remaining scope = %d", key, remaining)
		}
		seen[key] = remaining
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 26 {
		t.Fatalf("Billing V1 retirement count = %d, want 26", len(seen))
	}
	for key, remaining := range billingV1RetirementEvidence {
		if seen[key] != remaining {
			t.Fatalf("retirement %s remaining scope = %d, want %d",
				key, seen[key], remaining)
		}
	}
}

var billingV1RetirementEvidence = map[string]int64{
	"rq.alias.0822cbd7bcc64a6393d6888a1706bcceb344bd8aaaf57ef1d7b8dc98854d3af1": 4,
	"rq.alias.0ede27b5e0b01e0f3571bb5f37eb1476d49d0217a18fc3367cfd4c09ea9932dc": 4,
	"rq.alias.15f0d1a3d6dedbf0ee474c79c3e0646e3758b9d9a4b488fba0184ba6509f8eac": 4,
	"rq.alias.220f28df60a85ab64eca5a31e401719b07aadde91c40156944852595acabfe14": 4,
	"rq.alias.2526b296f4cbd48b6fe8968b90a3a943e8eba2ee392f81c2465bea77e7960a81": 4,
	"rq.alias.3aefe76d337363ab07313ade68d7b45c9ff4b47c9c7f3cacd5560e4592c52313": 4,
	"rq.alias.5a99293a048a151b8f76c28f20c799c0d853de41ce4098ed4e400e5b8538b85e": 3,
	"rq.alias.5f713d907f7a0c0b63bba718820050bcf3497492a996581e4c8c4023c37ca406": 3,
	"rq.alias.6ba785a4a37eec2eb65e6fd5c14a1155a428fc15d598a254304cee626ff63c79": 4,
	"rq.alias.707e4d28c5a1995bbb546c94d35bfae6a7878d87d6d4b239e71c84fe536d824e": 4,
	"rq.alias.7bb8ed7d09d7b09f3aa25a3f4ba0aeed24403f772a2bde3556d7aa55dd5d2cbf": 4,
	"rq.alias.7cb20259cf288ea30b4e29fb86084f9be0ff2db44b3ef2a61def7bd8808f8ed5": 2,
	"rq.alias.869d32a5cff513bb1644f9d9fcc31613a438b86578880a37209a1754d7e63d8f": 4,
	"rq.alias.8a197db1f851b65f7beecce6e4bb5151b697827582001a456df2a3f97340f5c1": 5,
	"rq.alias.91bcd921f0c9693bb931dcd2131f9c096863b109de70b6566e83cf8bb18077de": 3,
	"rq.alias.92478dcfe8e191ddabd6e13543a2a19aeb1c1d271a7e04e103b319d30d8a6c35": 4,
	"rq.alias.93e5400786a02340bd04c387af3aa518a7fb31973528262804b09ce4d7bc287f": 4,
	"rq.alias.9abdecb48a6a312c12eaa994e92d6b52c15bae3e50dd75ba882b3daa14e3108b": 4,
	"rq.alias.a67418ca56b2f99da5e9ce97ece2505b3020f083233bf5fc24a57a661f83a42e": 4,
	"rq.alias.aeeadbc48be1fe92202d2d5ad7b902c9358b60d5cc8eae032319cc186d5187e7": 4,
	"rq.alias.af9bc8b0b54c8c03defc4201d6a87c20f157f24f16ec0d2e908f1f7396cf74cc": 5,
	"rq.alias.c217f19b93c4d379b1d6d0fa9471df089a4362c42f94ab4cf5afa63f929d655e": 5,
	"rq.alias.c553539cbb76bef1ca61ac33deec5b44b81c336f184912f2f28e348b6aa1ea5f": 2,
	"rq.alias.d35c61f4158f271124d5928b0717576725490fcee158a36b8e77dfb1c6793836": 4,
	"rq.alias.d7e6a7ede9af7b310fa8b86e3db31e48f2bb145de0c5af44ba7681dc1a81154f": 4,
	"rq.alias.f944e8f0f3b8e6ac0b86f82a2bdd157a76e0c998f313a7a7f1317e1e431c7afc": 3,
}
