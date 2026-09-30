                      Схема работы

                         CAMUNDA

-----------------------------------------------------------------
│  1. POST /process-definition/key/application-process/start    │
│                                                               │
│                         │                                     │
│                 Создаётся Process Instance                    │
│                         │                                     │
│                    Start Event                                │
│                         │                                     │
│              Task: Проверить заявку                           │
│                         │                                     │
│                         │                                     │
│                  Exclusive Gateway                            │
│                  "Сумма >= 1000?"                             │
│                         │                                     │
│                    ┌────┴────┐                                │
│                    │         │                                │
│                 true│         │false                          │
│                    │         │                                │
│              Одобрить      Отклонить                          │
│               заявку         заявку                           │
│                    │         │                                │
│                    └────┬────┘                                │
│                         │                                     │
│                      End Event                                │
│                  "Заявка обработана"                          │
│                                                               │
-----------------------------------------------------------------

curl -X POST \
  "http://localhost:8080/engine-rest/process-definition/key/application-process/start" \
  -H "Content-Type: application/json" \
  -d '{
    "businessKey": "bpmn-test-001",
    "variables": {
      "amount": {
        "value": 1500,
        "type": "Integer"
      }
    }
  }'