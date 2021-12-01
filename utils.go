package main

// This whole package should have been a third party dependency, but I couldn't find one.
// Returns whether `fqdn` is a subdomain of `query`
func isSubdomainOf(fqdn string, query string) bool {
	if len(fqdn) < len(query) {
		return false
	}
	offset := len(fqdn) - len(query)
	if len(fqdn) > len(query) && fqdn[offset-1] != '.' {
		return false
	}
	trimmedFqdn := fqdn[offset:]
	return trimmedFqdn == query
}
