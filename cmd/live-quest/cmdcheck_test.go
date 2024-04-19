package main

import (
	"fmt"
	"testing"
)

// Test table
var testTable = []struct {
	Author    string
	URL       string
	Challenge Challenge

	Out *CheckResult

	WantErr bool
}{
	{
		Author:    "ronoaldopereira",
		URL:       "https://go.dev/play/p/XI2Iph8fANm",
		Challenge: Challenge{Type: ChallengeStatic, Backend: ChallengeBackendPlayGoDev, Output: "Olá Mundo\n"},
		Out:       &CheckResult{OK: true, Result: "OK"},
	},
	{
		Author:    "ronoaldopereira",
		URL:       "https://go.dev/play/p/apqTw4aZLKk",
		Challenge: Challenge{Type: ChallengeStatic, Backend: ChallengeBackendPlayGoDev, Output: "Olá Mundo\n"},
		Out:       nil,
		WantErr:   true,
	},
}

func TestCheck(t *testing.T) {
	for tn, tc := range testTable {
		t.Run(fmt.Sprintf("#%d", tn), func(t *testing.T) {
			t.Logf("Running Test Case %d: %v", tn, tc)

			// Exec
			out, err := Check(tc.Author, tc.URL, tc.Challenge)
			t.Logf("Output: %v, err: %v", out, err)

			// Asserts
			if tc.WantErr && err == nil {
				t.Errorf("Wanted error but got nil")
			}
			if !tc.WantErr && err != nil {
				t.Errorf("Wanted no error but got: '%v'", err)
			}
			if out == nil && tc.Out != nil {
				t.Fatalf("Wanted out but got nil")
			}
			if tc.Out != nil {
				if tc.Out.OK != out.OK {
					t.Errorf("Invalid Out.OK: Wanted='%v' Got='%v'", tc.Out.OK, out.OK)
				}
				if tc.Out.Result != "" && tc.Out.Result != out.Result {
					t.Errorf("Invalid Out.Result: Wanted='%v' Got='%v'", tc.Out.Result, out.Result)
				}
			}
		})
	}
}
