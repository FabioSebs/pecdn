package main

import (
	"net/http"
	"strings"

	"github.com/FabioSebs/infrastruc/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
)

var CFG = config.NewConfig()

var (
	REGION     = CFG.GetEnv("AWS_REGION")
	ACCESSKEY  = CFG.GetEnv("AWS_ACCESS_KEY_ID")
	SECRETKEY  = CFG.GetEnv("AWS_SECRET_ACCESS_KEY")
	BUCKETNAME = CFG.GetEnv("BUCKET_NAME")
)

func ConnectAWS() *session.Session {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(REGION),
			Credentials: credentials.NewStaticCredentials(
				ACCESSKEY,
				SECRETKEY,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}
	return sess
}

func UploadImage(c *gin.Context) {
	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)

	file, header, err := c.Request.FormFile("photo")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	filename := header.Filename
	ext := strings.Split(filename, ".")
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(BUCKETNAME),
		ACL:      aws.String("public-read"),
		Key:      aws.String(filename),
		Body:     file,
		Metadata: aws.StringMap(map[string]string{"Content-Type": "image/" + ext[1]}),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to upload file",
			"message":  err.Error(),
			"uploader": up,
		})
		return
	}
	filepath := "https://" + BUCKETNAME + "." + "s3.amazonaws.com/" + filename
	c.JSON(http.StatusOK, gin.H{
		"filepath": filepath,
	})
}

func SetUpGin() {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("sess", ConnectAWS())
		c.Next()
	})

	router.LoadHTMLGlob("templates/*")
	router.GET("/image", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})
	router.POST("/upload", UploadImage)

	router.Run(CFG.GetEnv("PORT"))
}

// //////////////MAIN//////////////////////
func main() {
	SetUpGin()
}
