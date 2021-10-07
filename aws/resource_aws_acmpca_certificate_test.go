package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAwsAcmpcaCertificate_RootCertificate(t *testing.T) {
	resourceName := "aws_acmpca_certificate.test"
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"

	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateConfig_RootCertificate(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttr(resourceName, "certificate_chain", ""),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA512WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/RootCACertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccAwsAcmpcaCertificate_SubordinateCertificate(t *testing.T) {
	resourceName := "aws_acmpca_certificate.test"
	rootCertificateAuthorityResourceName := "aws_acmpca_certificate_authority.root"
	subordinateCertificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"

	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateConfig_SubordinateCertificate(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", rootCertificateAuthorityResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_signing_request", subordinateCertificateAuthorityResourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA512WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/SubordinateCACertificate_PathLen0/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccAwsAcmpcaCertificate_EndEntityCertificate(t *testing.T) {
	resourceName := "aws_acmpca_certificate.test"

	csrDomain := acctest.RandomDomainName()
	csr, _ := acctest.TLSRSAX509CertificateRequestPEM(4096, csrDomain)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateConfig_EndEntityCertificate(domain, acctest.TLSPEMEscapeNewlines(csr)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
					resource.TestCheckResourceAttr(resourceName, "certificate_signing_request", csr),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/EndEntityCertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccAwsAcmpcaCertificate_Validity_EndDate(t *testing.T) {
	resourceName := "aws_acmpca_certificate.test"

	csrDomain := acctest.RandomDomainName()
	csr, _ := acctest.TLSRSAX509CertificateRequestPEM(4096, csrDomain)
	domain := acctest.RandomDomainName()
	later := time.Now().Add(time.Minute * 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateConfig_Validity_EndDate(domain, acctest.TLSPEMEscapeNewlines(csr), later),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
					resource.TestCheckResourceAttr(resourceName, "certificate_signing_request", csr),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", later),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "END_DATE"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/EndEntityCertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccAwsAcmpcaCertificate_Validity_Absolute(t *testing.T) {
	resourceName := "aws_acmpca_certificate.test"

	csrDomain := acctest.RandomDomainName()
	csr, _ := acctest.TLSRSAX509CertificateRequestPEM(4096, csrDomain)
	domain := acctest.RandomDomainName()
	later := time.Now().Add(time.Minute * 10).Unix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateConfig_Validity_Absolute(domain, acctest.TLSPEMEscapeNewlines(csr), later),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
					resource.TestCheckResourceAttr(resourceName, "certificate_signing_request", csr),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", strconv.FormatInt(later, 10)),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "ABSOLUTE"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/EndEntityCertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func testAccCheckAwsAcmpcaCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_certificate" {
			continue
		}

		input := &acmpca.GetCertificateInput{
			CertificateArn:          aws.String(rs.Primary.ID),
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.GetCertificate(input)
		if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
			return nil
		}
		if tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidStateException, "not in the correct state to have issued certificates") {
			// This is returned when checking root certificates and the certificate has not been associated with the certificate authority
			return nil
		}
		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("ACM PCA Certificate (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsAcmpcaCertificateExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn
		input := &acmpca.GetCertificateInput{
			CertificateArn:          aws.String(rs.Primary.ID),
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.GetCertificate(input)

		if err != nil {
			return err
		}

		if output == nil || output.Certificate == nil {
			return fmt.Errorf("ACM PCA Certificate %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsAcmpcaCertificateConfig_RootCertificate(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_partition" "current" {}
`, domain)
}

func testAccAwsAcmpcaCertificateConfig_SubordinateCertificate(domain string) string {
	return acctest.ConfigCompose(
		testAccAcmpcaCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/SubordinateCACertificate_PathLen0/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "SUBORDINATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "sub.%[1]s"
    }
  }
}
`, domain))
}

func testAccAwsAcmpcaCertificateConfig_EndEntityCertificate(domain, csr string) string {
	return acctest.ConfigCompose(
		testAccAcmpcaCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = "%[1]s"
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"

  validity {
    type  = "DAYS"
    value = 1
  }
}
`, csr))
}

func testAccAwsAcmpcaCertificateConfig_Validity_EndDate(domain, csr, expiry string) string {
	return acctest.ConfigCompose(
		testAccAcmpcaCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = "%[1]s"
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"

  validity {
    type  = "END_DATE"
    value = %[2]q
  }
}
`, csr, expiry))
}

func testAccAwsAcmpcaCertificateConfig_Validity_Absolute(domain, csr string, expiry int64) string {
	return acctest.ConfigCompose(
		testAccAcmpcaCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = "%[1]s"
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"

  validity {
    type  = "ABSOLUTE"
    value = %[2]d
  }
}
`, csr, expiry))
}

func testAccAcmpcaCertificateBaseRootCAConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "root" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority_certificate" "root" {
  certificate_authority_arn = aws_acmpca_certificate_authority.root.arn

  certificate       = aws_acmpca_certificate.root.certificate
  certificate_chain = aws_acmpca_certificate.root.certificate_chain
}

resource "aws_acmpca_certificate" "root" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = aws_acmpca_certificate_authority.root.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

data "aws_partition" "current" {}
  `, domain)
}

func TestValidateAcmPcaTemplateArn(t *testing.T) {
	validNames := []string{
		"arn:aws:acm-pca:::template/EndEntityCertificate/V1",                     // lintignore:AWSAT005
		"arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen0/V1",        // lintignore:AWSAT005
		"arn:aws-us-gov:acm-pca:::template/EndEntityCertificate/V1",              // lintignore:AWSAT005
		"arn:aws-us-gov:acm-pca:::template/SubordinateCACertificate_PathLen0/V1", // lintignore:AWSAT005
	}
	for _, v := range validNames {
		_, errors := validateAcmPcaTemplateArn(v, "template_arn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ACM PCA ARN: %q", v, errors)
		}
	}

	invalidNames := []string{
		"arn",
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                       // lintignore:AWSAT005
		"arn:aws:acm-pca:us-west-2::template/SubordinateCACertificate_PathLen0/V1", // lintignore:AWSAT003,AWSAT005
		"arn:aws:acm-pca::123456789012:template/EndEntityCertificate/V1",           // lintignore:AWSAT005
		"arn:aws:acm-pca:::not-a-template/SubordinateCACertificate_PathLen0/V1",    // lintignore:AWSAT005
	}
	for _, v := range invalidNames {
		_, errors := validateAcmPcaTemplateArn(v, "template_arn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ARN", v)
		}
	}
}

func TestExpandAcmpcaValidityValue(t *testing.T) {
	testCases := []struct {
		Type     string
		Value    string
		Expected int64
	}{
		{
			Type:     acmpca.ValidityPeriodTypeEndDate,
			Value:    "2021-02-26T16:04:00Z",
			Expected: 20210226160400,
		},
		{
			Type:     acmpca.ValidityPeriodTypeEndDate,
			Value:    "2021-02-26T16:04:00-08:00",
			Expected: 20210227000400,
		},
		{
			Type:     acmpca.ValidityPeriodTypeAbsolute,
			Value:    "1614385420",
			Expected: 1614385420,
		},
		{
			Type:     acmpca.ValidityPeriodTypeYears,
			Value:    "2",
			Expected: 2,
		},
	}

	for _, testcase := range testCases {
		i, _ := expandAcmpcaValidityValue(testcase.Type, testcase.Value)
		if i != testcase.Expected {
			t.Errorf("%s, %q: expected %d, got %d", testcase.Type, testcase.Value, testcase.Expected, i)
		}
	}

}
