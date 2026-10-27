package migration

import "fmt"

func validateDocument(document Document, expected expectedDocument) error {
	if document.owner != expected.owner {
		return fmt.Errorf("embedded owner differs from manifest")
	}
	if document.scope != expected.scope {
		return fmt.Errorf("embedded scope differs from manifest")
	}
	metadata, err := InspectDocument(document)
	if err != nil {
		return err
	}
	if metadata.SHA256 != expected.sha256 {
		return fmt.Errorf("raw SQL SHA-256 differs from manifest")
	}
	if metadata.TransactionMode != expected.transactionMode {
		return fmt.Errorf("Goose transaction directive differs from manifest")
	}
	return nil
}
