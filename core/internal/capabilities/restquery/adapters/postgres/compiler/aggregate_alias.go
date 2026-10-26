package compiler

import "strconv"

const (
	aggregateRawCTE   = `"__fc_raw"`
	aggregateTypedCTE = `"__fc_typed"`
	aggregateGuardCTE = `"__fc_guard"`
	aggregateGuardKey = `"__fc_total"`
)

func aggregateGroupAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 2 {
		return "", ErrContract
	}
	return `"__fc_g` + strconv.Itoa(ordinal) + `"`, nil
}

func aggregateMeasureAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 7 {
		return "", ErrContract
	}
	return `"__fc_m` + strconv.Itoa(ordinal) + `"`, nil
}

func aggregateRoundAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 7 {
		return "", ErrContract
	}
	return `"__fc_s` + strconv.Itoa(ordinal) + `"`, nil
}

func aggregateGuardAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 7 {
		return "", ErrContract
	}
	return `"__fc_c` + strconv.Itoa(ordinal) + `"`, nil
}

func aggregateAverageCountAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 7 {
		return "", ErrContract
	}
	return `"__fc_n` + strconv.Itoa(ordinal) + `"`, nil
}

func aggregateDivisorAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 7 {
		return "", ErrContract
	}
	return `"__fc_d` + strconv.Itoa(ordinal) + `"`, nil
}

func aggregateQuotientAlias(ordinal int) (string, error) {
	if ordinal < 0 || ordinal > 7 {
		return "", ErrContract
	}
	return `"__fc_q` + strconv.Itoa(ordinal) + `"`, nil
}
