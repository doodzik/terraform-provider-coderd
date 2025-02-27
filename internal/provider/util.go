package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
)

var (
	nameValidRegex           = regexp.MustCompile("^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$")
	templateVersionNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(?:[_.-]{1}[a-zA-Z0-9]+)*$`)
	displayNameRegex         = regexp.MustCompile(`^[^\s](.*[^\s])?$`)
)

func PtrTo[T any](v T) *T {
	return &v
}

func PrintOrNull(v any) string {
	if v == nil {
		return "null"
	}
	switch value := v.(type) {
	case *int32:
		if value == nil {
			return "null"
		}
		return fmt.Sprintf("%d", *value)
	case *int64:
		if value == nil {
			return "null"
		}
		return fmt.Sprintf("%d", *value)
	case *string:
		if value == nil {
			return "null"
		}
		out := fmt.Sprintf("%q", *value)
		return out
	case *bool:
		if value == nil {
			return "null"
		}
		return fmt.Sprintf(`%t`, *value)
	case *[]string:
		if value == nil {
			return "null"
		}
		var result string
		for i, role := range *value {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%q", role)
		}
		return fmt.Sprintf("[%s]", result)

	default:
		panic(fmt.Errorf("unknown type in template: %T", value))
	}
}

func computeDirectoryHash(directory string) (string, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// memberDiff returns the members to add and remove from the group, given the current members and the planned members.
// plannedMembers is deliberately our custom type, as Terraform cannot automatically produce `[]uuid.UUID` from a set.
func memberDiff(curMembers []uuid.UUID, plannedMembers []UUID) (add, remove []string) {
	curSet := make(map[uuid.UUID]struct{}, len(curMembers))
	planSet := make(map[uuid.UUID]struct{}, len(plannedMembers))

	for _, userID := range curMembers {
		curSet[userID] = struct{}{}
	}
	for _, plannedUserID := range plannedMembers {
		planSet[plannedUserID.ValueUUID()] = struct{}{}
		if _, exists := curSet[plannedUserID.ValueUUID()]; !exists {
			add = append(add, plannedUserID.ValueString())
		}
	}
	for _, curUserID := range curMembers {
		if _, exists := planSet[curUserID]; !exists {
			remove = append(remove, curUserID.String())
		}
	}
	return add, remove
}
