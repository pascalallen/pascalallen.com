package shell

type Style int

const (
	StylePlain Style = iota
	StyleBold
	StyleDim
	StyleAccent
	StyleSuccess
	StyleWarning
	StyleError
	StyleHeading
	StylePath
)

var styleNames = [...]string{"plain", "bold", "dim", "accent", "success", "warning", "error", "heading", "path"}

func (s Style) String() string {
	if s < 0 || int(s) >= len(styleNames) {
		return "unknown"
	}
	return styleNames[s]
}

type Span struct {
	Text  string
	Style Style
	Href  string
	Run   string
}

type Line []Span

type BlockKind int

const (
	BlockText BlockKind = iota
	BlockError
	BlockMedia
)

var blockKindNames = [...]string{"text", "error", "media"}

func (k BlockKind) String() string {
	if k < 0 || int(k) >= len(blockKindNames) {
		return "unknown"
	}
	return blockKindNames[k]
}

type Block struct {
	Kind  BlockKind
	Lines []Line
	Src   string
	Alt   string
}

type EffectKind int

const (
	EffectClear EffectKind = iota
	EffectOpen
	EffectFetch
	EffectTheme
)

var effectKindNames = [...]string{"clear", "open", "fetch", "theme"}

func (k EffectKind) String() string {
	if k < 0 || int(k) >= len(effectKindNames) {
		return "unknown"
	}
	return effectKindNames[k]
}

type Effect struct {
	Kind EffectKind
	Arg  string
}

type Result struct {
	Blocks  []Block
	Effects []Effect
}

type Completion struct {
	Line       string
	Cursor     int
	Candidates []string
}

type FeedItem struct {
	Name        string
	Description string
	URL         string
	Meta        string
}

type CommandInfo struct {
	Name     string
	Synopsis string
	Summary  string
}

const (
	FeedRepos    = "repos"
	FeedPackages = "packages"
)
