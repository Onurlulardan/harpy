package schemaexecutor

import schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"

var v2OperationContracts = map[schemav2.OperationID]struct {
	executor schemav2.ExecutorKind
	class    schemav2.TransactionClass
}{
	schemav2.ActivateGeneration:    {schemav2.FormsRuntimeExecutor, schemav2.Transactional},
	schemav2.AddConstraint:         {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.AddFieldColumn:        {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.AddRelationTable:      {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.BackfillFieldBatch:    {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.CleanupInvalidIndex:   {schemav2.MigratorExecutor, schemav2.ConcurrentIndex},
	schemav2.CreateFormTable:       {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.CreateIndex:           {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.CreateIndexConcurrent: {schemav2.MigratorExecutor, schemav2.ConcurrentIndex},
	schemav2.DisableOldWriter:      {schemav2.FormsRuntimeExecutor, schemav2.Transactional},
	schemav2.DropConstraint:        {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.DropFieldColumn:       {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.DropFormTable:         {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.DropIndex:             {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.DropIndexConcurrent:   {schemav2.MigratorExecutor, schemav2.ConcurrentIndex},
	schemav2.DropRelationTable:     {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.EnableDualWrite:       {schemav2.FormsRuntimeExecutor, schemav2.Transactional},
	schemav2.RetainGeneration:      {schemav2.FormsRuntimeExecutor, schemav2.Transactional},
	schemav2.SetFieldNotNull:       {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.ValidateConstraint:    {schemav2.MigratorExecutor, schemav2.Transactional},
	schemav2.VerifyStorage:         {schemav2.MigratorExecutor, schemav2.Transactional},
}
