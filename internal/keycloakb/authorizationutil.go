package keycloakb

import (
	"errors"

	api "github.com/cloudtrust/keycloak-bridge/api/management"
	"github.com/cloudtrust/keycloak-bridge/internal/dto"
	kc "github.com/cloudtrust/keycloak-client"
)

// TranslateGroupIDIntoGroupName convert the groupID of the provided Authorization array in GroupName
func TranslateGroupIDIntoGroupName(origin []dto.Authorization, groups []kc.GroupRepresentation) []dto.Authorization {
	// Build a mapping groupID -> groupName
	var groupIDMapping = make(map[string]string)
	groupIDMapping["*"] = "*"

	for _, group := range groups {
		groupIDMapping[*group.Id] = *group.Name
	}

	// Translate targetGroupId to targetGroupName
	var translatedAuthorizations = []dto.Authorization{}

	for _, auth := range origin {
		var targetGroup *string

		if auth.TargetGroupID != nil {
			v, ok := groupIDMapping[*auth.TargetGroupID]

			// Invalid values are ignored
			if !ok {
				continue
			}

			var groupName = string(v)
			targetGroup = &groupName
		}

		translatedAuthorizations = append(translatedAuthorizations, dto.Authorization{
			RealmID:       auth.RealmID,
			GroupID:       auth.GroupID,
			Action:        auth.Action,
			TargetRealmID: auth.TargetRealmID,
			TargetGroupID: targetGroup,
		})
	}

	return translatedAuthorizations
}

// TranslateGroupNameIntoGroupID convert the groupNAme of the provided Authorization array in GroupID
func TranslateGroupNameIntoGroupID(origin []dto.Authorization, mapper map[string]map[string]string) []dto.Authorization {
	// Convert groupName into groupID
	var convertedAuthorizations = []dto.Authorization{}
	for _, authz := range origin {
		var targetGroupIDPtr *string

		if authz.TargetRealmID != nil && authz.TargetGroupID != nil {
			targetGroupID, ok := mapper[*authz.TargetRealmID][*authz.TargetGroupID]

			// Invalid values are ignored
			if !ok {
				continue
			}

			copyGroupID := string(targetGroupID)
			targetGroupIDPtr = &copyGroupID
		}

		convertedAuthorizations = append(convertedAuthorizations, dto.Authorization{
			RealmID:       authz.RealmID,
			GroupID:       authz.GroupID,
			Action:        authz.Action,
			TargetRealmID: authz.TargetRealmID,
			TargetGroupID: targetGroupIDPtr,
		})
	}

	return convertedAuthorizations
}

// Validate the content of the provided array. Returns an error if any issue is detected
func Validate(authorizations []dto.Authorization, allowedTargetRealmsAndGroupIDs map[string]map[string]string) error {
	for _, auth := range authorizations {
		// Check TargetRealm
		if auth.TargetRealmID != nil {
			_, ok := allowedTargetRealmsAndGroupIDs[*auth.TargetRealmID]

			if !ok {
				return errors.New("Invalid target realm")
			}
		}

		// Check TargetGroupID
		if auth.TargetGroupID != nil {
			_, ok := allowedTargetRealmsAndGroupIDs[*auth.TargetRealmID][*auth.TargetGroupID]

			if !ok {
				return errors.New("Invalid target group")
			}
		}
	}

	var authZ = api.ConvertToAPIAuthorizations(authorizations)

	var checker = *authZ.Matrix
	for _, u := range checker {
		for realmID, v := range u {
			// Check if * as targetRealm, there is no other targetRealm rule
			if realmID == "*" && len(u) != 1 {
				return errors.New("If '*' is used as targetRealm, no other rules for this action are allowed")
			}

			// Check if * as targetGroupId, there is no other targetGroupID rule
			for targetGroup := range v {
				if targetGroup == "*" && len(v) != 1 {
					return errors.New("If '*' is used as targetGroupId, no other rules are allowed")
				}
			}
		}
	}

	return nil
}
