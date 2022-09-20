package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go/middleware"
	"reflect"
	"testing"
)

func Test_getServiceNames(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test-multiple-services",
			args: args{
				s: "foobar,whizbang,helloworld",
			},
			want: []string{"foobar", "whizbang", "helloworld"},
		},
		{
			name: "test-one-service",
			args: args{
				s: "helloworld",
			},
			want: []string{"helloworld"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getServiceNames(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getServiceNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGlobalInactiveEnvironment(t *testing.T) {
	tests := []struct {
		branch  string
		service string
		color   string
		err     error
	}{
		{branch: "dev1", service: "rnnt-global", color: "green", err: nil},
		{branch: "main", service: "rnnt-global", color: "blue", err: nil},
		{branch: "arst", service: "rnnt-global", color: "blue", err: errors.New("no secret arn found")},
	}
	manager := &secretManagerMock{}

	for _, test := range tests {
		env, err := getGlobalInactiveEnvironment(manager, test.branch, test.service)

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("err should match expected. Expected %v, got %v", test.err, err)
			}

			continue
		} else {
			if err != test.err {
				t.Errorf("err should match expected. Expected %v, got %v", test.err, err)
			}
		}

		if env != test.color {
			t.Errorf("color should match expected. Expected %v, got %v", test.color, env)
		}
	}

}

type listSecretsMock struct {
}

func (l *listSecretsMock) ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	devARN := "dev"
	devVal := "dev1"
	env := "env"
	prodARN := "production"
	prodVal := "production"

	secrets := &secretsmanager.ListSecretsOutput{
		NextToken: nil,
		SecretList: []types.SecretListEntry{
			{
				ARN: &devARN,
				Tags: []types.Tag{
					{Key: &env, Value: &devVal},
				},
			},
			{
				ARN: &prodARN,
				Tags: []types.Tag{
					{Key: &env, Value: &prodVal},
				},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}

	return secrets, nil
}

type getSecretMock struct {
}

func (l *getSecretMock) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	id := *params.SecretId
	var secretString string

	if id == "dev" {
		secretString = "{\"CURRENT_LIVE_ENVIRONMENT\": \"blue\"}"
	} else if id == "production" {
		secretString = "{\"CURRENT_LIVE_ENVIRONMENT\": \"green\"}"
	}

	return &secretsmanager.GetSecretValueOutput{
		SecretString: &secretString,
	}, nil
}

type secretManagerMock struct {
	getSecretMock
	listSecretsMock
}
