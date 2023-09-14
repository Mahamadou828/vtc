# VTC
> it's a small aggregator of cab API. The goal of this project is to test a combination of aws CDK and sam to build 
> serverless application

## Table of Contents
* [General Info](#general-information)
* [Technologies Used](#technologies-used)
* [Features](#features)
* [Screenshots](#screenshots)
* [Setup](#setup)
* [Project Status](#project-status)
* [Room for Improvement](#room-for-improvement)
<!-- * [License](#license) -->


## General Information
- VTC provide a small api to access cab provider such as mysam, uber, bolt, husk  (currently only mysam is available)
- The goal of this project for me is to test a new way to deploy serverless app using aws cdk and the aws sam template format. The end goal will be to create a small package re-usable for my other project. 
<!-- You don't have to answer all the questions - just the ones relevant to your project. -->


## Main Technologies Used
- Golang - version 1.8
- AWS CDK - version 2.78


## Features ( ready as of today)
- getOffers: allow to fetch offer across multiple provider.
- requestRide: request a given provider offer, depending on the provider spec you may request a live ride ( some provider don't possess a staging env ).
- login & signup: you can create an account, manage your payment method ( use stripe demo card ) and create payment. 



## Setup
There are two ways to start the project: either as a Restful api where every lambda function will be serve as a http endpoint 
or you can start the lambda function inside a docker container. 

To start the restful server you just need to use the `make serve` command

To start the lambdas as container you will need to use the `make lambda-start` command and provide 3 parameters: 

- endpointURL: the path to access your lambda function 
- event: a json file name containing a APIProxyRequest event, this event will be injected inside the lambda function at runtime 
- codeURI: the path to the lambda main.go executable 

### Requirement to start this project 
To start the project you will need to have access to the provider apis, that mean all api key and configuration secret should 
be provided by you. 


## Project Status
Project is: _in progress_


## Room for Improvement

Room for improvement:
- Re-write the unit test following the refactoring of the models package  
- Find a way to have a better error handling experience. I'm thinking about introducing a displayable error message for user and an error request trace for developer also I need to introduce custom error type. 
- Refactor the models package to be value semantic and not pointer semantic 
- Refactor uber type pollution

To do:
- Integration of Uber and Husk. 
- Implementation of the refresh ride logic


<!-- Optional -->
<!-- ## License -->
<!-- This project is open source and available under the [... License](). -->

<!-- You don't have to include all sections - just the one's relevant to your project -->