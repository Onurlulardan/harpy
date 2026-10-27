package application

import layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"

func defaultPresentationField(
	resource layoutapi.ResourceKind,
	id string,
	order int32,
) layoutapi.Field {
	en, tr := defaultFieldLabels[id][0], defaultFieldLabels[id][1]
	if en == "" {
		en, tr = "Display name", "Görünen ad"
	}
	return layoutapi.Field{
		FieldID: id, Interaction: layoutapi.InteractionAuto, Order: order,
		Widget: layoutapi.WidgetAuto,
		Translations: []layoutapi.Translation{
			{Locale: "en", Label: en}, {Locale: "tr", Label: tr},
		},
	}
}

func defaultSurfaceTranslations(profile surfaceProfile) []layoutapi.Translation {
	tr := map[layoutapi.SurfaceKind]string{
		layoutapi.SurfaceRecordCreate: "Oluştur",
		layoutapi.SurfaceRecordUpdate: "Güncelle",
		layoutapi.SurfaceRecordDetail: "Ayrıntılar",
		layoutapi.SurfaceRecordList:   "Liste",
		layoutapi.SurfaceRecordFilter: "Filtreler",
	}[profile.kind]
	return []layoutapi.Translation{
		{Locale: "en", Label: profile.label}, {Locale: "tr", Label: tr},
	}
}

func defaultSectionTranslations() []layoutapi.Translation {
	return []layoutapi.Translation{
		{Locale: "en", Label: "Fields"}, {Locale: "tr", Label: "Alanlar"},
	}
}

var defaultFieldLabels = map[string][2]string{
	userEmail:            {"Email", "E-posta"},
	userPortalIdentifier: {"Portal identifier", "Portal kimliği"},
	userStatus:           {"Status", "Durum"},
	userIdentityKind:     {"Identity kind", "Kimlik türü"},
	userRoleIDs:          {"Roles", "Roller"},
	userOrganizationIDs:  {"Organizations", "Organizasyonlar"},
	userCreatedAt:        {"Created at", "Oluşturulma zamanı"},
	userUpdatedAt:        {"Updated at", "Güncellenme zamanı"},
	roleScopeKind:        {"Scope", "Kapsam"},
	roleOrganizationID:   {"Organization", "Organizasyon"},
	roleName:             {"Name", "Ad"},
	roleDescription:      {"Description", "Açıklama"},
	roleStatus:           {"Status", "Durum"},
	roleProtected:        {"Protected", "Korumalı"},
	roleDefault:          {"Default role", "Varsayılan rol"},
	roleCreatedAt:        {"Created at", "Oluşturulma zamanı"},
	roleUpdatedAt:        {"Updated at", "Güncellenme zamanı"},
	organizationName:     {"Name", "Ad"},
	organizationSlug:     {"Slug", "Kısa ad"},
	organizationStatus:   {"Status", "Durum"},
	organizationParentID: {"Parent organization", "Üst organizasyon"},
	organizationOwnerIDs: {"Owners", "Sahipler"},
	organizationCreated:  {"Created at", "Oluşturulma zamanı"},
	organizationUpdated:  {"Updated at", "Güncellenme zamanı"},
}
