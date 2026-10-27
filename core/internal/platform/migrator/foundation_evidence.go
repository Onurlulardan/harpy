package migrator

import "time"

type foundationEvidence struct {
	receipts   []receiptEvidence
	extensions []extensionEvidence
}

type receiptEvidence struct {
	kind              string
	runID             string
	operationDigest   string
	canonicalDocument []byte
	canonicalSHA256   string
	signatureKeyID    string
	issuedAt          time.Time
}

type extensionEvidence struct {
	name              string
	version           string
	schemaName        string
	ownerRole         string
	controlFileSHA256 string
	capabilities      []string
	verifiedRunID     string
}
