codeURI := app/lambda/hello/main.go

# Vendor all the project dependencies.
tidy:
	go mod tidy
	go mod vendor

# Run all unit test
test:
	docker run -p 20000:27017 --detach --name thegoodseat_test \
	-e MONGO_INITDB_ROOT_USERNAME=user \
	-e MONGO_INITDB_ROOT_PASSWORD=password \
	mongo
	AWS_REGION=eu-west-1 go test ./... --count=1 || true
	docker stop thegoodseat_test
	docker rm thegoodseat_test

#=================================================== lambda
format-event:
	go run app/tools/test/main.go --endpointURL $(endpointURL)

build:
	docker build -t lambda:local --build-arg codeURI=$(codeURI) .

start:
	go run app/tools/test/main.go --endpointURL $(endpointURL)
	docker build -t lambda:local --build-arg codeURI=$(codeURI) .
	docker compose up

stop:
	docker compose down

#=================================================== AWS CDK
cdk-deploy:
	cdk deploy --all -O ./local/infra.spec.json

cdk-destroy:
	cdk destroy --all