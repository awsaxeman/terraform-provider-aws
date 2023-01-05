package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingAttachment_elb(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_elb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingAttachment_albTargetGroup(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_targetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByTargetGroupARNExists(resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingAttachment_multipleELBs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_attachment.test.0"
	resource11Name := "aws_autoscaling_attachment.test.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy,
		Steps: []resource.TestStep{
			// Create all the ELBs first.
			{
				Config: testAccAttachmentConfig_elbBase(rName, 11),
			},
			{
				Config: testAccAttachmentConfig_multipleELBs(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(resource1Name),
					testAccCheckAttachmentByLoadBalancerNameExists(resource11Name),
				),
			},
			{
				Config: testAccAttachmentConfig_elbBase(rName, 11),
			},
		},
	})
}

func TestAccAutoScalingAttachment_multipleALBTargetGroups(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_attachment.test.0"
	resource11Name := "aws_autoscaling_attachment.test.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy,
		Steps: []resource.TestStep{
			// Create all the target groups first.
			{
				Config: testAccAttachmentConfig_targetGroupBase(rName, 11),
			},
			{
				Config: testAccAttachmentConfig_multipleTargetGroups(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByTargetGroupARNExists(resource1Name),
					testAccCheckAttachmentByTargetGroupARNExists(resource11Name),
				),
			},
			{
				Config: testAccAttachmentConfig_targetGroupBase(rName, 11),
			},
		},
	})
}

func testAccCheckAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_attachment" {
			continue
		}

		var err error

		if targetGroupARN := rs.Primary.Attributes["lb_target_group_arn"]; targetGroupARN == "" {
			targetGroupARN = rs.Primary.Attributes["alb_target_group_arn"]

			err = tfautoscaling.FindAttachmentByTargetGroupARN(conn, rs.Primary.Attributes["autoscaling_group_name"], targetGroupARN)
		} else {
			err = tfautoscaling.FindAttachmentByLoadBalancerName(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["elb"])
		}

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Auto Scaling Group Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAttachmentByLoadBalancerNameExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

		return tfautoscaling.FindAttachmentByLoadBalancerName(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["elb"])
	}
}

func testAccCheckAttachmentByTargetGroupARNExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

		targetGroupARN := rs.Primary.Attributes["lb_target_group_arn"]
		if targetGroupARN == "" {
			targetGroupARN = rs.Primary.Attributes["alb_target_group_arn"]
		}

		return tfautoscaling.FindAttachmentByTargetGroupARN(conn, rs.Primary.Attributes["autoscaling_group_name"], targetGroupARN)
	}
}

func testAccAttachmentConfig_elbBase(rName string, elbCount int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_elb" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name               = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = data.aws_availability_zones.available.names
  max_size                  = 1
  min_size                  = 0
  desired_capacity          = 0
  health_check_grace_period = 300
  force_delete              = true
  name                      = %[1]q
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [load_balancers]
  }
}
`, rName, elbCount))
}

func testAccAttachmentConfig_targetGroupBase(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name     = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier       = aws_subnet.test[*].id
  max_size                  = 1
  min_size                  = 0
  desired_capacity          = 0
  health_check_grace_period = 300
  force_delete              = true
  name                      = %[1]q
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [target_group_arns]
  }
}
`, rName, targetGroupCount))
}

func testAccAttachmentConfig_elb(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_elbBase(rName, 1), `
resource "aws_autoscaling_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  elb                    = aws_elb.test[0].id
}
`)
}

func testAccAttachmentConfig_multipleELBs(rName string, n int) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_elbBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id
  elb                    = aws_elb.test[count.index].id
}
`, n))
}

func testAccAttachmentConfig_targetGroup(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_targetGroupBase(rName, 1), `
resource "aws_autoscaling_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  lb_target_group_arn    = aws_lb_target_group.test[0].arn
}
`)
}

func testAccAttachmentConfig_multipleTargetGroups(rName string, n int) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_targetGroupBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id
  lb_target_group_arn    = aws_lb_target_group.test[0].arn
}
`, n))
}
