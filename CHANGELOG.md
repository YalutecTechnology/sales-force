
<a name="v0.1.2"></a>
## [v0.1.2](https://bitbucket.org-eduardoochoa/yalochat/salesforce-integration/compare/v0.1.2..v0.1.1) (2021-10-03)

### Feat

* **Integration-Library Salesforce:** Create Salesforce Account. Add the CreateAccount and SearchAccount services. Closes #CSF-178
* **Integration-interconnection:** Envar for workflow configuration. Add envar  to config create case workflow. Closes #CSF-177
* **Integration-interconnection:** Ignore duplicated messages. Add condition to ignore duplicate messages in webhook. Closes #CSF-176


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

