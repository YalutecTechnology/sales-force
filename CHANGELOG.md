
<a name="v0.1.5"></a>
## [v0.1.5](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.5..v0.1.4) (2021-10-27)

### Feat

* **Integration - interconnections:** create setup url webhooks

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

