package basic

// Version function
func Version() string {
	return "<BUILD_ID>"
}

// BuildDate function
func BuildDate() string {
	return "<TIMESTAMP>"
}

// GitCommit function
func GitCommit() string {
	return "<GIT_COMMIT>"
}

// GitBranch function
func GitBranch() string {
	return "<GIT_BRANCH>"
}

// GitURL function
func GitURL() string {
	return "<GIT_URL>"
}
