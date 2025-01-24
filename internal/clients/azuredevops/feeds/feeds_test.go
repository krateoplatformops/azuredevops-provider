//go:build integration
// +build integration

package feeds

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/davecgh/go-spew/spew"
// 	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
// 	"github.com/lucasepe/dotenv"
// 	"github.com/lucasepe/httplib"
// )

// func TestFind(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Find(context.TODO(), cli, FindOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_NAME"),
// 		FeedName:     "test-feed-3",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if res == nil {
// 		return
// 	}

// 	fmt.Printf("Name: %s\n", res.Name)
// 	fmt.Printf("Id: %s\n", *res.Id)
// 	fmt.Printf("Url: %s\n", *res.Url)

// }
// func TestList(t *testing.T) {
// 	cli := createAzureDevopsClient()
// 	//cli.SetVerbose(true)

// 	all, err := List(context.TODO(), cli, ListOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_NAME"),
// 		IncludeUrls:  true,
// 	})
// 	if err != nil {
// 		if httplib.IsNotFoundError(err) {
// 			return
// 		}
// 		t.Fatal(err)
// 	}

// 	for _, el := range all {
// 		fmt.Printf("Name: %s\n", el.Name)
// 		fmt.Printf("Id: %s\n", *el.Id)
// 		fmt.Printf("Url: %s\n", *el.Url)
// 		fmt.Println("---------------")
// 	}

// }

// func TestGet(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Get(context.TODO(), cli, GetOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_NAME"),
// 		FeedId:       os.Getenv("FEED_ID"),
// 	})
// 	if err != nil {
// 		if httplib.IsNotFoundError(err) {
// 			t.Logf("NOT found\n")
// 			return
// 		}
// 		t.Fatal(err)
// 	}

// 	if res != nil {
// 		spew.Dump(res)
// 	}
// }

// func TestCreate(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Create(context.TODO(), cli, CreateOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_NAME"),

// 		Feed: &Feed{
// 			Name: "test-feed-3",
// 		},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if res == nil {
// 		return
// 	}
// 	spew.Dump(res)

// }

// func TestUpdate(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Update(context.TODO(), cli, UpdateOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_NAME"),
// 		FeedId:       os.Getenv("FEED_ID"),
// 		FeedUpdate: &FeedUpdate{
// 			//Name:          "test-feed-1",
// 			BadgesEnabled: true,
// 		},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	spew.Dump(res)
// }

// func createAzureDevopsClient() *azuredevops.Client {
// 	env, _ := dotenv.FromFile("../../../../.env")
// 	dotenv.PutInEnv(env, false)

// 	return azuredevops.NewClient(azuredevops.ClientOptions{
// 		Verbose: false,
// 		Token:   os.Getenv("TOKEN"),
// 	})
// }
