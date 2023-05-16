package queues

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/dotenv"
)

func TestDelete(t *testing.T) {
	cli := createAzureDevopsClient()

	err := Delete(context.TODO(), cli, DeleteOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		QueueId:      2085,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Get(context.TODO(), cli, GetOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		QueueId:      2085,
	})
	if err != nil {
		t.Fatal(err)
	}

	if res == nil {
		return
	}

	fmt.Printf("Id: %d\n", *res.Id)
	fmt.Printf("Name: %s\n", res.Name)
	fmt.Printf("Pool: (id=%d, name=%s)\n", *res.Pool.Id, res.Pool.Name)
}

func TestFindByNames(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := FindByNames(context.TODO(), cli, FindByNamesOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		QueueNames:   []string{"queue-1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, el := range res {
		fmt.Printf("Id: %d\n", *el.Id)
		fmt.Printf("Name: %s\n", el.Name)
		fmt.Printf("Pool: (id=%d, name=%s)\n", *el.Pool.Id, el.Pool.Name)
		fmt.Println("----------------------------")
	}
}

func TestAdd(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Add(context.TODO(), cli, AddOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		Queue: &TaskAgentQueue{
			Name: "queue-1",
			Pool: &TaskAgentPoolReference{
				Id: helpers.IntPtr(10),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if res == nil {
		return
	}
	spew.Dump(res)

}

func createAzureDevopsClient() *azuredevops.Client {
	env, _ := dotenv.FromFile("../../../../.env")
	dotenv.PutInEnv(env, false)

	return azuredevops.NewClient(azuredevops.ClientOptions{
		Verbose: false,
		Token:   os.Getenv("TOKEN"),
	})
}
