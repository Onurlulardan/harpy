package menu

import (
	"strings"
	"testing"
)

func TestMenuConfigSQLDoesNotRequireDeleteGrant(t *testing.T) {
	statements := []string{
		getActiveConfigSQL, deactivateActiveConfigsSQL, deactivateOtherConfigsSQL,
		insertConfigSQL, updateConfigSQL,
	}
	for _, statement := range statements {
		if strings.Contains(strings.ToUpper(statement), "DELETE") {
			t.Fatal("designer menu config SQL must not DELETE")
		}
	}
	if !strings.Contains(insertConfigSQL, "INSERT INTO catalog_menu_configs") {
		t.Fatal("create must insert catalog_menu_configs")
	}
	if !strings.Contains(updateConfigSQL, "UPDATE catalog_menu_configs") {
		t.Fatal("update must update catalog_menu_configs")
	}
}
