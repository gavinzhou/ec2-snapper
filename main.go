package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/kelseyhightower/envconfig"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// Specification set lambda env
type Specification struct {
	InstaincesList []string `envconfig: "INSTAINCESLIST"`
	CronENV        string   `envconfig: "CRONENV"`
	OWNER          string   `envconfig: "WONER"`
}

// SnapperInstance is snapper instance info
type SnapperInstance struct {
	InstanceID string
	tagName    string
	imageDate  string
	Owners     []*string
	cronenv    string
}

// CreateImageWithInstancesID is create ami from instanceID
func (s *SnapperInstance) CreateImageWithInstancesID(ec2svc ec2iface.EC2API) error {
	opts := &ec2.CreateImageInput{
		InstanceId:  aws.String(s.InstanceID),
		Description: aws.String(s.cronenv + "-" + s.tagName + "-" + s.imageDate),
		Name:        aws.String(s.cronenv + "-" + s.tagName + "-" + s.imageDate),
		NoReboot:    aws.Bool(true),
	}
	resp, err := ec2svc.CreateImage(opts)

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
		return err
	}
	fmt.Println("Response output: %s", resp)
	return nil
}

// GetInstanceTagName is get instainces tagName
func GetInstanceTagName(ec2svc ec2iface.EC2API, InstainceID string) string {
	res, err := ec2svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(InstainceID)},
	})
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, r := range res.Reservations {
		for _, i := range r.Instances {
			var tagName string
			for _, t := range i.Tags {
				if *t.Key == "Name" {
					tagName = *t.Value
				}
			}
			return tagName
		}
	}
	return ""
}

func ec2ami() {
	imageDate := time.Now().Format("2006010215")

	// InstaincesList := []string{"i-0419c138f7c51eea4", "i-0b65e6a087dabbf14"}
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
	for _, i := range sp.InstaincesList {
		s := SnapperInstance{
			InstanceID: i,
			tagName:    GetInstanceTagName(ec2Svc, i),
			imageDate:  imageDate,
			cronenv:    sp.CronENV,
		}
		s.CreateImageWithInstancesID(ec2Svc)
	}
}

func main() {
	lambda.Start(ec2ami)
}
