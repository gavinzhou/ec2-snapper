package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// Specification set lambda env
type Specification struct {
	OWNER        string `envconfig: "OWNER"`
	MonthlyPurge int    `envconfig: "MONTYLYPURGE"`
	DailyPurge   int    `envconfig: "DAILYPURGE"`
}

// CheckPurgeDays check ami is more than purge days
func CheckPurgeDays(CreationDate string, Purge int) bool {
	c, err := time.Parse("2006-01-02T15:04:05Z", CreationDate)
	if err != nil {
		return false
	}
	return int(time.Now().Sub(c).Hours()/24) > Purge
}

// ListAllBackupImages is get owners all images
func (s Specification) ListAllBackupImages(ec2svc ec2iface.EC2API) []string {
	var AMILists []string
	res, err := ec2svc.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String(s.OWNER)},
	})
	if err != nil {
		return AMILists
	}
	for _, i := range res.Images {
		if strings.Contains(*i.Name, "monthly-") && CheckPurgeDays(*i.CreationDate, s.DailyPurge) {
			AMILists = append(AMILists, *i.ImageId)
		}
		if strings.Contains(*i.Name, "daily-") && CheckPurgeDays(*i.CreationDate, s.MonthlyPurge) {
			AMILists = append(AMILists, *i.ImageId)
		}
	}
	return AMILists
}

// DeregisterImages is delelte images with daily
func DeregisterImages(ec2svc ec2iface.EC2API, InstainceID string) {
	_, err := ec2svc.DeregisterImage(&ec2.DeregisterImageInput{
		ImageId: aws.String(InstainceID),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occurred
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			// This case should never be hit, the SDK should always return an
			// error which satisfies the awserr.Error interface.
			fmt.Println(err.Error())
		}
	}
}

func purgeami() {
	var sp Specification
	if err := envconfig.Process("", &sp); err != nil {
		fmt.Println(err)
	}
	// Load session from shared config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create new ec2 client
	ec2Svc := ec2.New(sess)
	s := Specification{
		OWNER:        sp.OWNER,
		MonthlyPurge: sp.MonthlyPurge,
		DailyPurge:   sp.DailyPurge,
	}
	for _, i := range s.ListAllBackupImages(ec2Svc) {
		fmt.Printf("Deregistered image %s \n", i)
		DeregisterImages(ec2Svc, i)
	}
}

func main() {
	lambda.Start(purgeami)
}
