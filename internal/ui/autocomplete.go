package ui

import "strings"

// grammar maps a context key (built from the last 1-2 complete tokens) to
// the list of tokens that can follow. An empty key "" means top-level
// (beginning of the rule or after a completed expression + value).
var grammar = map[string][]string{
	// ── Top-level ────────────────────────────────────────────────────────────
	"": {
		"ip", "ip6", "inet", "arp", "bridge",
		"tcp", "udp", "udplite", "sctp", "dccp",
		"icmp", "icmpv6", "icmpx",
		"ct", "meta",
		"iif", "oif", "iifname", "oifname",
		"mark", "pkttype",
		"accept", "drop", "reject", "log", "counter",
		"limit", "return", "jump", "goto",
	},

	// ── IP / IP6 ──────────────────────────────────────────────────────────────
	"ip": {
		"saddr", "daddr", "protocol",
		"dscp", "ecn", "length",
		"id", "frag-off", "ttl", "checksum",
		"version", "hdrlength",
	},
	"ip6": {
		"saddr", "daddr", "nexthdr", "hoplimit",
		"flowlabel", "dscp", "ecn", "length", "version",
	},
	"ip protocol": {
		"tcp", "udp", "udplite", "sctp", "dccp",
		"gre", "esp", "ah", "icmp", "icmpv6",
	},
	"ip6 nexthdr": {
		"tcp", "udp", "udplite", "sctp", "dccp",
		"gre", "esp", "ah", "icmpv6", "ipv6-icmp",
	},

	// ── TCP / UDP / SCTP / DCCP ───────────────────────────────────────────────
	"tcp": {
		"sport", "dport",
		"sequence", "ackseq", "doff", "reserved",
		"flags", "window", "checksum", "urgptr",
	},
	"udp":  {"sport", "dport", "length", "checksum"},
	"sctp": {"sport", "dport", "checksum"},
	"dccp": {"sport", "dport"},

	// ── ICMP ──────────────────────────────────────────────────────────────────
	"icmp":   {"type", "code", "checksum", "id", "sequence"},
	"icmpv6": {"type", "code", "checksum", "parameter-problem", "packet-too-big", "mtu", "max-delay"},
	"icmpx":  {"type"},

	// ── CT ───────────────────────────────────────────────────────────────────
	"ct": {
		"state", "status", "direction", "mark",
		"expiration", "helper", "label",
		"l4proto", "proto-src", "proto-dst",
		"saddr", "daddr", "zone", "count",
		"bytes", "packets", "id",
	},
	"ct state": {
		"new", "established", "related", "invalid", "untracked",
	},
	"ct status": {
		"expected", "seen-reply", "assured", "confirmed",
		"snat", "dnat", "dying",
	},
	"ct direction": {"original", "reply"},
	"ct l4proto":   {"tcp", "udp", "sctp", "dccp", "icmp", "icmpv6"},

	// ── META ─────────────────────────────────────────────────────────────────
	"meta": {
		"length", "protocol", "priority", "random", "mark",
		"iif", "iifname", "iiftype",
		"oif", "oifname", "oiftype",
		"skuid", "skgid", "nfproto", "l4proto",
		"cgroup", "pkttype", "cpu",
		"iifgroup", "oifgroup", "secpath",
	},
	"meta l4proto": {"tcp", "udp", "udplite", "sctp", "dccp", "icmp", "icmpv6"},
	"meta nfproto": {"ipv4", "ipv6", "inet"},
	"meta pkttype": {"host", "broadcast", "multicast", "other"},

	// ── REJECT ───────────────────────────────────────────────────────────────
	"reject": {"with"},
	"reject with": {
		"icmp type port-unreachable",
		"icmp type host-unreachable",
		"icmp type net-unreachable",
		"icmp type admin-prohibited",
		"icmpv6 type no-route",
		"icmpv6 type admin-prohibited",
		"icmpv6 type addr-unreachable",
		"icmpv6 type port-unreachable",
		"icmpx type port-unreachable",
		"icmpx type admin-prohibited",
		"tcp reset",
	},

	// ── LOG ──────────────────────────────────────────────────────────────────
	"log":       {"prefix", "level", "flags", "group"},
	"log level": {"emerg", "alert", "crit", "err", "warn", "notice", "info", "debug", "audit"},
	"log flags": {"tcp sequence", "tcp options", "ip options", "skuid", "ether", "all"},

	// ── LIMIT ────────────────────────────────────────────────────────────────
	"limit":      {"rate", "rate over"},
	"limit rate": {"1/second", "5/second", "10/second", "100/second", "1/minute", "10/minute", "1/hour"},
}

// comparisonOperators are tokens that separate a field name from its value.
// We skip them when determining context so "tcp dport != " still gives
// suggestions for port-related contexts.
var comparisonOperators = map[string]bool{
	"==": true, "!=": true, "<": true, ">": true, "<=": true, ">=": true,
	"eq": true, "ne": true, "lt": true, "gt": true, "le": true, "ge": true,
}

// GetRuleSuggestions returns context-aware nftables keyword suggestions for
// the text the user has typed so far. It returns an empty slice when no useful
// suggestions are available.
func GetRuleSuggestions(input string) []string {
	prevTokens, prefix, endsWithSpace := tokenizeRuleInput(input)

	key := buildContextKey(prevTokens)
	completions := grammar[key]

	// If nothing matched fall back to top-level to let the user add another
	// expression or a verdict after a value token they just finished typing.
	if completions == nil {
		completions = grammar[""]
	}

	// If input ends with a space the prefix is empty; return all completions.
	if endsWithSpace {
		return completions
	}

	return filterByPrefix(completions, prefix)
}

// tokenizeRuleInput splits the input into already-complete tokens and the
// partial token currently being typed (prefix).
// e.g. "ip protocol tc" → (["ip", "protocol"], "tc", false)
//
//	"ip protocol "  → (["ip", "protocol"], "",    true)
func tokenizeRuleInput(input string) (prevTokens []string, prefix string, endsWithSpace bool) {
	parts := strings.Fields(input)
	if len(input) > 0 && input[len(input)-1] == ' ' {
		return parts, "", true
	}
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], false
}

// buildContextKey inspects prevTokens and returns the best matching grammar
// key. It tries the last two meaningful tokens first, then the last one.
func buildContextKey(tokens []string) string {
	// Strip trailing comparison operators and their values so we can still
	// suggest the next meaningful keyword.
	meaningful := stripTrailingValues(tokens)
	n := len(meaningful)

	if n >= 2 {
		key := meaningful[n-2] + " " + meaningful[n-1]
		if _, ok := grammar[key]; ok {
			return key
		}
	}
	if n >= 1 {
		key := meaningful[n-1]
		if _, ok := grammar[key]; ok {
			return key
		}
	}
	return ""
}

// stripTrailingValues removes value tokens from the end of the slice so that
// context detection is not confused by IP addresses, port numbers, etc.
// A token is treated as a "value" if it is a comparison operator or if it
// follows a comparison operator.
func stripTrailingValues(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	skipNext := false
	for _, t := range tokens {
		if skipNext {
			skipNext = false
			continue
		}
		if comparisonOperators[t] {
			skipNext = true
			continue
		}
		result = append(result, t)
	}
	// If the last token looks like a bare value (number, IP, etc.) and is not
	// in the grammar as a standalone key, drop it to reveal the preceding context.
	for len(result) > 0 {
		last := result[len(result)-1]
		if isValueToken(last) {
			result = result[:len(result)-1]
		} else {
			break
		}
	}
	return result
}

// isValueToken returns true when a token is likely a concrete value rather
// than an nftables keyword: numbers, IPs, CIDR blocks, named ports, etc.
func isValueToken(t string) bool {
	if t == "" {
		return false
	}
	// Starts with a digit → port number, IP address, etc.
	if t[0] >= '0' && t[0] <= '9' {
		return true
	}
	// Contains a dot → IP or CIDR (e.g. 192.168.1.0/24)
	if strings.ContainsAny(t, "./") {
		return true
	}
	// Quoted strings
	if t[0] == '"' || t[0] == '\'' {
		return true
	}
	// Check that it's not a known keyword in any grammar entry
	for _, completions := range grammar {
		for _, kw := range completions {
			if kw == t {
				return false
			}
		}
	}
	// Not a recognised keyword → treat as value
	return true
}

// filterByPrefix returns only those completions that start with prefix
// (case-insensitive).
func filterByPrefix(completions []string, prefix string) []string {
	if prefix == "" {
		return completions
	}
	lower := strings.ToLower(prefix)
	var out []string
	for _, c := range completions {
		if strings.HasPrefix(strings.ToLower(c), lower) {
			out = append(out, c)
		}
	}
	return out
}
