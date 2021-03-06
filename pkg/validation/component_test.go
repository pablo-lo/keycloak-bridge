package validation

import (
	"context"
	"errors"
	"testing"
	"time"

	log "github.com/cloudtrust/common-service/log"
	apikyc "github.com/cloudtrust/keycloak-bridge/api/kyc"
	api "github.com/cloudtrust/keycloak-bridge/api/validation"
	"github.com/cloudtrust/keycloak-bridge/internal/dto"
	"github.com/cloudtrust/keycloak-bridge/internal/keycloakb"
	"github.com/cloudtrust/keycloak-bridge/pkg/validation/mock"

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
		nationality   = "CH"
		docType       = "ID_CARD"
		docNumber     = "MEL123789654ABC"
		docExp        = "28.02.2050"
		docCountry    = "IL"
	)

	return apikyc.UserRepresentation{
		Gender:               &gender,
		FirstName:            &firstName,
		LastName:             &lastName,
		Email:                &email,
		PhoneNumber:          &phoneNumber,
		BirthDate:            &birthDate,
		BirthLocation:        &birthLocation,
		Nationality:          &nationality,
		IDDocumentType:       &docType,
		IDDocumentNumber:     &docNumber,
		IDDocumentExpiration: &docExp,
		IDDocumentCountry:    &docCountry,
	}
}

func TestGetUserComponent(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockUsersDB = mock.NewUsersDetailsDBModule(mockCtrl)
	var mockEventsDB = mock.NewEventsDBModule(mockCtrl)
	var mockTokenProvider = mock.NewTokenProvider(mockCtrl)
	var mockAccreditations = mock.NewAccreditationsModule(mockCtrl)

	var accessToken = "abcd-1234"
	var realm = "my-realm"
	var userID = ""

	var ctx = context.Background()

	var component = NewComponent(mockKeycloakClient, mockTokenProvider, mockUsersDB, nil, mockEventsDB, mockAccreditations, log.NewNopLogger())

	t.Run("Fails to retrieve token for technical user", func(t *testing.T) {
		var kcError = errors.New("kc error")
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return("", kcError)
		var _, err = component.GetUser(ctx, realm, userID)
		assert.NotNil(t, err)
	})

	t.Run("GetUser from Keycloak fails", func(t *testing.T) {
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return(accessToken, nil)
		var kcError = errors.New("kc error")
		mockKeycloakClient.EXPECT().GetUser(accessToken, realm, userID).Return(kc.UserRepresentation{}, kcError)
		var _, err = component.GetUser(ctx, realm, userID)
		assert.NotNil(t, err)
	})

	t.Run("GetUser from DB fails", func(t *testing.T) {
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return(accessToken, nil)
		mockKeycloakClient.EXPECT().GetUser(accessToken, realm, userID).Return(kc.UserRepresentation{}, nil)
		var dbError = errors.New("DB error")
		mockUsersDB.EXPECT().GetUserDetails(ctx, realm, userID).Return(dto.DBUser{}, dbError)
		var _, err = component.GetUser(ctx, realm, userID)
		assert.NotNil(t, err)
	})

	t.Run("No user found in DB", func(t *testing.T) {
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return(accessToken, nil)
		mockKeycloakClient.EXPECT().GetUser(accessToken, realm, userID).Return(kc.UserRepresentation{}, nil)
		mockUsersDB.EXPECT().GetUserDetails(ctx, realm, userID).Return(dto.DBUser{
			UserID: &userID,
		}, nil)
		var _, err = component.GetUser(ctx, realm, userID)
		assert.Nil(t, err)
	})

	t.Run("Date parsing error", func(t *testing.T) {
		var expirationDate = "01.01-2020"
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return(accessToken, nil)
		mockKeycloakClient.EXPECT().GetUser(accessToken, realm, userID).Return(kc.UserRepresentation{}, nil)
		mockUsersDB.EXPECT().GetUserDetails(ctx, realm, userID).Return(dto.DBUser{
			IDDocumentExpiration: &expirationDate,
		}, nil)
		var _, err = component.GetUser(ctx, realm, userID)
		assert.NotNil(t, err)
	})

	t.Run("Happy path", func(t *testing.T) {
		var expirationDate = "01.01.2020"
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return(accessToken, nil)
		mockKeycloakClient.EXPECT().GetUser(accessToken, realm, userID).Return(kc.UserRepresentation{}, nil)
		mockUsersDB.EXPECT().GetUserDetails(ctx, realm, userID).Return(dto.DBUser{
			IDDocumentExpiration: &expirationDate,
		}, nil)
		var _, err = component.GetUser(ctx, realm, userID)
		assert.Nil(t, err)
	})

}

func TestUpdateUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockUsersDB = mock.NewUsersDetailsDBModule(mockCtrl)
	var mockArchiveUsersDB = mock.NewArchiveDBModule(mockCtrl)
	var mockEventsDB = mock.NewEventsDBModule(mockCtrl)
	var mockTokenProvider = mock.NewTokenProvider(mockCtrl)
	var mockAccreditations = mock.NewAccreditationsModule(mockCtrl)

	var targetRealm = "cloudtrust"
	var userID = "abc789def"
	var accessToken = "abcdef"
	var ctx = context.TODO()

	var component = NewComponent(mockKeycloakClient, mockTokenProvider, mockUsersDB, mockArchiveUsersDB, mockEventsDB, mockAccreditations, log.NewNopLogger())

	t.Run("Fails to retrieve token for technical user", func(t *testing.T) {
		var user = api.UserRepresentation{
			FirstName: ptr("newFirstname"),
		}
		var kcError = errors.New("kc error")
		mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return("", kcError)
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.NotNil(t, err)
	})
	mockTokenProvider.EXPECT().ProvideToken(gomock.Any()).Return(accessToken, nil).AnyTimes()

	t.Run("No update needed", func(t *testing.T) {
		var user = api.UserRepresentation{}
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.Nil(t, err)
	})

	t.Run("Fails to update user in DB", func(t *testing.T) {
		var user = api.UserRepresentation{
			FirstName:      ptr("newFirstname"),
			IDDocumentType: ptr("type"),
		}
		mockUsersDB.EXPECT().GetUserDetails(ctx, targetRealm, userID).Return(dto.DBUser{
			UserID: &userID,
		}, nil)
		var dbError = errors.New("db error")
		mockUsersDB.EXPECT().StoreOrUpdateUserDetails(ctx, targetRealm, gomock.Any()).Return(dbError)
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.NotNil(t, err)
	})
	mockUsersDB.EXPECT().GetUserDetails(ctx, targetRealm, userID).Return(dto.DBUser{
		UserID: &userID,
	}, nil).AnyTimes()
	mockUsersDB.EXPECT().StoreOrUpdateUserDetails(ctx, targetRealm, gomock.Any()).Return(nil).AnyTimes()

	t.Run("Fails to get user from KC", func(t *testing.T) {
		var user = api.UserRepresentation{
			FirstName: ptr("newFirstname"),
		}
		var kcError = errors.New("kc error")
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kc.UserRepresentation{}, kcError)
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.NotNil(t, err)
	})
	mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kc.UserRepresentation{}, nil).AnyTimes()

	t.Run("Fails to update user in KC", func(t *testing.T) {
		var date = time.Now()
		var user = api.UserRepresentation{
			BirthDate: &date,
		}
		var kcError = errors.New("kc error")
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(kcError)
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.NotNil(t, err)
	})

	t.Run("Fails to update user in KC", func(t *testing.T) {
		var user = api.UserRepresentation{
			FirstName: ptr("newFirstname"),
		}
		var kcError = errors.New("kc error")
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(kcError)
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.NotNil(t, err)
	})
	mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, gomock.Any()).Return(nil).AnyTimes()

	t.Run("Failure to store event", func(t *testing.T) {
		var date = time.Now()
		var user = api.UserRepresentation{
			FirstName:            ptr("newFirstname"),
			IDDocumentExpiration: &date,
		}
		var e = errors.New("error")
		mockEventsDB.EXPECT().ReportEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(e)
		mockArchiveUsersDB.EXPECT().StoreUserDetails(ctx, targetRealm, gomock.Any()).Return(nil)
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.Nil(t, err)
	})
	mockEventsDB.EXPECT().ReportEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockArchiveUsersDB.EXPECT().StoreUserDetails(ctx, targetRealm, gomock.Any()).Return(nil)

	t.Run("Successful update", func(t *testing.T) {
		var user = api.UserRepresentation{
			FirstName:      ptr("newFirstname"),
			IDDocumentType: ptr("type"),
		}
		var err = component.UpdateUser(ctx, targetRealm, userID, user)
		assert.Nil(t, err)
	})
}

func TestCreateCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockUsersDB = mock.NewUsersDetailsDBModule(mockCtrl)
	var mockArchiveUsersDB = mock.NewArchiveDBModule(mockCtrl)
	var mockEventsDB = mock.NewEventsDBModule(mockCtrl)
	var mockTokenProvider = mock.NewTokenProvider(mockCtrl)
	var mockAccreditations = mock.NewAccreditationsModule(mockCtrl)

	var targetRealm = "cloudtrust"
	var userID = "abc789def"
	var accessToken = "the-access-token"
	var ctx = context.TODO()
	var datetime = time.Now()
	var check = api.CheckRepresentation{
		Operator: ptr("operator"),
		DateTime: &datetime,
		Status:   ptr("status"),
	}

	var component = NewComponent(mockKeycloakClient, mockTokenProvider, mockUsersDB, mockArchiveUsersDB, mockEventsDB, mockAccreditations, log.NewNopLogger())

	t.Run("Fails to store check in DB", func(t *testing.T) {
		var dbError = errors.New("db error")
		mockUsersDB.EXPECT().CreateCheck(ctx, targetRealm, userID, gomock.Any()).Return(dbError)
		var err = component.CreateCheck(ctx, targetRealm, userID, check)
		assert.NotNil(t, err)
	})

	t.Run("Can't get access token", func(t *testing.T) {
		check.Status = ptr("SUCCESS")
		mockUsersDB.EXPECT().CreateCheck(ctx, targetRealm, userID, gomock.Any()).Return(nil)
		mockTokenProvider.EXPECT().ProvideToken(ctx).Return("", errors.New("no token"))
		var err = component.CreateCheck(ctx, targetRealm, userID, check)
		assert.NotNil(t, err)
	})
	t.Run("Accreditation module fails", func(t *testing.T) {
		var kcUser kc.UserRepresentation
		check.Status = ptr("SUCCESS")
		mockUsersDB.EXPECT().CreateCheck(ctx, targetRealm, userID, gomock.Any()).Return(nil)
		mockTokenProvider.EXPECT().ProvideToken(ctx).Return(accessToken, nil)
		mockAccreditations.EXPECT().GetUserAndPrepareAccreditations(ctx, accessToken, targetRealm, userID, keycloakb.CredsIDNow).Return(kcUser, 0, errors.New("Accreds failed"))
		var err = component.CreateCheck(ctx, targetRealm, userID, check)
		assert.NotNil(t, err)
	})

	t.Run("Success w/o accreditations", func(t *testing.T) {
		check.Status = ptr("FRAUD_SUSPICION_CONFIRMED")
		mockUsersDB.EXPECT().CreateCheck(ctx, targetRealm, userID, gomock.Any()).Return(nil)
		mockEventsDB.EXPECT().ReportEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockTokenProvider.EXPECT().ProvideToken(ctx).Return(accessToken, nil)
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kc.UserRepresentation{}, nil)
		mockUsersDB.EXPECT().GetUserDetails(ctx, targetRealm, userID).Return(dto.DBUser{}, nil)
		mockArchiveUsersDB.EXPECT().StoreUserDetails(ctx, targetRealm, gomock.Any()).Return(nil)
		var err = component.CreateCheck(ctx, targetRealm, userID, check)
		assert.Nil(t, err)
	})
	t.Run("Computed accreditations, fails to store them in Keycloak", func(t *testing.T) {
		var kcUser kc.UserRepresentation
		check.Status = ptr("SUCCESS")
		mockUsersDB.EXPECT().CreateCheck(ctx, targetRealm, userID, gomock.Any()).Return(nil)
		mockTokenProvider.EXPECT().ProvideToken(ctx).Return(accessToken, nil)
		mockAccreditations.EXPECT().GetUserAndPrepareAccreditations(ctx, accessToken, targetRealm, userID, keycloakb.CredsIDNow).Return(kcUser, 1, nil)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, kcUser).Return(errors.New("KC fails"))
		mockKeycloakClient.EXPECT().GetUser(accessToken, targetRealm, userID).Return(kc.UserRepresentation{}, nil)
		mockUsersDB.EXPECT().GetUserDetails(ctx, targetRealm, userID).Return(dto.DBUser{}, nil)
		mockArchiveUsersDB.EXPECT().StoreUserDetails(ctx, targetRealm, gomock.Any()).Return(nil)
		var err = component.CreateCheck(ctx, targetRealm, userID, check)
		assert.NotNil(t, err)
	})
	t.Run("Success with accreditations", func(t *testing.T) {
		var kcUser kc.UserRepresentation
		check.Status = ptr("SUCCESS")
		mockUsersDB.EXPECT().CreateCheck(ctx, targetRealm, userID, gomock.Any()).Return(nil)
		mockTokenProvider.EXPECT().ProvideToken(ctx).Return(accessToken, nil)
		mockAccreditations.EXPECT().GetUserAndPrepareAccreditations(ctx, accessToken, targetRealm, userID, keycloakb.CredsIDNow).Return(kcUser, 1, nil)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, targetRealm, userID, kcUser).Return(nil)
		mockEventsDB.EXPECT().ReportEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		var err = component.CreateCheck(ctx, targetRealm, userID, check)
		assert.Nil(t, err)
	})
}

func ptr(value string) *string {
	return &value
}

func TestValidationContext(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()

	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockUsersDB = mock.NewUsersDetailsDBModule(mockCtrl)
	var mockArchiveUsersDB = mock.NewArchiveDBModule(mockCtrl)
	var mockTokenProvider = mock.NewTokenProvider(mockCtrl)
	var mockAccreditations = mock.NewAccreditationsModule(mockCtrl)
	var component = &component{
		keycloakClient:  mockKeycloakClient,
		usersDBModule:   mockUsersDB,
		archiveDBModule: mockArchiveUsersDB,
		tokenProvider:   mockTokenProvider,
		accredsModule:   mockAccreditations,
		logger:          log.NewNopLogger(),
	}

	var validationCtx = &validationContext{
		ctx:       context.TODO(),
		realmName: "my-realm",
		userID:    "abcd-4567",
		kcUser:    &kc.UserRepresentation{},
	}
	var accessToken = "abcd1234.efgh.5678ijkl"
	var anyError = errors.New("Any error")

	t.Run("updateKeycloakUser", func(t *testing.T) {
		t.Run("Fails to get access token", func(t *testing.T) {
			mockTokenProvider.EXPECT().ProvideToken(validationCtx.ctx).Return("", anyError)
			var err = component.updateKeycloakUser(validationCtx)
			assert.Equal(t, anyError, err)
		})
		t.Run("Fails to update user", func(t *testing.T) {
			mockTokenProvider.EXPECT().ProvideToken(validationCtx.ctx).Return(accessToken, nil)
			mockKeycloakClient.EXPECT().UpdateUser(accessToken, validationCtx.realmName, validationCtx.userID, gomock.Any()).Return(anyError)
			var err = component.updateKeycloakUser(validationCtx)
			assert.NotNil(t, err)
		})
		t.Run("Success", func(t *testing.T) {
			// already got an access token : won't retry
			mockKeycloakClient.EXPECT().UpdateUser(accessToken, validationCtx.realmName, validationCtx.userID, gomock.Any()).Return(nil)
			var err = component.updateKeycloakUser(validationCtx)
			assert.Nil(t, err)
		})
	})

	t.Run("getUserWithAccreditations", func(t *testing.T) {
		validationCtx.accessToken = nil
		validationCtx.kcUser = nil
		t.Run("Fails to get access token", func(t *testing.T) {
			mockTokenProvider.EXPECT().ProvideToken(validationCtx.ctx).Return("", anyError)
			var _, err = component.getUserWithAccreditations(validationCtx)
			assert.Equal(t, anyError, err)
		})
		t.Run("Fails to get user/accreditations", func(t *testing.T) {
			mockTokenProvider.EXPECT().ProvideToken(validationCtx.ctx).Return(accessToken, nil)
			mockAccreditations.EXPECT().GetUserAndPrepareAccreditations(validationCtx.ctx, accessToken, validationCtx.realmName,
				validationCtx.userID, gomock.Any()).Return(kc.UserRepresentation{}, 0, anyError)
			var _, err = component.getUserWithAccreditations(validationCtx)
			assert.Equal(t, anyError, err)
		})
		t.Run("Success", func(t *testing.T) {
			// already got an access token : won't retry
			mockAccreditations.EXPECT().GetUserAndPrepareAccreditations(validationCtx.ctx, accessToken, validationCtx.realmName,
				validationCtx.userID, gomock.Any()).Return(kc.UserRepresentation{}, 0, nil)
			var _, err = component.getUserWithAccreditations(validationCtx)
			assert.Nil(t, err)
		})
	})

	t.Run("Archive user", func(t *testing.T) {
		validationCtx.accessToken = &accessToken
		validationCtx.kcUser = nil
		validationCtx.dbUser = nil
		t.Run("get user from keycloak fails", func(t *testing.T) {
			mockKeycloakClient.EXPECT().GetUser(accessToken, validationCtx.realmName, validationCtx.userID).Return(kc.UserRepresentation{}, anyError)
			component.archiveUser(validationCtx)
		})
		mockKeycloakClient.EXPECT().GetUser(accessToken, validationCtx.realmName, validationCtx.userID).Return(kc.UserRepresentation{}, nil).AnyTimes()

		t.Run("get user from DB fails", func(t *testing.T) {
			mockUsersDB.EXPECT().GetUserDetails(validationCtx.ctx, validationCtx.realmName, validationCtx.userID).Return(dto.DBUser{}, anyError)
			component.archiveUser(validationCtx)
		})
		mockUsersDB.EXPECT().GetUserDetails(validationCtx.ctx, validationCtx.realmName, validationCtx.userID).Return(dto.DBUser{}, nil).AnyTimes()

		t.Run("success", func(t *testing.T) {
			mockArchiveUsersDB.EXPECT().StoreUserDetails(validationCtx.ctx, validationCtx.realmName, gomock.Any())
			component.archiveUser(validationCtx)
		})
	})
}
