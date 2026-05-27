package cache

import "fmt"

func TaskKey(id int) string {
	return fmt.Sprintf("task:%d", id)
}

func TaskListByUserKey(userID int) string {
	return fmt.Sprintf("tasks:user:%d", userID)
}
