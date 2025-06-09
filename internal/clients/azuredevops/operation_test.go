//go:build integration
// +build integration

package azuredevops

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"strings"
// )

// func ExampleClient_GetOperation() {
// 	cli := createAzureDevopsClient()

// 	res, err := cli.GetOperation(context.TODO(), GetOperationOpts{
// 		Organization: os.Getenv("ORG"),
// 		OperationId:  os.Getenv("OPERATION_ID"),
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	ok := strings.HasSuffix(res.Url, fmt.Sprintf("/_apis/operations/%s", os.Getenv("OPERATION_ID")))
// 	fmt.Printf("%t", ok)

// 	// Output:
// 	// true
// }
