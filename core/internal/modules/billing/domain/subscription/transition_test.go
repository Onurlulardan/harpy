package subscription

import "testing"

func TestTransitionMatrix(t *testing.T) {
	states := []State{Pending, Trialing, Active, PastDue, Paused, Cancelled, Expired}
	allowed := map[[2]State]bool{
		{Pending, Trialing}: true, {Pending, Active}: true, {Pending, Cancelled}: true,
		{Trialing, Active}: true, {Trialing, PastDue}: true,
		{Trialing, Cancelled}: true, {Trialing, Expired}: true,
		{Active, PastDue}: true, {Active, Paused}: true,
		{Active, Cancelled}: true, {Active, Expired}: true,
		{PastDue, Active}: true, {PastDue, Paused}: true,
		{PastDue, Cancelled}: true, {PastDue, Expired}: true,
		{Paused, Active}: true, {Paused, Cancelled}: true, {Paused, Expired}: true,
	}
	for _, from := range states {
		for _, to := range states {
			err := ValidateTransition(from, to)
			if (err == nil) != allowed[[2]State{from, to}] {
				t.Fatalf("transition %s -> %s error = %v", from, to, err)
			}
		}
	}
	if ValidateTransition(State("unknown"), Active) == nil {
		t.Fatal("unknown source state was accepted")
	}
}
