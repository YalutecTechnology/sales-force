
<a name="v1.1.6"></a>
## [v1.1.6](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.6..v1.1.5) (2022-04-29)

### Feat

* **Integration - Interconnection:** Decrease wait time in long polling. Decrease wait time in long polling  between each request to salesforce. Closes #UI-53
* **Integration - Interconnection:** Add pprof endpoints. Add profile endpoints to check api performance. Closes #UI-52


<a name="v1.1.5"></a>
## [v1.1.5](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.5..v1.1.4) (2022-03-31)

### Feat

* **Integration-interconnection:** Implement redis enterprise on salesforce and s1 | Staging. Add redis enterprise config connection. Closes #UI-27.

### Fix

* **Integration-interconnection:** Hotfix redis salesforce. Omit error redis nil to create chat. Closes #UI-37


<a name="v1.1.4"></a>
## [v1.1.4](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.4..v1.1.3) (2022-03-25)

### Feat

* **General improvemnts:** Show text when an image contains it. Closes #UI-36
* **Integration-interconnection:** End chat correctly. Change validation EndChat endpoint to validate interconnection in redis. Closes #UI-33
* **Integration-interconnection:** Improve handling redis errors. Add error redis handler in validate user function. Closes #UI-32


<a name="v1.1.3"></a>
## [v1.1.3](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.3..v1.1.2) (2022-03-02)

### Feat

* **Improve Salesforce implementation:** Upload files on cases. Upload files and add them to the case when there is an active chat and an attachment is sent on whatsapp and facebook. Closes #UI-30


<a name="v1.1.2"></a>
## [v1.1.2](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.2..v1.1.1) (2022-03-02)


<a name="v1.1.1"></a>
## [v1.1.1](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.1..v1.1.0) (2022-02-15)

### Feat

* **Improve Salesforce implementation:** Add reconnect on long-polling. Do a reconnection when a 503 http error occurs on long polling. Closes #UI-25
* **Integration-interconnection:** Redis singleton connection Salesforce. Change CreateManager to Redis singleton connection. Closes #UI-23.
* **Integration-interconnection:** Implement kafka. Add Kafka to the project to avoid lost messages. Closes #UI-8.

### Fix

* **Improve Salesforce implementation:** Adding producer to interconnections in restore. Change the order of the Kakfa producer declaration, to be done before restoring interconnects. Closes #UI-26


<a name="v1.1.0"></a>
## [v1.1.0](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v1.1.0..v0.1.9) (2022-02-03)

### Feat

* **Improve Salesforce implementation:** Fix error message on Salesforce clients. Correctly sending the response sent by the salesforce API on the errors returned, including datadog traces. Closes #UI-19
* **Improve Salesforce implementation:** Add userID and client on datadog traces Salesforce. Adding labels with values ​​such as sessionID, userID and client to more easily find errors in the traces. Closes #UI-18
* **Improve Salesforce implementation:** Add SIGTERM. The service must be able to close connections correctly when a shutdown occurs or the app ends. Closes #UI-6
* **Improve Salesforce implementation:** Avoid status messages on webhooks. We avoid status messages that integrations channels send when the bot is configured to send them. Closes #UI-10
* **Improve Salesforce implementation:** Handle errors on /chats/connect. Send a sentTo to the user when an error occurs when creating a chat with Salesforce. Closes #UI-2
* **Integration-api:**  Implement liveness and readiness check. Implements a readiness and liveness check. The kubernetes config must be able to check if the service is alive and try to recover, in addition to check if once is alive also is ready to receive incoming traffic. Closes #UI-4
* **Integration-envars:**  Add .env file as a example on the repo. Add .env example file to set envars. Closes #UI-5


<a name="v0.1.9"></a>
## [v0.1.9](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.9..v0.1.8) (2021-12-22)

### Feat

* **Validation - Bug fixing:** Add Consumed API Requests to Datadog. Add spans to requests made by salesforce API, botruuner, and integrations API http clients. Closes #CSF-224


<a name="v0.1.8"></a>
## [v0.1.8](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.8..v0.1.7) (2021-12-21)

### Chore

* Update gitignore to ignore main file
* Update Docker to run on Mac M1 (arm)

### Feat

* Change the way that we define the image's title that was sent by the end-user
* Add the capabitily to send custom fields when creating contacts Closes: DBR-1280
* Add the capability to set the timezone of the instance Closes DBR-1282
* Move predefined messages to env var. We create an envar where we move the predefined messages to change them at any time depending on each client. Closes #CSF-217
* **Validation - Bug fixing:** Add datadog to the project. Add datadog to the project, to be able to detect incidents in the endpoints of creating chat and webhooks, also in the sending of messages to Salesforce API and Integrations API Closes #CSF-220

### Fix

* **Validation - Bug fixing:** Remove from docker file the config to run on M1 processor. Closes #CSF-223
* **Validation - Bug fixing:** Fix missing custom messages, allow status and priority to be sent in extradata. Closes #CSF-222


<a name="v0.1.7"></a>
## [v0.1.7](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.7..v0.1.6) (2021-11-16)

### Chore

* add ratelimiter to Manager in tests

### Update

* adding rate limiter to channel consumption


<a name="v0.1.6"></a>
## [v0.1.6](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.6..v0.1.5) (2021-11-16)

### Chore

* **Validation-Bugfix:** Update documentation and userID to logs. Closes #CSF-206

### Feat

* **Validation-Bugfix:** Fix getContext and store context. Change the search to scan for the context and implement redis sets. Closes #CSF-207
* **Validation-Bugfix:** Retry sent messages. A variable of maximum number of retries is added, in case the sending of messages to a salesforce or integrations APi fails, we can send another message a number of times. Closes #CSF-214
* **Validation-Bugfix:** Sent  messages to go routines. Send salesforce messages with gorutines to the channel to make sending requests to API integrations faster. Closes #CSF-215
* **Validation-Bugfix:** Change dynamic key in redis to a static key in redis one per user. Add client to envar, change key redis to {{client}}:{{userID}}:interconnection for to store and  to retrieve interconnection. Closes #CSF-211

### Fix

* **Integration-interconnection:** Change to goroutine handle image. Add groutine when send image to salesforce. Closes #CSF-210
* **Validation-Bugfix:** Split handleInterconnection in 3 different interconnection. We separate the message reception channels for faster response speed. Closes #CSF-216
* **Validation-Bugfix:** Ignore audio template. Closes #CSF-209


<a name="v0.1.5"></a>
## [v0.1.5](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.5..v0.1.4) (2021-10-28)

### Feat

* **Endpoints:** endpoint to remove webhook from integrations. Closes #CSF-199
* **Integration - interconnections:** create setup url webhooks
* **Integration-interconnection:** Implement cron and api call to keep session alive. Add a cron to do api call to keep session alive. Closes #CSF-203
* **Validation-Bugfix:** Add env var with values to remove mx international code from phone number. Add SfcCodePhoneRemove envar, if exists it and phone number over 10 digits, remove phoneNumber international code Closes #CSF-204
* **Validation-Bugfix:** Move redis to a gorutine when doing chat connect. Move to a gorutine StoreInterconnectionInRedis. Closes #CSF-202

### Fix

* **Validation-Bugfix:** Fix finishChat endpoint. Update to the closed state of the interconnection in redis. Closes #CSF-200


<a name="v0.1.4"></a>
## [v0.1.4](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.4..v0.1.3) (2021-10-25)

### Feat

* **Integration-interconnection:** Default birthday date to env var. The environment variable SfcDefaultBirthDateAccount is added that defines a default date when creating personal accounts in Salesforce. Closes #CSF-201
* **Integration-interconnection:** Create library for studio ng and modify state varibles to make it work with it. Add library for studio ng to switch method to change state bot. Closes #CSF-194


<a name="v0.1.3"></a>
## [v0.1.3](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.3..v0.1.2) (2021-10-21)

### Feat

* **Integration-interconnection:** Add register webhook endpoint. Add register webhook endpoint and remove de createManager. Closes #CSF-196

### Fix

* **Integration-webhook:** Fixes when receiving requests from integrations. Fix response type message and save context concurrent. Closes #CSF-197


<a name="v0.1.2"></a>
## [v0.1.2](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.2..v0.1.1) (2021-10-14)

### Chore

* **Integration-Staging Deployment:** Create integration main documentation. Update the project README files with the necessary information to understand and build an instance of salesforce-integration, in addition to explaining the functions and services available. Closes #CSF-108

### Feat

* **Integration - Endpoints:** Add end chat endpoint. (New)
* **Integration-Library Salesforce:** Create Salesforce Account. Add the CreateAccount and SearchAccount services. Closes #CSF-178
* **Integration-interconnection:** Improve coverage and tests in package handlers. Add more unit test in package handlers . Closes #CSF-187
* **Integration-interconnection:** Improve coverage package helpers. Add more unit test in package helpers . Closes #CSF-186
* **Integration-interconnection:** Do composite when creating account. Change request to create account to composite request. Closes #CSF-185
* **Integration-interconnection:** Fix local cache from map to risttreto. Change local cache to ristretto library. Closes #CSF-183
* **Integration-interconnection:** Omit optional fields if not setted. Omit optional fields to others intregrations. Closes #CSF-193
* **Integration-interconnection:** Improve coverage package salesforce.go file. Add more unit test in package clients in salesforce.go file . Closes #CSF-188
* **Integration-interconnection:** Envar for workflow configuration. Add envar  to config create case workflow. Closes #CSF-177
* **Integration-interconnection:** Ignore duplicated messages. Add condition to ignore duplicate messages in webhook. Closes #CSF-176

### Fix

* **Integration-Backbone:** Fixing envar configurations, messages and cache. Add MessageCache again, "Tiempo de espera" messages were removed, cache time for duplicate incoming messages was increased, multiple file tests were fixed. Closes #CSF-180


<a name="v0.1.1"></a>
## [v0.1.1](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.1..v0.1.0-alpha) (2021-09-29)

### Feat

* **Integration-interconnection:** Create facebook passthrough integration. Add integration library Facebook  and flow webhook_fb interconnection. Add update status cache interconection redis in endChat Closes #CSF-171

### Fix

* **Integration-Library Salesforce:** Refresh Token fix. Moving return and fix not found reference. Closes #CSF-170


<a name="v0.1.0-alpha"></a>
## v0.1.0-alpha (2021-09-20)

### Feat

* **Integration - Backbone:** Saving Interconnection variables (Redis). We must save, update and delete from redis the interconnection between salesforce and the yalo bot. Closes #CSF-50
* **Integration - Backbone:** Switch buttonID for different issues. We must change the button when the origin of the bot is whatapp or facebook. Closes #CSF-159
* **Integration Enviroment:** Create skaffold configuration for stagging. Closes #CSF-23
* **Integration-Backbone:** Service to send images to Salesforce. Add endpoints to upload and associate an image to a case. Closes #CSF-60
* **Integration-Backbone:** Service to send message to Salesforce. Add endpoint to send a text message to Salesforce API. Closes #CSF-59
* **Integration-Backbone:** requester  to reconnect session for chat in  Salesforce. Add service to reconect session. Closes #CSF-56
* **Integration-Backbone:** requester  to create  contact  in  Salesfo Add service to create contact. Closes #CNF-150
* **Integration-Backbone:** Create main structure. Add project  base files : in-memory store library, proxy library, add endpoint healtcheck. Closes #CSF-43
* **Integration-Context:** Service to recover context from redis 24h. Add service to recover context of cache. Closes #CSF-143
* **Integration-Library Cache:**  Create main struct to save and read interaction. Add package to save and read interconnection Closes #CNF-52
* **Integration-Library Integrations:** Add integrations service library. Add integrations library with subscribe and unsubscribe webhook and send messages services. Closes #CSF-68
* **Integration-Library Salesforce:** Set to handle Salesforce messages (receiving and sending). We add start long polling service. Closes #CSF-65
* **Integration-Library Salesforce:** Set to Create Chat from Salesforce. We need to create an endpoint to create and initialize a chat with Salesforce. Closes #CSF-155
* **Integration-Library Salesforce:** requester  to end  chat  in  Salesforce. Add service to close chat. Closes #CSF-66
* **Integration-Library Salesforce:** Service to long polling messages from Salesforce. Add the request to obtain the agents' messages  of a chat, the security token is added to the request to obtain token. Closes #CSF-58
* **Integration-Library Salesforce:** Refresh Token when needed. We must update the salesforce library token when it returns a 401 error and we change the ownerID for the case according to the origin. Closes #CSF-169
* **Integration-Library Salesforce:** Service to retrieve contact from Salesforce. Add request for retrieve contactID from Salesforce and tests. Closes #CSF-151
* **Integration-Library Salesforce:** Set to request chat with Salesforce (Add Case) We add the caseId and the contactId to the chat request. Closes #CSF-158
* **Integration-Library Salesforce:** Set request to create chat sessions from Salesforce. Add request for create chat and tests. Closes #CSF-55
* **Integration-Manager:** Manage blocked user. If user is blocked from salesforce, we send it to the from-sf-blocked state. Closes #CSF-157
* **Integration-interconnection:** Interconnection for Message Handler(image). Add handler to send message to salesforce. Closes #CSF-82
* **Integration-interconnection:** Interconnection for Message Handler. Add handler to send message to salesforce . Closes #CSF-82
* **Integration-webhook:**  Set webhook in integrations. Add webhook endpoint  to save message form integrations Closes #CNF-140

