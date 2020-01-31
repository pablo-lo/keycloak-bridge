package kyc

import (
	"context"
	"errors"
	"testing"

	cs "github.com/cloudtrust/common-service"
	log "github.com/cloudtrust/common-service/log"
	apikyc "github.com/cloudtrust/keycloak-bridge/api/kyc"
	"github.com/cloudtrust/keycloak-bridge/internal/dto"
	"github.com/cloudtrust/keycloak-bridge/pkg/kyc/mock"
	kc "github.com/cloudtrust/keycloak-client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func createValidUser() apikyc.UserRepresentation {
	var (
		gender        = "M"
		firstName     = "Marc"
		lastName      = "El-Bichoun"
		email         = "marcel.bichon@elca.ch"
		phoneNumber   = "00 33 686 550011"
		birthDate     = "31.03.2001"
		birthLocation = "Montreux"
		docType       = "ID_CARD"
		docNumber     = "MEL123789654ABC"
		docExp        = "28.02.2050"
	)

	return apikyc.UserRepresentation{
		Gender:               &gender,
		FirstName:            &firstName,
		LastName:             &lastName,
		EmailAddress:         &email,
		PhoneNumber:          &phoneNumber,
		BirthDate:            &birthDate,
		BirthLocation:        &birthLocation,
		IDDocumentType:       &docType,
		IDDocumentNumber:     &docNumber,
		IDDocumentExpiration: &docExp,
	}
}

func TestGetActions(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockConfigDB = mock.NewConfigurationDBModule(mockCtrl)
	var mockUsersDB = mock.NewUsersDBModule(mockCtrl)
	var mockEventsDB = mock.NewEventsDBModule(mockCtrl)

	var component = NewComponent("realm", mockKeycloakClient, mockUsersDB, mockConfigDB, mockEventsDB, log.NewNopLogger())

	t.Run("GetActions", func(t *testing.T) {
		var res, err = component.GetActions(context.TODO())
		assert.Nil(t, err)
		assert.NotEqual(t, 0, len(res))
	})
}

func TestGetUserComponent(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockConfigDB = mock.NewConfigurationDBModule(mockCtrl)
	var mockUsersDB = mock.NewUsersDBModule(mockCtrl)
	var mockEventsDB = mock.NewEventsDBModule(mockCtrl)

	var accessToken = "abcd-1234"
	var realm = "my-realm"
	var username = "utr167x"
	var userID = "1234567890"
	var one = 1
	var kcUser = kc.UserRepresentation{
		Id:       &userID,
		Username: &username,
	}
	var kcSearchResp = kc.UsersPageRepresentation{Count: &one, Users: []kc.UserRepresentation{kcUser}}
	var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

	var component = NewComponent(realm, mockKeycloakClient, mockUsersDB, mockConfigDB, mockEventsDB, log.NewNopLogger())

	t.Run("GetUsers from Keycloak fails", func(t *testing.T) {
		var kcError = errors.New("kc error")
		mockKeycloakClient.EXPECT().GetUsers(accessToken, realm, realm, "username", username).Return(kc.UsersPageRepresentation{}, kcError)
		var _, err = component.GetUser(ctx, username)
		assert.NotNil(t, err)
	})

	t.Run("User not found in KC", func(t *testing.T) {
		var count = 0
		mockKeycloakClient.EXPECT().GetUsers(accessToken, realm, realm, "username", username).Return(kc.UsersPageRepresentation{Count: &count}, nil)
		var _, err = component.GetUser(ctx, username)
		assert.NotNil(t, err)
	})

	t.Run("GetUsers from DB fails", func(t *testing.T) {
		mockKeycloakClient.EXPECT().GetUsers(accessToken, realm, realm, "username", username).Return(kcSearchResp, nil)
		mockUsersDB.EXPECT().GetUser(ctx, realm, *kcUser.Id).Return(nil, errors.New("database"))
		var _, err = component.GetUser(ctx, username)
		assert.NotNil(t, err)
	})

	t.Run("GetUsers success", func(t *testing.T) {
		mockKeycloakClient.EXPECT().GetUsers(accessToken, realm, realm, "username", username).Return(kcSearchResp, nil)
		mockUsersDB.EXPECT().GetUser(ctx, realm, *kcUser.Id).Return(&dto.DBUser{}, nil)
		var user, err = component.GetUser(ctx, username)
		assert.Nil(t, err)
		assert.NotNil(t, user)
	})
}

func createUser(userID, username string, emailVerified bool, phoneNumberVerified bool) kc.UserRepresentation {
	var pnv = "false"
	if phoneNumberVerified {
		pnv = "true"
	}
	var attributes = map[string][]string{"phoneNumberVerified": []string{pnv}}
	return kc.UserRepresentation{
		Id:            &userID,
		Username:      &username,
		EmailVerified: &emailVerified,
		Attributes:    &attributes,
	}
}

func TestValidateUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockConfigDB = mock.NewConfigurationDBModule(mockCtrl)
	var mockUsersDB = mock.NewUsersDBModule(mockCtrl)
	var mockEventsDB = mock.NewEventsDBModule(mockCtrl)

	var targetRealm = "cloudtrust"
	var validUser = createValidUser()
	var userID = "abc789def"
	var username = "user_name"
	var kcUser = createUser(userID, username, true, true)
	var accessToken = "abcdef"
	var ctx = context.TODO()
	var dbUser = dto.DBUser{UserID: &userID}

	var component = NewComponent(targetRealm, mockKeycloakClient, mockUsersDB, mockConfigDB, mockEventsDB, log.NewNopLogger())

	ctx = context.WithValue(ctx, cs.CtContextAccessToken, accessToken)
	ctx = context.WithValue(ctx, cs.CtContextUsername, "operator")

	t.Run("Invalid user", func(t *testing.T) {
		var invalidUser = createValidUser()
		invalidUser.FirstName = nil
		var err = component.ValidateUser(ctx, userID, invalidUser)
		assert.NotNil(t, err)
	})

	t.Run("Email not verified", func(t *testing.T) {
		var searchResult = createUser(userID, username, false, true)
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(searchResult, nil)

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.NotNil(t, err)
	})

	t.Run("PhoneNumber not verified", func(t *testing.T) {
		var searchResult = createUser(userID, username, true, false)
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(searchResult, nil)

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.NotNil(t, err)
	})

	t.Run("SQL error when searching user in database", func(t *testing.T) {
		var sqlError = errors.New("sql error")
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kcUser, nil)
		mockUsersDB.EXPECT().GetUser(ctx, targetRealm, userID).Return(nil, sqlError)

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.NotNil(t, err)
	})

	t.Run("User not found in database", func(t *testing.T) {
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kcUser, nil)
		mockUsersDB.EXPECT().GetUser(ctx, targetRealm, userID).Return(nil, nil)

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.NotNil(t, err)
	})

	t.Run("Keycloak update fails", func(t *testing.T) {
		var kcError = errors.New("keycloak error")
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kcUser, nil)
		mockUsersDB.EXPECT().GetUser(ctx, targetRealm, userID).Return(&dbUser, nil)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(kcError)

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.Equal(t, kcError, err)
	})

	t.Run("Update user in DB fails", func(t *testing.T) {
		var dbError = errors.New("db update error")
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kcUser, nil)
		mockUsersDB.EXPECT().GetUser(ctx, targetRealm, userID).Return(&dbUser, nil)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(nil)
		mockUsersDB.EXPECT().StoreOrUpdateUser(ctx, targetRealm, gomock.Any()).Return(dbError)

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.Equal(t, dbError, err)
	})

	t.Run("ValidateUser is successful", func(t *testing.T) {
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kcUser, nil)
		mockUsersDB.EXPECT().GetUser(ctx, targetRealm, userID).Return(&dbUser, nil)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(nil)
		mockUsersDB.EXPECT().StoreOrUpdateUser(ctx, targetRealm, gomock.Any()).Return(nil)
		mockEventsDB.EXPECT().ReportEvent(gomock.Any(), "VALIDATE_USER", "back-office", gomock.Any())

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.Nil(t, err)
	})

	t.Run("ValidateUser is successful - Report event fails", func(t *testing.T) {
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kcUser, nil)
		mockUsersDB.EXPECT().GetUser(ctx, targetRealm, userID).Return(&dbUser, nil)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(nil)
		mockUsersDB.EXPECT().StoreOrUpdateUser(ctx, targetRealm, gomock.Any()).Return(nil)
		mockEventsDB.EXPECT().ReportEvent(gomock.Any(), "VALIDATE_USER", "back-office", gomock.Any()).Return(errors.New("report fails"))

		var err = component.ValidateUser(ctx, userID, validUser)
		assert.Nil(t, err)
	})
}

func TestIsPhoneNumberVerified(t *testing.T) {
	t.Run("No attributes", func(t *testing.T) {
		assert.False(t, isPhoneNumberVerified(nil))
	})

	var attrbs = make(map[string][]string)

	t.Run("No phoneNumberVerified attribute", func(t *testing.T) {
		assert.False(t, isPhoneNumberVerified(&attrbs))
	})

	t.Run("Invalid phoneNumberVerified attribute", func(t *testing.T) {
		attrbs["phoneNumberVerified"] = []string{"not a boolean"}
		assert.False(t, isPhoneNumberVerified(&attrbs))
	})

	t.Run("phoneNumberVerified is false", func(t *testing.T) {
		attrbs["phoneNumberVerified"] = []string{"false"}
		assert.False(t, isPhoneNumberVerified(&attrbs))
	})

	t.Run("phoneNumberVerified is true", func(t *testing.T) {
		attrbs["phoneNumberVerified"] = []string{"true"}
		assert.True(t, isPhoneNumberVerified(&attrbs))
	})
}