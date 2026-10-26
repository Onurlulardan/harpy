package componentowner

func legacyReceiptMatches(receipt Receipt, command Command) bool {
	return receipt.ID != "" && receipt.ID != command.OperationID &&
		receipt.OperationID == command.OperationID && receipt.Component == command.Component &&
		receipt.Action == command.Action && receipt.Ready && !receipt.CompletedAt.IsZero()
}

func errorOr(err, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
