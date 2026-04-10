package nft

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// run executes an nft command and returns combined output.
func run(args ...string) ([]byte, error) {
	cmd := exec.Command("nft", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nft %s: %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return out, nil
}

// ListRuleset returns the full parsed ruleset.
func ListRuleset() (*Ruleset, error) {
	out, err := run("-j", "list", "ruleset")
	if err != nil {
		return nil, err
	}
	var rs Ruleset
	if err := json.Unmarshal(out, &rs); err != nil {
		return nil, fmt.Errorf("parse ruleset: %w", err)
	}
	return &rs, nil
}

// ParseRuleset organises a flat ruleset into per-table TableData.
func ParseRuleset(rs *Ruleset) map[string]*TableData {
	tables := make(map[string]*TableData)
	key := func(family, name string) string { return family + "/" + name }

	for _, item := range rs.Nftables {
		switch {
		case item.Table != nil:
			t := item.Table
			k := key(t.Family, t.Name)
			tables[k] = &TableData{
				Table: *t,
				Rules: make(map[string][]Rule),
			}
		case item.Chain != nil:
			c := item.Chain
			k := key(c.Family, c.Table)
			if td, ok := tables[k]; ok {
				td.Chains = append(td.Chains, *c)
			}
		case item.Rule != nil:
			r := item.Rule
			k := key(r.Family, r.Table)
			if td, ok := tables[k]; ok {
				td.Rules[r.Chain] = append(td.Rules[r.Chain], *r)
			}
		case item.Set != nil:
			s := item.Set
			k := key(s.Family, s.Table)
			if td, ok := tables[k]; ok {
				td.Sets = append(td.Sets, *s)
			}
		case item.Map != nil:
			m := item.Map
			k := key(m.Family, m.Table)
			if td, ok := tables[k]; ok {
				td.Maps = append(td.Maps, *m)
			}
		case item.Flowtable != nil:
			f := item.Flowtable
			k := key(f.Family, f.Table)
			if td, ok := tables[k]; ok {
				td.Flowtables = append(td.Flowtables, *f)
			}
		case item.Counter != nil:
			c := item.Counter
			k := key(c.Family, c.Table)
			if td, ok := tables[k]; ok {
				td.Counters = append(td.Counters, *c)
			}
		case item.Quota != nil:
			q := item.Quota
			k := key(q.Family, q.Table)
			if td, ok := tables[k]; ok {
				td.Quotas = append(td.Quotas, *q)
			}
		case item.Limit != nil:
			l := item.Limit
			k := key(l.Family, l.Table)
			if td, ok := tables[k]; ok {
				td.Limits = append(td.Limits, *l)
			}
		}
	}
	return tables
}

// --- Table CRUD ---

func AddTable(family, name string) error {
	_, err := run("add", "table", family, name)
	return err
}

func DeleteTable(family, name string) error {
	_, err := run("delete", "table", family, name)
	return err
}

func FlushTable(family, name string) error {
	_, err := run("flush", "table", family, name)
	return err
}

// --- Chain CRUD ---

func AddChain(family, table, name string) error {
	_, err := run("add", "chain", family, table, name)
	return err
}

func AddBaseChain(family, table, name, chainType, hook, policy string, prio int) error {
	_, err := run("add", "chain", family, table, name,
		fmt.Sprintf("{ type %s hook %s priority %d ; policy %s ; }", chainType, hook, prio, policy))
	return err
}

func DeleteChain(family, table, name string) error {
	_, err := run("delete", "chain", family, table, name)
	return err
}

func FlushChain(family, table, name string) error {
	_, err := run("flush", "chain", family, table, name)
	return err
}

// --- Rule CRUD ---

// AddRule appends a rule to a chain. ruleText is the raw nft rule expression,
// e.g. "ip protocol tcp accept". comment is optional; pass "" to omit it.
func AddRule(family, table, chain, ruleText, comment string) error {
	args := []string{"add", "rule", family, table, chain, ruleText}
	if comment != "" {
		args = append(args, "comment", fmt.Sprintf("%q", comment))
	}
	_, err := run(args...)
	return err
}

// DeleteRule removes a rule by handle.
func DeleteRule(family, table, chain string, handle int) error {
	_, err := run("delete", "rule", family, table, chain, "handle", fmt.Sprintf("%d", handle))
	return err
}

// ReplaceRule replaces a rule at handle with new ruleText. comment is optional; pass "" to omit it.
func ReplaceRule(family, table, chain string, handle int, ruleText, comment string) error {
	args := []string{"replace", "rule", family, table, chain, "handle", fmt.Sprintf("%d", handle), ruleText}
	if comment != "" {
		args = append(args, "comment", fmt.Sprintf("%q", comment))
	}
	_, err := run(args...)
	return err
}

// InsertRule inserts a rule before the given handle (0 = at start).
func InsertRule(family, table, chain string, handle int, ruleText string) error {
	if handle > 0 {
		_, err := run("insert", "rule", family, table, chain, "handle", fmt.Sprintf("%d", handle), ruleText)
		return err
	}
	_, err := run("insert", "rule", family, table, chain, ruleText)
	return err
}

// --- Set CRUD ---

func AddSet(family, table, name, setType string, flags []string) error {
	flagStr := ""
	if len(flags) > 0 {
		flagStr = "flags " + strings.Join(flags, ",") + " ; "
	}
	_, err := run("add", "set", family, table, name,
		fmt.Sprintf("{ type %s ; %s}", setType, flagStr))
	return err
}

func DeleteSet(family, table, name string) error {
	_, err := run("delete", "set", family, table, name)
	return err
}

func FlushSet(family, table, name string) error {
	_, err := run("flush", "set", family, table, name)
	return err
}

// --- Map CRUD ---

func AddMap(family, table, name, keyType, valType string) error {
	_, err := run("add", "map", family, table, name,
		fmt.Sprintf("{ type %s : %s ; }", keyType, valType))
	return err
}

func DeleteMap(family, table, name string) error {
	_, err := run("delete", "map", family, table, name)
	return err
}

// --- Flowtable CRUD ---

func AddFlowtable(family, table, name string, devices []string) error {
	devStr := ""
	if len(devices) > 0 {
		devStr = "devices = { " + strings.Join(devices, ", ") + " } ; "
	}
	_, err := run("add", "flowtable", family, table, name,
		fmt.Sprintf("{ hook ingress priority 0 ; %s}", devStr))
	return err
}

func DeleteFlowtable(family, table, name string) error {
	_, err := run("delete", "flowtable", family, table, name)
	return err
}

// --- Counter CRUD ---

func AddCounter(family, table, name string) error {
	_, err := run("add", "counter", family, table, name)
	return err
}

func DeleteCounter(family, table, name string) error {
	_, err := run("delete", "counter", family, table, name)
	return err
}

// GetRuleText returns human-readable rule text for display.
// -a makes nft annotate each rule with "# handle N".
func GetRuleText(family, table, chain string, handle int) string {
	out, err := run("-a", "list", "chain", family, table, chain)
	if err != nil {
		return fmt.Sprintf("<handle %d>", handle)
	}
	lines := strings.Split(string(out), "\n")
	marker := fmt.Sprintf("# handle %d", handle)
	for _, l := range lines {
		if strings.Contains(l, marker) {
			idx := strings.Index(l, " # handle")
			if idx >= 0 {
				return strings.TrimSpace(l[:idx])
			}
			return strings.TrimSpace(l)
		}
	}
	return fmt.Sprintf("<handle %d>", handle)
}

// ListChainRuleTexts returns a map of handle -> human-readable rule text for a chain.
// -a makes nft annotate each rule with "# handle N".
func ListChainRuleTexts(family, table, chain string) map[int]string {
	out, err := run("-a", "list", "chain", family, table, chain)
	if err != nil {
		return nil
	}
	result := make(map[int]string)
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if !strings.Contains(trimmed, "# handle") {
			continue
		}
		idx := strings.Index(trimmed, " # handle ")
		if idx < 0 {
			continue
		}
		ruleText := strings.TrimSpace(trimmed[:idx])
		handleStr := strings.TrimSpace(trimmed[idx+len(" # handle "):])
		var handle int
		fmt.Sscanf(handleStr, "%d", &handle)
		result[handle] = ruleText
	}
	return result
}
