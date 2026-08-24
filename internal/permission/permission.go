package permission

import "couponbatch/internal/model"

const (
	View   = "coupon_batch:view"
	Create = "coupon_batch:create"
	Edit   = "coupon_batch:edit"
	Delete = "coupon_batch:delete"
	Export = "coupon_batch:export"
	Claim  = "coupon_batch:claim"
)

func Required(action string) string {
	switch action {
	case "view":
		return View
	case "create":
		return Create
	case "edit":
		return Edit
	case "delete":
		return Delete
	case "export":
		return Export
	case "claim":
		return Claim
	default:
		return ""
	}
}

func Allows(profile model.Profile, node string) bool {
	if !profile.Active {
		return false
	}
	for _, granted := range profile.Nodes {
		if granted == node || granted == "*" {
			return true
		}
	}
	for _, role := range profile.Roles {
		if role == "admin" && node != "" {
			return true
		}
	}
	return false
}

func DefaultAdmin(id string) model.Profile {
	return model.Profile{ID: id, DisplayName: "Coupon administrator", Roles: []string{"admin"}, Nodes: []string{View, Create, Edit, Delete, Export, Claim}, Active: true}
}
