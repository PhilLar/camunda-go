                      Схема работы 
                        
                         CAMUNDA
*---------------------------------------------------------------*
│                                                               │
│  1. POST /process-definition/key/application-process-go/start │
│                         │                                     │
│                 Создаётся Process Instance                    │
│                         │                                     │
│                    Start Event                                │
│                         │                                     │
│              Service Task: Проверить заявку                   │
│                         │                                     │
│                         │ topic = "check-application"         │
│                  External Task                                │
│                  ┌─────────────┐                              │
│                  │ WAIT / LOCK │<--------------               │               
│                  └─────────────┘               │              │
│                                                │              │
*---------------------------------------------------------------*
                                                 │
                                   fetchAndLock  │
                                                 │
                    *--------------------------------------*
                    │              GO WORKER               │
                    │                                      │
                    │  2. "Забирает задачу из              │
                    │      topic=check-application"        │
                    │             |                        │
                    │  3. Получает amount                  │
                    │             │                        │
                    │          approved?                   │
                    │             │                        │
                    │  4. POST /external-task/{id}/complete|
                    │             │                        │
                    *--------------------------------------*
                                  │
                         CAMUNDA получает approved
                                  │
                         Exclusive Gateway
                       "Заявка одобрена?"
                          │           │
                     true │           │ false 
                    Одобрить       Отклонить
                    заявку         заявку
                              │
                           End Event
                       "Заявка обработана"


1. curl -X POST \
  "http://localhost:8080/engine-rest/process-definition/key/application-process-go/start" \
  -H "Content-Type: application/json" \
  -d '{
    "businessKey": "go-test-001",
    "variables": {
      "amount": {
        "value": 1500,
        "type": "Integer"
      }
    }
  }'

2. curl -X POST "http://localhost:8080/engine-rest/external-task/fetchAndLock" \
  -H "Content-Type: application/json" \
  -d '{
    "workerId": "go-worker-1",
    "maxTasks": 1,
    "usePriority": false,
    "asyncResponseTimeout": 5000,
    "topics": [
      {
        "topicName": "check-application",
        "lockDuration": 10000,
        "variables": [
          "amount"
        ]
      }
    ]
  }'

4. curl -X POST "http://localhost:8080/engine-rest/external-task/780e664b-bcc3-11f1-896d-8a40c87aa5ce/complete" \
  -H "Content-Type: application/json" \
  -d '{
    "workerId": "go-worker-1",
    "variables": {
      "approved": {
        "value": true,
        "type": "Boolean"
      }
    }
  }'


Посмотреть результаты обработки

1. История данных процесса

curl "http://localhost:8080/engine-rest/history/variable-instance?processInstanceId=c887d319-bcd0-11f1-896d-8a40c87aa5ce" (processInstanceId берется из ответа /start)

Пример:
[
    {
        "type": "Integer",
        "value": 1500,
        "valueInfo": {},
        "id": "c887fa2a-bcd0-11f1-896d-8a40c87aa5ce",
        "name": "amount",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "caseDefinitionKey": null,
        "caseDefinitionId": null,
        "caseInstanceId": null,
        "caseExecutionId": null,
        "taskId": null,
        "errorMessage": null,
        "tenantId": null,
        "state": "CREATED",
        "createTime": "2026-09-30T13:13:52.877+0000",
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    },
    {
        "type": "Boolean",
        "value": true,
        "valueInfo": {},
        "id": "c889f601-bcd0-11f1-896d-8a40c87aa5ce",
        "name": "approved",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "caseDefinitionKey": null,
        "caseDefinitionId": null,
        "caseInstanceId": null,
        "caseExecutionId": null,
        "taskId": null,
        "errorMessage": null,
        "tenantId": null,
        "state": "CREATED",
        "createTime": "2026-09-30T13:13:52.890+0000",
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    }
]

2. История обработки процесса

curl "http://localhost:8080/engine-rest/history/activity-instance?processInstanceId=c887d319-bcd0-11f1-896d-8a40c87aa5ce"

Пример:
[
    {
        "id": "Activity_1976ibf:c88a4425-bcd0-11f1-896d-8a40c87aa5ce",
        "parentActivityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityId": "Activity_1976ibf",
        "activityName": "Одобрить заявку",
        "activityType": "task",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "taskId": null,
        "calledProcessInstanceId": null,
        "calledCaseInstanceId": null,
        "assignee": null,
        "startTime": "2026-09-30T13:13:52.892+0000",
        "endTime": "2026-09-30T13:13:52.892+0000",
        "durationInMillis": 0,
        "canceled": false,
        "completeScope": false,
        "tenantId": null,
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    },
    {
        "id": "Activity_1ej9p8f:c888484e-bcd0-11f1-896d-8a40c87aa5ce",
        "parentActivityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityId": "Activity_1ej9p8f",
        "activityName": "Проверить заявку",
        "activityType": "serviceTask",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c888484d-bcd0-11f1-896d-8a40c87aa5ce",
        "taskId": null,
        "calledProcessInstanceId": null,
        "calledCaseInstanceId": null,
        "assignee": null,
        "startTime": "2026-09-30T13:13:52.879+0000",
        "endTime": "2026-09-30T13:13:52.891+0000",
        "durationInMillis": 12,
        "canceled": false,
        "completeScope": false,
        "tenantId": null,
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    },
    {
        "id": "Event_16zvubf:c88a4426-bcd0-11f1-896d-8a40c87aa5ce",
        "parentActivityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityId": "Event_16zvubf",
        "activityName": "Заявка обработана",
        "activityType": "noneEndEvent",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "taskId": null,
        "calledProcessInstanceId": null,
        "calledCaseInstanceId": null,
        "assignee": null,
        "startTime": "2026-09-30T13:13:52.892+0000",
        "endTime": "2026-09-30T13:13:52.892+0000",
        "durationInMillis": 0,
        "canceled": false,
        "completeScope": true,
        "tenantId": null,
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    },
    {
        "id": "Gateway_0xgqq3h:c88a1d14-bcd0-11f1-896d-8a40c87aa5ce",
        "parentActivityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityId": "Gateway_0xgqq3h",
        "activityName": "Заявка одобрена?",
        "activityType": "exclusiveGateway",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "taskId": null,
        "calledProcessInstanceId": null,
        "calledCaseInstanceId": null,
        "assignee": null,
        "startTime": "2026-09-30T13:13:52.891+0000",
        "endTime": "2026-09-30T13:13:52.892+0000",
        "durationInMillis": 1,
        "canceled": false,
        "completeScope": false,
        "tenantId": null,
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    },
    {
        "id": "StartEvent_1:c888213c-bcd0-11f1-896d-8a40c87aa5ce",
        "parentActivityInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "activityId": "StartEvent_1",
        "activityName": "Заявка получена",
        "activityType": "startEvent",
        "processDefinitionKey": "application-process-go",
        "processDefinitionId": "application-process-go:1:44390d24-bcc3-11f1-896d-8a40c87aa5ce",
        "processInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "executionId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce",
        "taskId": null,
        "calledProcessInstanceId": null,
        "calledCaseInstanceId": null,
        "assignee": null,
        "startTime": "2026-09-30T13:13:52.878+0000",
        "endTime": "2026-09-30T13:13:52.878+0000",
        "durationInMillis": 0,
        "canceled": false,
        "completeScope": false,
        "tenantId": null,
        "removalTime": "2026-10-30T13:13:52.892+0000",
        "rootProcessInstanceId": "c887d319-bcd0-11f1-896d-8a40c87aa5ce"
    }
]