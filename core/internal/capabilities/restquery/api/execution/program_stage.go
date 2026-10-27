package execution

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"log/slog"
)

const (
	programStageSealDomain = "harpy.restquery.owner-program-stage.v1"
	maximumStagePrograms   = 4
)

// ProgramStage is an immutable, same-snapshot set of owner Programs.
type ProgramStage struct {
	programs  []Program
	stateSeal [sha256.Size]byte
	valid     bool
}

func NewProgramStage(programs []Program) (ProgramStage, error) {
	stage := ProgramStage{
		programs: append([]Program(nil), programs...), valid: true,
	}
	stage.stateSeal = sealProgramStage(stage)
	if stage.Validate() != nil || stage.validateReady() != nil {
		return ProgramStage{}, errExecutionContract
	}
	return stage, nil
}

func (stage ProgramStage) Validate() error {
	if !stage.valid || len(stage.programs) < 1 ||
		len(stage.programs) > maximumStagePrograms ||
		stage.stateSeal == [sha256.Size]byte{} ||
		stage.stateSeal != sealProgramStage(stage) {
		return errExecutionContract
	}
	if stage.programs[0].Validate() != nil {
		return errExecutionContract
	}
	states := make(map[*programState]struct{}, len(stage.programs))
	identity := stage.programs[0].state.identity
	for _, program := range stage.programs {
		if program.Validate() != nil || !identity.matches(program.state.identity) {
			return errExecutionContract
		}
		if _, duplicate := states[program.state]; duplicate {
			return errExecutionContract
		}
		states[program.state] = struct{}{}
	}
	return nil
}

func (stage ProgramStage) validateReady() error {
	if stage.Validate() != nil {
		return errExecutionContract
	}
	for _, program := range stage.programs {
		if programPhase(program.state.phase.Load()) != programReady ||
			program.state.context.Err() != nil {
			return errExecutionContract
		}
	}
	return nil
}

func sealProgramStage(stage ProgramStage) [sha256.Size]byte {
	digest := sha256.New()
	stageSealBytes(digest, []byte(programStageSealDomain))
	stageSealUint64(digest, uint64(len(stage.programs)))
	for _, program := range stage.programs {
		if program.state == nil {
			stageSealBytes(digest, nil)
			continue
		}
		stageSealBytes(digest, []byte(program.state.owner))
		stageSealBytes(digest, program.state.identity.tenantID[:])
		stageSealBytes(digest, program.state.identity.actorID[:])
		stageSealBytes(digest, []byte(program.state.identity.mode))
		stageSealBytes(digest, []byte(program.state.identity.operationID))
	}
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func stageSealBytes(digest hash.Hash, value []byte) {
	stageSealUint64(digest, uint64(len(value)))
	_, _ = digest.Write(value)
}

func stageSealUint64(digest hash.Hash, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = digest.Write(encoded[:])
}

func (ProgramStage) String() string   { return executionRedaction }
func (ProgramStage) GoString() string { return executionRedaction }
func (ProgramStage) LogValue() slog.Value {
	return executionLogValue()
}
func (stage ProgramStage) Format(state fmt.State, _ rune) {
	writeExecutionRedaction(state)
}
func (ProgramStage) MarshalJSON() ([]byte, error) { return rejectExecutionJSON() }
