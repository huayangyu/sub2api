package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ccVersionInBillingRe matches the full cc_version field including the fingerprint suffix.
// Format: cc_version=X.Y.Z.{3-char-hex-fp}
var ccVersionInBillingRe = regexp.MustCompile(`cc_version=\d+\.\d+\.\d+\.[a-f0-9]{3}`)

// cchPlaceholderRe matches the cch=00000 placeholder in billing header text,
// scoped to x-anthropic-billing-header to avoid touching user content.
var cchPlaceholderRe = regexp.MustCompile(`(x-anthropic-billing-header:[^"]*?\bcch=)(00000)(;)`)

const cchSeed uint64 = 0x6E52736AC806831E

// syncBillingHeaderVersion rewrites cc_version (including fingerprint suffix) in
// x-anthropic-billing-header system text blocks to match the version extracted from userAgent.
// The fingerprint suffix depends on the version, so both must be recomputed together.
func syncBillingHeaderVersion(body []byte, userAgent string) []byte {
	version := ExtractCLIVersion(userAgent)
	if version == "" {
		return body
	}

	systemResult := gjson.GetBytes(body, "system")
	if !systemResult.Exists() || !systemResult.IsArray() {
		return body
	}

	fp := computeClaudeCodeFingerprint(body, version)
	replacement := "cc_version=" + version + "." + fp
	idx := 0
	systemResult.ForEach(func(_, item gjson.Result) bool {
		text := item.Get("text")
		if text.Exists() && text.Type == gjson.String &&
			strings.HasPrefix(text.String(), "x-anthropic-billing-header") {
			newText := ccVersionInBillingRe.ReplaceAllString(text.String(), replacement)
			if newText != text.String() {
				if updated, err := sjson.SetBytes(body, fmt.Sprintf("system.%d.text", idx), newText); err == nil {
					body = updated
				}
			}
		}
		idx++
		return true
	})

	return body
}

// signBillingHeaderCCH computes the xxHash64-based CCH signature for the request
// body and replaces the cch=00000 placeholder with the computed 5-hex-char hash.
// The body must contain the placeholder when this function is called.
func signBillingHeaderCCH(body []byte) []byte {
	if !cchPlaceholderRe.Match(body) {
		return body
	}
	cch := fmt.Sprintf("%05x", xxHash64Seeded(body, cchSeed)&0xFFFFF)
	return cchPlaceholderRe.ReplaceAll(body, []byte("${1}"+cch+"${3}"))
}

// xxHash64Seeded computes xxHash64 of data with a custom seed.
func xxHash64Seeded(data []byte, seed uint64) uint64 {
	d := xxhash.NewWithSeed(seed)
	_, _ = d.Write(data)
	return d.Sum64()
}
