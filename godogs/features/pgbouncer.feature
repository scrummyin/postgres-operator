Feature: Pgbouncer support
	In order to be able to handle high ammount of connections
	As a dba
	I need add pgbouncer support from the cli

	Scenario: Create cluster with pgbouncer
		When I create a cluster named bouncetest with pgbouncer
		Then A pod labeled with "pg-cluster=bouncetest,crunchy-pgbouncer=true" should be up
