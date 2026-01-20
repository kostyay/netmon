package services

// serviceKey is a composite key for port + protocol lookup.
type serviceKey struct {
	port  int
	proto string // "tcp" or "udp"
}

// commonServices maps well-known port/protocol combinations to service names.
var commonServices = map[serviceKey]string{
	// System services
	{20, "tcp"}:    "ftp-data",
	{21, "tcp"}:    "ftp",
	{22, "tcp"}:    "ssh",
	{23, "tcp"}:    "telnet",
	{25, "tcp"}:    "smtp",
	{53, "tcp"}:    "dns",
	{53, "udp"}:    "dns",
	{67, "udp"}:    "dhcp",
	{68, "udp"}:    "dhcp",
	{69, "udp"}:    "tftp",
	{80, "tcp"}:    "http",
	{110, "tcp"}:   "pop3",
	{119, "tcp"}:   "nntp",
	{123, "udp"}:   "ntp",
	{137, "udp"}:   "netbios-ns",
	{138, "udp"}:   "netbios-dgm",
	{139, "tcp"}:   "netbios-ssn",
	{143, "tcp"}:   "imap",
	{161, "udp"}:   "snmp",
	{162, "udp"}:   "snmptrap",
	{179, "tcp"}:   "bgp",
	{194, "tcp"}:   "irc",
	{443, "tcp"}:   "https",
	{445, "tcp"}:   "smb",
	{465, "tcp"}:   "smtps",
	{514, "udp"}:   "syslog",
	{515, "tcp"}:   "lpr",
	{587, "tcp"}:   "submission",
	{636, "tcp"}:   "ldaps",
	{853, "tcp"}:   "dns-tls",
	{993, "tcp"}:   "imaps",
	{995, "tcp"}:   "pop3s",
	{1433, "tcp"}:  "mssql",
	{1521, "tcp"}:  "oracle",
	{1723, "tcp"}:  "pptp",
	{2049, "tcp"}:  "nfs",
	{2049, "udp"}:  "nfs",
	{3306, "tcp"}:  "mysql",
	{3389, "tcp"}:  "rdp",
	{5060, "tcp"}:  "sip",
	{5060, "udp"}:  "sip",
	{5222, "tcp"}:  "xmpp",
	{5432, "tcp"}:  "postgresql",
	{5900, "tcp"}:  "vnc",
	{6379, "tcp"}:  "redis",
	{6443, "tcp"}:  "k8s-api",
	{8080, "tcp"}:  "http-alt",
	{8443, "tcp"}:  "https-alt",
	{9200, "tcp"}:  "elasticsearch",
	{27017, "tcp"}: "mongodb",
}

// Lookup returns the service name for a port/protocol combination.
// Returns empty string if not found.
func Lookup(port int, proto string) string {
	return commonServices[serviceKey{port, proto}]
}

// LookupTCP returns the service name for a TCP port.
func LookupTCP(port int) string {
	return commonServices[serviceKey{port, "tcp"}]
}

// LookupUDP returns the service name for a UDP port.
func LookupUDP(port int) string {
	return commonServices[serviceKey{port, "udp"}]
}
