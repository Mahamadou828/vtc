// Code for the local API server. This will serve all the endpoint present inside the
// template.yml file through a http server running on 3000 localhost. All env variable
// will be parsed from the env.local file
// @version 1.0
// @author.name  Mahamadou Samake
// @author.email formationsamake@gmail.com
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"vtc/business/v1/web"
	"vtc/foundation/config"
	"vtc/foundation/lambda"

	getOffers "vtc/app/lambda/get-offers/handler"
	hello "vtc/app/lambda/hello/handler"
	login "vtc/app/lambda/login/handler"
	signup "vtc/app/lambda/signup/handler"
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

var mapFunctionNameHandler = map[string]web.Handler{
	"loginHandler":     login.Handler,
	"getOffersHandler": getOffers.Handler,
	"signupHandler":    signup.Handler,
	"helloHandler":     hello.Handler,
}

func main() {
	//parse env.local file
	log.Println("Parsing default env file and loading all env variable")
	if err := godotenv.Load(".env.local"); err != nil {
		log.Fatalf("failed to parse env file: %v", err)
	}

	//create a new config app
	log.Println("Creating new app config")
	app, err := config.NewApp()
	if err != nil {
		log.Fatalf("failed to create new app config: %v", err)
	}

	//parse the template.yml file
	log.Println("Parsing template.yml")
	file, err := os.ReadFile("template.yml")
	if err != nil {
		log.Fatalf("can't open template.yml: %v", err)
	}

	var template Template
	err = yaml.Unmarshal(file, &template)

	// serve all functions as http endpoint
	router := mux.NewRouter()

	for _, function := range template.Functions {
		log.Printf("Registering new route [%s] with path [%s]", function.Name, function.Path)
		router.HandleFunc("/"+function.Path, func(writer http.ResponseWriter, request *http.Request) {
			vars := mux.Vars(request)
			writer.Header().Set("Content-Type", "application/json")

			event, err := web.GetLocalRequestEvent()
			if err != nil {
				log.Fatalf("failed to create a mock event object: %v", err)
			}

			event.Path = request.URL.Path
			event.PathParameters = vars
			event.QueryStringParameters = vars

			bodyBytes, err := io.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				resp := struct {
					Err error `json:"error"`
				}{
					Err: err,
				}
				respBytes, _ := json.Marshal(resp)
				writer.Write(respBytes)
				return
			}
			event.Body = string(bodyBytes)

			handler, ok := mapFunctionNameHandler[function.Name]
			if !ok {
				writer.WriteHeader(http.StatusInternalServerError)
				resp := struct {
					Err string `json:"error"`
				}{
					Err: fmt.Sprintf("handler for the current route %v is missing", function.Name),
				}
				respBytes, _ := json.Marshal(resp)
				writer.Write(respBytes)
				return
			}

			//extract aggregator from header
			agg := request.Header.Get("aggregator")
			if len(agg) < 1 {
				writer.WriteHeader(http.StatusInternalServerError)
				resp := struct {
					Err string `json:"error"`
				}{
					Err: "missing aggregator code in request header",
				}
				respBytes, _ := json.Marshal(resp)
				writer.Write(respBytes)
				return
			}

			//Create a new request trace
			trace := lambda.RequestTrace{
				Now:        time.Now(),
				ID:         uuid.NewString(),
				Aggregator: agg,
			}

			//Put the new trace inside the context
			ctx := context.WithValue(context.Background(), lambda.CtxKey, &trace)

			resp, err := handler(ctx, event, app, &trace)

			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				resp := struct {
					Err string `json:"error"`
				}{
					Err: fmt.Sprintf("failed to call handler: %v", err),
				}
				respBytes, _ := json.Marshal(resp)
				writer.Write(respBytes)
				return
			}

			writer.WriteHeader(resp.StatusCode)

			writer.Write([]byte(resp.Body))

			return

		}).Methods(function.Method)
	}

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         ":3030",
		Handler:      router,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 5,
	}

	log.Println("Starting server on port :3030")
	if err := api.ListenAndServe(); err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
}
