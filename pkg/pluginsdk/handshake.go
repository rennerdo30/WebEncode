package pluginsdk

import "github.com/hashicorp/go-plugin"

// HandshakeConfig is the contract shared between Kernel and Plugins
// to ensure they are compatible.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "WEBENCODE_PLUGIN",
	MagicCookieValue: "webencode-protocol-v1",
}
