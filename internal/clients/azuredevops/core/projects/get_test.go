package projects

import (
	"context"
	"fmt"
	"testing"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/dotenv"
)

func TestGet(t *testing.T) {
	env, _ := dotenv.FromFile("../../../../../.env")

	cli := azuredevops.NewClient(azuredevops.ClientOpts{
		Verbose: false,
		BaseURL: env["BASE_URL"],
		Token:   env["TOKEN"],
	})

	_, err := Get(context.TODO(), cli, GetProjectOpts{
		Organization: env["ORG"],
		ProjectId:    "c1bf241f-d676-467a-9496-1bf1b910414a",
	})
	if err != nil {
		if IsNotFound(err) {
			fmt.Println("NOT FOUNDOK")
		}

		//var apierr *APIError
		//if errors.As(err, &apierr) {
		//	fmt.Println(apierr.Error())
		//}

	}

	//fmt.Printf("%v\n", res)
}
