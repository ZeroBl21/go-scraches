package scan_test

import (
	"errors"
	"os"
	"testing"

	"github.com/ZeroBl21/cli/ch07/pScan/scan"
)

func TestHostsList_Add(t *testing.T) {
	testCases := []struct {
		name string // description of this test case

		// Named input parameters for target function.
		host   string
		expLen int
		expErr error
	}{
		{"AddNew", "host2", 2, nil},
		{"AddExisting", "host1", 1, scan.ErrExists},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var hl scan.HostsList

			if err := hl.Add("host1"); err != nil {
				t.Fatal(err)
			}

			err := hl.Add(tc.host)
			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected error, got nil instead\n")
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error %q, got %q instead\n", tc.expErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %q instead\n", err)
			}

			if len(hl.Hosts) != tc.expLen {
				t.Errorf("Expected host name %q as index 1, got %q instead\n",
					tc.host, hl.Hosts[1])
			}
		})
	}
}

func TestHostsList_Remove(t *testing.T) {
	tests := []struct {
		name string // description of this test case

		// Named input parameters for target function.
		host   string
		expLen int
		expErr error
	}{
		{"RemoveExisting", "host1", 1, nil},
		{"RemoveNotFound", "host3", 1, scan.ErrNotExists},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var hl scan.HostsList

			for _, h := range []string{"host1", "host2"} {
				if err := hl.Add(h); err != nil {
					t.Fatal(err)
				}
			}

			err := hl.Remove(tc.host)
			if tc.expErr != nil {
				if err == nil {
					t.Fatal("Expected error, got nil instead\n")
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error %q, got %q instead\n", tc.expErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Expected not error, got %q instead\n", err)
			}

			if len(hl.Hosts) != tc.expLen {
				t.Errorf("Expected list length %d, got %d instead\n",
					tc.expLen, len(hl.Hosts))
			}

			if hl.Hosts[0] == tc.host {
				t.Errorf("Host name %q should not be in the list\n", tc.host)
			}
		})
	}
}

func TestHostsList_SaveLoad(t *testing.T) {
	var hl1 scan.HostsList
	var hl2 scan.HostsList

	hostName := "host1"

	hl1.Add(hostName)

	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	defer tf.Close()

	if err := hl1.Save(tf.Name()); err != nil {
		t.Fatalf("Error saving list to file: %s", err)
	}

	if err := hl2.Load(tf.Name()); err != nil {
		t.Fatalf("Error getting list to file: %s", err)
	}

	if hl1.Hosts[0] != hl2.Hosts[0] {
		t.Errorf("Host %q should match %q host.", hl1.Hosts[0], hl2.Hosts[1])
	}
}

func TestHostsList_Load_NoFile(t *testing.T) {
	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}

	if err := os.Remove(tf.Name()); err != nil {
		t.Fatalf("Error deleting temp file: %s", err)
	}

	var hl1 scan.HostsList
	if err := hl1.Load(tf.Name()); err != nil {
		t.Fatalf("Expected no error, got %q instead\n", err)
	}
}
