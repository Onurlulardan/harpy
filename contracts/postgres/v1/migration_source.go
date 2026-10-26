package postgresv1

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maximumMigrationSourceBytes = 1024 * 1024

var migrationSourceBOM = []byte{0xef, 0xbb, 0xbf}

type MigrationSourceMetadata struct {
	SHA256          string
	TransactionMode string
}

func InspectMigrationSource(content []byte) (MigrationSourceMetadata, error) {
	if len(content) == 0 || len(content) > maximumMigrationSourceBytes {
		return MigrationSourceMetadata{}, fmt.Errorf("migration SQL size is invalid")
	}
	if !utf8.Valid(content) || bytes.HasPrefix(content, migrationSourceBOM) ||
		bytes.ContainsRune(content, '\r') || content[len(content)-1] != '\n' {
		return MigrationSourceMetadata{}, fmt.Errorf(
			"migration SQL must be UTF-8, BOM-free, LF-only, and newline-terminated",
		)
	}
	transactionMode, err := inspectMigrationDirectives(string(content))
	if err != nil {
		return MigrationSourceMetadata{}, err
	}
	if err := validateGooseStatementTermination(string(content)); err != nil {
		return MigrationSourceMetadata{}, err
	}
	digest := sha256.Sum256(content)
	return MigrationSourceMetadata{
		SHA256: hex.EncodeToString(digest[:]), TransactionMode: transactionMode,
	}, nil
}

func validateGooseStatementTermination(content string) error {
	started, inBlock, buffered, blockContent, statementSeen := false, false, false, false, false
	for _, line := range strings.Split(content, "\n") {
		switch line {
		case "-- +goose NO TRANSACTION":
			continue
		case "-- +goose Up":
			started = true
			continue
		case "-- +goose StatementBegin":
			inBlock, blockContent = true, false
			continue
		case "-- +goose StatementEnd":
			if !inBlock || !blockContent {
				return fmt.Errorf("Goose statement block is empty")
			}
			inBlock, buffered, statementSeen = false, false, true
			continue
		}
		trimmed := strings.TrimSpace(line)
		if !started {
			if trimmed == "" || strings.HasPrefix(trimmed, "--") {
				continue
			}
			return fmt.Errorf("migration SQL appears before Goose Up directive")
		}
		if inBlock {
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				blockContent = true
			}
			continue
		}
		if !buffered && (trimmed == "" || strings.HasPrefix(trimmed, "--")) {
			continue
		}
		buffered = true
		if gooseLineEndsWithSemicolon(line) {
			buffered, statementSeen = false, true
		}
	}
	if inBlock || buffered || !statementSeen {
		return fmt.Errorf("migration SQL is empty or has an unterminated Goose statement")
	}
	return nil
}

func gooseLineEndsWithSemicolon(line string) bool {
	previous := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		previous = word
	}
	return strings.HasSuffix(previous, ";")
}

func inspectMigrationDirectives(content string) (string, error) {
	upCount, noTransactionCount, statementDepth, statementBlocks := 0, 0, 0, 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") || !strings.Contains(line, "+goose") {
			continue
		}
		switch line {
		case "-- +goose Up":
			if statementDepth != 0 {
				return "", fmt.Errorf("Goose Up directive is inside a statement block")
			}
			upCount++
		case "-- +goose NO TRANSACTION":
			if upCount != 0 || statementDepth != 0 {
				return "", fmt.Errorf("Goose NO TRANSACTION directive is out of order")
			}
			noTransactionCount++
		case "-- +goose StatementBegin":
			if upCount != 1 || statementDepth != 0 || statementBlocks != 0 {
				return "", fmt.Errorf("Goose statement block is invalid")
			}
			statementDepth, statementBlocks = 1, 1
		case "-- +goose StatementEnd":
			if statementDepth != 1 {
				return "", fmt.Errorf("Goose statement block is unbalanced")
			}
			statementDepth = 0
		case "-- +goose Down":
			return "", fmt.Errorf("Goose Down directive is forbidden")
		default:
			return "", fmt.Errorf("unknown Goose directive")
		}
	}
	if upCount != 1 || statementDepth != 0 || noTransactionCount > 1 {
		return "", fmt.Errorf("Goose directives are invalid")
	}
	if noTransactionCount == 1 {
		return "non_transactional", nil
	}
	return "transactional", nil
}
