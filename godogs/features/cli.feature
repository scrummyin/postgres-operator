# file: $GOPATH/godogs/features/godogs.feature
Feature: Manage pgo
	In order to be able to manage pgo
	As a sysadmin
	I need to be able to run commands from a terminal

	Scenario: Check pgo version
		When I run "pgo version"
		Then There should be matching version info for both client and server

	Scenario: List all clusters from pgo
		Given No clusters are currently running
		When I run "pgo show cluster all"
		Then Then pgo should have stdout containing "No clusters found."
