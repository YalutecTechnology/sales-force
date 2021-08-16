# salesforce-integration #

Description Application

## What is this repository for? ##

Description Service

The service has two main folders:
## App folder ##

This folder contains the main functionalities of the service together with the packages for the API server.

#### Base folder

This folder has all the support packages that will be used by the app folder. 

Here is the list of packages in the base folder with its explanation:

Package  | Description  | Implementations or sub-packages  
-------- | ------------ | -------------------------------- 
cache | Contains logic to retrieve and store line windows on cache. | Implementations: Redis
clients | Contains the clients to send requests to sent-to and the HSM api | 
helpers | It contains utilities that help us in the creation of endpoints | 

## How do I get set up? ##

* For this service you don't need to have a REDIS DB up and running. It's not necessary to have a an specific version or replication strategy.
*  It is necessary to create a Json Web Token through the ***/authenticate*** endpoint, sending a username and password in the body of the request.
* Here are the required ENV vars needed to run this service:


| **Name**             | **Description**                                   | **Required**      | **Defaults**      |
| --------------       | -----------------                                 | ----------------- | ----------------- |
| `SALESFORCE-INTEGRATION_APP_NAME`    | The connection name displayed on the application. | false             | **salesforce-integration**        |
| `SALESFORCE-INTEGRATION_HOST`        | The host for the service API.                     | false             | **localhost**     |
| `SALESFORCE-INTEGRATION_PORT`        | The port for the service API.                     | otherwise 8080    | **8080**          |
| `SALESFORCE-INTEGRATION_SENTRY_DSN`  | The project DSN for Sentry events.                | false             |                   |
| `SALESFORCE-INTEGRATION_ENVIRONMENT` | The environment used for Sentry events.           | false             | **dev**           |
| `SALESFORCE-INTEGRATION_MAIN_CONTEXT_TIME_OUT`| The time out for the main context in seconds.| false | **10** |
| `SALESFORCE-INTEGRATION_REDIS_ADDRESS`| The address of the Redis Instance (not sentinel).| Only if Redis Instance is not a sentinel Redis.| |
| `SALESFORCE-INTEGRATION_REDIS_MASTER`| The name of the master on a Redis Sentinel Instance (sentinel).| Only if Redis Instance is a sentinel Redis.| |
| `SALESFORCE-INTEGRATION_REDIS_SENTINEL_ADDRESS`| The address of the Redis Sentinel Instance (sentinel).| Only if Redis Instance is a sentinel Redis.| |
| `SALESFORCE-INTEGRATION_BOTRUNNER_URL`| Where we will send the requests to change states of converstations.| false | **http://botrunner** |
| `SALESFORCE-INTEGRATION_BOTRUNNER_TOKEN`| Access token if necessary to make requests to Sent-to.| false | |
| `SALESFORCE-INTEGRATION_YALO_USERNAME`| Username required to generate a JWT with ADMIN role through the */authenticate* endpoint.| true | **yaloUser** |
| `SALESFORCE-INTEGRATION_YALO_PASSWORD`| Password required to generate a JWT with ADMIN role through the */authenticate* endpoint. | true |  |
| `SALESFORCE-INTEGRATION_SECRET_KEY`| String required to sign the JWT that are created through the */authenticate* endpoint.| true |  |

## Running project ##

We must clone the repository with the ***develop*** branch or the branch with which we must work a *feature* or *bugfix* 

In order to run the project locally, we must go to the folder of our copy and use the command: 
``` sh
go run app/main.go
```
To run the tests of this project, we use the command:
``` sh
go test -coverprofile=coverage.txt -covermode=atomic ./...
```

## Running the project in Dev ##
In order to test our changes in dev, you run your service in a development cluster.

Your team has access to a development cluster, to which you can deploy the new service. If you followed the prerequisites installation in the [onboarding guide Part 1](https://www.notion.so/On-boarding-part-1-Setting-up-your-system-83b639f3ec3b4f5e8e322966960a4e1d), you will already have Skaffold, Visual Studio Code, and the Cloud Code extension for VS Code installed.

The stack will be in the Yalo staging cluster, accessible via:
```bash 
gcloud config configurations activate yalo-staging-env
gcloud container clusters get-credentials staging
kubectl cluster-info
```

Log in to docker using this command:

```bash 
gcloud auth configure-docker
```

Now change the namespace to one that we use to use this project: 

```bash 
kubectl config set-context --current --namespace=${nameSpaceAssigned}
```

Before you continue, run the following two commands to make sure the cluster and namespace are set correctly:

```bash
kubectl config current-context
```
Output: ***gke_yalo-staging-env_us-west2-a_staging***

```bash
kubectl config view | grep namespace
```
Output: ***${nameSpaceAssigned}***

Now run Skaffold in the staging environment, use this command

``` sh
skaffold dev --default-repo gcr.io/yalo-staging-env --port-forward
```

Before initializing skaffold we need to add in the ***hello.deployment.yaml*** file the envars that we need to use, all the required ones specified above must be added. Example:

``` yaml
    env:
        - name: SALESFORCE-INTEGRATION_HOST
          value: "0.0.0.0"
        - name: SALESFORCE-INTEGRATION_PORT
          value: "8080"
        - name: SALESFORCE-INTEGRATION_MONGO_CONNECTION_STRING
          value: "mongodb://mongodb-staging2-servers-vm-0.c.yalo-staging-env.internal:27017,mongodb-staging2-servers-vm-1.c.yalo-staging-env.internal:27017/?replicaSet=rs0"
        - name: SALESFORCE-INTEGRATION_REDIS_MASTER
          value: "mymaster"
        - name: SALESFORCE-INTEGRATION_REDIS_SENTINEL_ADDRESS
          value: "redis-redis-ha-announce-0.stage-1.svc.cluster.local:26379;redis-redis-ha-announce-1.stage-1.svc.cluster.local:26379;redis-redis-ha-announce-2.stage-1.svc.cluster.local:26379"
        - name: SALESFORCE-INTEGRATION_BOTRUNNER_URL
          value: "http://botrunner.stage-1:3000"    
        - name: SALESFORCE-INTEGRATION_YALO_USERNAME
          value: "${yaloUsername}"
        - name: SALESFORCE-INTEGRATION_YALO_PASSWORD
          value: "${yaloPassword}"
        - name: SALESFORCE-INTEGRATION_SECRET_KEY
          value: "${studioSecret}"
```

**Note:** *To see the previous steps in more detail, we can see the following document* [On-boarding part 2: Publishing an application](https://www.notion.so/On-boarding-part-2-Publishing-an-application-eac08ad3eaad435cb242340fe1a2bb98#2ba8af0501964491a730bb979fcd2ced)

## Who do I talk to? ##

* Gerardo Ezquerra Martín - **cat@underdog.mx**
* Armando Hernández Aguayo - **armando@yalochat.com**

