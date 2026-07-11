package ipm

import "testing"

func TestParsePackageRefDefaultsToMain(t *testing.T) {
	tree, pkg := parsePackageRef("demo")
	if tree != "main" || pkg != "demo" {
		t.Fatalf("unexpected package ref: tree=%q pkg=%q", tree, pkg)
	}
}

func TestParsePackageRefWithExplicitTree(t *testing.T) {
	tree, pkg := parsePackageRef("extra/demo")
	if tree != "extra" || pkg != "demo" {
		t.Fatalf("unexpected package ref: tree=%q pkg=%q", tree, pkg)
	}
}

func TestValidateCommandSafetyRejectsDangerousCommands(t *testing.T) {
	tests := []string{
		"rm -rf *",
		"sudo rm -rf /",
		"mkfs.ext4 /dev/sda",
		"dd if=image of=/dev/sda",
		"shutdown now",
	}
	for _, command := range tests {
		if err := validateCommandSafety(command, nil); err == nil {
			t.Fatalf("expected %q to be rejected", command)
		}
	}
}

func TestValidateCommandSafetyAllowsSafeCommands(t *testing.T) {
	if err := validateCommandSafety("echo hello", nil); err != nil {
		t.Fatalf("expected command to be allowed: %v", err)
	}
}
