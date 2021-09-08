
<a name="v0.0.1"></a>
## v0.0.1 (2021-09-06)

### Feat

* **Integration Enviroment:** Create skaffold configuration for stagging. Closes #CSF-23
* **Integration-Backbone:** requester  to create  contact  in  Salesfo Add service to create contact. Closes #CNF-150
* **Integration-Backbone:** Service to send images to Salesforce. Add endpoints to upload and associate an image to a case. Closes #CSF-60
* **Integration-Backbone:** Service to send message to Salesforce. Add endpoint to send a text message to Salesforce API. Closes #CSF-59
* **Integration-Backbone:** requester  to reconnect session for chat in  Salesforce. Add service to reconect session. Closes #CSF-56
* **Integration-Backbone:** Create main structure. Add project  base files : in-memory store library, proxy library, add endpoint healtcheck. Closes #CSF-43
* **Integration-Context:** Service to recover context from redis 24h. Add service to recover context of cache. Closes #CSF-143
* **Integration-Library Cache:**  Create main struct to save and read interaction. Add package to save and read interconnection Closes #CNF-52
* **Integration-Library Integrations:** Add integrations service library. Add integrations library with subscribe and unsubscribe webhook and send messages services. Closes #CSF-68
* **Integration-Library Salesforce:** requester  to end  chat  in  Salesforce. Add service to close chat. Closes #CSF-66
* **Integration-Library Salesforce:** Service to retrieve contact from Salesforce. Add request for retrieve contactID from Salesforce and tests. Closes #CSF-151
* **Integration-Library Salesforce:** Set to handle Salesforce messages (receiving and sending). We add start long polling service. Closes #CSF-65
* **Integration-Library Salesforce:** Set request to create chat sessions from Salesforce. Add request for create chat and tests. Closes #CSF-55
* **Integration-Library Salesforce:** Service to long polling messages from Salesforce. Add the request to obtain the agents' messages  of a chat, the security token is added to the request to obtain token. Closes #CSF-58
* **Integration-Library Salesforce:** Set to Create Chat from Salesforce. We need to create an endpoint to create and initialize a chat with Salesforce. Closes #CSF-155
* **Integration-interconnection:** Interconnection for Message Handler. Add handler to send message to salesforce . Closes #CSF-82
* **Integration-webhook:**  Set webhook in integrations. Add webhook endpoint  to save message form integrations Closes #CNF-140

