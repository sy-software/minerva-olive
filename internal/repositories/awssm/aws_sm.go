package awssm

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	nilUtils "github.com/sy-software/minerva-go-utils/nil"
)

type AWSSM struct {
	mngr *secretsmanager.SecretsManager
}

func NewAWSSM() *AWSSM {
	session := session.Must(session.NewSession())

	region, exists := os.LookupEnv("AWS_REGION")
	if !exists {
		region, exists = os.LookupEnv("AWS_DEFAULT_REGION")
		if !exists {
			region = "us-east-1"
		}
	}

	mngr := secretsmanager.New(session, aws.NewConfig().WithRegion(region))

	return &AWSSM{
		mngr: mngr,
	}
}

func (aws *AWSSM) Get(name string) (string, error) {
	// TODO: Implement secret versioning an stage
	input := secretsmanager.GetSecretValueInput{
		SecretId: &name,
	}

	output, err := aws.mngr.GetSecretValue(&input)

	if err != nil {
		return "", err
	}

	return nilUtils.CoalesceStr(output.SecretString, ""), nil
}
