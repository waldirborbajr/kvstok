package config

// All variables will be replaced on release workflow defined on .goreleaser.yaml
// go build -v -ldflags="-X 'github.com/waldirborbajr/kvstok/cmd.Verzion=v2.0.0'"
// Yeah it is VerZion and not VerSion it is not wrong
var (
	Verzion = "-dev0.0.0"
	// Commit     = "x"
	// CommitDate = "x"
	// BuiltBy    = "x"
)
