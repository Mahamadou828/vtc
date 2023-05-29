#Build the go binary for the given lambda function
FROM golang:1.19 as app

ARG codeURI

#Disable CGO to assure that the binary is not bind to anything
ENV CGO_ENABLED 0

RUN mkdir -p /service

COPY . /service

WORKDIR /service

RUN go build -o main ${codeURI}


#Run the binary inside the lambda RIE ( Runtime Interface Emulator )
FROM public.ecr.aws/lambda/go:latest
COPY --from=app /service ${LAMBDA_TASK_ROOT}

COPY ./event.local.json ${LAMBDA_TASK_ROOT}/event.local.json

CMD ["main"]