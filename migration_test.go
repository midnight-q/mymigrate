package mymigrate

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MyMigrateNewNames(t *testing.T) {
	testError := errors.New("test error")

	type testCase struct {
		newNames      map[string]UpFunc
		appliedNames  []string
		applyErr      error
		expectedErr   error
		expectedNames []string
	}

	testCases := []testCase{
		{
			newNames:      map[string]UpFunc{},
			appliedNames:  []string{},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{},
		},
		{
			newNames: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
			},
			appliedNames:  []string{},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{"0"},
		},
		{
			newNames: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
			},
			appliedNames:  []string{"0"},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{},
		},
		{
			newNames: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
				"1": UpFunc(func(db *sql.DB) error { return nil }),
				"2": UpFunc(func(db *sql.DB) error { return nil }),
			},
			appliedNames:  []string{"0", "2"},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{"1"},
		},
		{
			newNames: map[string]UpFunc{
				"2": UpFunc(func(db *sql.DB) error { return nil }),
				"1": UpFunc(func(db *sql.DB) error { return nil }),
				"0": UpFunc(func(db *sql.DB) error { return nil }),
			},
			appliedNames:  []string{"0"},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{"1", "2"},
		},
		{
			newNames: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
			},
			appliedNames:  []string{"2"},
			applyErr:      testError,
			expectedErr:   testError,
			expectedNames: []string{},
		},
	}

	for i, c := range testCases {
		resetMigrations()
		resetAppliedFunc()
		resetMarkAppliedFunc()

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			getApplied = func(db *sql.DB) ([]string, error) {
				return c.appliedNames, c.applyErr
			}

			for name, mig := range c.newNames {
				Add(name, mig)
			}

			result, err := NewNames()
			if err != c.expectedErr {
				t.Errorf("I expected to get error \"%v\" but got \"%v\"", c.expectedErr, err)
				return
			}

			if !isEqualSlices(result, c.expectedNames) {
				t.Errorf("I expected to get names %v but got %v", c.expectedNames, result)
			}
		})
	}

}

func Test_MyMigrateApply(t *testing.T) {
	applyErr := errors.New("apply err")
	markAppliedErr := errors.New("mark applied err")
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	type testCase struct {
		migrations       map[string]UpFunc
		applyErr         error
		markAppliedErr   error
		expectedErr      error
		expectMarkedCall map[string]bool
	}

	testCases := []testCase{
		{
			migrations:       map[string]UpFunc{},
			applyErr:         nil,
			markAppliedErr:   nil,
			expectedErr:      nil,
			expectMarkedCall: map[string]bool{},
		},
		{
			migrations:       map[string]UpFunc{},
			applyErr:         applyErr,
			markAppliedErr:   nil,
			expectedErr:      applyErr,
			expectMarkedCall: map[string]bool{},
		},
		{
			migrations:       map[string]UpFunc{},
			applyErr:         nil,
			markAppliedErr:   markAppliedErr,
			expectedErr:      nil,
			expectMarkedCall: map[string]bool{},
		},
		{
			migrations: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
			},
			applyErr:         applyErr,
			markAppliedErr:   markAppliedErr,
			expectedErr:      applyErr,
			expectMarkedCall: map[string]bool{},
		},
		{
			migrations: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
			},
			applyErr:         nil,
			markAppliedErr:   markAppliedErr,
			expectedErr:      markAppliedErr,
			expectMarkedCall: map[string]bool{"0": true},
		},
		{
			migrations: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return err1 }),
				"1": UpFunc(func(db *sql.DB) error { return err2 }),
			},
			applyErr:         nil,
			markAppliedErr:   nil,
			expectedErr:      err1,
			expectMarkedCall: map[string]bool{},
		},
		{
			migrations: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
				"1": UpFunc(func(db *sql.DB) error { return err2 }),
			},
			applyErr:         nil,
			markAppliedErr:   nil,
			expectedErr:      err2,
			expectMarkedCall: map[string]bool{"0": true},
		},
		{
			migrations: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return err1 }),
				"1": UpFunc(func(db *sql.DB) error { return nil }),
			},
			applyErr:         nil,
			markAppliedErr:   nil,
			expectedErr:      err1,
			expectMarkedCall: map[string]bool{},
		},
		{
			migrations: map[string]UpFunc{
				"0": UpFunc(func(db *sql.DB) error { return nil }),
				"1": UpFunc(func(db *sql.DB) error { return nil }),
			},
			applyErr:         nil,
			markAppliedErr:   nil,
			expectedErr:      nil,
			expectMarkedCall: map[string]bool{"0": true, "1": true},
		},
	}

	for i, c := range testCases {
		resetMigrations()
		resetAppliedFunc()
		resetMarkAppliedFunc()

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// we've already tested NewNames() function
			// so, here we will return always empty slice
			getApplied = func(db *sql.DB) ([]string, error) {
				return []string{}, c.applyErr
			}

			markedCall := make(map[string]bool)
			markApplied = func(db *sql.DB, name string) error {
				if !c.expectMarkedCall[name] {
					t.Errorf("I didn't excpect that mig '%s' will be marked as aplied", name)
				}

				markedCall[name] = true

				return c.markAppliedErr
			}

			for name, mig := range c.migrations {
				Add(name, mig)
			}

			err := Apply()
			if err != c.expectedErr {
				t.Errorf("I expected to get error \"%v\" but got \"%v\"", c.expectedErr, err)
			}

			if len(markedCall) != len(c.expectMarkedCall) {
				t.Errorf("I expected that these migrations (%+v) we will try to mark as applied. "+
					"But we tried to mark as applied %+v", c.expectMarkedCall, markedCall)
			}
		})
	}
}

func Test_MyMigrateDown(t *testing.T) {
	reset := func() {
		resetMigrations()
		resetMarkAppliedFunc()
		resetAppliedFunc()
		resetDownFunc()
	}

	type testCase struct {
		applied       []string
		appliedErr    error
		downCount     int
		expDownNames  []string
		downErr       error
		expErr        error
		expDownCalled bool
	}

	testCases := map[string]testCase{
		"applied error": {
			applied:       []string{},
			appliedErr:    errors.New("hello"),
			downCount:     0,
			expDownNames:  []string{},
			downErr:       nil,
			expErr:        errors.New("hello"),
			expDownCalled: false,
		},
		"applied zero migrations and ask do down 10": {
			applied:       []string{},
			appliedErr:    nil,
			downCount:     10,
			expDownNames:  []string{},
			downErr:       nil,
			expErr:        nil,
			expDownCalled: false,
		},
		"applied two and ask do down 10": {
			applied:       []string{"mig_001", "mig_002"},
			appliedErr:    nil,
			downCount:     10,
			expDownNames:  []string{"mig_001", "mig_002"},
			downErr:       nil,
			expErr:        nil,
			expDownCalled: true,
		},
		"applied three and ask do down 2": {
			applied:       []string{"mig_001", "mig_002", "mig_003"},
			appliedErr:    nil,
			downCount:     2,
			expDownNames:  []string{"mig_001", "mig_002"},
			downErr:       nil,
			expErr:        nil,
			expDownCalled: true,
		},
		"applied three and ask do down all of them by 0 number": {
			applied:       []string{"mig_001", "mig_002", "mig_003"},
			appliedErr:    nil,
			downCount:     0,
			expDownNames:  []string{"mig_001", "mig_002", "mig_003"},
			downErr:       nil,
			expErr:        nil,
			expDownCalled: true,
		},
		"applied three and ask do down all of them by exact number": {
			applied:       []string{"mig_001", "mig_002", "mig_003"},
			appliedErr:    nil,
			downCount:     3,
			expDownNames:  []string{"mig_001", "mig_002", "mig_003"},
			downErr:       nil,
			expErr:        nil,
			expDownCalled: true,
		},
		"applied and downed with error": {
			applied:       []string{"mig_001", "mig_002", "mig_003"},
			appliedErr:    nil,
			downCount:     1,
			expDownNames:  []string{"mig_001"},
			downErr:       errors.New("down error"),
			expErr:        errors.New("down error"),
			expDownCalled: true,
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			defer reset()

			getApplied = func(db *sql.DB) ([]string, error) {
				return tc.applied, tc.appliedErr
			}

			isDownCalled := false
			down = func(db *sql.DB, names []string) error {
				assert.EqualValues(t, tc.expDownNames, names, "проверка на ожидаемые миграции для отката")
				isDownCalled = true
				return tc.downErr
			}

			err := Down(tc.downCount)
			assert.EqualValues(t, tc.expErr, err, "проверка на ожидаемую ошибку")
			assert.EqualValues(t, tc.expDownCalled, isDownCalled, "проверка навызов функции down")
		})
	}

}

func isEqualSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for j, n := range a {
		if n != b[j] {
			return false
		}
	}

	return true
}
