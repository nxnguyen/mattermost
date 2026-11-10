// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package oauthgoogle

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

func TestGoogleProviderGetUserFromJSON(t *testing.T) {
	rctx := request.TestContext(t)
	provider := &GoogleProvider{}
	t.Setenv(AllowedEmailsEnv, "NguyenNX@bachasoft.com")

	userJSON := `{"sub":"google-sub-123","email":"NguyenNX@Bachasoft.com","email_verified":true,"name":"Nguyen Nguyen"}`
	user, err := provider.GetUserFromJSON(rctx, strings.NewReader(userJSON), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "nguyennx@bachasoft.com", user.Email)
	assert.Equal(t, "nguyennx@bachasoft.com", user.FirstName)
	assert.Empty(t, user.LastName)
	assert.Equal(t, model.ServiceGoogle, user.AuthService)
	assert.Equal(t, "google-sub-123", user.GetAuthData())
	assert.True(t, user.EmailVerified)
}

func TestGoogleProviderRejectsUnverifiedOrNonAllowlistedProfiles(t *testing.T) {
	rctx := request.TestContext(t)
	provider := &GoogleProvider{}
	t.Setenv(AllowedEmailsEnv, "nguyennx@bachasoft.com")

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "unverified email",
			body: `{"sub":"google-sub-123","email":"nguyennx@bachasoft.com","email_verified":false}`,
		},
		{
			name: "email outside allowlist",
			body: `{"sub":"google-sub-123","email":"other@bachasoft.com","email_verified":true}`,
		},
		{
			name: "missing stable subject",
			body: `{"email":"nguyennx@bachasoft.com","email_verified":true}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.GetUserFromJSON(rctx, strings.NewReader(tc.body), nil, nil)
			require.Error(t, err)
		})
	}
}

func TestGoogleProviderRequiresConfiguredAllowlist(t *testing.T) {
	rctx := request.TestContext(t)
	provider := &GoogleProvider{}
	t.Setenv(AllowedEmailsEnv, "")

	_, err := provider.GetUserFromJSON(rctx, strings.NewReader(`{"sub":"google-sub-123","email":"nguyennx@bachasoft.com","email_verified":true}`), nil, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrEmailNotAllowed)
}

func TestGoogleProviderAllowlistRejectionIsDistinctFromInvalidProfile(t *testing.T) {
	t.Setenv(AllowedEmailsEnv, "nguyennx@bachasoft.com,uyenltt@bachasoft.com")
	provider := &GoogleProvider{}
	for _, tc := range []struct {
		name   string
		body   string
		denied bool
	}{
		{"verified disallowed email", `{"sub":"other-sub","email":"outside@bachasoft.com","email_verified":true}`, true},
		{"unverified disallowed email", `{"sub":"other-sub","email":"outside@bachasoft.com","email_verified":false}`, false},
		{"missing subject", `{"email":"outside@bachasoft.com","email_verified":true}`, false},
		{"malformed response", `{`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.GetUserFromJSON(request.TestContext(t), strings.NewReader(tc.body), nil, nil)
			require.Error(t, err)
			if tc.denied {
				require.ErrorIs(t, err, ErrEmailNotAllowed)
			} else {
				require.NotErrorIs(t, err, ErrEmailNotAllowed)
			}
		})
	}
	user, err := provider.GetUserFromJSON(request.TestContext(t), strings.NewReader(`{"sub":"uyen-sub","email":"UyenLTT@BachaSoft.com","email_verified":true}`), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "uyenltt@bachasoft.com", user.Email)
}

func TestIsGoogleOnlyModeRequiresEnabledGoogleCredentialsAndAllowlist(t *testing.T) {
	t.Setenv(GoogleOnlyModeEnv, "false")
	config := &model.Config{}
	config.SetDefaults()
	*config.GoogleSettings.Enable = true
	*config.GoogleSettings.Id = "google-client-id"
	*config.GoogleSettings.Secret = "google-client-secret"

	t.Setenv(AllowedEmailsEnv, "nguyennx@bachasoft.com")
	assert.True(t, IsGoogleOnlyMode(config))

	t.Setenv(AllowedEmailsEnv, "not-an-email")
	assert.False(t, IsGoogleOnlyMode(config))

	t.Setenv(AllowedEmailsEnv, "nguyennx@bachasoft.com")
	*config.GoogleSettings.Secret = ""
	assert.False(t, IsGoogleOnlyMode(config))

	*config.GoogleSettings.Secret = "google-client-secret"
	*config.GoogleSettings.Enable = false
	assert.False(t, IsGoogleOnlyMode(config))

	t.Setenv(GoogleOnlyModeEnv, "true")
	assert.True(t, IsGoogleOnlyMode(config), "explicit Google-only mode should fail closed during incomplete setup")
}

func TestGoogleProviderUsesVerifiedOpenIDUserInfo(t *testing.T) {
	config := &model.Config{}
	config.SetDefaults()

	settings, err := (&GoogleProvider{}).GetSSOSettings(request.TestContext(t), config, model.ServiceGoogle)
	require.NoError(t, err)
	assert.Equal(t, "https://openidconnect.googleapis.com/v1/userinfo", model.SafeDereference(settings.UserAPIEndpoint))
	assert.Contains(t, strings.Fields(model.SafeDereference(settings.Scope)), "openid")
}
