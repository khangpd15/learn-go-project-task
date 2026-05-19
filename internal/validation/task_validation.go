package validation

func IsValidStatus(status string) bool {
	validStatus := map[string]bool{
		"TODO":        true,
		"IN_PROGRESS": true,
		"DONE":        true,
	}
	return validStatus[status]
}

func IsValidId(id int) bool {

	return id > 0
}