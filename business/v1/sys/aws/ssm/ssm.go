// Package ssm implement the aws secret manager service using the secret manager package.
// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.44.256/service/secretsmanager
package ssm

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// GetSecrets retrieve all the secrets for a given ssm pool
func GetSecrets(sess *session.Session, poolName string) (map[string]string, error) {
	var res map[string]string

	svc := secretsmanager.New(sess)

	secret, err := svc.GetSecretValue(
		&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(poolName),
		},
	)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				return res, fmt.Errorf("failed to retrieve secret: %s, error: %s, %s", *secret.Name, secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			case secretsmanager.ErrCodeInvalidParameterException:
				return res, fmt.Errorf("failed to retrieve secret: %s, error: %s, %s", *secret.Name, secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				return res, fmt.Errorf("failed to retrieve secret: %s, error: %s, %s", *secret.Name, secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
			case secretsmanager.ErrCodeDecryptionFailure:
				return res, fmt.Errorf("failed to retrieve secret: %s, error: %s, %s", *secret.Name, secretsmanager.ErrCodeDecryptionFailure, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				return res, fmt.Errorf("failed to retrieve secret: %s, error: %s, %s", *secret.Name, secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			default:
				return res, fmt.Errorf(aerr.Error())
			}
		} else {
			return res, err
		}
	}

	if err := json.Unmarshal([]byte(*secret.SecretString), &res); err != nil {
		return res, fmt.Errorf("failed to unmarshal secret: %s, error: %s", poolName, err)
	}

	return res, nil
}
