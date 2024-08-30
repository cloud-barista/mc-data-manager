package models

import (
	"testing"
)

func TestTaskService(t *testing.T) {
	service := NewTaskService()

	// 1. CreateTask test
	task1 := Task{
		TaskID:      "task_1",
		TaskName:    "Data Backup",
		Description: "Backup data from source to destination.",
		Status:      "active",
		Schedule:    "0 2 * * *",
	}
	service.CreateTask(task1)

	// 2. GetTask test
	retrievedTask, err := service.GetTask("task_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrievedTask.TaskID != task1.TaskID {
		t.Errorf("expected task ID %s, got %s", task1.TaskID, retrievedTask.TaskID)
	}

	// 3. GetTaskList test
	tasks := service.GetTaskList()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	// 4. UpdateTask test
	task1.Status = "inactive"
	err = service.UpdateTask(task1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 5. GetTask test (after update)
	updatedTask, err := service.GetTask("task_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedTask.Status != "inactive" {
		t.Errorf("expected status inactive, got %s", updatedTask.Status)
	}

	// 6. DeleteTask test
	err = service.DeleteTask("task_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 7. GetTask test (after deletion)
	_, err = service.GetTask("task_1")
	if err == nil {
		t.Fatal("expected error, got none")
	}

	// 8. CreateFlow test
	flow := Flow{
		FlowID:   "flow_1",
		FlowName: "Daily Backup Flow",
		Tasks:    []Task{task1},
		Status:   "active",
	}
	service.CreateFlow(flow)

	// 9. GetFlow test
	retrievedFlow, err := service.GetFlow("flow_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrievedFlow.FlowID != flow.FlowID {
		t.Errorf("expected flow ID %s, got %s", flow.FlowID, retrievedFlow.FlowID)
	}

	// 10. GetFlowList test
	flows := service.GetFlowList()
	if len(flows) != 1 {
		t.Errorf("expected 1 flow, got %d", len(flows))
	}

	// 11. CreateSchedule test
	schedule := Schedule{
		ScheduleID: "schedule_1",
		FlowID:     flow.FlowID,
		Cron:       "0 2 * * *",
		Status:     "active",
	}
	service.CreateSchedule(schedule)

	// 12. GetSchedule test
	retrievedSchedule, err := service.GetSchedule("schedule_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrievedSchedule.ScheduleID != schedule.ScheduleID {
		t.Errorf("expected schedule ID %s, got %s", schedule.ScheduleID, retrievedSchedule.ScheduleID)
	}

	// 13. GetScheduleList test
	schedules := service.GetScheduleList()
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(schedules))
	}
}
