package billingv1

import (
	"fmt"
	"net/http"
)

type operation func(http.ResponseWriter, *http.Request)

func (handler *Handler) BindRoutes() ([]RouteBinding, error) {
	if handler == nil {
		return nil, fmt.Errorf("Billing V1 compatibility handler is required")
	}
	operations := handler.operations()
	routes := Routes()
	bindings := make([]RouteBinding, 0, len(routes))
	for _, route := range routes {
		action, found := operations[route.ID]
		if !found || action == nil {
			return nil, fmt.Errorf("Billing V1 operation %s is required", route.ID)
		}
		current, currentAction := route, action
		bindings = append(bindings, RouteBinding{
			Method: current.Method, Pattern: current.RuntimePattern,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := handler.availability.AdmitRoute(r.Context(), current); err != nil {
					handler.fail(w, current.ID, err)
					return
				}
				if err := handler.recordHit(r.Context(), current.ID); err != nil {
					handler.fail(w, current.ID, ErrRetirementEvidence)
					return
				}
				currentAction(w, r)
			}),
		})
	}
	return bindings, nil
}
