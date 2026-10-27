package iam

import (
	"context"
	"errors"
	"io"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type systemTransferExportRows struct {
	ctx         context.Context
	scope       platformtx.TenantScope
	service     *queryapp.Service
	plan        systemTransferExportPlan
	positions   map[string]int
	projections []systemTransferExportProjection
	seen        map[uuid.UUID]struct{}
	rows        [][]systemtransfer.ExportCell
	next        int
	finished    bool
}

func newSystemTransferExportRows(
	ctx context.Context,
	scope platformtx.TenantScope,
	service *queryapp.Service,
	plan systemTransferExportPlan,
) (*systemTransferExportRows, error) {
	projections, err := systemTransferExportProjections(plan.columns)
	if ctx == nil || ctx.Err() != nil || service == nil || err != nil ||
		len(plan.columns) != len(plan.kinds) {
		return nil, systemtransfer.ErrExportQueryInvalid
	}
	positions := make(map[string]int, len(plan.columns))
	for index, column := range plan.columns {
		positions[column] = index
	}
	return &systemTransferExportRows{ctx: ctx, scope: scope, service: service,
		plan: plan, positions: positions, projections: projections,
		seen: make(map[uuid.UUID]struct{})}, nil
}

func (source *systemTransferExportRows) Next() ([]systemtransfer.ExportCell, error) {
	if source == nil || source.ctx == nil || source.ctx.Err() != nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	if source.next < len(source.rows) {
		row := append([]systemtransfer.ExportCell(nil), source.rows[source.next]...)
		source.next++
		return row, nil
	}
	if source.finished {
		return nil, io.EOF
	}
	if err := source.loadPage(); err != nil {
		return nil, err
	}
	if len(source.rows) == 0 {
		return nil, io.EOF
	}
	row := append([]systemtransfer.ExportCell(nil), source.rows[0]...)
	source.next = 1
	return row, nil
}

func (source *systemTransferExportRows) loadPage() error {
	var expected []uuid.UUID
	var expectedMore bool
	rows := [][]systemtransfer.ExportCell(nil)
	for index := range source.projections {
		projection := &source.projections[index]
		request, err := queryapp.NewPageRequest(
			source.plan.request, projection.fields, 500, projection.after(),
		)
		if err != nil {
			return systemtransfer.ErrExportQueryInvalid
		}
		result, err := source.service.RunCatalog(
			source.ctx, source.scope, source.plan.published, request,
		)
		if err != nil {
			return systemTransferExportQueryError(err)
		}
		records, err := systemTransferExportRecords(result.Items(), projection.fields)
		if err != nil {
			return err
		}
		if index == 0 {
			expected, rows, expectedMore, err = source.firstPage(records, result.HasMore())
			if err != nil {
				return err
			}
		} else if !systemTransferExportSamePage(expected, records,
			expectedMore, result.HasMore()) {
			return systemtransfer.ErrExportStateConflict
		}
		if err = source.applyProjection(rows, records, *projection); err != nil {
			return err
		}
		if err = advanceSystemTransferProjection(projection, result); err != nil {
			return err
		}
	}
	for _, row := range rows {
		for _, cell := range row {
			if cell.Validate() != nil {
				return systemtransfer.ErrExportQueryInvalid
			}
		}
	}
	source.rows, source.next, source.finished = rows, 0, !expectedMore
	return nil
}

func systemTransferExportQueryError(err error) error {
	var status interface{ HTTPStatus() int }
	if errors.As(err, &status) {
		switch status.HTTPStatus() {
		case 400:
			return systemtransfer.ErrExportQueryInvalid
		case 401, 403:
			return systemtransfer.ErrExportAuthorizationRevoked
		}
	}
	return errors.Join(systemtransfer.ErrExportUnavailable, err)
}

var _ systemtransfer.ExportRowSource = (*systemTransferExportRows)(nil)
