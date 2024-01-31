package sdkv2provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/consts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func init() {
	resource.AddTestSweepers("cloudflare_hyperdrive_config", &resource.Sweeper{
		Name: "cloudflare_hyperdrive_config",
		F:    testSweepCloudflareHyperdriveConfigSweeper,
	})
}

func testSweepCloudflareHyperdriveConfigSweeper(r string) error {
	ctx := context.Background()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	client, clientErr := sharedClient()
	if clientErr != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to create Cloudflare client: %s", clientErr))
	}

	resp, err := client.ListHyperdriveConfigs(context.Background(), cloudflare.AccountIdentifier(accountID), cloudflare.ListHyperdriveConfigParams{})
	if err != nil {
		return err
	}

	for _, q := range resp {
		err := client.DeleteHyperdriveConfig(ctx, cloudflare.AccountIdentifier(accountID), q.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func testHyperdriveConfigConfig(
	rnd, accountId, name string, password string, origin cloudflare.HyperdriveConfigOrigin, caching cloudflare.HyperdriveConfigCaching,
) string {
	return fmt.Sprintf(`
		resource "cloudflare_hyperdrive_config" "%[1]s" {
			account_id = "%[2]s"
			name       = "%[3]s"
			password   = "%[4]s"
			origin     = {
				database = "%[5]s"
				host     = "%[6]s"
				port     = "%[7]s"
				scheme   = "%[8]s"
				user     = "%[9]s"
			}
			caching = {
				disabled               = %[10]s
				max_age                = %[11]s
				stale_while_revalidate = %[12]s
			}		  
		}`,
		rnd, accountId, name, password, origin.Database, origin.Host, fmt.Sprintf("%d", origin.Port), origin.Scheme, origin.User, fmt.Sprintf("%t", *caching.Disabled), fmt.Sprintf("%d", caching.MaxAge), fmt.Sprintf("%d", caching.StaleWhileRevalidate),
	)
}

func TestAccCloudflareHyperdriveConfig_Basic(t *testing.T) {
	t.Parallel()

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	rnd := generateRandomResourceName()

	resourceName := "cloudflare_hyperdrive_config." + rnd

	var origin = cloudflare.HyperdriveConfigOrigin{
		Database: "database",
		Host:     "host",
		Port:     1234,
		Scheme:   "scheme",
		User:     "user",
	}

	var disabled = false

	var caching = cloudflare.HyperdriveConfigCaching{
		Disabled:             &disabled,
		MaxAge:               1,
		StaleWhileRevalidate: 1,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCloudflareHyperdriveConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testHyperdriveConfigConfig(
					rnd,
					accountID,
					rnd,
					"password",
					origin,
					caching,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rnd),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttr(resourceName, "password", "password"),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.database", "database"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.host", "host"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.port", "port"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.scheme", "scheme"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.user", "user"),
					resource.TestCheckResourceAttr(resourceName, "caching.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "caching.0.disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "caching.0.max_age", "1"),
					resource.TestCheckResourceAttr(resourceName, "caching.0.stale_while_revalidate", "1"),
				),
			},
			{
				Config: testHyperdriveConfigConfig(
					rnd,
					accountID,
					rnd+"-updated",
					"updated-password",
					origin,
					caching,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rnd+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttr(resourceName, "password", "updated-password"),
				),
			},
			{
				ImportState:         true,
				ImportStateVerify:   true,
				ResourceName:        resourceName,
				ImportStateIdPrefix: fmt.Sprintf("%s/", accountID),
			},
		},
	})
}

func testAccCloudflareHyperdriveConfigDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_hyperdrive_config" {
			continue
		}

		accountID := rs.Primary.Attributes[consts.AccountIDSchemaKey]
		resp, err := client.ListHyperdriveConfigs(context.Background(), cloudflare.AccountIdentifier(accountID), cloudflare.ListHyperdriveConfigParams{})
		if err != nil {
			return err
		}

		for _, n := range resp {
			if n.ID == rs.Primary.ID {
				return fmt.Errorf("hyperdrive config still exists but should not")
			}
		}
	}

	return nil
}
