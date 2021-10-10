package basic

import (
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
)

// HomeDir returns the home directory
func HomeDir() string {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env)
}

// CallInfo type
type CallInfo struct {
	ProjectName string
	FileName    string
	PackageName string
	FuncName    string
	Line        int
}

// GetCallInfo returns the call info
func GetCallInfo(skip int) (*CallInfo, bool) {
	pc, fileName, line, ok := runtime.Caller(skip)
	if !ok {
		return nil, false
	}

	fullfunctionName := runtime.FuncForPC(pc).Name()

	segments := strings.Split(fullfunctionName, "/")
	projectName := segments[2]
	segment := segments[len(segments)-1]
	names := strings.Split(segment, ".")

	packageName := names[0]
	funcName := names[1]

	return &CallInfo{
		ProjectName: projectName,
		FileName:    fileName,
		Line:        line,
		PackageName: packageName,
		FuncName:    funcName,
	}, true
}

var checkStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

// IsStringAlphanumeric checks the characters are valid for an ID
func IsStringAlphanumeric(s string) bool {
	return checkStringAlphabetic(s)
}

// CheckCharactersInID checks the characters are valid for an ID
func CheckCharactersInID(s string) error {
	for _, r := range s {

		ok := false
		if unicode.IsLetter(r) {
			ok = true
		} else if unicode.IsDigit(r) {
			ok = true
		}

		if !ok {
			return codeerror.NewBadRequest("Invalid ID")
		}
	}
	return nil
}

// Contains function
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// CheckStringArraysAreEqual tells whether a and b contain the same elements NOT in-order order
func CheckStringArraysAreEqual(x, y []string) bool {

	if x == nil {
		return y == nil
	} else if y == nil {
		return false
	}

	if len(x) != len(y) {
		return false
	}

	xMap := make(map[string]int)
	yMap := make(map[string]int)

	for _, xElem := range x {
		xMap[xElem]++
	}
	for _, yElem := range y {
		yMap[yElem]++
	}

	for xMapKey, xMapVal := range xMap {
		if yMap[xMapKey] != xMapVal {
			return false
		}
	}
	return true
}

// CheckStringArraysAreEqualInOrder tells whether a and b contain the same elements IN order
func CheckStringArraysAreEqualInOrder(x, y []string) bool {

	if len(x) != len(y) {
		return false
	}
	for i, v := range x {
		if v != y[i] {
			return false
		}
	}
	return true
}

// Quote makes the string safe for database input
func Quote(s string) string {
	if CheckSubstrings(s, "'", `"`, "`", "\\") {
		tag := "$@@@$"
		return "'" + tag + s + tag + "'"
	}
	return "'" + s + "'"
}

// CheckSubstrings checks a string for substrings
func CheckSubstrings(str string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}

func GetEnvInteger(name string, def int) (int, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return def, nil
	}
	return strconv.Atoi(value)
}

func GetEnvString(name string, def string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return def, nil
	}
	return value, nil
}
