package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"os/exec"
	"strings"
)

type Skipp3r struct {
	Bucket  *string
	Prefix  *string
	Version *string
	Svc     *s3.S3
}

// Unfortunately, the AWS SDK does not have a sync function like the AWS CLI.  Rather than program
// that in, we will simply exec the cli from within this function.
func (skipp3r Skipp3r) set(srcPath *string, destPath *string, doDelete *bool) error {
	cmdString := "--no-verify-ssl s3 sync " + *srcPath
	cmdString += " s3://" + *skipp3r.Bucket

	if skipp3r.Prefix != nil {
		cmdString += "/" + *skipp3r.Prefix
	}

	cmdString += "/" + *skipp3r.Version

	if *doDelete {
		cmdString += " --delete"
	}

	foo := strings.Split(cmdString, " ")
	return exec.Command("aws", foo...).Run()
}

func (skipp3r Skipp3r) get(pathIn *string) string {
	retVal := "{"

	path := ""
	if skipp3r.Prefix != nil {
		path += *skipp3r.Prefix
	}

	key := path + "/" + *skipp3r.Version + "/common/config"
	// Get common config first
	resp, err := skipp3r.Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(*skipp3r.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		retVal += "\"common\": " + string(b)
	}

	if *pathIn != "" && *pathIn != "common" {
		retVal += ","
		resp, err = skipp3r.Svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(*skipp3r.Bucket),
			Key:    aws.String(path + "/" + *skipp3r.Version + "/" + *pathIn + "/config"),
		})
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}

		if b, err := ioutil.ReadAll(resp.Body); err == nil {
			retVal += "\"" + *pathIn + "\": " + string(b)
		}
	}

	retVal += "}"
	return retVal
}
