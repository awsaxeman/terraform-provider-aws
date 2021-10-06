package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// add sweeper to delete known test servicecat products
func init() {
	resource.AddTestSweepers("aws_servicecatalog_product", &resource.Sweeper{
		Name: "aws_servicecatalog_product",
		Dependencies: []string{
			"aws_servicecatalog_provisioning_artifact",
		},
		F: testSweepServiceCatalogProducts,
	})
}

func testSweepServiceCatalogProducts(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).ServiceCatalogConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.SearchProductsAsAdminInput{}

	err = conn.SearchProductsAsAdminPages(input, func(page *servicecatalog.SearchProductsAsAdminOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd == nil || pvd.ProductViewSummary == nil {
				continue
			}

			id := aws.StringValue(pvd.ProductViewSummary.ProductId)

			r := ResourceProduct()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Products for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Products for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Products sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "distributor", "distributör"),
					resource.TestCheckResourceAttr(resourceName, "has_default_path", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", "ägare"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.disable_template_validation", "true"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_url"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.StatusCreated),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_email", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "support_url", domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, rName, rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_update(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "ny beskrivning", "ny supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "ny beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "ny supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_updateTags(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_updateTags(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "natural"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_physicalID(rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_physical_id"),
					acctest.MatchResourceAttrRegionalARN(
						resourceName,
						"provisioning_artifact_parameters.0.template_physical_id",
						"cloudformation",
						regexp.MustCompile(fmt.Sprintf(`stack/%s/.*`, rName)),
					),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProductDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product" {
			continue
		}

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeProductAsAdmin(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Product (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProductExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProductAsAdmin(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProductConfigTemplateURLBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}
`, rName)
}

func testAccAWSServiceCatalogProductConfig_basic(rName, description, supportDescription, domain, email string) string {
	return acctest.ConfigCompose(testAccAWSServiceCatalogProductConfigTemplateURLBase(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = %[2]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[3]q
  support_email       = %[5]q
  support_url         = %[4]q

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, description, supportDescription, domain, email))
}

func testAccAWSServiceCatalogProductConfig_updateTags(rName, description, supportDescription, domain, email string) string {
	return acctest.ConfigCompose(testAccAWSServiceCatalogProductConfigTemplateURLBase(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = %[2]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[3]q
  support_email       = %[5]q
  support_url         = %[4]q

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Yak         = %[1]q
    Environment = "natural"
  }
}
`, rName, description, supportDescription, domain, email))
}

func testAccAWSServiceCatalogProductConfig_physicalID(rName, domain, email string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = %[3]q
  support_url         = %[2]q

  provisioning_artifact_parameters {
    description          = "artefaktbeskrivning"
    name                 = %[1]q
    template_physical_id = aws_cloudformation_stack.test.id
    type                 = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, domain, email)
}
