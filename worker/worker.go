package worker

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/kataras/iris/v12"
	"github.com/lampnick/doctron/conf"
	"github.com/lampnick/doctron/converter"
	"github.com/lampnick/doctron/converter/doctron_core"
	"github.com/lampnick/doctron/uploader"
)

const (
	BucketName = "fir-949c0.appspot.com"
)

type ClientUploader struct {
	Cl         *storage.Client
	ProjectID  string
	BucketName string
	UploadPath string
}

var Pool *tunny.Pool

var ErrNoNeedToUpload = errors.New("no need to upload")
var (
	ErrWrongDoctronParam = errors.New("wrong doctron params given")
)

func log(ctx iris.Context, format string, args ...interface{}) {
	ctx.Application().Logger().Infof(format, args...)
}

type DoctronOutputDTO struct {
	Buf []byte
	Url string
	Err error
}

func DoctronHandler(params interface{}) interface{} {
	doctronOutputDTO := DoctronOutputDTO{}
	doctronConfig, ok := params.(converter.DoctronConfig)
	if !ok {
		doctronOutputDTO.Err = ErrWrongDoctronParam
		return doctronOutputDTO
	}

	doctron := doctron_core.NewDoctron(doctronConfig.Ctx, doctronConfig.DoctronType, doctronConfig.ConvertConfig)

	convertBytes, err := doctron.Convert()

	out, err := os.Create("test.pdf")
	if err != nil {
		return err
	}
	defer out.Close()
	file := bytes.NewReader(convertBytes)

	_, err = io.Copy(out, file)
	if err != nil {
		return err
	}

	log(doctronConfig.IrisCtx, "uuid:[%s],doctron.Convert Elapsed [%s],url:[%s]", doctronConfig.TraceId, doctron.GetConvertElapsed(), doctronConfig.IrisCtx.Request().RequestURI)
	if err != nil {
		doctronOutputDTO.Err = err
		return doctronOutputDTO
	}

	doctronOutputDTO.Buf = convertBytes
	if doctronConfig.UploadKey == "" {
		storageClient, err := storage.NewClient(context.Background(), option.WithCredentialsFile("serviceAccount.json"))
		if err != nil {
			panic(fmt.Sprintf("google storage client cannot be initiated, err: %v", err))
		}

		uploader := &ClientUploader{
			Cl:         storageClient,
			BucketName: BucketName,
			UploadPath: "test-files/",
		}

		file, err := os.Open("test.pdf")
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		err = uploader.UploadFile(file, "test.pdf")
		if err != nil {
			return err
		}

		doctronOutputDTO.Err = ErrNoNeedToUpload
		return doctronOutputDTO
	} else {
		doctronUploader := uploader.NewDoctronUploader(
			doctronConfig.Ctx,
			conf.LoadedConfig.Doctron.Uploader,
			uploader.UploadConfig{Key: doctronConfig.UploadKey, Stream: convertBytes},
		)
		uploadUrl, err := doctronUploader.Upload()
		log(doctronConfig.IrisCtx, "uuid:[%s],doctron.Upload Elapsed [%s],url:[%s]", doctronConfig.TraceId, doctronUploader.GetUploadElapsed(), doctronConfig.IrisCtx.Request().RequestURI)
		if err != nil {
			doctronOutputDTO.Err = err
			return doctronOutputDTO
		}
		doctronOutputDTO.Url = uploadUrl
		return doctronOutputDTO
	}
}

func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.Cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
