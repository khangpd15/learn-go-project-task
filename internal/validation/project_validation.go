package validation

func IsValidProjectName(name string) bool {
	return len(name) > 0
}

func IsValidProjectDescription(description string) bool {
	return len(description) <= 500
}

func IsValidProjectId(id int) bool {
	return id > 0
}

func IsValidProjectOwnerId(ownerId int) bool {
	return ownerId > 0
}
