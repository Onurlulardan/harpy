package permissionrequest

import "github.com/google/uuid"

func (record DecisionRecord) Validate() error {
	if record.OperationID == uuid.Nil || record.RequestSHA256 == [32]byte{} ||
		record.GroupID == uuid.Nil || record.ReviewerID == uuid.Nil ||
		record.ReviewTaskID == uuid.Nil || record.Result.Replayed ||
		record.Result.Revision < 1 || record.Result.AuthorizationClock < 1 {
		return ErrInvalid
	}
	want := "iam.permission_request.approved"
	if record.Decision == DecisionReject {
		want = "iam.permission_request.rejected"
	} else if record.Decision != DecisionApprove {
		return ErrInvalid
	}
	if record.Result.Code != want {
		return ErrInvalid
	}
	return nil
}
