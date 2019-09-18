Feature: Manage clusters
	In order to be able to manage clusters
	As a dba
	I need manage clusters from the cli

	Scenario: Create cluster from pgo
		When I create a cluster named jkcluster
		Then A primary pod labeled with "pg-cluster=jkcluster" should be up

	Scenario: Delete cluster from pgo
		Given An existing cluster named deleteme
		When I run "pgo delete cluster deleteme" and type "yes"
		Then No pods with label "pg-cluster=deleteme" should exist

	Scenario: pgo can list a specific cluster
		Given An existing cluster named singleone
		When I run "pgo show cluster singleone"
		Then Then pgo should have stdout containing "cluster : singleone"
