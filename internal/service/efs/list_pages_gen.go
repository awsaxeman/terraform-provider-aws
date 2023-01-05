// Code generated by "internal/generate/listpages/main.go -ListOps=DescribeMountTargets -InputPaginator=Marker -OutputPaginator=NextMarker"; DO NOT EDIT.

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
)

func describeMountTargetsPages(conn *efs.EFS, input *efs.DescribeMountTargetsInput, fn func(*efs.DescribeMountTargetsOutput, bool) bool) error {
	return describeMountTargetsPagesWithContext(context.Background(), conn, input, fn)
}

func describeMountTargetsPagesWithContext(ctx context.Context, conn *efs.EFS, input *efs.DescribeMountTargetsInput, fn func(*efs.DescribeMountTargetsOutput, bool) bool) error {
	for {
		output, err := conn.DescribeMountTargetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.NextMarker
	}
	return nil
}
