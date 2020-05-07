package photos

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/Dimitriy14/staff-manager/util"

	awservices "github.com/Dimitriy14/staff-manager/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	serverSideEncryption = "AES256"
	acl                  = "public-read"
)

type Uploader interface {
	Upload(ctx context.Context, fileExt string, content []byte) (url string, err error)
}

func NewPhotosUploader(s3 *awservices.S3Manager, storageUrl string, bucketName string) *uploaderImpl {
	return &uploaderImpl{
		s3:         s3,
		storageUrl: storageUrl,
		bucketName: bucketName,
	}
}

type uploaderImpl struct {
	s3         *awservices.S3Manager
	storageUrl string
	bucketName string
}

func (u *uploaderImpl) Upload(ctx context.Context, fileExt string, content []byte) (string, error) {
	var (
		user     = util.GetUserAccessFromCtx(ctx)
		fileName = fmt.Sprintf("staff/%s%s", user.UserID, fileExt)
	)

	_, err := u.s3.Uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(u.bucketName),
		Key:                  aws.String(fileName),
		ACL:                  aws.String(acl),
		Body:                 bytes.NewReader(content),
		ContentType:          aws.String(http.DetectContentType(content)),
		ServerSideEncryption: aws.String(serverSideEncryption),
	})

	return fmt.Sprintf("%s/%s", u.storageUrl, fileName), err
}
