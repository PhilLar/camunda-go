package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const camundaURL = "http://localhost:8080/engine-rest"

type FetchAndLockRequest struct {
	WorkerID string `json:"workerId"`
	// Максимальное количество задач,
	// которое мы хотим получить за один запрос
	MaxTasks int `json:"maxTasks"`
	// Использовать ли при выборе задач приоритет
	UsePriority bool           `json:"usePriority"`
	Topics      []TopicRequest `json:"topics"`
	// Если задач сейчас нет, Camunda может подождать
	// прежде чем вернуть пустой результат
	AsyncResponseTimeout int `json:"asyncResponseTimeout"`
}

type TopicRequest struct {
	TopicName string `json:"topicName"`
	// На сколько миллисекунд Camunda блокирует задачу
	// Пока задача заблокирована этим worker,
	// другой worker не должен её получить
	LockDuration int64 `json:"lockDuration"`
	// Переменные процесса (amount)
	Variables []string `json:"variables"`
}

type ExternalTask struct {
	// ID самой External Task в Camunda
	ID string `json:"id"`
	// Topic задачи
	TopicName string `json:"topicName"`
	WorkerID  string `json:"workerId"`
	// ID экземпляра процесса.
	ProcessInstanceID string `json:"processInstanceId"`
	// Key определения процесса
	ProcessDefinitionKey string `json:"processDefinitionKey"`
	// Business Key нашего бизнес-объекта
	BusinessKey string `json:"businessKey"`
	// Переменные процесса.
	// "amount": {
	//     "type": "Integer",
	//     "value": 1500
	// }
	Variables map[string]Variable `json:"variables"`
}

type Variable struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func main() {
	workerID := "go-worker-1"

	for {
		// Пытаемся получить одну External Task.
		task, err := fetchTask(workerID)
		if err != nil {
			fmt.Println("fetch error:", err)
			continue
		}

		if task == nil {
			continue
		}

		fmt.Println("получили External Task:", task.ID)
		fmt.Println("businessKey:", task.BusinessKey)
		fmt.Println("processInstanceId:", task.ProcessInstanceID)

		amount := int(task.Variables["amount"].Value.(float64))

		fmt.Println("amount:", amount)

		approved := amount >= 1000

		fmt.Println("approved:", approved)

		err = completeTask(workerID, task.ID, approved)
		if err != nil {
			fmt.Println("complete error:", err)
			continue
		}

		fmt.Println("External Task завершена")
	}
}

// fetchTask забирает задачу из Camunda
//
// POST /external-task/fetchAndLock ->
// Camunda ->
// находит External Task ->
// блокирует её за нашим worker
func fetchTask(workerID string) (*ExternalTask, error) {
	reqBody := FetchAndLockRequest{
		WorkerID:    workerID,
		MaxTasks:    1,
		UsePriority: false,
		// Ждать задачу до 5 секунд,
		// если её прямо сейчас нет.
		AsyncResponseTimeout: 5000,
		Topics: []TopicRequest{
			{
				TopicName: "check-application",
				// Заблокировать задачу за worker
				LockDuration: 10000,
				// Camunda должна прислать
				// переменную amount.
				Variables: []string{"amount"},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		camundaURL+"/external-task/fetchAndLock",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"Camunda returned %s: %s",
			resp.Status,
			string(data),
		)
	}

	var tasks []ExternalTask

	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	// 	MaxTasks: 1 - так что придет только одна задача
	return &tasks[0], nil
}

// completeTask заканчивает обработку задачи и присылает результат в Camunda
func completeTask(workerID string, taskID string, approved bool) error {
	body := map[string]interface{}{
		"workerId": workerID,
		"variables": map[string]interface{}{
			"approved": map[string]interface{}{
				"value": approved,
				"type":  "Boolean",
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		camundaURL+"/external-task/"+taskID+"/complete",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"Camunda returned %s: %s",
			resp.Status,
			string(data),
		)
	}

	return nil
}
