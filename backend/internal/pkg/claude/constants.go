// Package claude provides constants and helpers for Claude API integration.
package claude

// Claude Code 客户端相关常量

// Beta header 常量
//
// 这里的常量对齐真实 Claude Code CLI 的最新流量（截至 2026-04）。
// 选型参考：与 Parrot (src/transform/cc_mimicry.py) 的 BETAS 保持一致，
// 原因：Anthropic 上游会基于 anthropic-beta 的完整集合判定请求来源；
// 缺少任何"官方 Claude Code 请求才会带"的 beta，都会被降级到第三方额度，
// 对应报错：`Third-party apps now draw from your extra usage, not your plan limits.`
const (
	BetaOAuth                    = "oauth-2025-04-20"
	BetaClaudeCode               = "claude-code-20250219"
	BetaInterleavedThinking      = "interleaved-thinking-2025-05-14"
	BetaFineGrainedToolStreaming = "fine-grained-tool-streaming-2025-05-14"
	BetaTokenCounting            = "token-counting-2024-11-01"
	BetaContext1M                = "context-1m-2025-08-07"
	BetaFastMode                 = "fast-mode-2026-02-01"

	// 新增（对齐官方 CLI 2.1.9x 以来的流量）
	BetaPromptCachingScope = "prompt-caching-scope-2026-01-05"
	BetaEffort             = "effort-2025-11-24"
	BetaRedactThinking     = "redact-thinking-2026-02-12"
	BetaContextManagement  = "context-management-2025-06-27"
	BetaExtendedCacheTTL   = "extended-cache-ttl-2025-04-11"
)

// DroppedBetas 是转发时需要从 anthropic-beta header 中移除的 beta token 列表。
// 这些 token 是客户端特有的，不应透传给上游 API。
var DroppedBetas = []string{}

// DefaultBetaHeader Claude Code 客户端默认的 anthropic-beta header
const DefaultBetaHeader = BetaClaudeCode + "," + BetaOAuth + "," + BetaInterleavedThinking + "," + BetaFineGrainedToolStreaming

// MessageBetaHeaderNoTools /v1/messages 在无工具时的 beta header
//
// NOTE: Claude Code OAuth credentials are scoped to Claude Code. When we "mimic"
// Claude Code for non-Claude-Code clients, we must include the claude-code beta
// even if the request doesn't use tools, otherwise upstream may reject the
// request as a non-Claude-Code API request.
const MessageBetaHeaderNoTools = BetaClaudeCode + "," + BetaOAuth + "," + BetaInterleavedThinking

// MessageBetaHeaderWithTools /v1/messages 在有工具时的 beta header
const MessageBetaHeaderWithTools = BetaClaudeCode + "," + BetaOAuth + "," + BetaInterleavedThinking

// CountTokensBetaHeader count_tokens 请求使用的 anthropic-beta header
const CountTokensBetaHeader = BetaClaudeCode + "," + BetaOAuth + "," + BetaInterleavedThinking + "," + BetaTokenCounting

// HaikuBetaHeader Haiku 模型使用的 anthropic-beta header（不需要 claude-code beta）
const HaikuBetaHeader = BetaOAuth + "," + BetaInterleavedThinking

// APIKeyBetaHeader API-key 账号建议使用的 anthropic-beta header（不包含 oauth）
const APIKeyBetaHeader = BetaClaudeCode + "," + BetaInterleavedThinking + "," + BetaFineGrainedToolStreaming

// APIKeyHaikuBetaHeader Haiku 模型在 API-key 账号下使用的 anthropic-beta header（不包含 oauth / claude-code）
const APIKeyHaikuBetaHeader = BetaInterleavedThinking

// DefaultCacheControlTTL 是网关代理为自己生成的 cache_control 块默认使用的 ttl。
// 真实 Claude Code CLI 当前使用 "1h"，但本仓策略是"客户端透传 ttl 优先；
// 客户端缺省时统一使用 5m"，这样既不浪费 1h 缓存额度，也保留客户端自定义能力。
const DefaultCacheControlTTL = "5m"

// CLICurrentVersion 是 sub2api 当前对外伪装的 Claude Code CLI 版本号（三段 semver）。
// 用于 billing attribution block 中的 cc_version=X.Y.Z.{fp} 前缀以及 fingerprint 计算。
// 必须与 DefaultHeaders["User-Agent"] 中的版本号严格一致；不一致会被 Anthropic 判第三方。
const CLICurrentVersion = "2.1.92"

// CLIVersionPool 是可用的 CLI 版本池，按 accountID 确定性选择。
// 多版本分散降低所有请求集中在单一版本号的风险。
// 每个版本对应一个 SDK package version，保持一致性。
var CLIVersionPool = []VersionProfile{
	{CLIVersion: "2.1.92", PackageVersion: "0.70.0", RuntimeVersion: "v24.13.0"},
	{CLIVersion: "2.1.91", PackageVersion: "0.69.1", RuntimeVersion: "v22.14.0"},
	{CLIVersion: "2.1.90", PackageVersion: "0.69.0", RuntimeVersion: "v22.12.0"},
	{CLIVersion: "2.1.89", PackageVersion: "0.68.2", RuntimeVersion: "v22.11.0"},
	{CLIVersion: "2.1.88", PackageVersion: "0.68.0", RuntimeVersion: "v24.13.0"},
}

// VersionProfile 将 CLI 版本与对应的 SDK / runtime 版本绑定，确保三者一致。
type VersionProfile struct {
	CLIVersion     string
	PackageVersion string
	RuntimeVersion string
}

// SelectVersionProfile 根据 accountID 确定性选择版本 profile。
func SelectVersionProfile(accountID int64) VersionProfile {
	if len(CLIVersionPool) == 0 {
		return VersionProfile{CLIVersion: CLICurrentVersion, PackageVersion: "0.70.0", RuntimeVersion: "v24.13.0"}
	}
	idx := int(accountID) % len(CLIVersionPool)
	if idx < 0 {
		idx = -idx
	}
	return CLIVersionPool[idx]
}

// FullClaudeCodeMimicryBetas 返回最"像"真实 Claude Code CLI 的完整 beta 列表，
// 用于 OAuth 账号伪装成 Claude Code 时使用。
// 顺序与真实 CLI 抓包一致。
//
// 使用建议：
//   - OAuth 账号 + 非 haiku：追加这整份列表，再按需保留 client 带来的 beta。
//   - OAuth 账号 + haiku：Anthropic 对 haiku 不做 third-party 判定，使用 HaikuBetaHeader 即可。
//   - API-key 账号：不要使用本函数，参见 APIKeyBetaHeader。
//   - 不默认加入 redact-thinking，避免上游抹除 thinking 内容；客户端显式传入时由合并逻辑保留。
//
// 2026-05 更新：补齐 fine-grained-tool-streaming 和 token-counting，
// 真实 CLI 2.1.88+ 的 /v1/messages 请求均携带这两个 beta。
func FullClaudeCodeMimicryBetas() []string {
	return []string{
		BetaClaudeCode,
		BetaOAuth,
		BetaInterleavedThinking,
		BetaFineGrainedToolStreaming,
		BetaPromptCachingScope,
		BetaEffort,
		BetaContextManagement,
		BetaExtendedCacheTTL,
		BetaTokenCounting,
	}
}

// DefaultHeaders 是 Claude Code 客户端默认请求头。
var DefaultHeaders = map[string]string{
	// Keep these in sync with recent Claude CLI traffic to reduce the chance
	// that Claude Code-scoped OAuth credentials are rejected as "non-CLI" usage.
	// 版本参考：对齐 Parrot (src/transform/cc_mimicry.py:49) 的 CLI_USER_AGENT。
	"User-Agent":                                "claude-cli/2.1.92 (external, cli)",
	"X-Stainless-Lang":                          "js",
	"X-Stainless-Package-Version":               "0.70.0",
	"X-Stainless-OS":                            "Linux",
	"X-Stainless-Arch":                          "arm64",
	"X-Stainless-Runtime":                       "node",
	"X-Stainless-Runtime-Version":               "v24.13.0",
	"X-Stainless-Retry-Count":                   "0",
	"X-Stainless-Timeout":                       "600",
	"X-App":                                     "cli",
	"Anthropic-Dangerous-Direct-Browser-Access": "true",
}

// FingerprintProfile 定义一组完整的客户端环境指纹。
// 每个 profile 模拟一个真实的 Claude Code 用户环境。
type FingerprintProfile struct {
	OS             string // macOS, Linux, Windows
	Arch           string // arm64, x64
	Runtime        string // node
	RuntimeVersion string // v22.x, v24.x
}

// FingerprintProfilePool 是可用的环境指纹池。
// 真实用户群体中 macOS/Linux/Windows 和 arm64/x64 都有分布。
var FingerprintProfilePool = []FingerprintProfile{
	{OS: "macOS", Arch: "arm64", Runtime: "node", RuntimeVersion: "v24.13.0"},
	{OS: "Linux", Arch: "x64", Runtime: "node", RuntimeVersion: "v22.14.0"},
	{OS: "macOS", Arch: "arm64", Runtime: "node", RuntimeVersion: "v22.12.0"},
	{OS: "Linux", Arch: "arm64", Runtime: "node", RuntimeVersion: "v24.13.0"},
	{OS: "Windows_NT", Arch: "x64", Runtime: "node", RuntimeVersion: "v22.14.0"},
	{OS: "macOS", Arch: "x64", Runtime: "node", RuntimeVersion: "v24.13.0"},
	{OS: "Linux", Arch: "x64", Runtime: "node", RuntimeVersion: "v24.13.0"},
	{OS: "Windows_NT", Arch: "arm64", Runtime: "node", RuntimeVersion: "v22.12.0"},
}

// SelectFingerprintProfile 根据 accountID 确定性选择环境指纹 profile。
func SelectFingerprintProfile(accountID int64) FingerprintProfile {
	if len(FingerprintProfilePool) == 0 {
		return FingerprintProfile{OS: "Linux", Arch: "arm64", Runtime: "node", RuntimeVersion: "v24.13.0"}
	}
	idx := int(accountID) % len(FingerprintProfilePool)
	if idx < 0 {
		idx = -idx
	}
	return FingerprintProfilePool[idx]
}

// BuildDefaultHeadersForAccount 根据 accountID 生成该账号专属的默认请求头。
// 版本和环境指纹均从 pool 中确定性选择，确保同一账号始终使用相同的指纹。
// TODO: 待后续 PR 接入到 gateway 转发逻辑中，替换全局 DefaultHeaders
func BuildDefaultHeadersForAccount(accountID int64) map[string]string {
	vp := SelectVersionProfile(accountID)
	fp := SelectFingerprintProfile(accountID)
	return map[string]string{
		"User-Agent":                                "claude-cli/" + vp.CLIVersion + " (external, cli)",
		"X-Stainless-Lang":                          "js",
		"X-Stainless-Package-Version":               vp.PackageVersion,
		"X-Stainless-OS":                            fp.OS,
		"X-Stainless-Arch":                          fp.Arch,
		"X-Stainless-Runtime":                       fp.Runtime,
		"X-Stainless-Runtime-Version":               fp.RuntimeVersion,
		"X-Stainless-Retry-Count":                   "0",
		"X-Stainless-Timeout":                       "600",
		"X-App":                                     "cli",
		"Anthropic-Dangerous-Direct-Browser-Access": "true",
	}
}

// Model 表示一个 Claude 模型
type Model struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

// DefaultModels Claude Code 客户端支持的默认模型列表
var DefaultModels = []Model{
	{
		ID:          "claude-opus-4-5-20251101",
		Type:        "model",
		DisplayName: "Claude Opus 4.5",
		CreatedAt:   "2025-11-01T00:00:00Z",
	},
	{
		ID:          "claude-opus-4-6",
		Type:        "model",
		DisplayName: "Claude Opus 4.6",
		CreatedAt:   "2026-02-06T00:00:00Z",
	},
	{
		ID:          "claude-opus-4-7",
		Type:        "model",
		DisplayName: "Claude Opus 4.7",
		CreatedAt:   "2026-04-17T00:00:00Z",
	},
	{
		ID:          "claude-sonnet-4-6",
		Type:        "model",
		DisplayName: "Claude Sonnet 4.6",
		CreatedAt:   "2026-02-18T00:00:00Z",
	},
	{
		ID:          "claude-sonnet-4-5-20250929",
		Type:        "model",
		DisplayName: "Claude Sonnet 4.5",
		CreatedAt:   "2025-09-29T00:00:00Z",
	},
	{
		ID:          "claude-haiku-4-5-20251001",
		Type:        "model",
		DisplayName: "Claude Haiku 4.5",
		CreatedAt:   "2025-10-01T00:00:00Z",
	},
}

// DefaultModelIDs 返回默认模型的 ID 列表
func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, m := range DefaultModels {
		ids[i] = m.ID
	}
	return ids
}

// DefaultTestModel 测试时使用的默认模型
const DefaultTestModel = "claude-sonnet-4-5-20250929"

// ModelIDOverrides Claude OAuth 请求需要的模型 ID 映射
var ModelIDOverrides = map[string]string{
	"claude-sonnet-4-5": "claude-sonnet-4-5-20250929",
	"claude-opus-4-5":   "claude-opus-4-5-20251101",
	"claude-haiku-4-5":  "claude-haiku-4-5-20251001",
}

// ModelIDReverseOverrides 用于将上游模型 ID 还原为短名
var ModelIDReverseOverrides = map[string]string{
	"claude-sonnet-4-5-20250929": "claude-sonnet-4-5",
	"claude-opus-4-5-20251101":   "claude-opus-4-5",
	"claude-haiku-4-5-20251001":  "claude-haiku-4-5",
}

// NormalizeModelID 根据 Claude OAuth 规则映射模型
func NormalizeModelID(id string) string {
	if id == "" {
		return id
	}
	if mapped, ok := ModelIDOverrides[id]; ok {
		return mapped
	}
	return id
}

// DenormalizeModelID 将上游模型 ID 转换为短名
func DenormalizeModelID(id string) string {
	if id == "" {
		return id
	}
	if mapped, ok := ModelIDReverseOverrides[id]; ok {
		return mapped
	}
	return id
}
