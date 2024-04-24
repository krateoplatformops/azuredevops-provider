package securefiles

import (
	"context"
	"net/http"
	"path"
	"reflect"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type SecureFileResource struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	CreatedBy  interface{} `json:"createdBy"`
	CreatedOn  string      `json:"createdOn"`
	ModifiedBy interface{} `json:"modifiedBy"`
	ModifiedOn string      `json:"modifiedOn"`
}

type GetOptions struct {
	Organization string
	Project      string
	SecretFileId string
}

// Get retrieves information about a secureFile.
// GET https://dev.azure.com/{{organization}}/{{project}}/_apis/distributedtask/securefiles/{{secureid}}?api-version=7.0-preview.1
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*SecureFileResource, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/distributedtask/securefiles", opts.SecretFileId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &SecureFileResource{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if err != nil {
		return nil, err
	}
	if reflect.DeepEqual(*val, SecureFileResource{}) {
		return nil, err
	}

	return val, nil
}

type ListOptions struct {
	Organization      string
	Project           string
	ContinuationToken *string
}

type ListResponse struct {
	Count             int                  `json:"count"`
	Value             []SecureFileResource `json:"value"`
	ContinuationToken *string              `json:"continuationToken"`
}

// GET https://dev.azure.com/{{organization}}/{{project}}/_apis/distributedtask/securefiles?api-version={{api_version}}
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListResponse, error) {
	queryparams := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	if opts.ContinuationToken != nil {
		queryparams = append(queryparams, "continuationToken", helpers.String(opts.ContinuationToken))
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/distributedtask/securefiles"),
		Params:  queryparams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &ListResponse{
		Value: []SecureFileResource{},
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if err != nil {
		return nil, err
	}
	if reflect.DeepEqual(*val, SecureFileResource{}) {
		return nil, err
	}

	return val, nil
}

type FindOptions struct {
	ListOptions
	SecureFileName string
}

func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*SecureFileResource, error) {
	var continuationToken *string
	opts.ListOptions.ContinuationToken = continuationToken
	for {
		secureFiles, err := List(ctx, cli, opts.ListOptions)
		if err != nil {
			return nil, err
		}
		for _, secureFile := range secureFiles.Value {
			if secureFile.Name == opts.SecureFileName {
				return &secureFile, nil
			}
		}
		if secureFiles.ContinuationToken == nil {
			break
		}
		opts.ListOptions.ContinuationToken = secureFiles.ContinuationToken
	}
	return nil, nil
}

type DeleteOptions struct {
	Project      string
	Organization string
	SecureFileId string
}

// DELETE https://dev.azure.com/{{organization}}/{{project}}/_apis/distributedtask/securefiles/{{secureid}}?api-version={{api_version}}
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/distributedtask/securefiles", opts.SecureFileId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"},
	}).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusNoContent),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
