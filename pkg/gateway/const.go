package gateway

const (
	jsonErrFailedToDecodeBody = "{\"error\":\"Failed to decode json body\", \"message\": %q}"
	jsonErrFailedToFetchJob   = "{\"error\":\"Failed to fetch job\", \"message\": %q}"
	jsonErrFailedToSaveJob    = "{\"error\":\"Failed to save job\", \"message\": %q}"
	jsonErrJobNotFound        = "{\"error\":\"Job %s not found\", \"message\": %q}"
)