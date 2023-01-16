# Salesforce-Integration Runbook

| Description | How to recover Salesforce-Integrations app after an error, it might be related to an internal or external issue |
| --- | --- |
| Created At | January 27, 2021, 9:00 AM |
| Created By | Underdog @Uriel Díaz @Eduardo Ochoa Martínez  |
| Updated By | @Nathali Aguayo |

This Runbook describes how to recover Salesforce - Integrations middleware when an error occurs with any resource involved.

# Troubleshooting

## Recover when Redis fails

When Redis fails, please follow the next steps

1. Please contact Ops ASAP, the first step is to notify that the service is failing because it could be a general issue that can affect many projects so just to keep everyone synced, try to tell your leader to create a ticket.
2. Let’s take a look at the log to recognize the error, if it is something related to a connection issue like the next,
    
    ```go
    dial tcp XX.XX.XXX.XXX:6379: connect: no route to host
    ```
    
    Visit the rancher2 site where the application is allocated (`redis-sentinel-node`).
    
    For this implementation there are three Pods, let’s confirm if each one is up and working correctly by having a look into the logs. If for some reason at least one have a similar restarting problem like the next:
    
    ```go
    CrashLoopBackOff: back-off 5m0s restarting failed container= redis pod=redis-sentinel-node-2
    ```
    Example Image of this error in Rancher

   ![rancher-error.png](/docs/images/rancher-error.png)

    The recommendation to restart Redis ASAP is to set all the replicas to 0, this will remove all the sentinel nodes so the next step is to reload new ones, for that it is necessary to set replicas to three again to have new sentinel nodes ready just in seconds.
    

## Recover when no messages are received from Integrations

When messages are missing from Integrations Core, please keep the following steps:

1) Please contact Platform Team ASAP, the first step is to notify that the service is failing because it could be already tracked so just to keep everyone synced, try to look for Alex Baquiax, Juan Flores, or Manuel Cordón.

2) Confirm what microservice is failing, could be Dispatcher, Freeway, API, or Channel, just look into the logs to recover the error and have some clue at hand.

3) Confirm with Integrations Team if the version of the Webhooks is correct, sometimes should be a lower version to work correctly this is due to release control that has to be followed.

- Webhook document example
    
    ![webhook-document.png](/docs/images/webhook-document.png)
    

## Restart the middleware when it was shot down

This deployment is running as a [K8s Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/).

Just to have a quick look, let’s check if `salesforce-integration` pod is running on K8s, following the next command:

```bash
kubectl get pods

# Sample output, do not copy
salesforce-integration-8d5b456d4-gpzhv                            1/1     Running   0          3h35m
```

If for some reason does not appear on the list or the status is down, here are a few steps to follow in order to start it up again:

1. Visit the rancher2 site at `workloads`.
    
2. Now look for `salesforce-integration` deployment and check the state of the pod just to confirm.
3. Go to the logs, click on the three points button, and select *View Logs* to track the issue and have evidence of what the instance knocked down
4. After tracking the error now it is time to restart the instance. Go to the three points button at the top of the page and look for Redeploy option similar to the next image
    
    ![rancher-redeploy.png](/docs/images/rancher-redeploy.png)
    

**NOTE**: If time is not in your favor please skip the number three go directly to the fourth step and try to restart the pod.

## Recover when Kafka fails

When messages are missing from Kafka due to an error with the connection o something else, please follow the steps below:

1) Please contact DevOps Team ASAP, the first step is to notify that the service is failing because it could be already tracked so just to keep everyone synced.

2) If the Dev has access to Prod Confluent account, try to find if the group and the topics already 

exist, if not then we need to create them with the required settings, partitions, retention time and so on.

![confluent-topics.png](/docs/images/confluent-topics.png)

3) If everything is alright with the topics then let’s have a look into the messages and observe the traffic that they are/were receiving, something that makes sense to the regular load.

![confluent-topics2.png](/docs/images/confluent-topics2.png)

4) If you don’t have access to Confluent, please come back to the step one, and ask for the access.

# Common Create Chat Errors

If the customer reports that a chat cannot be created, this can be for several reasons, mainly a change in the configuration of the integration user or a change in some configuration.

Therefore, the first steps to follow if the customer reports that a chat is failing are to go to the Datadog dashboards to identify the problem.

## What to do when the Integration is returning the next error: `could not create chat in salesforce: this contact is blocked`

![blocked-contact.png](/docs/images/blocked-contact.png)

This means that an agent or admin of the Salesforce platform on the client side blocked this user. So it must be unblocked on the Salesforce side.

If the problem persists validate with Salesforce that the user by email or phone number has this property in the client DB `CP_BlockedChatYalo__c = false`.

## What to do when the Integration is returning the next error: `INVALID_OPERATION_WITH_EXPIRED_PASSWORD`

This means for some reason the client has changed the Salesforce’s password, to fix this we have to follow the steps below:

1. Contact Client Admin, if is the case of Coppel you can contact: [ricardo.olivas@coppel.com](mailto:ricardo.olivas@coppel.com), describe the situation and ask for a new access to Salesforce.
2. First locate the salesforce user in the integrations who you can be found in the following env var: **SALESFORCE-INTEGRATION_SFC_USERNAME.**
3. Once we have the user, the next step is provide it to the Client Admin so that he should send a mail in order to be able to reset the password and as a result a new **security_token** will be generated. 
    
    **NOTE**: It is important to provide the mail account where the reset-password mail will be sent.
    
4. Once we have the new **password** and **security_token** values we must update those secrets values in *Google Secret Manager*, the Devops team must be asked to update the secrets for the following environments of the instance:
    
    ```yaml
    SALESFORCE-INTEGRATION_SFC_PASSWORD
    SALESFORCE-INTEGRATION_SFC_SECURITY_TOKEN
    ```
    
5. Once the corresponding secrets have been updated in the instance then synchronize in ArgoCD and redeploy the instance in Rancher.
    
    ![rancher-redeploy2.png](/docs/images/rancher-redeploy2.png)
    
6. The final step is to review on Datadog if the chats are creating successfully, for that visit the Dashboard page.
    

## What to do when the Integration is returning the next error:   `INVALID_AUTH_HEADER message:INVALID_HEADER_TYPE`

If we get this error it could mean that the salesforce API access token was not generated correctly and is sending a "" value as a token.

![invalid-header-type.png](/docs/images/invalid-header-type.png)

First let's validate that an empty token is sent. For this we go to the following dashboard: in order to can validate that in a request with error a header without token is being sent:

![header-without-token.png](/docs/images/header-without-token.png)

If a Bearer is sent without the token, check the section **Error Getting Token**

## Error Getting Token

Validate in the Datadog dashboard with the operation name **get_token** if we get errors 400 as in the following image means that the access token for the integration is not being generated, then check that the **username** and **password** of the integration user are correct, as well as the **client_secret** and **client_id**, have not been updated.

![datadog-dashboard.png](/docs/images/datadog-dashboard.png)

We can also validate the data sent in the payload as in the following image, and validate the data with the client.

![token-error-traces.png](/docs/images/token-error-traces.png)

If the data is correct or we want to make a test, we can make a request to obtain a token with postman to validate what error it gives us, for example:

```bash
curl --location --request POST '[https://login.salesforce.com/services/oauth2/token](https://login.salesforce.com/services/oauth2/token)' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'grant_type=password' \
--data-urlencode 'client_id=${client_id}' \
--data-urlencode 'client_secret=${client_secret}' \
--data-urlencode 'username=${username}' \
--data-urlencode 'password=${password}'
```

**Note**: The values must be replaced with the values of the payload that are in the Datadog registry. Also validate this URL [https://login.salesforce.com](https://login.salesforce.com/) , you can change by client or environment.

If the error returned is the one in the image, it means that the **password**, **security_token** or **username** is invalid:

```json
{
    "error": "invalid_grant",
    "error_description": "authentication failure"
}
```

If you get the following error, it is because the client_secret is not the one you are sending

```json
{
    "error": "invalid_client",
    "error_description": "invalid client credentials"
}
```

If the following is displayed, the client_id is not correct.

```json
{
    "error": "invalid_client_id",
    "error_description": "client identifier invalid"
}
```

# Long Polling

Below are the possible errors that you can see in the execution of the long polling. 

### What is Long Polling?

This is a routine that makes requests to a Salesforce API constantly to review messages and events of a chat, if a chat was created through the endpoint **chats/connect** does not mean that the chat is already being attended by any of the agents, with this long polling service validates whether the chat was rejected, successful or there is an agent attending.

## What to do when the Integration is returning the next error: `[403] - Error call with status : Session required but was invalid or not found`

![long-polling-error-traces.png](/docs/images/long-polling-error-traces.png)

This error means that the chat ended because the session expired or for some reason was deleted from Salesforce.
This error can have many reasons why it happens, here are the possible reasons:

- If it is peak hour it may be that it takes goroutines more than 45s to make requests to Salesforce **/get/messages** service, but this bug has been fixed before.
- It may be that it is peak time, and the session expires automatically after a long time of waiting for the agent. You have to ask the client admin about that.
- The session expired or for some reason was removed from the Salesforce side, due to maintenance or overcoming some recent configuration or changes, etc.

If this error occurs, the end-user will be returned to the **state time out** with the message that they cannot be attended at the moment.

If this error occurs for a few users it may be a momentary API error, but if it is very constant for all users or **more than half of the chats**, you should look for the source of the problem, first validating with Salesforce if there is an error reported on their side for the client.

## What to do when the Integration is returning the next error: `[503] - Error call with status :`

![error-call-with-status-503.png](/docs/images/error-call-with-status-503.png)

This error means that the Salesforce API responds with a 503 Service Unavailable code, possibly the Salesforce API had an error at that time, if it happens in many long-pollings in a row may be that the service is momentarily down if it responds with 503 to all requests for more than 5 min, you should contact the Salesforce team to validate what is happening.

## What to do when the Integration is returning the next error: `Event [ChatRequestFail] : [Unavailable]`

![chatRequestFail-error.png](/docs/images/chatRequestFail-error.png)

This in itself is not an error, there are times when the chats are created, but do not reach the agent and show this message in the long-polling, it may mean the following:

- That there are no agents available, in the attention queue.
- If there are agents available and they do not arrive, possibly the agents are not assigned to the queue destination of the chat. So the first step is to check which queue the chat is sent to and then the Salesforce admin validates that the agents are associated with this destination queue. If they are not assigned they should be. If they are, check with the Salesforce team, possibly the chat routing is failing.

## What to do when the Integration is returning the next error: `[500] - Error call with status : Internal Server Error`

![internal-server-error.png](/docs/images/internal-server-error.png)

This error means that the Salesforce API responds with a `501` Internal Server Error code, possibly the Salesforce API had an error at that time, if it happens in many long-pollings in a row may be that the service was momentarily down, if it responds with `503` to all requests for more than 5 min, you should contact the Salesforce team to validate what is happening.

## What to do when the Integration is returning the next error: `Event [ChatRequestFail] : [Blocked]`

There are times when the chats are created but do not reach the agent and show this message in the long-polling, this error has happened twice in Coppel, which means that the integration user is blocked to make requests, it may have been blocked by mistake or some change in permission policies affected the integration user.

To solve it we must notify the Salesforce platform admin and ask him to unblock it or allow the user to make a request to the Salesforce Chat Agent Live API.

To get the integration user to let's look for the value of this environment variable: **SALESFORCE-INTEGRATION_SFC_USERNAME.**

## What to do when we do not receive LiveChat API events and the chats are not created:

If the chats do not reach Salesforce and the users of the WA and FB bots get stuck waiting for an agent, we check Datadog but there are no error messages, we do the following:

1.- We check interconection.longPolling at the **Datadog Dashboard** operation to validate if there are errors.
2.- If there are no errors but all the messages that respond to us are 204 http status and there are no events.

- We need to report to the Salesforce team or the customer if the LiveAgent API url has not changed on the customer instance. If so, update the environment variable: **SALESFORCE-INTEGRATION_SFC_CHAT_URL**
- If it is correct, report to the Salesforce team that the client's LiveChatApi does not respond to us with any Chat events.
We must share the LiveChatApi url and share the request and response of what responds to us when creating a sessionID.

Example:

``` JSON
GET 'https://d.la4-c1-ia4.salesforceliveagent.com/chat/rest/System/SessionId'
 
{"msg":"Get Session sucessfully","response":{"clientPollTimeout":40,"key":"${key}","affinityToken":"${affinityToken}","id":"${id}"},"severity":"info","time":"2022-03-21T21:31:27Z"}
```