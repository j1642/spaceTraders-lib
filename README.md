This is a local HTTP server and GUI for accessing the [Space Traders API](https://docs.spacetraders.io/).

To use, run `go run server.go` and enter the URL `localhost:8080` in your browser.

HTMX is required and is not included in the repo at the moment.

composites/ - concurrent ship management functions using the API responses

objects/ - structs for unmarshalling the API JSON responses

requests/ - HTTP requests for the API endpoints
