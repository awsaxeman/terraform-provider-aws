// Code generated by internal/generate/tags/main.go; DO NOT EDIT.

package rds

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// ListTags lists rds service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn *rds.RDS, identifier string) (tftags.KeyValueTags, error) {
	input := &rds.ListTagsForResourceInput{
		ResourceName: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.TagList), nil
}

// []*SERVICE.Tag handling

// Tags returns rds service tags.
func Tags(tags tftags.KeyValueTags) []*rds.Tag {
	result := make([]*rds.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &rds.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from rds service tags.
func KeyValueTags(tags []*rds.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates rds service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *rds.RDS, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &rds.RemoveTagsFromResourceInput{
			ResourceName: aws.String(identifier),
			TagKeys:      aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.RemoveTagsFromResource(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &rds.AddTagsToResourceInput{
			ResourceName: aws.String(identifier),
			Tags:         Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.AddTagsToResource(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
