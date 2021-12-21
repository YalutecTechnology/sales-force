# salesforce-integration #

This service will be used as a solution for Coppel, it will allow communication between users of WhatsApp or Facebook bots and Coppel's human advisor through a live chat on the Salesforce platform.

The requirements of this solution are documented in the following [RFC-155.](https://www.notion.so/yalo/RFC-155-Salesforce-integration-204cbcfdacd44fd2a20e41a4348df35c).

We can have more detailed information on the flow of this solution in the following documentation:

- [Salesforce Integration Middleware](https://www.notion.so/yalo/Salesforce-Integration-Middleware-fca28101435948a8a52b61733c955973)

## What is this repository for? ##

This service will serve as a middleware that will connect a user of a Yalo bot when requesting assistance with a human agent through a live chat on the Salesforce platform.

This solution aims to be used for different implementations with any client in a simple way, in which it is necessary to establish a live chat, through a Yalo bot and the Salesforce CRM platform.

The flow implemented for creating chat between an end user and a human agent in Salesforce is as follows:

1. The user is interacting with a Yalo bot.
2. If the user requests human assistance by typing the keyword ***"Ayuda"***, Yalo Core, puts the artificial intelligence in a waiting state and Yalo Component creates a Chat session with Salesforce by making a request to the endpoint [`v1/chats/connect`](https://www.notion.so/yalo/Salesforce-Integrations-Endpoints-57ff093479084b89a9990a5bd0faa7b1) of this solution.
3. The endpoint must consume the following minimum information to create a chat:

| Field | Description
| :---  | :--- |
| userID | User identifier phone in whatsapp and facebook Id in messenger. |
| botSlug  | Bot name. |
| botId |  Phone for whatsapp bot or pageId in messenger bot. |
| name |  User name. |
| provider | Chat origin, the allowed values are **whatsapp** or **facebook**. |
| email |  User's email to search or register the contact or person account in Salesforce.|
| phoneNumber | User's phone number to search or register the contact or personal account in Salesforce allowed in whatsapp bot, only if it is enabled. |
| extraData | It mainly sends the custom fields that the customer has in Salesforce to add them to the cases and they are sent if we have the information. If necessary you can send metadata for custom implementations. |

4. First, it is validated that the user does not have a live chat in existence at the time of making the request.

5. First, it is validated that the user does not have a live chat in existence at the time of making the request.

6. We search through the email or phone received that there is a **contact** or **personal account** in the customer's Salesforce account, if it does not exist, the contact or personal account is created.

7. We validate that the user is not blocked in Salesforce, if so, the corresponding chat is not created and through the botrunner or studiong client the user is changed to a *blocked user state*, specified in the environment variable ***BLOCKED_USER_STATE***

8. Once the contact has been validated in Salesforce, we create a query case with the data provided to add it and request a chat. [ChatRequest in LiveAgent API](https://developer.salesforce.com/docs/atlas.en-us.live_agent_rest.meta/live_agent_rest/live_agent_rest_ChasitorInit.htm)

9. After successfully requesting the chat session with Salesforce, we save the information necessary to maintain the session in an **Interconnection** object in Redis.

10. Finally, we initialize in the ***Interconnection*** a goroutine that starts the [***Long Polling***](https://developer.salesforce.com/docs/atlas.en-us.live_agent_rest.meta/live_agent_rest/live_agent_rest_http_long_polling_loop.htm) process with Salesforce, with which we consume a service to identify the different events that happen in the chat every 5 seconds.

11. According to the events received, **salesforce-integration** performs the following actions:

* ***204 HTTP Status :*** It means that there are no changes in the chat.
* ***503 HTTP Status :*** An attempt will be made to reconnect the session because the connection to Salesforce was lost. More details [ReconnectSession](https://developer.salesforce.com/docs/atlas.en-us.live_agent_rest.meta/live_agent_rest/live_agent_rest_ReconnectSession.htm)
* ***Other HTTP Status:*** The **Interconnection** is ended and through the **Botrunner or Studing client** the user is changed to a time-ended state, specified in the **TIMEOUT_STATE** environment variable
* ***ChatRequestFail:*** The chat was not created correctly in the Salesforce console, due to no agents available in the queue the chat was sent to or some unknown error.
* ***ChatRequestSuccess :*** The chat was created successfully and the **Interconnection** goes to `OnHold` status waiting for a human agent to accept the chat. The service sends the user a `Esperando agente` message via [**Integrations API**](https://www.notion.so/yalo/Integrations-API-00643b2f689943bb9daa1c5d064f0b32#fea506d44c434aeebb8399fe6225e3c1).
* ***ChatEstablished :*** An agent accepts the chat and the **Interconnection** goes to `Active` status, at this time both can send messages to each other, but before that, the service sends the context of the query as the first message, that is, the previous conversation between the bot's AI and the user before requesting a live chat.
* ***ChatMessage :*** The agent sent a message to the user, then through [**Integrations API**](https://www.notion.so/yalo/Integrations-API-00643b2f689943bb9daa1c5d064f0b32#fea506d44c434aeebb8399fe6225e3c1) we forward the message to the user to their *WhatsApp* or *Messenger*.
* ***ChatEnded :*** The agent ended the chat and the **Interconnection** goes to `Closed` status, closing the session in the integration and through the **Botrunner or Studiong client** the user is changed to a successfully *finished chat state*, specified in the environment variable **SUCCESS_STATE**.

***How do I send a user's response to an agent in the Salesforce console?***

When the instance is raised, we need to register the webhooks through the endpoint `/v1/integrations/webhook/register/{{provider}}`, then the webhooks are registered in the [Integrations API](https://www.notion.so/yalo/Integrations-API-00643b2f689943bb9daa1c5d064f0b32#de3c9ae04829401e86f756092e0e58e2) that will receive the messages between the users and the bot.

If we receive a text or image message from a user, we validate the following:

* If the user does not have an Active or OnHold chat, this message is saved as context in Redis.

* If there is an Active chat then through our [LiveAgent API client](https://developer.salesforce.com/docs/atlas.en-us.live_agent_rest.meta/live_agent_rest/live_agent_rest_ChatMessage_request.htm) we send that message to the agent's console. If the message is of type Image, a message is sent to the agent saying `"El usuario adjunto una imagen al caso"`, and the image through our Salesforce API client is attached to the case.

Any other type of message other than TEXT or IMAGE in the webhooks is ignored.

The service has two main folders:

## App folder ##

This folder contains the main functionalities of the service together with the packages for the API server.

The **api/hadlers** package contains the endpoints available for this integration, mainly the endpoint ***/chats/connect*** to create the chats with Salesforce, as well as the endpoints that we need to register as webhooks in **integrations-api** so that **integrations-channels** can send us the messages between users and bots of *WhatsApp* and *Faceebook*.

Each of them is detailed in the [README](https://bitbucket.org/yalochat/salesforce-integration/src/develop/app/api/handlers/README.md) file in ***app/api/handlers*** or in [Saleforce Integrations Endpoints](https://www.notion.so/yalo/Salesforce-Integrations-Endpoints-57ff093479084b89a9990a5bd0faa7b1).

In **/manager** folder we will find all the logic implemented to be able to manage the chats between the end users and the agents in Salesforce. For this we explain some components that we need to be able to perform these functions:

* ***interconnectionMap:*** This map is a local cache that will store the **Interconnections** on ONHOLD and ACTIVE. Once the chats are finished they are removed from this map and the `long-polling` service with Salesforce is closed.

* [***Interconnection:***](https://www.notion.so/yalo/Interconnection-5a09f5048671404994c895e574cea8c6) Structure that represents a live chat session between a Yalo bot user and a human agent in the Salesforce console. This object is stored in Redis and contains all the necessary information to be able to start the `Long-Polling` service with Salesforce. In the following table we see the necessary information that is needed and stored in Redis:

| Field | Description
| :---  | :--- |
| userID | User identifier phone in whatsapp and facebook Id in messenger. |
| sessionId | Chat session identifier in Salesforce |
| sessionKey | Security key for the chat session required for requests. |
| affinityToken | The affinity token for the session that’s passed in the header for all future requests. |
| status  | The three states are: OnHold, Active and Closed. |
| timestamp  | Timestamp of the creation of the inteconnection. |
| botSlug  | Bot name. |
| botId |  Phone for whatsapp bot or pageId in messenger bot. |
| name |  User name. |
| provider | Chat origin, the allowed values are **whatsapp** or **facebook**. |
| email |  Email received to create the chat. |
| phoneNumber | Phone received to create the chat. |
| caseId | The ID of the case created in Salesforce for this chat session. |
| extraData | The custom fields that the customer has in Salesforce to add them to the cases or metadata for custom implementations. |

* ***contextcache:*** This redis client will store the context of each user of all the messages sent between him and the bot, received through the webhook.

* ***context:*** The context is the messages sent between the user and the bot before a live chat is requested. This struct represents the information that we store in redis whose time-to-live is only 24 hours.

| Field | Description
| :---  | :--- |
| userID | User identifier phone in whatsapp and facebook Id in messenger. |
| timestamp  | Timestamp of sending the message. |
| url  | Image url if it is an image type message. |
| mimeType |  MIME Type of the image. |
| text |  Message text. |
| from | Who sent the message the user or the bot. |

* ***SalesforceService:*** This component contains the clients to connect to the Salesforce APIs and the functions used with Salesforce:
  1. Authorization API
  2. [Live Agent Rest API](https://developer.salesforce.com/docs/atlas.en-us.live_agent_rest.meta/live_agent_rest/live_agent_rest_API_requests.htm)
  3. [Data Salesforce API](https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_quickstart_intro.htm)

* ***IntegrationsClient:*** Client to use message sending and webhook logging endpoints with [Integrations API](https://developer.yalo.com/yalo/reference/integrations-api-overview).

* ***BotrunnnerClient:*** SentTo client to be able to change a user's status in the bot flow.

#### Base folder

This folder has all the support packages that will be used by the app folder.

Here is the list of packages in the base folder with its explanation:

Package  | Description  | Implementations or sub-packages
-------- | ------------ | --------------------------------
cache | Contains logic to retrieve and store mensages of context and interconnections on cache. | Implementations: Redis (ContextCache, InterconnectionCache) and [Ristretto Cache](https://github.com/dgraph-io/ristretto)
clients | Contains the clients to send requests to Botrunner sent-to, Studiong sent-to, Integrations API, AgentLiveChat API and Saleforce API |
constants | It will contain the common constants and errors that we use in the service|
helpers | It contains utilities that help us in the creation of endpoints or file encoding, and other functionalities.|
models | Contains base structures that we will use for this service.  |

## How do I get set up? ##

* For this service you don't need to have a REDIS DB up and running. It's not necessary to have a an specific version or replication strategy.
* It is necessary to create a Json Web Token through the `/authenticate`  endpoint, sending a username and password in the body of the request.
* This integration can change the state of a user in the bot through botrunner or studiong, you just have to add the ENV API vars you want to use.
* This integration can send messages to the bot through integrations API.
* Here are the required ENV vars needed to run this service:


| **Name**  | **Description**  | **Required**  | **Defaults**      |
| -------------- | ---------------------- | ----------------- | -------- |
| `SALESFORCE-INTEGRATION_APP_NAME`    | The connection name displayed on the application. | true             | **salesforce-integration**        |
| `SALESFORCE-INTEGRATION_HOST`        | The host for the service API.                     | true             | **localhost**     |
| `SALESFORCE-INTEGRATION_PORT`        | The port for the service API.                     | true otherwise 8080    | **8080**          |
| `SALESFORCE-INTEGRATION_SENTRY_DSN`  | The project DSN for Sentry events.                | false             |                   |
| `SALESFORCE-INTEGRATION_ENVIRONMENT` | The environment used for Sentry events or also for some functions that are required in dev or prod environment.           | false             | **dev**           |
| `SALESFORCE-INTEGRATION_MAIN_CONTEXT_TIME_OUT`| The time out for the main context in seconds.| false | **10** |
| `SALESFORCE-INTEGRATION_REDIS_ADDRESS`| The address of the Redis Instance (not sentinel).| Only if Redis Instance is not a sentinel Redis.| |
| `SALESFORCE-INTEGRATION_REDIS_MASTER`| The name of the master on a Redis Sentinel Instance (sentinel).| Only if Redis Instance is a sentinel Redis.| |
| `SALESFORCE-INTEGRATION_REDIS_SENTINEL_ADDRESS`| The address of the Redis Sentinel Instance (sentinel).| Only if Redis Instance is a sentinel Redis.| |
| `SALESFORCE-INTEGRATION_BOTRUNNER_URL`| **Sent-to** API URL, used to change the status of a user in the bot flow. | false | **http://botrunner** |
| `SALESFORCE-INTEGRATION_BOTRUNNER_TOKEN`| Access token if necessary to make requests to **Sent-to**.| false | |
| `SALESFORCE-INTEGRATION_BOTRUNNER_TIMEOUT`| Number of seconds to wait to send a request to **Sent-to**.| false | **4** |
| `SALESFORCE-INTEGRATION_BLOCKED_USER_STATE`| Status of the bot to send with **Botrunner Client** when a user is blocked by salesforce.| true | **whatsapp:from-sf-blocked,facebook:from-sf-blocked** |
| `SALESFORCE-INTEGRATION_TIMEOUT_STATE`| Status of the bot to send with **Botrunner Client** when the chat is rejected because there are no agents, because the waiting time for an agent in salesforce to accept the chat ended or there was an unknown error in the long polling.| true | **whatsapp:from-sf-timeout,facebook:from-sf-timeout** |
| `SALESFORCE-INTEGRATION_SUCCESS_STATE`| Status of the bot to send with **Botrunner Client** when the chat ended successfully between a user and an agent.| true | **whatsapp:from-sf-success,facebook:from-sf-success** |
| `SALESFORCE-INTEGRATION_YALO_USERNAME`| Username required to generate a JWT with YALO role through the `/authenticate` endpoint.| true | **yaloUser** |
| `SALESFORCE-INTEGRATION_YALO_PASSWORD`| Password required to generate a JWT with YALO role through the `/authenticate` endpoint. | true |  |
| `SALESFORCE-INTEGRATION_SALESFORCE_USERNAME`| Username required to generate a JWT with SALESFORCE role through the `/authenticate` endpoint.| true | **salesforceUser** |
| `SALESFORCE-INTEGRATION_SALESFORCE_PASSWORD`| Password required to generate a JWT with SALESFORCE role through the `/authenticate` endpoint. | true |  |
| `SALESFORCE-INTEGRATION_SECRET_KEY`| String required to sign the JWT that are created through the `/authenticate` endpoint.| true |  |
| `SALESFORCE-INTEGRATION_SFC_CLIENT_ID`| String with the value of the client ID to obtain an accesstoken and connect to the salesforce API.| true *(request it to salesforce team or client)* |  |
| `SALESFORCE-INTEGRATION_SFC_CLIENT_SECRET`| String with the value of the client Secret to obtain an accesstoken and connect to the salesforce API.| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_USERNAME`| String with the value of the username of the api user to obtain an accesstoken and connect to the salesforce API.| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_PASSWORD`| String with the value of the password of the api user to obtain an accesstoken and connect to the salesforce API.| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_SECURITY_TOKEN`| String with the value of the user's security token to obtain an accesstoken and connect to the salesforce API. This value must be obtained from the Salesforce console in the section *profile* -> *configuration* -> [*Reset security token*](https://onlinehelp.coveo.com/en/ces/7.0/administrator/getting_the_security_token_for_your_salesforce_account.htm) .| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_BASE_URL`| Value of the base URL to connect with the salesforce API to be able to create contacts, cases and image upload. | true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_CHAT_URL`| Base URL value to connect to Agent Live API to request chats. | true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_LOGIN_URL`| Authorization API URL value. | true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_API_VERSION`| The Salesforce API version for the request, we should preferably use the most current.| true *(request it to salesforce team or client)*  | **52** |
| `SALESFORCE-INTEGRATION_SFC_ORGANIZATION_ID`| organization_id identifier for chat requests.| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_ DEPLOYMENT_ID`| deployment_id identifier for chat requests.| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_RECORD_TYPE_ID`| record_type_id identifier for case requests.| true *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_ACCOUNT_RECORD_TYPE_ID`| record_type_id identifier for submitting requests to create person accounts in Salesforce. If this value does not exist, only the contact is created.| false *(request it to salesforce team or client)*  |  |
| `SALESFORCE-INTEGRATION_SFC_DEFAULT_BIRTH_DATE_ACCOUNT`| Indicates the default date to create personal accounts in salesforce, only if the send ACCOUNT_RECORD_TYPE exists, in case the client requests a special date we can change it.| false | **1921-01-01T00:00:00** |
| `SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT`| It contains a map with the custom values of the client to identify the `bot source flow`, such as the values of the case subject, the buttonIDs, and the ownerIds, you can see an example [here](https://www.notion.so/yalo/Envar-SfcSourceFlowBot-d5ea5a94fca34659996fbf1bf76e33db). | true  |  |
| `SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_FIELD`| Name of the parameter that defines the reason for the chat query, this value must be sent in the chat request within the extraData map. Required to redirect to the corresponding queue in Salesforce. | true   | ***source_flow_bot*** |
| `SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_FIELD`| Name of the parameter that defines the reason for the chat query, this value must be sent in the chat request within the extraData map. Required to redirect to the corresponding queue in Salesforce. | true   | ***source_flow_bot*** |
| `SALESFORCE-INTEGRATION_SFC_BLOCKED_CHAT_FIELD`| Name of the parameter that tells us if we should validate the BlockedChatYalo attribute created by Salesforce to block a user. With this we activate reject a chat if the user was blocked in Salesforce for any reason and send it to the user status blocked in the bot. | true | ***false*** |
| `SALESFORCE-INTEGRATION_SFC_CUSTOM_FIELDS_CASE`| Contains a value map with the customer's custom fields to create a case on your salesforce platform, you can see an example [here](https://www.notion.so/yalo/Envar-SfcSourceFlowBot-d5ea5a94fca34659996fbf1bf76e33db#a264a1c925b94e9e8a4bd23ac097a278). | false   |  |
| `SALESFORCE-INTEGRATION_SFC_CUSTOM_FIELDS_CONTACT`| Contains a value map with the contact's custom fields to create a contact on your salesforce platform. | false   |  |
| `SALESFORCE-INTEGRATION_TIMEZONE`| Contains a string value to define the timezone of the service. | false   | ***America/Mexico_City*** |
| `SALESFORCE-INTEGRATION_SEND_IMAGE_NAME_IN_MESSAGE`| Contains a boolean to define if the service should send the image's name in the chat, when the end-user upload an image. | false   | false |
| `SALESFORCE-INTEGRATION_SFC_CODE_PHONE_REMOVE`| Indicates the codes of the phones to be deleted if the phone number is greater than 10 digits. By default, the 521 and 52 corresponding to Mexico are eliminated, more codes can be added to this shipment, example: "521,52,54,57,1". | false   | ***521,52*** |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_WA_CHANNEL`| Type of channel for sending messages to the WhatsApp bot . | false  | **outgoing_webhook** |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_FB_CHANNEL`| Type of channel for sending messages to the Facebook bot . | false  | **passthrough** |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_WA_BOT_ID`| Bot ID of the WhatsApp bot . | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_FB_BOT_ID`| Bot ID of the Facebook bot . | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_WA_BOT_JWT`| Json Web Token of the WhatsApp bot for make requests to ***Integrations API*** . | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_FB_BOT_JWT`| Json Web Token of the Facebook bot for make requests to  ***Integrations API***. | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_BASE_URL`| Base URL of the  ***Integrations API*** to make requests. | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_SIGNATURE`| Security signature, which must be sent in the header of the request received in our webhook to validate that the signature received in the header matches this value. | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_WA_BOT_PHONE`| Phone number to register the whatsapp bot webhook in integrations API. | false  |  |
| `SALESFORCE-INTEGRATION_INTEGRATIONS_FB_BOT_PHONE`| Phone number to register the facebook bot webhook in integrations API. In this case the phone number is the facebookId of the bot page. | false  |  |
| `SALESFORCE-INTEGRATION_WEBHOOK_BASE_URL`| Url base of our webhooks where integrations channels will send us the messages received by the bot. | false  |  |
| `SALESFORCE-INTEGRATION_KEYWORDS_RESTART`| Reserved words of the bot to restart the flow, in case these words are received in the webhook, this integration if it detects that there is an active chat, this will be closed on the side of salesforce and the integration, this function is only active in the environment of development. | false  | ***coppelbot,regresar,reiniciar,restart*** |
| `SALESFORCE-INTEGRATION_STUDIO_NG_URL`| **Sent-to** API URL, used to change the status of a user in the bot flow with studiong. | false | **http://studiong** |
| `SALESFORCE-INTEGRATION_STUDIO_NG_TOKEN`| Access token of studiong if necessary to make requests to **Sent-to**.| false | |
| `SALESFORCE-INTEGRATION_STUDIO_NG_TIMEOUT`| Number of seconds to wait to send a request to **Sent-to** to studiong.| false | **4** |
| `SALESFORCE-INTEGRATION_SPEC_SCHEDULE`| It is the time to configure the cron that allows making a request to the Salesforce API, so that our token does not expire due to inactivity. By default every 59 min. Note: remember that the inactivity time is 2 hours for the token to expire.| false | **@every 59m** |

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
        - name: SALESFORCE-INTEGRATION_ENVIRONMENT
          value: "dev"
        - name: SALESFORCE-INTEGRATION_REDIS_MASTER
          value: "mymaster"
        - name: SALESFORCE-INTEGRATION_REDIS_SENTINEL_ADDRESS
          value: "{{redis-address}}"
        - name: SALESFORCE-INTEGRATION_BOTRUNNER_URL
          value: "http://botrunner.stage-1:3000"
        - name: SALESFORCE-INTEGRATION_BLOCKED_USER_STATE
          value: "whatsapp:from-sf-blocked,facebook:from-sf-blocked"
        - name: SALESFORCE-INTEGRATION_TIMEOUT_STATE
          value: "whatsapp:from-sf-timeout,facebook:from-sf-timeout"
        - name: SALESFORCE-INTEGRATION_SUCCESS_STATE
          value: "whatsapp:from-sf-success,facebook:from-sf-success"
        - name: SALESFORCE-INTEGRATION_YALO_USERNAME
          value: "{{yalo_user}}"
        - name: SALESFORCE-INTEGRATION_YALO_PASSWORD
          value: "{{yalo_user_password}}"
        - name: SALESFORCE-INTEGRATION_SALESFORCE_USERNAME
          value: "{{salesforce_user}}"
        - name: SALESFORCE-INTEGRATION_SALESFORCE_PASSWORD
          value: "{{salesforce_user_password}}"
        - name: SALESFORCE-INTEGRATION_SECRET_KEY
          value: "{{secret_key}}"
        - name: SALESFORCE-INTEGRATION_SFC_CLIENT_ID
          value: "{{client_id_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_CLIENT_SECRET
          value: "{{client_secret_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_USERNAME
          value: "api_user_salesforce_enviroment"
        - name: SALESFORCE-INTEGRATION_SFC_PASSWORD
          value: "{{api_user_password_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_SECURITY_TOKEN
          value: "{{api_user_security_token_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_BASE_URL
          value: "{{data_api_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_CHAT_URL
          value: "{{agent_live_api_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_LOGIN_URL
          value: "{{oauth_api_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_ORGANIZATION_ID
          value: "{{organization_id_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_DEPLOYMENT_ID
          value: "{{deployment_id_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_RECORD_TYPE_ID
          value: "{{record_type_id_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_ACCOUNT_RECORD_TYPE_ID
          value: "{{account_record_type_id_salesforce_enviroment}}"
        - name: SALESFORCE-INTEGRATION_SFC_CUSTOM_FIELDS_CASE
          value: "{{customs_fields_names_customer}}"
        - name: SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT
          value: 'default={"subject":"{{subject_case}}","providers":{"whatsapp":{"button_id":"{{button_id}}","owner_id":"{{owner_id}}"},"facebook":{"button_id":"{{button_id}}","owner_id":"{{owner_id}}"}}}'
        - name: SALESFORCE-INTEGRATION_SFC_BLOCKED_CHAT_FIELD
          value: "false"
        - name: SALESFORCE-INTEGRATION_SFC_CODE_PHONE_REMOVE
          value: "521,52"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_WA_CHANNEL
          value: "outgoing_webhook"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_WA_BOT_ID
          value: "coppel-wa-staging"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_FB_CHANNEL
          value: "passthrough"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_FB_BOT_ID
          value: "coppel-msn-staging"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_FB_BOT_JWT
          value: "{{fb_bot_jwt}}"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_WA_BOT_JWT
          value: "{{wa_bot_jwt}}"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_BASE_URL
          value: "https://api-staging2.yalochat.com/underdog-integrations-api"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_WA_BOT_PHONE
          value: "{{phone_number_wa_bot_with_prefix_+52}}"
        - name: SALESFORCE-INTEGRATION_INTEGRATIONS_FB_BOT_PHONE
          value: "{{page_id_fb_bot}}"
        - name: SALESFORCE-INTEGRATION_WEBHOOK_BASE_URL
          value: "{{deployment_url}}"
```

**Note:** *To see the previous steps in more detail, we can see the following document* [On-boarding part 2: Publishing an application](https://www.notion.so/On-boarding-part-2-Publishing-an-application-eac08ad3eaad435cb242340fe1a2bb98#2ba8af0501964491a730bb979fcd2ced)

## Who do I talk to? ##

* Gerardo Ezquerra Martín - **cat@underdog.mx**
* Armando Hernández Aguayo - **armando@yalochat.com**
