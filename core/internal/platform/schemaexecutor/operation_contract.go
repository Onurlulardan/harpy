package schemaexecutor

import schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"

var operationContracts = map[schemav1.OperationID]struct {
	executor schemav1.ExecutorKind
	class    schemav1.TransactionClass
}{
	schemav1.ActivateGeneration:    {schemav1.FormsRuntimeExecutor, schemav1.Transactional},
	schemav1.AddConstraint:         {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.AddFieldColumn:        {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.AddRelationTable:      {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.BackfillFieldBatch:    {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.CleanupInvalidIndex:   {schemav1.MigratorExecutor, schemav1.ConcurrentIndex},
	schemav1.CreateFormTable:       {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.CreateIndex:           {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.CreateIndexConcurrent: {schemav1.MigratorExecutor, schemav1.ConcurrentIndex},
	schemav1.DisableOldWriter:      {schemav1.FormsRuntimeExecutor, schemav1.Transactional},
	schemav1.DropConstraint:        {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.DropFieldColumn:       {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.DropFormTable:         {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.DropIndex:             {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.DropIndexConcurrent:   {schemav1.MigratorExecutor, schemav1.ConcurrentIndex},
	schemav1.DropRelationTable:     {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.EnableDualWrite:       {schemav1.FormsRuntimeExecutor, schemav1.Transactional},
	schemav1.RetainGeneration:      {schemav1.FormsRuntimeExecutor, schemav1.Transactional},
	schemav1.ValidateConstraint:    {schemav1.MigratorExecutor, schemav1.Transactional},
	schemav1.VerifyStorage:         {schemav1.MigratorExecutor, schemav1.Transactional},
}
