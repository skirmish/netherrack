package netherrack

import (
	"github.com/thinkofdeath/soulsand/locale"
)

func setDefaultLocaleStrings() {
	locale.Put("en_GB", "message.player.connect", "§e%s joined the server")
	locale.Put("en_GB", "message.player.disconnect", "§e%s left the server")

	locale.Put("en_GB", "command.error.unknown", "§cError: Unknown command %s")
	locale.Put("en_GB", "command.error.parse", "§cError: %s")
	locale.Put("en_GB", "command.error.int.range", "Number must be between %d-%d")
	locale.Put("en_GB", "command.error.string.length", "String length is longer max length (%d)")
	locale.Put("en_GB", "command.error.float.range", "Number must be between %.2f-%.2f")
	locale.Put("en_GB", "command.usage.command", "§7Command usage for: %s")
	locale.Put("en_GB", "command.usage.int.range", "[Integer(%d-%d)]")
	locale.Put("en_GB", "command.usage.int.norange", "[Integer]")
	locale.Put("en_GB", "command.usage.string.norange", "[String]")
	locale.Put("en_GB", "command.usage.string.range", "[String(Max Length:%d)]")
	locale.Put("en_GB", "command.usage.float.range", "[Float(%.2f-%.2f)]")
	locale.Put("en_GB", "command.usage.float.norange", "[Float]")

	locale.Put("en_GB", "disconnect.reason.loggedin", "Someone is already logged in with that name")
	locale.Put("en_GB", "disconnect.reason.unknown", "Disconnected")
}