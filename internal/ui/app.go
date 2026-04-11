package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pzindyaev/nfrows/internal/nft"
)

type viewKind int

const (
	viewTables viewKind = iota
	viewTableDetail
	viewChainRules
)

type modalKind int

const (
	modalNone modalKind = iota
	modalAddTable
	modalDeleteTable
	modalFlushTable
	modalAddChain
	modalAddBaseChain
	modalDeleteChain
	modalFlushChain
	modalAddRule
	modalEditRule
	modalDeleteRule
	modalAddSet
	modalDeleteSet
	modalAddMap
	modalDeleteMap
	modalAddFlowtable
	modalDeleteFlowtable
	modalAddCounter
	modalDeleteCounter
)

// tabKind enumerates the table-detail tabs.
type tabKind int

const (
	tabChains tabKind = iota
	tabSets
	tabMaps
	tabFlowtables
	tabCounters
	tabCount
)

var tabNames = []string{"Chains", "Sets", "Maps", "Flowtables", "Counters"}

// App is the root bubbletea model.
type App struct {
	width  int
	height int

	// Data.
	tables    map[string]*nft.TableData
	tableKeys []string // sorted
	loading   bool
	flashMsg  string
	flashErr  bool

	// Navigation.
	view viewKind

	// Tables view.
	tablesTbl TableWidget

	// Table-detail view.
	activeTableKey string
	activeTab      tabKind
	chainsTbl      TableWidget
	setsTbl        TableWidget
	mapsTbl        TableWidget
	flowtablesTbl  TableWidget
	countersTbl    TableWidget

	// Chain-rules view.
	activeChain  string
	ruleTexts    map[int]string // handle → text
	rulesTbl     TableWidget

	// Modal.
	modal     modalKind
	form      Form
	confirm   ConfirmDialog

	// Vim key-sequence state.
	pendingG bool // true after a lone 'g' press, waiting for 'gg'
}

func NewApp() App {
	return App{
		loading: true,
		view:    viewTables,
		tablesTbl: TableWidget{
			Columns: []Column{
				{Title: "Family", Width: 8},
				{Title: "Name", Width: 20, Flex: true},
				{Title: "Chains", Width: 7},
				{Title: "Sets", Width: 6},
				{Title: "Maps", Width: 6},
			},
			Height: 20,
		},
	}
}

func (a App) Init() tea.Cmd {
	return cmdLoadRuleset()
}

// ---------- Update ----------

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.resizeTables()
		return a, nil

	case rulesetLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.flashMsg = msg.err.Error()
			a.flashErr = true
		} else {
			a.tables = msg.tables
			a.buildTableKeys()
			a.rebuildCurrentView()
		}
		return a, nil

	case ruleTextsMsg:
		a.ruleTexts = msg.texts
		a.rebuildRulesTable()
		return a, nil

	case opResultMsg:
		if msg.err != nil {
			a.flashMsg = msg.err.Error()
			a.flashErr = true
		} else {
			a.flashMsg = "Done."
			a.flashErr = false
		}
		a.modal = modalNone
		return a, cmdLoadRuleset()

	case tea.KeyMsg:
		// Modal takes priority.
		if a.modal != modalNone {
			return a.updateModal(msg)
		}
		return a.updateNav(msg)
	}

	return a, nil
}

func (a App) updateNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// --- Vim gg sequence ---
	if a.pendingG {
		a.pendingG = false
		if isKey(msg, "g") {
			a.activeTabTableOrCurrent().GoToTop()
			return a, nil
		}
		// Not a gg sequence; fall through and handle the new key normally.
	}

	switch {
	case isKey(msg, "ctrl+c"), isKey(msg, "q"):
		if a.view == viewTables {
			return a, tea.Quit
		}
		a.goBack()
		return a, nil

	case isKey(msg, "esc"):
		a.pendingG = false
		a.goBack()
		return a, nil

	case isKey(msg, "r"):
		a.loading = true
		a.flashMsg = ""
		return a, cmdLoadRuleset()

	// --- Movement ---
	case isKey(msg, "up"), isKey(msg, "k"):
		a.moveUp()
		return a, nil

	case isKey(msg, "down"), isKey(msg, "j"):
		return a.moveDown()

	case isKey(msg, "G"):
		a.activeTabTableOrCurrent().GoToBottom()
		return a, nil

	case isKey(msg, "g"):
		a.pendingG = true
		return a, nil

	case isKey(msg, "ctrl+u"):
		a.activeTabTableOrCurrent().HalfPageUp()
		return a, nil

	case isKey(msg, "ctrl+d"):
		a.activeTabTableOrCurrent().HalfPageDown()
		return a, nil

	// --- Tab switching with h/l ---
	case isKey(msg, "h"):
		if a.view == viewTableDetail {
			a.activeTab = (a.activeTab - 1 + tabCount) % tabCount
		}
		return a, nil

	case isKey(msg, "l"):
		if a.view == viewTableDetail {
			a.activeTab = (a.activeTab + 1) % tabCount
		}
		return a, nil

	case isKey(msg, "enter"):
		return a.handleEnter()

	case isKey(msg, "tab"):
		if a.view == viewTableDetail {
			a.activeTab = (a.activeTab + 1) % tabCount
		}
		return a, nil

	case isKey(msg, "shift+tab"):
		if a.view == viewTableDetail {
			a.activeTab = (a.activeTab - 1 + tabCount) % tabCount
		}
		return a, nil

	case isKey(msg, "a"), isKey(msg, "o"):
		return a.handleAdd()

	case isKey(msg, "d"):
		return a.handleDelete()

	case isKey(msg, "e"), isKey(msg, "i"):
		return a.handleEdit()

	case isKey(msg, "f"):
		return a.handleFlush()
	}
	return a, nil
}

func (a App) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.modal {
	case modalDeleteTable, modalDeleteChain, modalDeleteRule,
		modalFlushTable, modalFlushChain,
		modalDeleteSet, modalDeleteMap, modalDeleteFlowtable, modalDeleteCounter:
		// Confirm dialog.
		switch msg.String() {
		case "y":
			return a.executeModal()
		case "n", "esc":
			a.modal = modalNone
		}
		return a, nil
	}

	// Form modals.
	switch msg.String() {
	case "enter":
		if !a.form.Validate() {
			return a, nil
		}
		return a.executeModal()
	case "esc":
		// Let the form dismiss its autocomplete dropdown first.
		if a.form.ConsumeEsc() {
			return a, nil
		}
		a.modal = modalNone
		return a, nil
	}
	var cmd tea.Cmd
	a.form, cmd = a.form.Update(msg)
	return a, cmd
}

func (a App) executeModal() (tea.Model, tea.Cmd) {
	vals := a.form.Values()
	td := a.activeTableData()

	switch a.modal {

	// --- Table ---
	case modalAddTable:
		family, name := vals[0], vals[1]
		return a, cmdOp(func() error { return nft.AddTable(family, name) })
	case modalDeleteTable:
		if td == nil {
			return a, nil
		}
		t := td.Table
		return a, cmdOp(func() error { return nft.DeleteTable(t.Family, t.Name) })
	case modalFlushTable:
		if td == nil {
			return a, nil
		}
		t := td.Table
		return a, cmdOp(func() error { return nft.FlushTable(t.Family, t.Name) })

	// --- Chain ---
	case modalAddChain:
		if td == nil {
			return a, nil
		}
		t := td.Table
		name := vals[0]
		return a, cmdOp(func() error { return nft.AddChain(t.Family, t.Name, name) })
	case modalAddBaseChain:
		if td == nil {
			return a, nil
		}
		t := td.Table
		name, chainType, hook, policy, prioStr := vals[0], vals[1], vals[2], vals[3], vals[4]
		var prio int
		fmt.Sscanf(prioStr, "%d", &prio)
		return a, cmdOp(func() error {
			return nft.AddBaseChain(t.Family, t.Name, name, chainType, hook, policy, prio)
		})
	case modalDeleteChain:
		if td == nil {
			return a, nil
		}
		t := td.Table
		chain := a.activeChain
		if chain == "" {
			if row, ok := a.chainsTbl.SelectedRow(); ok {
				chain = row[1]
			}
		}
		return a, cmdOp(func() error { return nft.DeleteChain(t.Family, t.Name, chain) })
	case modalFlushChain:
		if td == nil {
			return a, nil
		}
		t := td.Table
		chain := a.activeChain
		if chain == "" {
			if row, ok := a.chainsTbl.SelectedRow(); ok {
				chain = row[1]
			}
		}
		return a, cmdOp(func() error { return nft.FlushChain(t.Family, t.Name, chain) })

	// --- Rule ---
	case modalAddRule:
		if td == nil {
			return a, nil
		}
		t := td.Table
		chain := a.activeChain
		ruleText, comment := vals[0], vals[1]
		return a, cmdOp(func() error { return nft.AddRule(t.Family, t.Name, chain, ruleText, comment) })
	case modalEditRule:
		if td == nil {
			return a, nil
		}
		t := td.Table
		chain := a.activeChain
		row, ok := a.rulesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		var handle int
		fmt.Sscanf(row[0], "%d", &handle)
		ruleText, comment := vals[0], vals[1]
		return a, cmdOp(func() error { return nft.ReplaceRule(t.Family, t.Name, chain, handle, ruleText, comment) })
	case modalDeleteRule:
		if td == nil {
			return a, nil
		}
		t := td.Table
		chain := a.activeChain
		row, ok := a.rulesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		var handle int
		fmt.Sscanf(row[0], "%d", &handle)
		return a, cmdOp(func() error { return nft.DeleteRule(t.Family, t.Name, chain, handle) })

	// --- Set ---
	case modalAddSet:
		if td == nil {
			return a, nil
		}
		t := td.Table
		name, setType, flagStr := vals[0], vals[1], vals[2]
		var flags []string
		if strings.TrimSpace(flagStr) != "" {
			for _, f := range strings.Split(flagStr, ",") {
				flags = append(flags, strings.TrimSpace(f))
			}
		}
		return a, cmdOp(func() error { return nft.AddSet(t.Family, t.Name, name, setType, flags) })
	case modalDeleteSet:
		if td == nil {
			return a, nil
		}
		t := td.Table
		row, ok := a.setsTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		name := row[1]
		return a, cmdOp(func() error { return nft.DeleteSet(t.Family, t.Name, name) })

	// --- Map ---
	case modalAddMap:
		if td == nil {
			return a, nil
		}
		t := td.Table
		name, keyType, valType := vals[0], vals[1], vals[2]
		return a, cmdOp(func() error { return nft.AddMap(t.Family, t.Name, name, keyType, valType) })
	case modalDeleteMap:
		if td == nil {
			return a, nil
		}
		t := td.Table
		row, ok := a.mapsTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		name := row[1]
		return a, cmdOp(func() error { return nft.DeleteMap(t.Family, t.Name, name) })

	// --- Flowtable ---
	case modalAddFlowtable:
		if td == nil {
			return a, nil
		}
		t := td.Table
		name, devStr := vals[0], vals[1]
		var devices []string
		for _, d := range strings.Split(devStr, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				devices = append(devices, d)
			}
		}
		return a, cmdOp(func() error { return nft.AddFlowtable(t.Family, t.Name, name, devices) })
	case modalDeleteFlowtable:
		if td == nil {
			return a, nil
		}
		t := td.Table
		row, ok := a.flowtablesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		name := row[1]
		return a, cmdOp(func() error { return nft.DeleteFlowtable(t.Family, t.Name, name) })

	// --- Counter ---
	case modalAddCounter:
		if td == nil {
			return a, nil
		}
		t := td.Table
		name := vals[0]
		return a, cmdOp(func() error { return nft.AddCounter(t.Family, t.Name, name) })
	case modalDeleteCounter:
		if td == nil {
			return a, nil
		}
		t := td.Table
		row, ok := a.countersTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		name := row[1]
		return a, cmdOp(func() error { return nft.DeleteCounter(t.Family, t.Name, name) })
	}

	return a, nil
}

// ---------- Navigation helpers ----------

func (a *App) goBack() {
	switch a.view {
	case viewChainRules:
		a.view = viewTableDetail
		a.activeChain = ""
		a.ruleTexts = nil
	case viewTableDetail:
		a.view = viewTables
		a.activeTableKey = ""
	}
}

func (a *App) moveUp() {
	switch a.view {
	case viewTables:
		a.tablesTbl.MoveUp()
	case viewTableDetail:
		a.activeTabTable().MoveUp()
	case viewChainRules:
		a.rulesTbl.MoveUp()
	}
}

func (a App) moveDown() (tea.Model, tea.Cmd) {
	switch a.view {
	case viewTables:
		a.tablesTbl.MoveDown()
	case viewTableDetail:
		a.activeTabTable().MoveDown()
	case viewChainRules:
		a.rulesTbl.MoveDown()
	}
	return a, nil
}

func (a App) handleEnter() (tea.Model, tea.Cmd) {
	switch a.view {
	case viewTables:
		row, ok := a.tablesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		// row: family, name, chains, sets, maps
		key := row[0] + "/" + row[1]
		a.activeTableKey = key
		a.activeTab = tabChains
		a.view = viewTableDetail
		a.rebuildDetailTables()

	case viewTableDetail:
		if a.activeTab == tabChains {
			row, ok := a.chainsTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.activeChain = row[1]
			a.view = viewChainRules
			a.ruleTexts = nil
			a.rebuildRulesTable()
			td := a.activeTableData()
			if td != nil {
				return a, cmdLoadRuleTexts(td.Table.Family, td.Table.Name, a.activeChain)
			}
		}
	}
	return a, nil
}

func (a App) handleAdd() (tea.Model, tea.Cmd) {
	switch a.view {
	case viewTables:
		a.modal = modalAddTable
		a.form = NewForm("Add Table", []FormField{
			{Label: "Family", Required: true,
				Options: []string{"ip", "ip6", "inet", "arp", "bridge", "netdev"}},
			{Label: "Name", Placeholder: "my_table", Required: true},
		})
		a.form.Width = 50

	case viewTableDetail:
		switch a.activeTab {
		case tabChains:
			a.modal = modalAddBaseChain
			a.form = NewForm("Add Chain", []FormField{
				{Label: "Name", Placeholder: "my_chain", Required: true},
				{Label: "Type", Options: []string{"filter", "nat", "route"}, Required: true},
				{Label: "Hook", Options: []string{"prerouting", "input", "forward", "output", "postrouting", "ingress"}, Required: true},
				{Label: "Policy", Options: []string{"accept", "drop"}, Required: true},
				{Label: "Priority", Placeholder: "0", Value: "0", Required: true},
			})
			a.form.Width = 60
		case tabSets:
			a.modal = modalAddSet
			a.form = NewForm("Add Set", []FormField{
				{Label: "Name", Placeholder: "my_set", Required: true},
				{Label: "Type", Placeholder: "ipv4_addr", Required: true},
				{Label: "Flags (csv)", Placeholder: "interval,timeout"},
			})
			a.form.Width = 55
		case tabMaps:
			a.modal = modalAddMap
			a.form = NewForm("Add Map", []FormField{
				{Label: "Name", Placeholder: "my_map", Required: true},
				{Label: "Key type", Placeholder: "ipv4_addr", Required: true},
				{Label: "Value type", Placeholder: "verdict", Required: true},
			})
			a.form.Width = 55
		case tabFlowtables:
			a.modal = modalAddFlowtable
			a.form = NewForm("Add Flowtable", []FormField{
				{Label: "Name", Placeholder: "my_ft", Required: true},
				{Label: "Devices (csv)", Placeholder: "eth0,eth1"},
			})
			a.form.Width = 55
		case tabCounters:
			a.modal = modalAddCounter
			a.form = NewForm("Add Counter", []FormField{
				{Label: "Name", Placeholder: "my_counter", Required: true},
			})
			a.form.Width = 45
		}

	case viewChainRules:
		a.modal = modalAddRule
		a.form = NewForm("Add Rule", []FormField{
			{Label: "Rule", Placeholder: "ip protocol tcp accept", Required: true, Autocomplete: true},
			{Label: "Comment", Placeholder: "optional description"},
		})
		a.form.Width = 70
	}
	return a, nil
}

func (a App) handleDelete() (tea.Model, tea.Cmd) {
	switch a.view {
	case viewTables:
		row, ok := a.tablesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		a.modal = modalDeleteTable
		a.confirm = ConfirmDialog{
			Message: fmt.Sprintf("Delete table %s/%s?", row[0], row[1]),
			Width:   a.width / 2,
		}
	case viewTableDetail:
		switch a.activeTab {
		case tabChains:
			row, ok := a.chainsTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.modal = modalDeleteChain
			a.confirm = ConfirmDialog{
				Message: fmt.Sprintf("Delete chain %q?", row[1]),
				Width:   a.width / 2,
			}
		case tabSets:
			row, ok := a.setsTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.modal = modalDeleteSet
			a.confirm = ConfirmDialog{Message: fmt.Sprintf("Delete set %q?", row[1]), Width: a.width / 2}
		case tabMaps:
			row, ok := a.mapsTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.modal = modalDeleteMap
			a.confirm = ConfirmDialog{Message: fmt.Sprintf("Delete map %q?", row[1]), Width: a.width / 2}
		case tabFlowtables:
			row, ok := a.flowtablesTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.modal = modalDeleteFlowtable
			a.confirm = ConfirmDialog{Message: fmt.Sprintf("Delete flowtable %q?", row[1]), Width: a.width / 2}
		case tabCounters:
			row, ok := a.countersTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.modal = modalDeleteCounter
			a.confirm = ConfirmDialog{Message: fmt.Sprintf("Delete counter %q?", row[1]), Width: a.width / 2}
		}
	case viewChainRules:
		row, ok := a.rulesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		a.modal = modalDeleteRule
		a.confirm = ConfirmDialog{
			Message: fmt.Sprintf("Delete rule handle %s?", row[0]),
			Width:   a.width / 2,
		}
	}
	return a, nil
}

func (a App) handleEdit() (tea.Model, tea.Cmd) {
	if a.view == viewChainRules {
		row, ok := a.rulesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		existingComment := row[2]
		if existingComment == "-" {
			existingComment = ""
		}
		a.modal = modalEditRule
		a.form = NewForm("Edit Rule", []FormField{
			{Label: "Rule", Value: row[1], Required: true, Autocomplete: true},
			{Label: "Comment", Value: existingComment, Placeholder: "optional description"},
		})
		a.form.Width = 70
	}
	return a, nil
}

func (a App) handleFlush() (tea.Model, tea.Cmd) {
	switch a.view {
	case viewTables:
		row, ok := a.tablesTbl.SelectedRow()
		if !ok {
			return a, nil
		}
		a.modal = modalFlushTable
		a.confirm = ConfirmDialog{
			Message: fmt.Sprintf("Flush all rules in table %s/%s?", row[0], row[1]),
			Width:   a.width / 2,
		}
	case viewTableDetail:
		if a.activeTab == tabChains {
			row, ok := a.chainsTbl.SelectedRow()
			if !ok {
				return a, nil
			}
			a.modal = modalFlushChain
			a.confirm = ConfirmDialog{
				Message: fmt.Sprintf("Flush all rules in chain %q?", row[1]),
				Width:   a.width / 2,
			}
		}
	}
	return a, nil
}

// ---------- Data helpers ----------

func (a *App) buildTableKeys() {
	keys := make([]string, 0, len(a.tables))
	for k := range a.tables {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	a.tableKeys = keys
	a.rebuildTablesTable()
}

func (a *App) rebuildCurrentView() {
	switch a.view {
	case viewTables:
		a.rebuildTablesTable()
	case viewTableDetail:
		a.rebuildDetailTables()
	case viewChainRules:
		a.rebuildRulesTable()
	}
}

func (a *App) rebuildTablesTable() {
	rows := make([]Row, 0, len(a.tableKeys))
	for _, k := range a.tableKeys {
		td := a.tables[k]
		rows = append(rows, Row{
			td.Table.Family,
			td.Table.Name,
			fmt.Sprintf("%d", len(td.Chains)),
			fmt.Sprintf("%d", len(td.Sets)),
			fmt.Sprintf("%d", len(td.Maps)),
		})
	}
	a.tablesTbl.SetRows(rows)
}

func (a *App) rebuildDetailTables() {
	td := a.activeTableData()
	if td == nil {
		return
	}

	// Chains.
	chainRows := make([]Row, 0, len(td.Chains))
	for _, c := range td.Chains {
		hook := c.Hook
		if hook == "" {
			hook = "-"
		}
		policy := c.Policy
		if policy == "" {
			policy = "-"
		}
		chainType := c.Type
		if chainType == "" {
			chainType = "regular"
		}
		rules := len(td.Rules[c.Name])
		chainRows = append(chainRows, Row{
			fmt.Sprintf("%d", c.Handle),
			c.Name,
			chainType,
			hook,
			policy,
			fmt.Sprintf("%d", rules),
		})
	}
	a.chainsTbl = TableWidget{
		Columns: []Column{
			{Title: "Handle", Width: 7},
			{Title: "Name", Width: 20, Flex: true},
			{Title: "Type", Width: 8},
			{Title: "Hook", Width: 12},
			{Title: "Policy", Width: 8},
			{Title: "Rules", Width: 6},
		},
		Rows:   chainRows,
		Height: a.contentHeight(),
		Width:  a.width - 2,
	}

	// Sets.
	setRows := make([]Row, 0, len(td.Sets))
	for _, s := range td.Sets {
		setRows = append(setRows, Row{
			fmt.Sprintf("%d", s.Handle),
			s.Name,
			fmt.Sprintf("%v", s.Type),
			strings.Join(s.Flags, ","),
		})
	}
	a.setsTbl = TableWidget{
		Columns: []Column{
			{Title: "Handle", Width: 7},
			{Title: "Name", Width: 24, Flex: true},
			{Title: "Type", Width: 16},
			{Title: "Flags", Width: 20},
		},
		Rows:   setRows,
		Height: a.contentHeight(),
		Width:  a.width - 2,
	}

	// Maps.
	mapRows := make([]Row, 0, len(td.Maps))
	for _, m := range td.Maps {
		mapRows = append(mapRows, Row{
			fmt.Sprintf("%d", m.Handle),
			m.Name,
			fmt.Sprintf("%v", m.Type),
			fmt.Sprintf("%v", m.Map),
		})
	}
	a.mapsTbl = TableWidget{
		Columns: []Column{
			{Title: "Handle", Width: 7},
			{Title: "Name", Width: 24, Flex: true},
			{Title: "Key type", Width: 16},
			{Title: "Val type", Width: 16},
		},
		Rows:   mapRows,
		Height: a.contentHeight(),
		Width:  a.width - 2,
	}

	// Flowtables.
	ftRows := make([]Row, 0, len(td.Flowtables))
	for _, f := range td.Flowtables {
		ftRows = append(ftRows, Row{
			fmt.Sprintf("%d", f.Handle),
			f.Name,
			f.Hook,
			strings.Join(f.Devices, ","),
		})
	}
	a.flowtablesTbl = TableWidget{
		Columns: []Column{
			{Title: "Handle", Width: 7},
			{Title: "Name", Width: 24, Flex: true},
			{Title: "Hook", Width: 12},
			{Title: "Devices", Width: 20},
		},
		Rows:   ftRows,
		Height: a.contentHeight(),
		Width:  a.width - 2,
	}

	// Counters.
	cntRows := make([]Row, 0, len(td.Counters))
	for _, c := range td.Counters {
		cntRows = append(cntRows, Row{
			fmt.Sprintf("%d", c.Handle),
			c.Name,
			fmt.Sprintf("%d", c.Packets),
			fmt.Sprintf("%d", c.Bytes),
		})
	}
	for _, q := range td.Quotas {
		cntRows = append(cntRows, Row{
			fmt.Sprintf("%d", q.Handle),
			q.Name + " (quota)",
			"-",
			fmt.Sprintf("%d/%d", q.Used, q.Bytes),
		})
	}
	for _, l := range td.Limits {
		cntRows = append(cntRows, Row{
			fmt.Sprintf("%d", l.Handle),
			l.Name + " (limit)",
			"-",
			fmt.Sprintf("%d/%s", l.Rate, l.Per),
		})
	}
	a.countersTbl = TableWidget{
		Columns: []Column{
			{Title: "Handle", Width: 7},
			{Title: "Name", Width: 24, Flex: true},
			{Title: "Packets", Width: 12},
			{Title: "Bytes/Info", Width: 16},
		},
		Rows:   cntRows,
		Height: a.contentHeight(),
		Width:  a.width - 2,
	}
}

func (a *App) rebuildRulesTable() {
	td := a.activeTableData()
	if td == nil {
		return
	}
	rules := td.Rules[a.activeChain]
	rows := make([]Row, 0, len(rules))
	for _, r := range rules {
		text := fmt.Sprintf("<handle %d>", r.Handle)
		if a.ruleTexts != nil {
			if t, ok := a.ruleTexts[r.Handle]; ok {
				text = t
			}
		}
		comment := r.Comment
		if comment == "" {
			comment = "-"
		}
		rows = append(rows, Row{
			fmt.Sprintf("%d", r.Handle),
			text,
			comment,
		})
	}
	a.rulesTbl = TableWidget{
		Columns: []Column{
			{Title: "Handle", Width: 7},
			{Title: "Rule", Flex: true},
			{Title: "Comment", Width: 20},
		},
		Rows:   rows,
		Height: a.contentHeight(),
		Width:  a.width - 2,
	}
}

func (a *App) activeTableData() *nft.TableData {
	if a.activeTableKey == "" {
		return nil
	}
	return a.tables[a.activeTableKey]
}

func (a *App) activeTabTable() *TableWidget {
	switch a.activeTab {
	case tabChains:
		return &a.chainsTbl
	case tabSets:
		return &a.setsTbl
	case tabMaps:
		return &a.mapsTbl
	case tabFlowtables:
		return &a.flowtablesTbl
	case tabCounters:
		return &a.countersTbl
	}
	return &a.chainsTbl
}

// activeTabTableOrCurrent returns the focused table widget regardless of view.
func (a *App) activeTabTableOrCurrent() *TableWidget {
	switch a.view {
	case viewTables:
		return &a.tablesTbl
	case viewTableDetail:
		return a.activeTabTable()
	case viewChainRules:
		return &a.rulesTbl
	}
	return &a.tablesTbl
}

func (a *App) resizeTables() {
	h := a.contentHeight()
	inner := a.width - 2 // border takes 1 char on each side
	a.tablesTbl.Width = inner
	a.tablesTbl.Height = h
	a.chainsTbl.Width = inner
	a.chainsTbl.Height = h
	a.setsTbl.Width = inner
	a.setsTbl.Height = h
	a.mapsTbl.Width = inner
	a.mapsTbl.Height = h
	a.flowtablesTbl.Width = inner
	a.flowtablesTbl.Height = h
	a.countersTbl.Width = inner
	a.countersTbl.Height = h
	a.rulesTbl.Width = inner
	a.rulesTbl.Height = h
}

// contentHeight returns the number of data rows the table body can show.
//
// Actual rendered line budget per view (sections joined with "\n"):
//
//	All views:    header(1) + breadcrumb(1) + flash-or-blank(1) + statusbar(1)  = 4
//	              + border-top(1) + col-header(1) + separator(1) + border-bot(1) = 4
//	viewTableDetail only: tabs-bar(1) + tabs-border-bottom(1)                   = 2
//
// Total fixed overhead: 8 (tables/chain-rules) or 10 (table-detail).
func (a *App) contentHeight() int {
	overhead := 8
	if a.view == viewTableDetail {
		overhead = 10
	}
	h := a.height - overhead
	if h < 5 {
		h = 5
	}
	return h
}

// ---------- View ----------

func (a App) View() string {
	if a.loading {
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("Loading nftables ruleset…"),
		)
	}

	var sections []string

	// Header.
	sections = append(sections, a.viewHeader())

	// Breadcrumb.
	sections = append(sections, a.viewBreadcrumb())

	// Tabs (only in table detail and chain views).
	if a.view == viewTableDetail {
		sections = append(sections, a.viewTabs())
	}

	// Main content.
	sections = append(sections, a.viewContent())

	// Flash message.
	sections = append(sections, a.viewFlash())

	// Status bar / key hints.
	sections = append(sections, a.viewStatusBar())

	body := strings.Join(sections, "\n")

	// Modal overlay.
	if a.modal != modalNone {
		body = a.overlayModal(body)
	}

	return body
}

func (a App) viewHeader() string {
	left := styleTitle.Render("nfrows")
	right := lipgloss.NewStyle().Foreground(colorSubtext).Padding(0, 1).
		Render("nftables TUI")
	gap := strings.Repeat(" ", max(0, a.width-lipgloss.Width(left)-lipgloss.Width(right)))
	return lipgloss.NewStyle().
		Background(colorSecondary).
		Render(left + gap + right)
}

func (a App) viewBreadcrumb() string {
	sep := styleKeyDesc.Render(" › ")
	parts := []string{styleBreadcrumbActive.Render("tables")}

	if a.activeTableKey != "" {
		td := a.activeTableData()
		if td != nil {
			label := familyStyle(td.Table.Family).Render(td.Table.Family) +
				styleBreadcrumb.Render(":") +
				styleBreadcrumbActive.Render(td.Table.Name)
			parts = append(parts, label)
		}
	}
	if a.view == viewTableDetail {
		parts = append(parts, styleBreadcrumbActive.Render(tabNames[a.activeTab]))
	}
	if a.view == viewChainRules && a.activeChain != "" {
		parts = append(parts, styleBreadcrumbActive.Render(tabNames[tabChains]))
		parts = append(parts, styleBreadcrumbActive.Render(a.activeChain))
	}

	return styleBreadcrumb.Render(strings.Join(parts, sep))
}

func (a App) viewTabs() string {
	tabs := make([]string, len(tabNames))
	for i, name := range tabNames {
		if tabKind(i) == a.activeTab {
			tabs[i] = styleTabActive.Render(name)
		} else {
			tabs[i] = styleTab.Render(name)
		}
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	line := lipgloss.NewStyle().
		Width(a.width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottomForeground(colorBorder).
		Render(bar)
	return line
}

func (a App) viewContent() string {
	switch a.view {
	case viewTables:
		return styleTableBorder.Width(a.width - 2).Render(a.tablesTbl.View())
	case viewTableDetail:
		tbl := a.activeTabTable()
		return styleTableBorder.Width(a.width - 2).Render(tbl.View())
	case viewChainRules:
		return styleTableBorder.Width(a.width - 2).Render(a.rulesTbl.View())
	}
	return ""
}

func (a App) viewFlash() string {
	if a.flashMsg == "" {
		return ""
	}
	if a.flashErr {
		return styleError.Render("  " + a.flashMsg)
	}
	return styleSuccess.Render("  " + a.flashMsg)
}

func (a App) viewStatusBar() string {
	var hints []string
	add := func(k, desc string) {
		hints = append(hints, styleKeyHint.Render(k)+" "+styleKeyDesc.Render(desc))
	}

	add("r", "refresh")
	add("a/o", "add")
	add("j/k", "down/up")
	add("gg/G", "top/bottom")
	add("^u/^d", "½page")

	switch a.view {
	case viewTables:
		add("d", "delete")
		add("f", "flush")
		add("enter", "open")
		add("q", "quit")
	case viewTableDetail:
		add("d", "delete")
		if a.activeTab == tabChains {
			add("f", "flush chain")
			add("enter", "open chain")
		}
		add("h/l", "prev/next tab")
		add("esc", "back")
	case viewChainRules:
		add("d", "delete rule")
		add("e/i", "edit rule")
		add("esc", "back")
	}

	bar := strings.Join(hints, "  ")
	return styleStatusBar.Width(a.width).Render(bar)
}

func (a App) overlayModal(behind string) string {
	var overlay string

	isConfirm := a.modal == modalDeleteTable || a.modal == modalDeleteChain ||
		a.modal == modalDeleteRule || a.modal == modalFlushTable ||
		a.modal == modalFlushChain || a.modal == modalDeleteSet ||
		a.modal == modalDeleteMap || a.modal == modalDeleteFlowtable ||
		a.modal == modalDeleteCounter

	if isConfirm {
		overlay = a.confirm.View()
	} else {
		overlay = a.form.View()
	}

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}),
	)
}

// key is a convenience matcher.
func isKey(msg tea.KeyMsg, k string) bool {
	return msg.String() == k
}
