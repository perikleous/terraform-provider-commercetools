package commercetools

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestAPIExtensionUnmarshallExtensionDestination(t *testing.T) {
	rawDestination := map[string]interface{}{
		"type":          "AWSLambda",
		"arn":           "arn:aws:lambda:eu-west-1:111111111:function:api_extensions",
		"access_key":    "ABCSDF123123123",
		"access_secret": "****abc/",
	}

	resourceDataMap := map[string]interface{}{
		"id":             "2845b936-e407-4f29-957b-f8deb0fcba97",
		"version":        1,
		"createdAt":      "2018-12-03T16:13:03.969Z",
		"lastModifiedAt": "2018-12-04T09:06:59.491Z",
		"destination":    []interface{}{rawDestination},
		"triggers": []interface{}{
			map[string]interface{}{
				"triggers": []interface{}{"Create", "Update"},
			},
		},
		"timeout_in_ms": 1,
		"key":           "create-order",
	}

	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, resourceDataMap)
	destination, _ := unmarshallExtensionDestination(d)
	lambdaDestination, ok := destination.(platform.AWSLambdaDestination)

	assert.True(t, ok)
	assert.Equal(t, lambdaDestination.Arn, "arn:aws:lambda:eu-west-1:111111111:function:api_extensions")
	assert.Equal(t, lambdaDestination.AccessKey, "ABCSDF123123123")
	assert.Equal(t, lambdaDestination.AccessSecret, "****abc/")
}

func TestAPIExtensionUnmarshallExtensionDestinationAuthentication(t *testing.T) {
	var input = map[string]interface{}{
		"authorization_header": "12345",
		"azure_authentication": "AzureKey",
	}

	auth, err := unmarshallExtensionDestinationAuthentication(input)
	assert.Nil(t, auth)
	assert.NotNil(t, err)

	input = map[string]interface{}{
		"authorization_header": "12345",
	}

	auth, err = unmarshallExtensionDestinationAuthentication(input)
	httpAuth, ok := auth.(*platform.AuthorizationHeaderAuthentication)
	assert.True(t, ok)
	assert.Equal(t, "12345", httpAuth.HeaderValue)
	assert.NotNil(t, auth)
	assert.Nil(t, err)
}

func TestUnmarshallExtensionTriggers(t *testing.T) {
	resourceDataMap := map[string]interface{}{
		"id":             "2845b936-e407-4f29-957b-f8deb0fcba97",
		"version":        1,
		"createdAt":      "2018-12-03T16:13:03.969Z",
		"lastModifiedAt": "2018-12-04T09:06:59.491Z",
		"trigger": []interface{}{
			map[string]interface{}{
				"resource_type_id": "cart",
				"actions":          []interface{}{"Create", "Update"},
			},
		},
		"timeout_in_ms": 1,
		"key":           "create-order",
	}

	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, resourceDataMap)
	triggers := unmarshallExtensionTriggers(d)

	assert.Len(t, triggers, 1)
	assert.Equal(t, triggers[0].ResourceTypeId, platform.ExtensionResourceTypeIdCart)
	assert.Len(t, triggers[0].Actions, 2)
}

func TestAccAPIExtension_basic(t *testing.T) {
	fmt.Println("RUN IT")
	name := fmt.Sprintf("extension_%s", acctest.RandString(5))
	timeoutInMs := acctest.RandIntRange(200, 1800)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAPIExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIExtensionConfig(name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists("ext"),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "key", name),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "trigger.0.actions.#", "1"),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "trigger.0.actions.0", "Create"),
				),
			},
			{
				Config: testAccAPIExtensionUpdate(name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists("ext"),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "key", name),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "trigger.0.actions.#", "2"),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "trigger.0.actions.0", "Create"),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "trigger.0.actions.1", "Update"),
				),
			},
		},
	})
}

func testAccAPIExtensionConfig(name string, timeoutInMs int) string {
	return fmt.Sprintf(`
resource "commercetools_api_extension" "ext" {
  key = "%s"
  timeout_in_ms = %d

  destination {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions = ["Create"]
  }
}
`, name, timeoutInMs)
}

func testAccAPIExtensionUpdate(name string, timeoutInMs int) string {
	return fmt.Sprintf(`
resource "commercetools_api_extension" "ext" {
  key = "%s"
  timeout_in_ms = %d

  destination {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions = ["Create", "Update"]
  }
}
`, name, timeoutInMs)
}

func testAccAPIExtensionExists(n string) resource.TestCheckFunc {
	name := fmt.Sprintf("commercetools_api_extension.%s", n)
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Extension ID is set")
		}
		client := getClient(testAccProvider.Meta())
		result, err := client.Extensions().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("Extension not found")
		}

		return nil
	}
}

func testAccCheckAPIExtensionDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_api_extension" {
			continue
		}
		response, err := client.Extensions().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("api extension (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
