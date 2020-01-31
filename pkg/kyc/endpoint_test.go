package kyc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	apikyc "github.com/cloudtrust/keycloak-bridge/api/kyc"
	"github.com/cloudtrust/keycloak-bridge/pkg/kyc/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMakeGetActionsEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockKYCComponent := mock.NewComponent(mockCtrl)

	var m = map[string]string{}
	var expectedError = errors.New("get-actions")

	t.Run("GetActions - success case", func(t *testing.T) {
		mockKYCComponent.EXPECT().GetActions(gomock.Any()).Return([]apikyc.ActionRepresentation{}, nil)
		_, err := MakeGetActionsEndpoint(mockKYCComponent)(context.Background(), m)
		assert.Nil(t, err)
	})

	t.Run("GetActions - failure case", func(t *testing.T) {
		mockKYCComponent.EXPECT().GetActions(gomock.Any()).Return([]apikyc.ActionRepresentation{}, expectedError)
		_, err := MakeGetActionsEndpoint(mockKYCComponent)(context.Background(), m)
		assert.Equal(t, expectedError, err)
	})
}

func TestMakeGetUserEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockKYCComponent := mock.NewComponent(mockCtrl)

	var realm = "master"
	var username = "user1234"
	var m = map[string]string{"realm": realm, "username": username}
	var expectedError = errors.New("get-user")

	t.Run("GetUser - success case", func(t *testing.T) {
		mockKYCComponent.EXPECT().GetUser(gomock.Any(), username).Return(apikyc.UserRepresentation{}, nil)
		_, err := MakeGetUserEndpoint(mockKYCComponent)(context.Background(), m)
		assert.Nil(t, err)
	})

	t.Run("GetUser - failure case", func(t *testing.T) {
		mockKYCComponent.EXPECT().GetUser(gomock.Any(), username).Return(apikyc.UserRepresentation{}, expectedError)
		_, err := MakeGetUserEndpoint(mockKYCComponent)(context.Background(), m)
		assert.Equal(t, expectedError, err)
	})
}

func TestMakeValidateUserEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockKYCComponent := mock.NewComponent(mockCtrl)

	var realm = "master"
	var first = "John"
	var last = "Doe"
	var userID = "ux467913"
	var user = apikyc.UserRepresentation{UserID: &userID, FirstName: &first, LastName: &last}
	var m = map[string]string{}

	t.Run("ValidateUser - success case", func(t *testing.T) {
		var bytes, _ = json.Marshal(user)
		m["realm"] = realm
		m["body"] = string(bytes)
		m["userId"] = userID
		mockKYCComponent.EXPECT().ValidateUser(gomock.Any(), userID, user).Return(nil).Times(1)
		_, err := MakeValidateUserEndpoint(mockKYCComponent)(context.Background(), m)
		assert.Nil(t, err)
	})

	t.Run("ValidateUser - failure case", func(t *testing.T) {
		m["realm"] = realm
		m["body"] = "{"
		_, err := MakeValidateUserEndpoint(mockKYCComponent)(context.Background(), m)
		assert.NotNil(t, err)
	})
}