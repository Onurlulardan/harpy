package devjournal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func readRecords(path string) ([]journalRecord, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open development authority journal")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 512*1024)
	var records []journalRecord
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		canonical, err := canonicaljsonv1.CanonicalizeObject(line, 512*1024)
		if err != nil || !bytes.Equal(line, canonical) {
			return nil, fmt.Errorf("development authority journal is not canonical")
		}
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		var record journalRecord
		if decoder.Decode(&record) != nil || decoder.Decode(&struct{}{}) != io.EOF {
			return nil, fmt.Errorf("development authority journal record is invalid")
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read development authority journal")
	}
	return records, nil
}

func appendRecord(path string, record journalRecord) error {
	document, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode development authority record")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 512*1024)
	if err != nil {
		return fmt.Errorf("canonicalize development authority record")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("append development authority journal")
	}
	defer file.Close()
	if _, err := file.Write(append(canonical, '\n')); err != nil {
		return fmt.Errorf("write development authority journal")
	}
	return file.Sync()
}
