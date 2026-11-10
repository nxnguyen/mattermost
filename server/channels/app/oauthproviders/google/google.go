// Copyright (c) 2026-present Amobi Collab. All Rights Reserved.

package oauthgoogle

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
)

const (
	// AllowedEmailsEnv is a comma-separated list of exact Google email addresses allowed to sign in.
	AllowedEmailsEnv = "AMOBI_GOOGLE_ALLOWED_EMAILS"
	// GoogleOnlyModeEnv can keep all non-Google sign-in methods disabled during incomplete setup.
	GoogleOnlyModeEnv = "AMOBI_GOOGLE_ONLY"
	userInfoEndpoint  = "https://openidconnect.googleapis.com/v1/userinfo"
	maxUserInfoSize   = 64 * 1024
)

type GoogleProvider struct{}

// ErrEmailNotAllowed distinguishes an access policy rejection from invalid Google user info.
var ErrEmailNotAllowed = errors.New("Google account is not allowlisted")

type googleUserInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func init() {
	einterfaces.RegisterOAuthProvider(model.ServiceGoogle, &GoogleProvider{})
}

func (p *GoogleProvider) GetUserFromJSON(rctx request.CTX, data io.Reader, _ *model.User, _ *model.SSOSettings) (*model.User, error) {
	decoder := json.NewDecoder(io.LimitReader(data, maxUserInfoSize))
	var info googleUserInfo
	if err := decoder.Decode(&info); err != nil {
		return nil, err
	}

	info.Email = normalizeEmail(info.Email)
	if info.Subject == "" {
		return nil, errors.New("Google user info is missing a subject")
	}
	if len(info.Subject) > model.UserAuthDataMaxLength {
		return nil, errors.New("Google user info subject is too long")
	}
	if !isEmailAddress(info.Email) || !info.EmailVerified {
		return nil, errors.New("Google account must have a verified email address")
	}
	if !isAllowedEmail(info.Email, os.Getenv(AllowedEmailsEnv)) {
		return nil, ErrEmailNotAllowed
	}

	username := strings.SplitN(info.Email, "@", 2)[0]
	username = model.CleanUsername(rctx.Logger(), username)

	return &model.User{
		Username:      username,
		FirstName:     info.Email,
		Email:         info.Email,
		EmailVerified: true,
		AuthData:      &info.Subject,
		AuthService:   model.ServiceGoogle,
	}, nil
}

// IsGoogleOnlyMode reports whether Google-only login is explicitly enabled or fully configured.
// The explicit setting keeps other authentication methods disabled during incomplete setup.
func IsGoogleOnlyMode(config *model.Config) bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv(GoogleOnlyModeEnv)), "true") {
		return true
	}
	if config == nil || !model.SafeDereference(config.GoogleSettings.Enable) {
		return false
	}
	if strings.TrimSpace(model.SafeDereference(config.GoogleSettings.Id)) == "" || strings.TrimSpace(model.SafeDereference(config.GoogleSettings.Secret)) == "" {
		return false
	}
	return HasAllowedEmails(os.Getenv(AllowedEmailsEnv))
}

// HasAllowedEmails reports whether the allowlist contains at least one usable email address.
func HasAllowedEmails(rawAllowlist string) bool {
	for _, candidate := range strings.Split(rawAllowlist, ",") {
		if isEmailAddress(normalizeEmail(candidate)) {
			return true
		}
	}
	return false
}

func (p *GoogleProvider) GetSSOSettings(_ request.CTX, config *model.Config, _ string) (*model.SSOSettings, error) {
	if config == nil {
		return nil, errors.New("Mattermost Google settings are unavailable")
	}

	settings := config.GoogleSettings
	scope := strings.Fields(model.SafeDereference(settings.Scope))
	for _, requiredScope := range []string{"openid", "email", "profile"} {
		if !contains(scope, requiredScope) {
			scope = append(scope, requiredScope)
		}
	}
	settings.Scope = model.NewPointer(strings.Join(scope, " "))
	settings.UserAPIEndpoint = model.NewPointer(userInfoEndpoint)
	return &settings, nil
}

// Google identity is established through Google's userinfo endpoint using the access token
// exchanged by Mattermost. The ID token isn't needed to build the Mattermost user record.
func (p *GoogleProvider) GetUserFromIdToken(_ request.CTX, _ string) (*model.User, error) {
	return nil, nil
}

func (p *GoogleProvider) IsSameUser(_ request.CTX, dbUser, oauthUser *model.User) bool {
	return dbUser != nil && oauthUser != nil && dbUser.GetAuthData() != "" && dbUser.GetAuthData() == oauthUser.GetAuthData()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isAllowedEmail(email string, rawAllowlist string) bool {
	for _, candidate := range strings.Split(rawAllowlist, ",") {
		if isEmailAddress(email) && isEmailAddress(normalizeEmail(candidate)) && normalizeEmail(candidate) == email {
			return true
		}
	}
	return false
}

func isEmailAddress(email string) bool {
	if strings.ContainsAny(email, " \t\r\n<>") || strings.Count(email, "@") != 1 {
		return false
	}
	parts := strings.SplitN(email, "@", 2)
	return parts[0] != "" && parts[1] != ""
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
