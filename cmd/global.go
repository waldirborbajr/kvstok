package cmd

// All variables will be replaced on release workflow defined on .goreleaser.yaml
// go build -v -ldflags="-X 'github.com/waldirborbajr/kvstok/cmd.Verzion=v2.0.0'"
// Yeah it is VerZion and not VerSion it is not wrong
var (
	Verzion = "x.x.x"
	// Commit     = "x"
	// CommitDate = "x"
	// BuiltBy    = "x"
)
