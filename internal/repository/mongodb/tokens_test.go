package mongodb

import (
	"context"
	"testing"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	"github.com/stretchr/testify/require"
)

func generateToken(t *testing.T, id string) *model.Token {
	token, err := generateAPIToken(id)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.Len(t, token.Plaintext, 26)
	require.Len(t, token.Hash, 32)
	require.Equal(t, id, string(token.ServiceID))

	return token
}

func TestGenerateAPIToken(t *testing.T) {
	serviceID := utils.RandomServiceID()
	generateToken(t, serviceID)
}

func TestNewAPIToken(t *testing.T) {
	serviceID := utils.RandomServiceID()
	_, err := tokenRes.NewAPIKey(context.Background(), serviceID)
	require.NoError(t, err)

	_, err = tokenRes.NewAPIKey(context.Background(), serviceID)
	require.Error(t, err)
	require.EqualError(t, err, model.ErrDuplicateService.Error())
}

func TestUpdateToken(t *testing.T) {
	serviceID := utils.RandomServiceID()
	token1, err := tokenRes.NewAPIKey(context.Background(), serviceID)
	require.NoError(t, err)
	require.NotEmpty(t, token1)

	token2, err := tokenRes.UpdateToken(context.Background(), model.ServiceID(serviceID))
	require.NoError(t, err)
	require.NotEmpty(t, token2)

	require.Len(t, token2.Plaintext, 26)
	require.Len(t, token2.Hash, 32)
	require.NotEqual(t, token1.Plaintext, token2.Plaintext)
	require.NotEqual(t, token1.Hash, token2.Hash)
	require.Equal(t, token1.ServiceID, token2.ServiceID)

	unknownToken, err := tokenRes.UpdateToken(context.Background(), model.AnonymousService)
	require.Error(t, err)
	require.EqualError(t, err, model.ErrRecordNotFound.Error())
	require.Nil(t, unknownToken)
}

func TestGetTokenByAPIKey(t *testing.T) {
	serviceID := utils.RandomServiceID()
	token1, err := tokenRes.NewAPIKey(context.Background(), serviceID)
	require.NoError(t, err)
	require.NotEmpty(t, token1)

	returnedServiceID, err := tokenRes.GetTokenByAPIKey(context.Background(), token1.Plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, returnedServiceID)

	require.Equal(t, model.ServiceID(serviceID), *returnedServiceID)

	unknownServiceID, err := tokenRes.GetTokenByAPIKey(context.Background(), utils.RandomString(26))
	require.Error(t, err)
	require.EqualError(t, err, model.ErrRecordNotFound.Error())
	require.Nil(t, unknownServiceID)
}
