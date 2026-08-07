// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/githubappinstallations"
)

func fetchGithubAppInstallationByNameOrGHID(ctx context.Context, config ConfiguredClient, name string, installationID int) (id string, instID int, instName string, err error) {
	if name == "" && installationID == 0 {
		return "", 0, "", fmt.Errorf("invalid parameters, either name or installation id must have a value")
	}

	pageSize := int32(100)
	queryParams := &githubappinstallations.GithubAppInstallationsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}
	if name != "" {
		queryParams.Filtername = &name
	}
	if installationID != 0 {
		idStr := strconv.Itoa(installationID)
		queryParams.Filterinstallation_id = &idStr
	}

	for {
		list, listErr := config.ClientV2.API.GithubAppInstallations().Get(ctx, withQueryParams(queryParams))
		if listErr != nil {
			return "", 0, "", fmt.Errorf("error retrieving Github App Installations: %w", listErr)
		}

		for _, item := range list.GetData() {
			attrs := item.GetAttributes()
			if attrs == nil {
				continue
			}
			iName := valueOrZero(attrs.GetName())
			iID := 0
			if attrs.GetInstallationId() != nil {
				iID = int(*attrs.GetInstallationId())
			}

			match := false
			switch {
			case name != "" && installationID != 0:
				match = iName == name && iID == installationID
			case name != "":
				match = iName == name
			case installationID != 0:
				match = iID == installationID
			}

			if match {
				return valueOrZero(item.GetId()), iID, iName, nil
			}
		}

		nextPage := nextPageFromMeta(list.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return "", 0, "", tfeV2.ErrNotFound
}
