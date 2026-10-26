package postgres

import (
	"context"
	"errors"
)

var ErrInstallationBindingUnavailable = errors.New(
	"platform authority installation binding unavailable",
)

type InstallationRow interface{ Scan(...any) error }

type InstallationDatabase struct {
	QueryRow func(context.Context, string, ...any) InstallationRow
}

type InstallationBinding struct {
	queryRow func(context.Context, string, ...any) InstallationRow
}

func BindInstallation(database InstallationDatabase) (InstallationBinding, error) {
	if database.QueryRow == nil {
		return InstallationBinding{}, ErrInstallationBindingUnavailable
	}
	return InstallationBinding{queryRow: database.QueryRow}, nil
}

func (binding InstallationBinding) BindOwner(
	ctx context.Context,
	arguments ...any,
) (bool, error) {
	return binding.queryBoolean(ctx, installationBindOwnerSQL, arguments...)
}

func (binding InstallationBinding) PinHead(
	ctx context.Context,
	arguments ...any,
) (bool, error) {
	return binding.queryBoolean(ctx, installationPinHeadSQL, arguments...)
}

func (binding InstallationBinding) queryBoolean(
	ctx context.Context,
	statement string,
	arguments ...any,
) (bool, error) {
	if binding.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return false, ErrInstallationBindingUnavailable
	}
	var replay bool
	err := binding.queryRow(ctx, statement, arguments...).Scan(&replay)
	return replay, err
}
