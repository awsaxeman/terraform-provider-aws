// Code generated by internal/generate/tags/main.go; DO NOT EDIT.

package cloudfront

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// ListTags lists cloudfront service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn *cloudfront.CloudFront, identifier string) (tftags.KeyValueTags, error) {
	input := &cloudfront.ListTagsForResourceInput{
		Resource: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.Tags.Items), nil
}

// []*SERVICE.Tag handling

// Tags returns cloudfront service tags.
func Tags(tags tftags.KeyValueTags) []*cloudfront.Tag {
	result := make([]*cloudfront.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &cloudfront.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from cloudfront service tags.
func KeyValueTags(tags []*cloudfront.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates cloudfront service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *cloudfront.CloudFront, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &cloudfront.UntagResourceInput{
			Resource: aws.String(identifier),
			TagKeys:  &cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAWS().Keys())},
		}

		_, err := conn.UntagResource(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &cloudfront.TagResourceInput{
			Resource: aws.String(identifier),
			Tags:     &cloudfront.Tags{Items: Tags(updatedTags.IgnoreAWS())},
		}

		_, err := conn.TagResource(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
