package azuredevops

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type DeleteDefinitionOptions struct {
	Organization string
	Project      string
	DefinitionId string
}

func getDefinitionsAPIVersion(cli *Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.Definitions
		if apiVersion != nil {
			if strings.EqualFold(*apiVersion, "none") {
				apiVersionParams = nil
				isNone = true
			} else {
				apiVersionParams = []string{ApiVersionKey, helpers.String(apiVersion)}
			}
		}
	}
	return apiVersionParams, isNone
}

// Delete definition and all associated builds.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/build/definitions/{definitionId}?api-version=7.0
func (c *Client) DeleteDefinition(ctx context.Context, opts DeleteDefinitionOptions) error {
	apiVersionParams, isNone := getDefinitionsAPIVersion(c)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{ApiVersionKey, ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.BaseURL(Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/build/definitions/", opts.DefinitionId),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	return httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod: c.authMethod,
		Verbose:    c.verbose,
		Validators: []httplib.HandleResponseFunc{
			httplib.CheckStatus(http.StatusOK, http.StatusNoContent),
		},
	})
}
