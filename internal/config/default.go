package config

// defaultConfigContent contains the default TOML configuration that is written
// when no config file is found.
const defaultConfigContent = `# Catalyst Configuration
# --------------------
#
# runecraft_host: The SSH hostname or IP address for the RuneCraft server.
#                 This is the server Catalyst will connect to.
#
# Example:
# runecraft_host = "runecraft.example.com"

runecraft_host = "localhost"
`
