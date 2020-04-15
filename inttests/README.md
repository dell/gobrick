# GOBRICK inttests

###Usage

1. Create volume on storage array
2. Publish volume to the node where you want to run the tests
3. From `iscsi|fc_test_data.json_example` create your own `iscsi|fc_test_data.json`
4. Run tests:  `go test . -v`