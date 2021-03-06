package apiaccount

import (
	"context"
	"encoding/json"

	"github.com/cloudtrust/common-service/validation"
	"github.com/cloudtrust/keycloak-bridge/internal/constants"
	msg "github.com/cloudtrust/keycloak-bridge/internal/constants"
	"github.com/cloudtrust/keycloak-bridge/internal/keycloakb"
	kc "github.com/cloudtrust/keycloak-client"
)

// AccountRepresentation struct
type AccountRepresentation struct {
	Gender               *string                        `json:"gender,omitempty"`
	FirstName            *string                        `json:"firstName,omitempty"`
	LastName             *string                        `json:"lastName,omitempty"`
	Username             *string                        `json:"username,omitempty"`
	Email                *string                        `json:"email,omitempty"`
	EmailVerified        *bool                          `json:"emailVerified,omitempty"`
	PhoneNumber          *string                        `json:"phoneNumber,omitempty"`
	PhoneNumberVerified  *bool                          `json:"phoneNumberVerified,omitempty"`
	BirthDate            *string                        `json:"birthDate,omitempty"`
	BirthLocation        *string                        `json:"birthLocation,omitempty"`
	Nationality          *string                        `json:"nationality,omitempty"`
	IDDocumentType       *string                        `json:"idDocumentType,omitempty"`
	IDDocumentNumber     *string                        `json:"idDocumentNumber,omitempty"`
	IDDocumentExpiration *string                        `json:"idDocumentExpiration,omitempty"`
	IDDocumentCountry    *string                        `json:"idDocumentCountry,omitempty"`
	Locale               *string                        `json:"locale,omitempty"`
	Accreditations       *[]AccreditationRepresentation `json:"accreditations,omitempty"`
}

// AccreditationRepresentation is a representation of accreditations
type AccreditationRepresentation struct {
	Type       *string `json:"type"`
	ExpiryDate *string `json:"expiryDate"`
	Expired    *bool   `json:"expired,omitempty"`
	Revoked    *bool   `json:"revoked,omitempty"`
}

// CredentialRepresentation struct
type CredentialRepresentation struct {
	ID             *string `json:"id,omitempty"`
	Type           *string `json:"type,omitempty"`
	UserLabel      *string `json:"userLabel,omitempty"`
	CreatedDate    *int64  `json:"createdDate,omitempty"`
	CredentialData *string `json:"credentialData,omitempty"`
	Value          *string `json:"value,omitempty"`
	Temporary      *bool   `json:"temporary,omitempty"`
}

// Configuration struct
type Configuration struct {
	EditingEnabled                    *bool           `json:"editing_enabled"`
	ShowAuthenticatorsTab             *bool           `json:"show_authenticators_tab"`
	ShowPasswordTab                   *bool           `json:"show_password_tab"`
	ShowProfileTab                    *bool           `json:"show_profile_tab"`
	ShowAccountDeletionButton         *bool           `json:"show_account_deletion_button"`
	RedirectSuccessfulRegistrationURL *string         `json:"redirect_successful_registration_url"`
	AvailableChecks                   map[string]bool `json:"available-checks"`
	BarcodeType                       *string         `json:"barcode_type"`
}

// UpdatePasswordBody is the definition of the expected body content of UpdatePassword method
type UpdatePasswordBody struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

// LabelBody struct
type LabelBody struct {
	Label string `json:"label,omitempty"`
}

// ConvertCredential creates an API credential from a KC credential
func ConvertCredential(credKc *kc.CredentialRepresentation) CredentialRepresentation {
	var cred CredentialRepresentation
	cred.ID = credKc.ID
	cred.Type = credKc.Type
	cred.UserLabel = credKc.UserLabel
	cred.CreatedDate = credKc.CreatedDate
	cred.CredentialData = credKc.CredentialData
	cred.Temporary = credKc.Temporary
	// credKc.Value is ignored. It may contains secret for legacy credential thus we don't want to transmit it.
	return cred
}

// ConvertToAPIAccount creates an API account representation from a KC user representation
func ConvertToAPIAccount(ctx context.Context, userKc kc.UserRepresentation, logger keycloakb.Logger) AccountRepresentation {
	var userRep AccountRepresentation

	userRep.Username = userKc.Username
	userRep.Email = userKc.Email
	userRep.EmailVerified = userKc.EmailVerified
	userRep.FirstName = userKc.FirstName
	userRep.LastName = userKc.LastName

	if value := userKc.GetAttributeString(constants.AttrbPhoneNumber); value != nil {
		userRep.PhoneNumber = value
	}
	if verified, err := userKc.GetAttributeBool(constants.AttrbPhoneNumberVerified); err == nil && verified != nil {
		userRep.PhoneNumberVerified = verified
	}
	if value := userKc.GetAttributeString(constants.AttrbGender); value != nil {
		userRep.Gender = value
	}
	if value := userKc.GetAttributeDate(constants.AttrbBirthDate, constants.SupportedDateLayouts); value != nil {
		userRep.BirthDate = value
	}
	if value := userKc.GetAttributeString(constants.AttrbLocale); value != nil {
		userRep.Locale = value
	}
	if values := userKc.GetAttribute(constants.AttrbAccreditations); len(values) > 0 {
		userRep.Accreditations = convertToAccreditations(ctx, values, logger)
	}

	return userRep
}

func convertToAccreditations(ctx context.Context, values []string, logger keycloakb.Logger) *[]AccreditationRepresentation {
	var accreds []AccreditationRepresentation
	for _, accredJSON := range values {
		var accred AccreditationRepresentation
		if json.Unmarshal([]byte(accredJSON), &accred) == nil {
			// If accreditation is revoked, considers it as expired so that this version is backward compatible with backoffice
			if accred.Revoked != nil && *accred.Revoked {
				accred.Expired = accred.Revoked
			} else {
				accred.Expired = keycloakb.IsDateInThePast(accred.ExpiryDate)
			}
			accred.Revoked = nil
			accreds = append(accreds, accred)
		} else {
			logger.Warn(ctx, "msg", "Can't unmarshall JSON", "json", accredJSON)
		}
	}
	return &accreds
}

// ConvertToKCUser creates a KC user representation from an API user
func ConvertToKCUser(user AccountRepresentation) kc.UserRepresentation {
	var userRep kc.UserRepresentation

	userRep.Username = user.Username
	userRep.Email = user.Email
	userRep.FirstName = user.FirstName
	userRep.LastName = user.LastName

	var attributes = make(kc.Attributes)
	attributes.SetStringWhenNotNil(constants.AttrbPhoneNumber, user.PhoneNumber)
	attributes.SetStringWhenNotNil(constants.AttrbLocale, user.Locale)

	if len(attributes) > 0 {
		userRep.Attributes = &attributes
	}

	return userRep
}

// Validators

// Validate is a validator for AccountRepresentation
func (user AccountRepresentation) Validate() error {
	return validation.NewParameterValidator().
		ValidateParameterRegExp(msg.Username, user.Username, RegExpUsername, false).
		ValidateParameterRegExp(msg.Email, user.Email, RegExpEmail, false).
		ValidateParameterRegExp(msg.Firstname, user.FirstName, RegExpFirstName, false).
		ValidateParameterRegExp(msg.Lastname, user.LastName, RegExpLastName, false).
		ValidateParameterRegExp(msg.PhoneNumber, user.PhoneNumber, RegExpPhoneNumber, false).
		ValidateParameterRegExp(msg.Locale, user.Locale, RegExpLocale, false).
		ValidateParameterRegExp(msg.Gender, user.Gender, constants.RegExpGender, false).
		ValidateParameterDateMultipleLayout(msg.Birthdate, user.BirthDate, constants.SupportedDateLayouts, false).
		ValidateParameterRegExp(msg.BirthLocation, user.BirthLocation, constants.RegExpNameSpecialChars, false).
		ValidateParameterRegExp(msg.Nationality, user.Nationality, constants.RegExpCountryCode, false).
		ValidateParameterIn(msg.IDDocumentType, user.IDDocumentType, constants.AllowedDocumentTypes, false).
		ValidateParameterRegExp(msg.IDDocumentNumber, user.IDDocumentNumber, constants.RegExpIDDocumentNumber, false).
		ValidateParameterLength(msg.IDDocumentNumber, user.IDDocumentNumber, 1, 50, false).
		ValidateParameterDateMultipleLayout(msg.IDDocumentExpiration, user.IDDocumentExpiration, constants.SupportedDateLayouts, false).
		ValidateParameterRegExp(msg.IDDocumentCountry, user.IDDocumentCountry, constants.RegExpCountryCode, false).
		Status()
}

// Validate is a validator for UpdatePasswordBody
func (updatePwd UpdatePasswordBody) Validate() error {
	return validation.NewParameterValidator().
		ValidateParameterRegExp(msg.CurrentPassword, &updatePwd.CurrentPassword, RegExpPassword, true).
		ValidateParameterRegExp(msg.NewPassword, &updatePwd.NewPassword, RegExpPassword, true).
		ValidateParameterRegExp(msg.ConfirmPassword, &updatePwd.ConfirmPassword, RegExpPassword, true).
		Status()
}

// Validate is a validator for CredentialRepresentation
func (credential CredentialRepresentation) Validate() error {
	return validation.NewParameterValidator().
		ValidateParameterRegExp(msg.ID, credential.ID, RegExpID, false).
		ValidateParameterRegExp(msg.Type, credential.Type, RegExpType, false).
		ValidateParameterRegExp(msg.Label, credential.UserLabel, RegExpLabel, false).
		Status()
}

// Regular expressions for parameters validation
const (
	RegExpID         = constants.RegExpID
	RegExpIDNullable = `^([a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12})$|^(null)$`
	RegExpLabel      = `^.{0,255}$`
	RegExpType       = `^[a-zA-Z0-9-_]{1,128}$`
	RegExpRealmName  = constants.RegExpRealmName

	// Password
	RegExpPassword = constants.RegExpPassword
	// User
	RegExpUsername    = constants.RegExpUsername
	RegExpEmail       = constants.RegExpEmail
	RegExpFirstName   = constants.RegExpFirstName
	RegExpLastName    = constants.RegExpLastName
	RegExpPhoneNumber = constants.RegExpPhoneNumber
	RegExpLocale      = `^\w{2}(-\w{2})?$`
)
