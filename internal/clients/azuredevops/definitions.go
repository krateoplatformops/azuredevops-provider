package azuredevops

import (
	"context"
	"net/http"
	"path"

	"github.com/lucasepe/httplib"
)

type DeleteDefinitionOptions struct {
	Organization string
	Project      string
	DefinitionId string
}

// Delete definition and all associated builds.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/build/definitions/{definitionId}?api-version=7.0
func (c *Client) DeleteDefinition(ctx context.Context, opts DeleteDefinitionOptions) error {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.BaseURL(Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/build/definitions/", opts.DefinitionId),
		Params:  []string{ApiVersionKey, ApiVersionVal},
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
