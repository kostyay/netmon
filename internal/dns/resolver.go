package dns

import (
	"context"
	"net"
	"time"
)

// ResolveResult is returned by async resolution.
type ResolveResult struct {
	IP       string
	Hostname string
	Err      error
}

// ResolveAsync performs reverse DNS lookup asynchronously.
// Returns a channel that receives the result.
func ResolveAsync(ctx context.Context, ip string) <-chan ResolveResult {
	ch := make(chan ResolveResult, 1)

	go func() {
		defer close(ch)

		// Set timeout
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		names, err := net.DefaultResolver.LookupAddr(ctx, ip)
		if err != nil || len(names) == 0 {
			ch <- ResolveResult{IP: ip, Err: err}
			return
		}

		// Remove trailing dot from hostname
		hostname := names[0]
		if len(hostname) > 0 && hostname[len(hostname)-1] == '.' {
			hostname = hostname[:len(hostname)-1]
		}

		ch <- ResolveResult{IP: ip, Hostname: hostname}
	}()

	return ch
}

// Resolve performs synchronous reverse DNS lookup.
func Resolve(ctx context.Context, ip string) (string, error) {
	result := <-ResolveAsync(ctx, ip)
	return result.Hostname, result.Err
}
