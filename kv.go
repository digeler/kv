package main

/*
 * You need to set four environment variables before using the app:
 * AZURE_TENANT_ID: Your Azure tenant ID
 * AZURE_CLIENT_ID: Your Azure client ID. This will be an app ID from your AAD.
 * AZURE_CLIENT_SECRET: The secret for the client ID above.
 * KVAULT: The name of your vault (just the name, not the full URL/path)
 *
 * Usage
 * List the secrets currently in the vault (not the values though):
 * kv-pass
 *
 * Get the value for a secret in the vault:
 * kv-pass YOUR_SECRETS_NAME
 *
 * Add or Update a secret in the vault:
 * kv-pass -edit YOUR_NEW_VALUE YOUR_SECRETS_NAME
 *
 * Delete a secret in the vault:
 * kv-pass -delete YOUR_SECRETS_NAME
 */

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
)

var (
	interVal  = flag.Int("interval", 30, "set value for request")
	setDebug  = flag.Bool("debug", false, "debug")
	vaultName string
)

func main() {
	flag.Parse()

	if os.Getenv("AZURE_TENANT_ID") == "" || os.Getenv("AZURE_CLIENT_ID") == "" || os.Getenv("AZURE_CLIENT_SECRET") == "" || os.Getenv("KVAULT") == "" || os.Getenv("SECNAME") == "" {
		fmt.Println(os.Getenv("AZURE_TENANT_ID"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), os.Getenv("KVAULT"), os.Getenv("SECNAME"))
		fmt.Println("env vars not set, exiting...")
		os.Exit(1)
	}
	vaultName = os.Getenv("KVAULT")

	authorizer, err := kvauth.NewAuthorizerFromEnvironment()
	if err != nil {
		fmt.Printf("unable to create vault authorizer: %v\n", err)
		os.Exit(1)
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer

	if *setDebug {
		basicClient.RequestInspector = logRequest()
		basicClient.ResponseInspector = logResponse()
	}

	getSecret(basicClient, os.Getenv("SECNAME"))

}

func getSecret(basicClient keyvault.BaseClient, secname string) {
	dur := *interVal
	for i := 0; i < dur; i++ {

		time.Sleep(5 * time.Second)
		secretResp, err := basicClient.GetSecret(context.Background(), "https://"+vaultName+".vault.azure.net", secname, "")
		if err != nil {
			fmt.Printf("unable to get value for secret: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(*secretResp.Value)
	}
}

func logRequest() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err != nil {
				log.Println(err)
			}
			dump, _ := httputil.DumpRequestOut(r, true)
			log.Println(string(dump))
			return r, err
		})
	}
}

func logResponse() autorest.RespondDecorator {
	return func(p autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(r *http.Response) error {
			err := p.Respond(r)
			if err != nil {
				log.Println(err)
			}
			dump, _ := httputil.DumpResponse(r, true)
			log.Println(string(dump))
			return err
		})
	}
}
