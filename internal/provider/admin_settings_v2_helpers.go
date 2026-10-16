// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
)

// adminAvatarSource is a flattened view of a single avatar source entry from
// the generated go-tfe v2 admin customization settings model.
type adminAvatarSource struct {
	Source  string `json:"source"`
	Enabled bool   `json:"enabled"`
}

// adminCustomizationSettings is a flattened view of the generated go-tfe v2
// admin customization settings model.
type adminCustomizationSettings struct {
	ID           string
	SupportEmail string
	// SupportURLAddress is a hostname without a scheme, e.g. "support.example.com".
	SupportURLAddress        string
	InternalSupportURL       string
	InternalDocumentationURL string
	InternalTutorialsURL     string
	LoginHelp                string
	Footer                   string
	Error                    string
	NewUser                  string
	AvatarSources            []adminAvatarSource
}

// adminCustomizationSettingsUpdate holds the writable customization attributes.
// Nil fields are omitted from the PATCH request and left unchanged.
type adminCustomizationSettingsUpdate struct {
	SupportEmail             *string
	SupportURLAddress        *string
	InternalSupportURL       *string
	InternalDocumentationURL *string
	InternalTutorialsURL     *string
	LoginHelp                *string
	Footer                   *string
	Error                    *string
	NewUser                  *string
	// AvatarSources is sent when non-nil (an empty slice sends an empty list).
	AvatarSources []adminAvatarSource
}

func (opts adminCustomizationSettingsUpdate) isEmpty() bool {
	return opts.SupportEmail == nil &&
		opts.SupportURLAddress == nil &&
		opts.InternalSupportURL == nil &&
		opts.InternalDocumentationURL == nil &&
		opts.InternalTutorialsURL == nil &&
		opts.LoginHelp == nil &&
		opts.Footer == nil &&
		opts.Error == nil &&
		opts.NewUser == nil &&
		opts.AvatarSources == nil
}

// adminGeneralSettings is a flattened view of the generated go-tfe v2 admin
// general settings model (only the attributes managed by the provider).
type adminGeneralSettings struct {
	ID                       string
	EnablePublicRegistration bool
}

// adminCustomizationSettingsUpdateFrom builds an update that writes every
// attribute of s, used to restore a previously captured snapshot.
func adminCustomizationSettingsUpdateFrom(s *adminCustomizationSettings) adminCustomizationSettingsUpdate {
	sources := s.AvatarSources
	if sources == nil {
		sources = []adminAvatarSource{}
	}
	return adminCustomizationSettingsUpdate{
		SupportEmail:             ptr(s.SupportEmail),
		SupportURLAddress:        ptr(s.SupportURLAddress),
		InternalSupportURL:       ptr(s.InternalSupportURL),
		InternalDocumentationURL: ptr(s.InternalDocumentationURL),
		InternalTutorialsURL:     ptr(s.InternalTutorialsURL),
		LoginHelp:                ptr(s.LoginHelp),
		Footer:                   ptr(s.Footer),
		Error:                    ptr(s.Error),
		NewUser:                  ptr(s.NewUser),
		AvatarSources:            sources,
	}
}

func readAdminCustomizationSettings(ctx context.Context, client *tfev2.Client) (*adminCustomizationSettings, error) {
	env, err := client.API.Admin().CustomizationSettings().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	return adminCustomizationSettingsFromEnvelope(env), nil
}

func patchAdminCustomizationSettings(ctx context.Context, client *tfev2.Client, opts adminCustomizationSettingsUpdate) (*adminCustomizationSettings, error) {
	if opts.isEmpty() {
		return readAdminCustomizationSettings(ctx, client)
	}

	attrs := models.NewAdminCustomizationSettings_attributes()
	attrs.SetSupportEmailAddress(opts.SupportEmail)
	attrs.SetSupportUrlAddress(opts.SupportURLAddress)
	attrs.SetInternalSupportUrl(opts.InternalSupportURL)
	attrs.SetInternalDocumentationUrl(opts.InternalDocumentationURL)
	attrs.SetInternalTutorialsUrl(opts.InternalTutorialsURL)
	attrs.SetLoginHelp(opts.LoginHelp)
	attrs.SetFooter(opts.Footer)
	attrs.SetError(opts.Error)
	attrs.SetNewUser(opts.NewUser)

	if opts.AvatarSources != nil {
		sources := make([]models.AdminCustomizationSettings_attributes_avatarSourcesable, 0, len(opts.AvatarSources))
		for _, s := range opts.AvatarSources {
			src := models.NewAdminCustomizationSettings_attributes_avatarSources()
			src.SetSource(parseEnumPtr[models.AdminCustomizationSettings_attributes_avatarSources_source](
				models.ParseAdminCustomizationSettings_attributes_avatarSources_source, s.Source))
			src.SetEnabled(ptr(s.Enabled))
			sources = append(sources, src)
		}
		attrs.SetAvatarSources(sources)
	}

	data := models.NewAdminCustomizationSettings()
	data.SetTypeEscaped(ptr(models.CUSTOMIZATIONSETTINGS_ADMINCUSTOMIZATIONSETTINGS_TYPE))
	data.SetAttributes(attrs)

	body := models.NewAdminCustomizationSettingsEnvelope()
	body.SetData(data)

	env, err := client.API.Admin().CustomizationSettings().Patch(ctx, body, nil)
	if err != nil {
		return nil, err
	}
	return adminCustomizationSettingsFromEnvelope(env), nil
}

func adminCustomizationSettingsFromEnvelope(env models.AdminCustomizationSettingsEnvelopeable) *adminCustomizationSettings {
	out := &adminCustomizationSettings{}
	if env == nil || env.GetData() == nil {
		return out
	}
	data := env.GetData()
	out.ID = valueOrZero(data.GetId())

	attrs := data.GetAttributes()
	if attrs == nil {
		return out
	}
	out.SupportEmail = valueOrZero(attrs.GetSupportEmailAddress())
	out.SupportURLAddress = valueOrZero(attrs.GetSupportUrlAddress())
	out.InternalSupportURL = valueOrZero(attrs.GetInternalSupportUrl())
	out.InternalDocumentationURL = valueOrZero(attrs.GetInternalDocumentationUrl())
	out.InternalTutorialsURL = valueOrZero(attrs.GetInternalTutorialsUrl())
	out.LoginHelp = valueOrZero(attrs.GetLoginHelp())
	out.Footer = valueOrZero(attrs.GetFooter())
	out.Error = valueOrZero(attrs.GetError())
	out.NewUser = valueOrZero(attrs.GetNewUser())

	for _, s := range attrs.GetAvatarSources() {
		if s == nil || s.GetSource() == nil {
			continue
		}
		out.AvatarSources = append(out.AvatarSources, adminAvatarSource{
			Source:  enumStringOrEmpty(s.GetSource()),
			Enabled: valueOrZero(s.GetEnabled()),
		})
	}
	return out
}

func readAdminGeneralSettings(ctx context.Context, client *tfev2.Client) (*adminGeneralSettings, error) {
	env, err := client.API.Admin().GeneralSettings().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	return adminGeneralSettingsFromEnvelope(env), nil
}

func patchAdminGeneralSettings(ctx context.Context, client *tfev2.Client, enablePublicRegistration *bool) (*adminGeneralSettings, error) {
	if enablePublicRegistration == nil {
		return readAdminGeneralSettings(ctx, client)
	}

	attrs := models.NewAdminGeneralSettings_attributes()
	attrs.SetEnablePublicRegistration(enablePublicRegistration)

	data := models.NewAdminGeneralSettings()
	data.SetTypeEscaped(ptr(models.GENERALSETTINGS_ADMINGENERALSETTINGS_TYPE))
	data.SetAttributes(attrs)

	body := models.NewAdminGeneralSettingsEnvelope()
	body.SetData(data)

	env, err := client.API.Admin().GeneralSettings().Patch(ctx, body, nil)
	if err != nil {
		return nil, err
	}
	return adminGeneralSettingsFromEnvelope(env), nil
}

func adminGeneralSettingsFromEnvelope(env models.AdminGeneralSettingsEnvelopeable) *adminGeneralSettings {
	out := &adminGeneralSettings{}
	if env == nil || env.GetData() == nil {
		return out
	}
	data := env.GetData()
	out.ID = valueOrZero(data.GetId())
	if attrs := data.GetAttributes(); attrs != nil {
		out.EnablePublicRegistration = valueOrZero(attrs.GetEnablePublicRegistration())
	}
	return out
}
