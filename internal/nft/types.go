package nft

// Ruleset is the top-level nft JSON structure.
type Ruleset struct {
	Nftables []RulesetItem `json:"nftables"`
}

type RulesetItem struct {
	Metainfo  *Metainfo  `json:"metainfo,omitempty"`
	Table     *Table     `json:"table,omitempty"`
	Chain     *Chain     `json:"chain,omitempty"`
	Rule      *Rule      `json:"rule,omitempty"`
	Set       *Set       `json:"set,omitempty"`
	Map       *Map       `json:"map,omitempty"`
	Flowtable *Flowtable `json:"flowtable,omitempty"`
	Counter   *Counter   `json:"counter,omitempty"`
	Quota     *Quota     `json:"quota,omitempty"`
	Limit     *Limit     `json:"limit,omitempty"`
}

type Metainfo struct {
	Version           string `json:"version"`
	ReleaseName       string `json:"release_name"`
	JsonSchemaVersion int    `json:"json_schema_version"`
}

type Table struct {
	Family string `json:"family"`
	Name   string `json:"name"`
	Handle int    `json:"handle"`
}

type Chain struct {
	Family   string  `json:"family"`
	Table    string  `json:"table"`
	Name     string  `json:"name"`
	Handle   int     `json:"handle"`
	Type     string  `json:"type,omitempty"`
	Hook     string  `json:"hook,omitempty"`
	Prio     *int    `json:"prio,omitempty"`
	Policy   string  `json:"policy,omitempty"`
}

type Rule struct {
	Family  string      `json:"family"`
	Table   string      `json:"table"`
	Chain   string      `json:"chain"`
	Handle  int         `json:"handle"`
	Comment string      `json:"comment,omitempty"`
	Expr    interface{} `json:"expr,omitempty"`
}

type Set struct {
	Family string      `json:"family"`
	Table  string      `json:"table"`
	Name   string      `json:"name"`
	Handle int         `json:"handle"`
	Type   interface{} `json:"type,omitempty"`
	Flags  []string    `json:"flags,omitempty"`
	Elem   interface{} `json:"elem,omitempty"`
}

type Map struct {
	Family string      `json:"family"`
	Table  string      `json:"table"`
	Name   string      `json:"name"`
	Handle int         `json:"handle"`
	Type   interface{} `json:"type,omitempty"`
	Map    interface{} `json:"map,omitempty"`
	Flags  []string    `json:"flags,omitempty"`
	Elem   interface{} `json:"elem,omitempty"`
}

type Flowtable struct {
	Family  string   `json:"family"`
	Table   string   `json:"table"`
	Name    string   `json:"name"`
	Handle  int      `json:"handle"`
	Hook    string   `json:"hook,omitempty"`
	Prio    *int     `json:"prio,omitempty"`
	Devices []string `json:"dev,omitempty"`
}

type Counter struct {
	Family  string `json:"family"`
	Table   string `json:"table"`
	Name    string `json:"name"`
	Handle  int    `json:"handle"`
	Packets int64  `json:"packets,omitempty"`
	Bytes   int64  `json:"bytes,omitempty"`
}

type Quota struct {
	Family string `json:"family"`
	Table  string `json:"table"`
	Name   string `json:"name"`
	Handle int    `json:"handle"`
	Bytes  int64  `json:"bytes,omitempty"`
	Used   int64  `json:"used,omitempty"`
	Inv    bool   `json:"inv,omitempty"`
}

type Limit struct {
	Family   string `json:"family"`
	Table    string `json:"table"`
	Name     string `json:"name"`
	Handle   int    `json:"handle"`
	Rate     int64  `json:"rate,omitempty"`
	Per      string `json:"per,omitempty"`
	Burst    int64  `json:"burst,omitempty"`
	Unit     string `json:"unit,omitempty"`
	Inv      bool   `json:"inv,omitempty"`
}

// TableData holds all entities for a given table.
type TableData struct {
	Table      Table
	Chains     []Chain
	Rules      map[string][]Rule // keyed by chain name
	Sets       []Set
	Maps       []Map
	Flowtables []Flowtable
	Counters   []Counter
	Quotas     []Quota
	Limits     []Limit
}
