// Code of the architecture for the tgs in go project. The infrastructure use the aws cdk package in go
// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2@v2.73.0#section-readme
// @version 1.0
// @author.name  Mahamadou Samake
// @author.email formationsamake@gmail.com
package main

/*
	@todo add many env to the infra
*/

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"vtc/business/v1/sys/aws/ssm"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	agw "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	cognito "github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	docdb "github.com/aws/aws-cdk-go/awscdk/v2/awsdocdb"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	iam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	identitypool "github.com/aws/aws-cdk-go/awscdkcognitoidentitypoolalpha/v2"
	lambda "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"gopkg.in/yaml.v2"
)

type Template struct {
	TemplateVersion string `yaml:"TemplateVersion"`
	Description     string `yaml:"Description"`
	APIVersion      string `yaml:"APIVersion"`
	Globals         struct {
		SSMPoolName string `yaml:"SSMPoolName"`
		Api         struct {
			Cors struct {
				AllowMethods     []string `yaml:"AllowMethods"`
				AllowHeaders     []string `yaml:"AllowHeaders"`
				AllowOrigin      []string `yaml:"AllowOrigin"`
				AllowCredentials bool     `yaml:"AllowCredentials"`
			} `yaml:"Cors"`
		} `yaml:"Api"`
		Function struct {
			MemorySize float64 `yaml:"MemorySize"`
			Timeout    float64 `yaml:"Timeout"`
		} `yaml:"Function"`
	} `yaml:"Globals"`

	Functions map[string]Function `yaml:"Resources"`
}

type Function struct {
	CodeURI     string `yaml:"CodeURI"`
	Path        string `yaml:"Path"`
	Name        string `yaml:"Name"`
	Description string `yaml:"Description"`
	Method      string `yaml:"Method"`
	Environment struct {
		Variables map[string]string `yaml:"Variables"`
	} `yaml:"Environment"`
}

type InfraStackProps struct {
	awscdk.StackProps
}

func NewInfraStack(scope constructs.Construct, id string, props *InfraStackProps) awscdk.Stack {
	//Initialize a new aws session
	sess, err := session.NewSession(&aws.Config{
		Region: props.Env.Region,
	})
	if err != nil {
		log.Fatalf("can't create a new aws session: %v", err)
	}

	//Initialize a aws stack
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	//================================================================= VPC
	//create a vpc with a public and a private subnet
	vpc := ec2.NewVpc(stack, jsii.String("tgs-with-go-vpc"), &ec2.VpcProps{
		VpcName: jsii.String("tgs-with-go-vpc"),
		SubnetConfiguration: &[]*ec2.SubnetConfiguration{
			{
				Name:       jsii.String("tgs-with-go-public-tgs"),
				SubnetType: ec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	//================================================================= Cognito
	//upload the pre-signup lambda trigger
	preSignFn := lambda.NewGoFunction(
		stack,
		jsii.String(fmt.Sprintf("tgs-with-go-cognito-presignup")),
		&lambda.GoFunctionProps{
			Entry:        jsii.String("app/lambda/cognitopresignup"),
			FunctionName: jsii.String("tgs-with-go-cognito-presignup"),
		},
	)

	//create a cognito pool for user
	c := cognito.NewUserPool(stack, jsii.String("tgs-with-go-pool"), &cognito.UserPoolProps{
		UserPoolName:  jsii.String("tgs-with-go-pool"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		SignInAliases: &cognito.SignInAliases{
			Username:          jsii.Bool(true),
			PreferredUsername: jsii.Bool(true),
		},
		AutoVerify: &cognito.AutoVerifiedAttrs{
			Phone: jsii.Bool(true),
		},
		StandardAttributes: &cognito.StandardAttributes{
			Email: &cognito.StandardAttribute{
				Required: jsii.Bool(true),
				Mutable:  jsii.Bool(true),
			},
			PhoneNumber: &cognito.StandardAttribute{
				Required: jsii.Bool(true),
				Mutable:  jsii.Bool(true),
			},
			Fullname: &cognito.StandardAttribute{
				Required: jsii.Bool(true),
				Mutable:  jsii.Bool(true),
			},
		},
		PasswordPolicy: &cognito.PasswordPolicy{
			MinLength:        jsii.Number(12),
			RequireLowercase: jsii.Bool(true),
			RequireUppercase: jsii.Bool(true),
			RequireDigits:    jsii.Bool(true),
			RequireSymbols:   jsii.Bool(true),
		},
		AccountRecovery:   cognito.AccountRecovery_PHONE_ONLY_WITHOUT_MFA,
		SelfSignUpEnabled: jsii.Bool(true),
		LambdaTriggers: &cognito.UserPoolTriggers{
			PreSignUp: preSignFn,
		},
	})

	//create a new app client
	poolClient := c.AddClient(jsii.String("tgs-with-go-api"), &cognito.UserPoolClientOptions{
		AuthFlows: &cognito.AuthFlow{
			AdminUserPassword: jsii.Bool(true),
			Custom:            jsii.Bool(true),
			UserPassword:      jsii.Bool(true),
			UserSrp:           jsii.Bool(true),
		},
		GenerateSecret: jsii.Bool(false),
	})

	identitypool.NewUserPoolAuthenticationProvider(&identitypool.UserPoolAuthenticationProviderProps{
		UserPool:       c,
		UserPoolClient: poolClient,
	})

	identitypool.NewIdentityPool(stack, jsii.String("tgs-with-go-identitypool"), &identitypool.IdentityPoolProps{
		AllowUnauthenticatedIdentities: jsii.Bool(true),
		AuthenticationProviders: &identitypool.IdentityPoolAuthenticationProviders{
			UserPools: &[]identitypool.IUserPoolAuthenticationProvider{
				identitypool.NewUserPoolAuthenticationProvider(&identitypool.UserPoolAuthenticationProviderProps{
					UserPool:       c,
					UserPoolClient: poolClient,
				}),
			},
		},
		IdentityPoolName: jsii.String("tgs-with-go-identity-pool"),
	})

	//================================================================= Database-DocumentDB
	//create the document db database
	docdb.NewDatabaseCluster(stack, jsii.String("tgs-with-go-db"), &docdb.DatabaseClusterProps{
		MasterUser: &docdb.Login{
			Username:   jsii.String("master"),
			SecretName: jsii.String("tgs-with-go-db-secret"),
		},
		InstanceType:  ec2.InstanceType_Of(ec2.InstanceClass_MEMORY5, ec2.InstanceSize_LARGE),
		Vpc:           vpc,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	//================================================================= ApiGateway
	//create a iam role to attribute to all lambda functions
	role := iam.Role_FromRoleArn(
		stack,
		jsii.String("tgs-with-go-lambda-role"),
		jsii.String("arn:aws:iam::685367675161:role/TGS-WITH-GO"),
		nil,
	)

	//open and parse the template file
	file, err := os.ReadFile("template.yml")
	if err != nil {
		log.Fatalf("can't open template.yml: %v", err)
	}

	var template Template
	err = yaml.Unmarshal(file, &template)

	//create a new api gateway
	api := agw.NewRestApi(stack, jsii.String("tgswithgoapi"), &agw.RestApiProps{
		DefaultCorsPreflightOptions: &agw.CorsOptions{
			AllowMethods:     jsii.Strings(template.Globals.Api.Cors.AllowMethods...),
			AllowHeaders:     jsii.Strings(template.Globals.Api.Cors.AllowHeaders...),
			AllowOrigins:     jsii.Strings(template.Globals.Api.Cors.AllowOrigin...),
			AllowCredentials: jsii.Bool(template.Globals.Api.Cors.AllowCredentials),
		},
	})

	//extract secret from aws secret manager
	secrets, err := ssm.GetSecrets(sess, template.Globals.SSMPoolName)
	if err != nil {
		log.Fatalf("can't get secrets from provided pool: %s err: %v", template.Globals.SSMPoolName, err)
	}

	for _, function := range template.Functions {
		//create a new endpoint
		endpoint := api.Root().AddResource(jsii.String(function.Path), nil)

		//extract all environment variables
		env := map[string]*string{}
		for name, value := range function.Environment.Variables {
			env[name] = jsii.String(value)
		}

		//extract secret from aws secret manager and inject them into the environment variables.
		for value, secret := range secrets {
			env[secret] = jsii.String(value)
		}

		//put default environment variables
		env["COGNITO_USER_POOL_ID"] = c.UserPoolId()
		env["COGNITO_CLIENT_POOL_ID"] = poolClient.UserPoolClientId()

		//create the new lambda function
		lambdaFn := lambda.NewGoFunction(
			stack,
			jsii.String(function.Name),
			&lambda.GoFunctionProps{
				Entry:        jsii.String(function.CodeURI),
				FunctionName: jsii.String(function.Name),
				MemorySize:   jsii.Number(template.Globals.Function.MemorySize),
				Description:  jsii.String(function.Description),
				Environment:  &env,
				Timeout:      awscdk.Duration_Seconds(jsii.Number(template.Globals.Function.Timeout)),
				Role:         role,
			},
		)

		//adding endpoint and linking the function to it
		endpoint.AddMethod(
			jsii.String(function.Method),
			agw.NewLambdaIntegration(
				lambdaFn,
				nil,
			),
			&agw.MethodOptions{},
		)
	}

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewInfraStack(app, "TgsWithGoStack", &InfraStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Region: jsii.String("eu-west-1"),
	}
}
