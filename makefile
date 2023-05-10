codeURI := app/lambda/hello/main.go

# Vendor all the project dependencies.
tidy:
	go mod tidy
	go mod vendor

test:
	go test ./... --count=1

#=================================================== lambda
start:
	docker build -t lambda:local --build-arg codeURI=$(codeURI) .
	docker compose up

stop:
	docker compose down

#=================================================== AWS CDK
cdk-deploy:
	cdk deploy --all -O ./local/infra.spec.json

cdk-destroy:
	cdk destroy --all