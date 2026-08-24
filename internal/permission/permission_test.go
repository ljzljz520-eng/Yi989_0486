package permission

import (
	"couponbatch/internal/model"
	"testing"
)

func TestPermissionNodes(t *testing.T) {
	admin := DefaultAdmin("a")
	for _, node := range []string{View, Create, Edit, Delete, Export, Claim} {
		if !Allows(admin, node) {
			t.Fatal(node)
		}
	}
	if Allows(model.Profile{ID: "off", Active: false, Nodes: []string{"*"}}, View) {
		t.Fatal("inactive profile allowed")
	}
}
