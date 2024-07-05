package checksum

func ExampleCalculate() {
	payload := []byte{0, 1, 2, 3, 4, 5, 6}
	var checksum Checksum

	// checksum contains SHA256 hash of the payload
	Calculate(&checksum, SHA256, payload)

	// checksum contains TZ hash of the payload
	Calculate(&checksum, TZ, payload)
}
