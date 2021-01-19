// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"

// AddRepositoryContext adds the given repository with the baseurl to the list of repositories.
// If the last repository is the same baseUrl and type it will not be added.
func AddRepositoryContext(repoCtx []cdv2.RepositoryContext, rType, baseUrl string) []cdv2.RepositoryContext {
	if len(baseUrl) == 0 {
		return repoCtx
	}
	if len(repoCtx) == 0 {
		return []cdv2.RepositoryContext{{
			Type:    rType,
			BaseURL: baseUrl,
		}}
	}
	effectiveCtx := repoCtx[len(repoCtx)-1]
	if effectiveCtx.Type == rType && effectiveCtx.BaseURL == baseUrl {
		return repoCtx
	}
	return append(repoCtx, cdv2.RepositoryContext{
		Type:    rType,
		BaseURL: baseUrl,
	})
}
