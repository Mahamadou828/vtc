# Vendor all the project dependencies.
tidy:
	go mod tidy
	go mod vendor

test:
	go test ./... --count=1

#=================================================== AWS CDK
cdk-deploy:
	cdk deploy --all -O ./local/infra.spec.json

cdk-destroy:
	cdk destroy --all