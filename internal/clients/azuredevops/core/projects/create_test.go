package projects

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/dotenv"
)

func TestCreate(t *testing.T) {
	env, _ := dotenv.FromFile("../../../../../.env")

	cli := azuredevops.NewClient(azuredevops.ClientOpts{
		Verbose: true,
		BaseURL: env["BASE_URL"],
		Token:   env["TOKEN"],
	})

	res, err := Create(context.TODO(), cli, CreateProjectOpts{
		Organization: env["ORG"],
		TeamProject: &TeamProject{
			Name:        helpers.StringPtr("ProjectCreatedViaRestApi"),
			Description: helpers.StringPtr("Test Project Created by Go!"),
			Capabilities: &map[string]map[string]string{
				"versioncontrol": {
					"sourceControlType": "Git",
				},
				"processTemplate": {
					"templateTypeId": "6b724908-ef14-45cf-84f8-768b5384da45",
				},
			},
		},
	})
	if err != nil {
		var apierr *APIError
		if errors.As(err, &apierr) {
			fmt.Println(apierr.Error())
		}
	}

	fmt.Printf("%v\n", res)
}
