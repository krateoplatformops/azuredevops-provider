//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestGetOperation(t *testing.T) {
	cli := setupClient()

	res, err := cli.GetOperation(context.TODO(), GetOperationOpts{
		Organization: os.Getenv("ORG"),
		OperationId:  "ebf2b78f-1d81-418c-8559-3c867660776a",
	})
	if err != nil {
		if IsNotFound(err) {
			fmt.Println("OP NOT FOUND")
		}

		//var apierr *APIError
		//if errors.As(err, &apierr) {
		//	fmt.Println(apierr.Error())
		//}

	}

	fmt.Printf("%v\n", res)
}
