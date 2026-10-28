package storagecatalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Onurlulardan/harpy/tools/internal/strictjson"
)

func loadOwner(
	repository *os.Root,
	content []byte,
	include string,
	expectedName string,
) (Owner, int64, []string, error) {
	var version documentVersion
	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&version); err != nil {
		return Owner{}, 0, nil, ownerDecodeError(include, err)
	}
	switch version.Version {
	case inlineOwnerVersion:
		var document inlineOwnerDocument
		if err := strictjson.Decode(content, &document); err != nil {
			return Owner{}, 0, nil, ownerDecodeError(include, err)
		}
		return document.Owner, 0, nil, nil
	case modularOwnerVersion:
		var document modularOwnerDocument
		if err := strictjson.Decode(content, &document); err != nil {
			return Owner{}, 0, nil, ownerDecodeError(include, err)
		}
		return loadOwnerFragments(repository, document, expectedName)
	default:
		return Owner{}, 0, nil, fmt.Errorf(
			"unsupported storage owner document version %d", version.Version,
		)
	}
}

func loadOwnerFragments(
	repository *os.Root,
	document modularOwnerDocument,
	expectedName string,
) (Owner, int64, []string, error) {
	if document.Owner.Name != expectedName {
		return Owner{}, 0, nil, fmt.Errorf(
			"storage owner name %q differs from fragment name %q",
			document.Owner.Name, expectedName,
		)
	}
	includes, err := validateOwnerFragmentIncludes(expectedName, document.Includes)
	if err != nil {
		return Owner{}, 0, nil, err
	}
	owner := document.Owner
	var totalBytes int64
	for _, include := range includes {
		content, err := readBoundedFile(repository, include, maximumFragmentBytes)
		if err != nil {
			return Owner{}, 0, nil, fmt.Errorf("read storage owner fragment: %w", err)
		}
		totalBytes += int64(len(content))
		var fragment ownerFragmentDocument
		if err := strictjson.Decode(content, &fragment); err != nil {
			return Owner{}, 0, nil, ownerDecodeError(include, err)
		}
		if fragment.Version != ownerFragmentVersion {
			return Owner{}, 0, nil, fmt.Errorf(
				"unsupported storage owner fragment version %d", fragment.Version,
			)
		}
		if fragment.Owner != expectedName {
			return Owner{}, 0, nil, fmt.Errorf(
				"storage owner fragment %q declares owner %q", include, fragment.Owner,
			)
		}
		if err := mergeOwnerFragment(&owner, fragment); err != nil {
			return Owner{}, 0, nil, fmt.Errorf("merge storage owner fragment %q: %w", include, err)
		}
	}
	return owner, totalBytes, includes, nil
}

func ownerDecodeError(include string, err error) error {
	return fmt.Errorf("decode storage owner fragment %q: %w", include, err)
}
