package services

import (
	wafer "github.com/anthropics/wafer-sdk-go"
)

// CryptoClient provides typed access to the WAFER cryptographic capability
// for hashing, signing, verification, and random byte generation. All
// operations are sent as messages through the context.
type CryptoClient struct {
	ctx *wafer.Context
}

// NewCryptoClient creates a new CryptoClient bound to the given context.
func NewCryptoClient(ctx *wafer.Context) *CryptoClient {
	return &CryptoClient{ctx: ctx}
}

// Hash computes a hash of the given password (or other input string). The
// hashing algorithm is determined by the runtime implementation (e.g., bcrypt).
// Returns the hash string on success.
//
// Message kind: "svc.crypto.hash"
// Data: password string
func (c *CryptoClient) Hash(password string) (string, error) {
	msg := &wafer.Message{
		Kind: "svc.crypto.hash",
		Data: []byte(password),
	}
	result := c.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return "", result.Err
	}
	if result.Response == nil {
		return "", nil
	}
	return string(result.Response.Data), nil
}

// CompareHash compares a plaintext password against a previously computed hash.
// Returns true if they match, false otherwise.
//
// Message kind: "svc.crypto.compare_hash"
// Data: password string
// Meta: [["hash", hash]]
func (c *CryptoClient) CompareHash(password, hash string) (bool, error) {
	msg := &wafer.Message{
		Kind: "svc.crypto.compare_hash",
		Data: []byte(password),
		Meta: map[string]string{
			"hash": hash,
		},
	}
	result := c.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return false, result.Err
	}
	// A Respond action means the comparison succeeded (match).
	// A Continue action means no match.
	return result.Action == wafer.Respond, nil
}

// Sign creates a cryptographic signature over the given data using the
// algorithm and key specified in metadata. Returns the signature bytes.
//
// Message kind: "svc.crypto.sign"
// Data: data to sign
// Meta: [["algorithm", algorithm], ["key", key]]
func (c *CryptoClient) Sign(data []byte, algorithm, key string) ([]byte, error) {
	msg := &wafer.Message{
		Kind: "svc.crypto.sign",
		Data: data,
		Meta: map[string]string{
			"algorithm": algorithm,
			"key":       key,
		},
	}
	result := c.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil {
		return nil, nil
	}
	return result.Response.Data, nil
}

// Verify checks a cryptographic signature against the given data. Returns true
// if the signature is valid.
//
// Message kind: "svc.crypto.verify"
// Data: data that was signed
// Meta: [["algorithm", algorithm], ["key", key], ["signature", signature]]
func (c *CryptoClient) Verify(data []byte, algorithm, key, signature string) (bool, error) {
	msg := &wafer.Message{
		Kind: "svc.crypto.verify",
		Data: data,
		Meta: map[string]string{
			"algorithm": algorithm,
			"key":       key,
			"signature": signature,
		},
	}
	result := c.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return false, result.Err
	}
	return result.Action == wafer.Respond, nil
}

// RandomBytes generates cryptographically secure random bytes of the given
// length. Returns the random bytes.
//
// Message kind: "svc.crypto.random_bytes"
// Meta: [["length", length]]
func (c *CryptoClient) RandomBytes(length int) ([]byte, error) {
	msg := &wafer.Message{
		Kind: "svc.crypto.random_bytes",
		Meta: map[string]string{
			"length": intToString(length),
		},
	}
	result := c.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil {
		return nil, nil
	}
	return result.Response.Data, nil
}

// intToString converts an int to its decimal string representation without
// importing strconv to keep the binary small.
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// Reverse.
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
