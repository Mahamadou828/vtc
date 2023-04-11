package main

import (
	"log"
	"os"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	agw "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
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
		Api struct {
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

	Functions map[string]Function `yaml:"Functions"`
}

type Function struct {
	CodeURI     string `yaml:"CodeURI"`
	Path        string `yaml:"Path"`
	Name        string `yaml:"Name"`
	Description string `yaml:"Description"`
	Method      string `yaml:"Method"`
	Environment struct {
		Variables map[string]string `yaml:"Variables"`
		Secrets   map[string]string `yaml:"Secrets"`
	} `yaml:"Environment"`
}

type InfraStackProps struct {
	awscdk.StackProps
}

func NewInfraStack(scope constructs.Construct, id string, props *InfraStackProps) awscdk.Stack {
	//Initialize a aws stack
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	//open and parse the template file
	file, err := os.ReadFile("app/infra/template.yml")
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

	for _, function := range template.Functions {
		//create a new endpoint
		endpoint := api.Root().AddResource(jsii.String(function.Path), nil)

		//extract all environment variables
		env := map[string]*string{}
		for name, value := range function.Environment.Secrets {
			env[name] = jsii.String(value)
		}

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
